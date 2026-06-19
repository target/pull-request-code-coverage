package reporter

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/target/pull-request-code-coverage/internal/plugin/domain"
)

func TestStepSummary_Write_RendersMarkdownReport(t *testing.T) {

	var buf bytes.Buffer

	e := NewStepSummary(&buf).Write(domain.SourceLineCoverageReport{
		domain.SourceLineCoverage{
			CoverageData: domain.CoverageData{
				CoveredInstructionCount: 1,
			},
		},
	})

	assert.NoError(t, e)

	out := buf.String()
	assert.Contains(t, out, commentMarker)
	assert.Contains(t, out, "Patch Coverage Report")
}

func TestStepSummary_Write_NoDataWritesNothing(t *testing.T) {

	var buf bytes.Buffer

	e := NewStepSummary(&buf).Write(domain.SourceLineCoverageReport{})

	assert.NoError(t, e)
	assert.Empty(t, buf.String())
}

func TestStepSummary_GetName(t *testing.T) {
	assert.Equal(t, "github step summary reporter", NewStepSummary(&strings.Builder{}).GetName())
}
