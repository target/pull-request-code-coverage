package calculator

import (
	"git.target.com/search-product-team/pull-request-code-coverage/internal/plugin/coverage"
	"git.target.com/search-product-team/pull-request-code-coverage/internal/plugin/domain"
	"github.com/pkg/errors"
)

type Coverage interface {
	DetermineCoverage(changedLines []domain.SourceLine, coverageReport coverage.Report) ([]domain.SourceLineCoverage, error)
}

func NewCoverage() *DefaultCoverage {
	return &DefaultCoverage{}
}

type DefaultCoverage struct{}

func (*DefaultCoverage) DetermineCoverage(changedLines []domain.SourceLine, coverageReport coverage.Report) ([]domain.SourceLineCoverage, error) {
	var result []domain.SourceLineCoverage

	for _, sl := range changedLines {

		if sl.Module != coverageReport.Name() {
			return nil, errors.Errorf("Coverage report %v did not match the module %v", sl.Module, coverageReport.Name())
		}

		coverageData, found := coverageReport.GetCoverageData(sl.Module, sl.SrcDir, sl.Pkg, sl.FileName, sl.LineNumber)

		if found {
			result = append(result, domain.SourceLineCoverage{
				SourceLine:   sl,
				CoverageData: *coverageData,
			})
		} else {
			result = append(result, domain.SourceLineCoverage{
				SourceLine:   sl,
				CoverageData: domain.CoverageData{},
			})
		}
	}

	return result, nil
}