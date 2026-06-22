package reporter

import (
	"io"

	"github.com/pkg/errors"
	"github.com/target/pull-request-code-coverage/internal/plugin/domain"
)

// StepSummary writes the Markdown coverage report to a writer backed by the
// GitHub Actions job summary file ($GITHUB_STEP_SUMMARY). The summary shows on
// the workflow run page, so coverage is visible even on fork PRs where the
// GITHUB_TOKEN is read-only and the PR comment cannot be posted.
type StepSummary struct {
	out io.Writer
}

func NewStepSummary(out io.Writer) *StepSummary {
	return &StepSummary{
		out: out,
	}
}

func (s *StepSummary) Write(changedLinesWithCoverage domain.SourceLineCoverageReport) error {
	if changedLinesWithCoverage.TotalLinesWithData() == 0 {
		return nil
	}

	if _, err := io.WriteString(s.out, buildMarkdownReport(changedLinesWithCoverage)); err != nil {
		return errors.Wrap(err, "Failed writing job summary")
	}

	return nil
}

func (s *StepSummary) GetName() string {
	return "github step summary reporter"
}
