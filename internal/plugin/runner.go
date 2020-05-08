package plugin

import (
	"fmt"
	"git.target.com/search-product-team/pull-request-code-coverage/internal/plugin/calculator"
	"git.target.com/search-product-team/pull-request-code-coverage/internal/plugin/coverage"
	"git.target.com/search-product-team/pull-request-code-coverage/internal/plugin/coverage/cobertura"
	"git.target.com/search-product-team/pull-request-code-coverage/internal/plugin/coverage/jacoco"
	"git.target.com/search-product-team/pull-request-code-coverage/internal/plugin/pluginhttp"
	"git.target.com/search-product-team/pull-request-code-coverage/internal/plugin/pluginjson"
	"git.target.com/search-product-team/pull-request-code-coverage/internal/plugin/reporter"
	"git.target.com/search-product-team/pull-request-code-coverage/internal/plugin/sourcelines/unifieddiff"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"io"
	"strconv"
	"strings"
)

type DefaultRunner struct{}

func NewRunner() *DefaultRunner {
	return &DefaultRunner{}
}

//Run ...
//nolint: gocyclo
func (*DefaultRunner) Run(propertyGetter func(string) (string, bool), changedSourceLinesSource io.Reader, reportDefaultOut io.Writer) error {

	rawSourceDirs, found := getActiveEnvVariable(propertyGetter, "PLUGIN_SOURCE_DIRS", "PARAMETER_SOURCE_DIRS")
	if !found {
		return errors.New("Missing property PARAMETER_SOURCE_DIRS")
	}
	logrus.Info(fmt.Sprintf("PLUGIN_SOURCE_DIRS set to %v", rawSourceDirs))

	sourceDirs := parseSourceDirs(rawSourceDirs)

	coverageType, found := getActiveEnvVariable(propertyGetter, "PLUGIN_COVERAGE_TYPE", "PARAMETER_COVERAGE_TYPE")
	if !found {
		return errors.New("Missing property PARAMETER_COVERAGE_TYPE")
	}

	loader, getLoaderErr := getCoverageReportLoader(coverageType, sourceDirs)
	if getLoaderErr != nil {
		return errors.Wrap(getLoaderErr, "Failed opening coverage loader")
	}

	coverageFile, found := getActiveEnvVariable(propertyGetter, "PLUGIN_COVERAGE_FILE", "PARAMETER_COVERAGE_FILE")
	if !found {
		return errors.New("Missing property PARAMETER_COVERAGE_FILE")
	}

	module, found := getActiveEnvVariable(propertyGetter, "PLUGIN_MODULE", "PARAMETER_MODULE")
	if !found {
		logrus.Info("PARAMETER_MODULE was missing defaulting to false")
		module = ""
	}

	ghAPIKey, ghAPIKeyFound := getActiveEnvVariable(propertyGetter, "PLUGIN_GH_API_KEY", "PARAMETER_GH_API_KEY")
	if !ghAPIKeyFound {
		logrus.Info("PARAMETER_GH_API_KEY was missing, will not send report to PR comments")
	}

	ghAPIBaseURL, ghAPIBaseURLFound := getActiveEnvVariable(propertyGetter, "PLUGIN_GH_API_BASE_URL", "PARAMETER_GH_API_BASE_URL")
	if !ghAPIBaseURLFound {
		logrus.Info("PARAMETER_GH_API_BASE_URL was missing, will not send report to PR comments")
	}

	repoPR, repoPRFound := getActiveEnvVariable(propertyGetter, "DRONE_PULL_REQUEST", "BUILD_PULL_REQUEST_NUMBER")
	if !repoPRFound {
		logrus.Info("BUILD_PULL_REQUEST_NUMBER was missing, will not send report to PR comments")
	}
	repoOwner, repoOwnerFound := getActiveEnvVariable(propertyGetter, "DRONE_REPO_OWNER", "REPOSITORY_ORG")
	if !repoOwnerFound {
		logrus.Info("REPOSITORY_ORG was missing, will not send report to PR comments")
	}
	repoName, repoNameFound := getActiveEnvVariable(propertyGetter, "DRONE_REPO_NAME", "REPOSITORY_NAME")
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

	debugStr, found := getActiveEnvVariable(propertyGetter, "PLUGIN_DEBUG", "PARAMETER_DEBUG")
	if !found {
		logrus.Info("PLUGIN_DEBUG was missing defaulting to false")
		debugStr = "false"
	}

	debug, parseDebugErr := strconv.ParseBool(debugStr)
	if parseDebugErr != nil {
		logrus.Info("PLUGIN_DEBUG was invalid, defaulting to false")
		debug = false
	}

	if debug {
		logrus.SetLevel(logrus.DebugLevel)
	}

	changedLinesWithCoverage := calculator.NewCoverage().DetermineCoverage(changedLines, coverageReport)

	reporters := []reporter.Reporter{reporter.NewSimple(reportDefaultOut)}

	if ghAPIKeyFound && ghAPIBaseURLFound && repoPRFound && repoOwnerFound && repoNameFound {
		reporters = append(reporters, reporter.NewGithubPullRequest(ghAPIKey, ghAPIBaseURL, repoPR, repoOwner, repoName, &pluginhttp.DefaultClient{}, &pluginjson.DefaultClient{}))
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
	default:
		return jacoco.NewReportLoader(), nil
	}

}

func getActiveEnvVariable(propertyGetter func(string) (string, bool), droneVar, velaVar string) (string, bool) {
	value, droneVarFound := propertyGetter(droneVar)
	if !droneVarFound {
		return propertyGetter(velaVar)
	}
	return value, droneVarFound
}
