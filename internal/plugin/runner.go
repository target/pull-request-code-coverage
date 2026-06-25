package plugin

import (
	"fmt"
	"io"
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
	"github.com/target/pull-request-code-coverage/internal/plugin/domain"
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

	// Master on/off switch. When PARAMETER_ENABLED is explicitly false the plugin
	// does nothing and succeeds, so it can be left wired into a pipeline and
	// toggled off without editing the step. Absent or unparseable means enabled,
	// preserving the original always-on behavior.
	if enabledStr, enabledFound := propertyGetter("PARAMETER_ENABLED"); enabledFound {
		if enabled, parseErr := strconv.ParseBool(strings.TrimSpace(enabledStr)); parseErr != nil {
			logrus.Infof("PARAMETER_ENABLED %q is not a valid boolean, defaulting to enabled", enabledStr)
		} else if !enabled {
			logrus.Info("PARAMETER_ENABLED is false, skipping coverage report")
			return nil
		}
	}

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
	logrus.Info("enabled reporters are ")
	for _, eachOne := range reporters {
		logrus.Info(eachOne.GetName())
	}

	if writeErr := reporter.NewForking(reporters).Write(changedLinesWithCoverage); writeErr != nil {
		return writeErr
	}

	return enforceMinCoverage(propertyGetter, changedLinesWithCoverage)
}

// enforceMinCoverage fails the build when PARAMETER_MIN_COVERAGE is set and the
// diff coverage falls below it. The report is always written first (so reviewers
// still see the numbers); only the exit status changes. When the parameter is
// absent or empty the gate is disabled and the run succeeds, preserving the
// original report-only behavior.
func enforceMinCoverage(propertyGetter func(string) (string, bool), report domain.SourceLineCoverageReport) error {
	raw, found := propertyGetter("PARAMETER_MIN_COVERAGE")
	if !found || strings.TrimSpace(raw) == "" {
		logrus.Info("PARAMETER_MIN_COVERAGE was missing, coverage gate disabled")
		return nil
	}

	threshold, parseErr := strconv.ParseFloat(strings.TrimSpace(raw), 64)
	if parseErr != nil {
		return errors.Wrapf(parseErr, "PARAMETER_MIN_COVERAGE %q is not a valid number", raw)
	}

	actual := report.DiffCoveragePercent()
	if actual < threshold {
		return errors.Errorf("diff coverage %.f%% is below the required minimum of %.f%%", actual, threshold)
	}

	logrus.Infof("diff coverage %.f%% meets the required minimum of %.f%%", actual, threshold)

	return nil
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
