package plugin

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/target/pull-request-code-coverage/internal/plugin/calculator"
	"github.com/target/pull-request-code-coverage/internal/plugin/coverage"
	"github.com/target/pull-request-code-coverage/internal/plugin/coverage/cobertura"
	"github.com/target/pull-request-code-coverage/internal/plugin/coverage/jacoco"
	"github.com/target/pull-request-code-coverage/internal/plugin/coverage/lcov"
	"github.com/target/pull-request-code-coverage/internal/plugin/coverage/pythoncov"
	"github.com/target/pull-request-code-coverage/internal/plugin/githubdiff"
	"github.com/target/pull-request-code-coverage/internal/plugin/gitdiff"
	"github.com/target/pull-request-code-coverage/internal/plugin/pluginhttp"
	"github.com/target/pull-request-code-coverage/internal/plugin/pluginjson"
	"github.com/target/pull-request-code-coverage/internal/plugin/reporter"
	"github.com/target/pull-request-code-coverage/internal/plugin/sourcelines/unifieddiff"
)

// defaultGithubAPIBaseURL is the public GitHub REST API root. GitHub Enterprise
// users should set PARAMETER_GH_API_BASE_URL to their API root (e.g.
// https://git.example.com/api/v3).
const defaultGithubAPIBaseURL = "https://api.github.com"

type DefaultRunner struct{}

func NewRunner() *DefaultRunner {
	return &DefaultRunner{}
}

