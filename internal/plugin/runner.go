package plugin

import (
	"git.target.com/search-product-team/pull-request-code-coverage/internal/plugin/calculator"
	"git.target.com/search-product-team/pull-request-code-coverage/internal/plugin/coverage"
	"git.target.com/search-product-team/pull-request-code-coverage/internal/plugin/coverage/cobertura"
	"git.target.com/search-product-team/pull-request-code-coverage/internal/plugin/coverage/jacoco"
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

func (*DefaultRunner) Run(propertyGetter func(string) (string, bool), changedSourceLinesSource io.Reader, reportDefaultOut io.Writer) error {

	coverageFile, found := propertyGetter("PLUGIN_COVERAGE_FILE")
	if !found {
		return errors.New("Missing property PLUGIN_COVERAGE_FILE")
	}

	////TODO !found
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

	reporter.NewSimple(reportDefaultOut).Write(changedLinesWithCoverage)

	return nil
}

func GetCoverageReportLoader(coverageType string, sourceDir string) coverage.Loader {
	switch coverageType {
	case "cobertura":
		return cobertura.NewReportLoader(sourceDir)
	default:
		return jacoco.NewReportLoader()
	}

}
