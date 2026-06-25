package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSourceLineCoverageReport_DiffCoveragePercent(t *testing.T) {
	tts := []struct {
		name     string
		report   SourceLineCoverageReport
		expected float64
	}{
		{
			name:     "no measurable instructions defaults to 100",
			report:   SourceLineCoverageReport{{CoverageData: CoverageData{}}},
			expected: 100,
		},
		{
			name:     "fully covered",
			report:   SourceLineCoverageReport{{CoverageData: CoverageData{CoveredInstructionCount: 4}}},
			expected: 100,
		},
		{
			name:     "fully missed",
			report:   SourceLineCoverageReport{{CoverageData: CoverageData{MissedInstructionCount: 3}}},
			expected: 0,
		},
		{
			name: "partial coverage",
			report: SourceLineCoverageReport{
				{CoverageData: CoverageData{CoveredInstructionCount: 3}},
				{CoverageData: CoverageData{MissedInstructionCount: 1}},
			},
			expected: 75,
		},
	}

	for _, tt := range tts {
		t.Run(tt.name, func(t *testing.T) {
			assert.InDelta(t, tt.expected, tt.report.DiffCoveragePercent(), 0.0001)
		})
	}
}
