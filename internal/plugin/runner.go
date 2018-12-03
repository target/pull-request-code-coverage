package plugin

import (
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
)

type DefaultRunner struct{}

func NewRunner() *DefaultRunner {
	return &DefaultRunner{}
}

//Run ...
//nolint: gocyclo
func (*DefaultRunner) Run(propertyGetter func(string) (string, bool), changedSourceLinesSource io.Reader, reportDefaultOut io.Writer) error {

	coverageFile, found := propertyGetter("PLUGIN_COVERAGE_FILE")
	if !found {
		return errors.New("Missing property PLUGIN_COVERAGE_FILE")
	}

	coverageType, found := propertyGetter("PLUGIN_COVERAGE_TYPE")
	if !found {
		return errors.New("Missing property PLUGIN_COVERAGE_TYPE")
	}

	sourceDir, found := propertyGetter("PLUGIN_SOURCE_DIR")
	if !found {
		return errors.New("Missing property PLUGIN_SOURCE_DIR")
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

	coverageReport, loadCoverageErr := GetCoverageReportLoader(coverageType, sourceDir).Load(coverageFile)
	if loadCoverageErr != nil {
		return errors.Wrap(loadCoverageErr, "Failed loading coverage report")
	}

	changedLines, changedLinesErr := unifieddiff.NewChangedSourceLinesLoader(module, sourceDir).Load(changedSourceLinesSource)
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

func GetCoverageReportLoader(coverageType string, sourceDir string) coverage.Loader {
	switch coverageType {
	case "cobertura":
		return cobertura.NewReportLoader(sourceDir)
	default:
		return jacoco.NewReportLoader()
	}

}
