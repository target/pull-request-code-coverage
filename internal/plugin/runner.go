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
	"strings"
)

type DefaultRunner struct{}

func NewRunner() *DefaultRunner {
	return &DefaultRunner{}
}

//Run ...
//nolint: gocyclo
func (*DefaultRunner) Run(propertyGetter func(string) (string, bool), changedSourceLinesSource io.Reader, reportDefaultOut io.Writer) error {

	rawSourceDirs, found := propertyGetter("PLUGIN_SOURCE_DIRS")
	if !found {
		return errors.New("Missing property PLUGIN_SOURCE_DIRS")
	}
	logrus.Info(fmt.Sprintf("PLUGIN_SOURCE_DIRS set to %v", rawSourceDirs))

	sourceDirs := parseSourceDirs(rawSourceDirs)

	coverageType, found := propertyGetter("PLUGIN_COVERAGE_TYPE")
	if !found {
		return errors.New("Missing property PLUGIN_COVERAGE_TYPE")
	}

	loader, getLoaderErr := getCoverageReportLoader(coverageType, sourceDirs)
	if getLoaderErr != nil {
		return errors.Wrap(getLoaderErr, "Failed opening coverage loader")
	}

	coverageFile, found := propertyGetter("PLUGIN_COVERAGE_FILE")
	if !found {
		return errors.New("Missing property PLUGIN_COVERAGE_FILE")
	}

	module, found := propertyGetter("PLUGIN_MODULE")
	if !found {
		logrus.Info("PLUGIN_MODULE was missing defaulting to false")
		module = ""
	}

	ghAPIKey, ghAPIKeyFound := propertyGetter("PLUGIN_GH_API_KEY")
	if !ghAPIKeyFound {
		logrus.Info("PLUGIN_GH_API_KEY was missing, will not send report to PR comments")
	}

	ghAPIBaseURL, ghAPIBaseURLFound := propertyGetter("PLUGIN_GH_API_BASE_URL")
	if !ghAPIBaseURLFound {
		logrus.Info("PLUGIN_GH_API_BASE_URL was missing, will not send report to PR comments")
	}

	dronePR, dronePRFound := propertyGetter("DRONE_PULL_REQUEST")
	if !dronePRFound {
		logrus.Info("DRONE_PULL_REQUEST was missing, will not send report to PR comments")
	}

	droneOwner, droneOwnerFound := propertyGetter("DRONE_REPO_OWNER")
	if !droneOwnerFound {
		logrus.Info("DRONE_REPO_OWNER was missing, will not send report to PR comments")
	}

	droneRepo, droneRepoFound := propertyGetter("DRONE_REPO_NAME")
	if !droneRepoFound {
		logrus.Info("DRONE_REPO_NAME was missing, will not send report to PR comments")
	}

	coverageReport, loadCoverageErr := loader.Load(coverageFile)
	if loadCoverageErr != nil {
		return errors.Wrap(loadCoverageErr, "Failed loading coverage report")
	}

	changedLines, changedLinesErr := unifieddiff.NewChangedSourceLinesLoader(module, sourceDirs).Load(changedSourceLinesSource)
	if changedLinesErr != nil {
		return errors.Wrap(changedLinesErr, "Failed loading changed lines")
	}

	changedLinesWithCoverage, determineCoverageErr := calculator.NewCoverage().DetermineCoverage(changedLines, coverageReport)
	if determineCoverageErr != nil {
		return errors.Wrap(determineCoverageErr, "Failed determining overall coverage for changed lines")
	}

	reporters := []reporter.Reporter{reporter.NewSimple(reportDefaultOut)}

	if ghAPIKeyFound && ghAPIBaseURLFound && dronePRFound && droneOwnerFound && droneRepoFound {
		reporters = append(reporters, reporter.NewGithubPullRequest(ghAPIKey, ghAPIBaseURL, dronePR, droneOwner, droneRepo, &pluginhttp.DefaultClient{}, &pluginjson.DefaultClient{}))
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
