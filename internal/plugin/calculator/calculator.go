package calculator

import (
	"github.com/sirupsen/logrus"
	"github.com/target/pull-request-code-coverage/internal/plugin/coverage"
	"github.com/target/pull-request-code-coverage/internal/plugin/domain"
)

type Coverage interface {
	DetermineCoverage(changedLines []domain.SourceLine, coverageReport coverage.Report) ([]domain.SourceLineCoverage, error)
}

func NewCoverage() *DefaultCoverage {
	return &DefaultCoverage{}
}

type DefaultCoverage struct{}

func (*DefaultCoverage) DetermineCoverage(changedLines []domain.SourceLine, coverageReport coverage.Report) []domain.SourceLineCoverage {
	var result []domain.SourceLineCoverage

	for _, sl := range changedLines {

		coverageData, found := coverageReport.GetCoverageData(sl.Module, sl.SrcDir, sl.Pkg, sl.FileName, sl.LineNumber)

		if found {
			logrus.Debugf("%v found: %v", sl, coverageData)
			result = append(result, domain.SourceLineCoverage{
				SourceLine:   sl,
				CoverageData: *coverageData,
			})
		} else {
			logrus.Debugf("%v not found", sl)
			result = append(result, domain.SourceLineCoverage{
				SourceLine:   sl,
				CoverageData: domain.CoverageData{},
			})
		}
	}

	return result
}