// Run ...
// nolint: gocyclo
func (*DefaultRunner) Run(propertyGetter func(string) (string, bool), changedSourceLinesSource io.Reader, reportDefaultOut io.Writer) error {

	logrus.Info("starting pull-request-code-coverage run")

	rawSourceDirs, found := propertyGetter("PARAMETER_SOURCE_DIRS")
	if !found {
		return errors.New("Missing property PARAMETER_SOURCE_DIRS")
	}
	logrus.Info(fmt.Sprintf("PARAMETER_SOURCE_DIRS set to %v", rawSourceDirs))

	sourceDirs := parseSourceDirs(rawSourceDirs)

	coverageType, found := propertyGetter("PARAMETER_COVERAGE_TYPE")
	if !found {
		return errors.New("Missing property PARAMETER_COVERAGE_TYPE")
	}

	loader, getLoaderErr := getCoverageReportLoader(coverageType, sourceDirs)
	if getLoaderErr != nil {
		return errors.Wrap(getLoaderErr, "Failed opening coverage loader")
	}

	coverageFile, found := propertyGetter("PARAMETER_COVERAGE_FILE")
	if !found {
		return errors.New("Missing property PARAMETER_COVERAGE_FILE")
	}

	module, found := propertyGetter("PARAMETER_MODULE")
	if !found {
		logrus.Info("PARAMETER_MODULE was missing defaulting to false")
		module = ""
	}

	ghAPIKey, ghAPIKeyFound := propertyGetter("PARAMETER_GH_API_KEY")
	if !ghAPIKeyFound {
		ghAPIKey, ghAPIKeyFound = propertyGetter("PLUGIN_GH_API_KEY")
		if !ghAPIKeyFound {
			logrus.Info("PARAMETER_GH_API_KEY/PLUGIN_GH_API_KEY was missing, will not send report to PR comments")
		}
	}

	ghAPIBaseURL, ghAPIBaseURLFound := propertyGetter("PARAMETER_GH_API_BASE_URL")
	if !ghAPIBaseURLFound || ghAPIBaseURL == "" {
		ghAPIBaseURL = defaultGithubAPIBaseURL
		logrus.Info(fmt.Sprintf("PARAMETER_GH_API_BASE_URL was missing, defaulting to %v", ghAPIBaseURL))
	}

	repoPR, repoPRFound := propertyGetter("BUILD_PULL_REQUEST_NUMBER")
	if !repoPRFound {
		logrus.Info("BUILD_PULL_REQUEST_NUMBER was missing, will not send report to PR comments")
	}
	repoOwner, repoOwnerFound := propertyGetter("REPOSITORY_ORG")
	if !repoOwnerFound {
		logrus.Info("REPOSITORY_ORG was missing, will not send report to PR comments")
	}
	repoName, repoNameFound := propertyGetter("REPOSITORY_NAME")
	if !repoNameFound {
		logrus.Info("REPOSITORY_NAME was missing, will not send report to PR comments")
	}
	coverageReport, loadCoverageErr := loader.Load(coverageFile)
	if loadCoverageErr != nil {
		return errors.Wrap(loadCoverageErr, "Failed loading coverage report")
	}

	diffSource, found := propertyGetter("PARAMETER_DIFF_SOURCE")
	if !found || diffSource == "" {
		// Auto-detect: when running in Vela CI (VELA_PULL_REQUEST_TARGET is set by
		// the Vela runtime) fall back to the git diff mode so pipelines that relied
		// on the old start.sh entrypoint continue to work without any config change.
		if velaTarget, hasVelaTarget := propertyGetter("VELA_PULL_REQUEST_TARGET"); hasVelaTarget && velaTarget != "" {
			logrus.Info("PARAMETER_DIFF_SOURCE was missing but VELA_PULL_REQUEST_TARGET is set, defaulting to git")
			diffSource = "git"
		} else {
			logrus.Info("PARAMETER_DIFF_SOURCE was missing, defaulting to stdin")
			diffSource = "stdin"
		}
	}

	switch diffSource {
	case "stdin":
		// changedSourceLinesSource already points at the piped-in diff (stdin);
		// nothing to do. This is the original, default behavior.
	case "git":
		// Runs git fetch + git diff locally, reproducing what start.sh did before
		// it was removed in v1.0.1. The base branch is read from PARAMETER_BASE_BRANCH
		// first, then VELA_PULL_REQUEST_TARGET for backward compatibility.
		baseBranch, hasBranch := propertyGetter("PARAMETER_BASE_BRANCH")
		if !hasBranch || baseBranch == "" {
			baseBranch, _ = propertyGetter("VELA_PULL_REQUEST_TARGET")
		}

		logrus.Infof("PARAMETER_DIFF_SOURCE is git, diffing against origin/%s", baseBranch)

		diffReader, fetchErr := gitdiff.NewLoader(baseBranch, module).Load()
		if fetchErr != nil {
			return errors.Wrap(fetchErr, "Failed fetching diff via git")
		}

		changedSourceLinesSource = diffReader
	case "github":
		if !ghAPIKeyFound || !repoPRFound || !repoOwnerFound || !repoNameFound {
			return errors.New("PARAMETER_DIFF_SOURCE=github requires a GitHub API key (PARAMETER_GH_API_KEY), BUILD_PULL_REQUEST_NUMBER, REPOSITORY_ORG and REPOSITORY_NAME")
		}

		logrus.Info("PARAMETER_DIFF_SOURCE is github, fetching diff from the GitHub API")

		diffReader, fetchErr := githubdiff.NewLoader(ghAPIKey, ghAPIBaseURL, repoPR, repoOwner, repoName, &pluginhttp.DefaultClient{}).Load()
		if fetchErr != nil {
			return errors.Wrap(fetchErr, "Failed fetching diff from github")
		}

		changedSourceLinesSource = diffReader
	default:
		return errors.Errorf("Unknown PARAMETER_DIFF_SOURCE %q (expected \"stdin\", \"git\", or \"github\")", diffSource)
	}

	changedLines, changedLinesErr := unifieddiff.NewChangedSourceLinesLoader(module, sourceDirs).Load(changedSourceLinesSource)
	if changedLinesErr != nil {
		return errors.Wrap(changedLinesErr, "Failed loading changed lines")
	}

	debugStr, found := propertyGetter("PARAMETER_DEBUG")
	if !found {
		logrus.Info("PARAMETER_DEBUG was missing defaulting to false")
		debugStr = "false"
	}

	debug, parseDebugErr := strconv.ParseBool(debugStr)
	if parseDebugErr != nil {
		logrus.Info("PARAMETER_DEBUG was invalid, defaulting to false")
		debug = false
	}

	if debug {
		logrus.SetLevel(logrus.DebugLevel)
	}

	changedLinesWithCoverage := calculator.NewCoverage().DetermineCoverage(changedLines, coverageReport)

	reporters := []reporter.Reporter{reporter.NewSimple(reportDefaultOut)}

	if ghAPIKeyFound && repoPRFound && repoOwnerFound && repoNameFound {
		reporters = append(reporters, reporter.NewGithubPullRequest(ghAPIKey, ghAPIBaseURL, repoPR, repoOwner, repoName, &pluginhttp.DefaultClient{}, &pluginjson.DefaultClient{}))
	}

	// GitHub Actions sets GITHUB_STEP_SUMMARY to a file whose Markdown is rendered
	// on the run's summary page. Writing there surfaces coverage even when no PR
	// comment can be posted (e.g. fork PRs with a read-only token).
	if summaryPath, found := propertyGetter("GITHUB_STEP_SUMMARY"); found && summaryPath != "" {
		// nolint: gosec // path comes from the trusted GitHub Actions runner env
		summaryFile, openErr := os.OpenFile(summaryPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if openErr != nil {
			return errors.Wrap(openErr, "Failed opening GITHUB_STEP_SUMMARY file")
		}
		defer func() {
			_ = summaryFile.Close()
		}()

		reporters = append(reporters, reporter.NewStepSummary(summaryFile))
	}

	logrus.Info("enabled reporters are ")
	for _, eachOne := range reporters {
		logrus.Info(eachOne.GetName())
	}

	return reporter.NewForking(reporters).Write(changedLinesWithCoverage)
}

func parseSourceDirs(rawSourceDirStr string) []string {
	rawSourceDirs := strings.Split(rawSourceDirStr, ",")
	resultSoruceDirs := make([]string, len(rawSourceDirs))

	for idx, rs := range rawSourceDirs {
		resultSoruceDirs[idx] = strings.TrimSpace(rs)
	}

	return resultSoruceDirs
}

func getCoverageReportLoader(coverageType string, sourceDirs []string) (coverage.Loader, error) {
	switch coverageType {
	case "cobertura":
		if len(sourceDirs) != 1 {
			return nil, errors.Errorf("Currently not supporting multiple source dirs with cobertura coverage report type.")
		}

		return cobertura.NewReportLoader(sourceDirs[0]), nil
	case "python":
		return pythoncov.NewReportLoader(), nil
	case "lcov", "javascript", "typescript":
		return lcov.NewReportLoader(), nil
	default:
		return jacoco.NewReportLoader(), nil
	}

}
