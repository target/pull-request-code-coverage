package reporter

import (
	"fmt"

	"io"
	"log"

	"github.com/target/pull-request-code-coverage/internal/plugin/domain"
)

type Simple struct {
	Out          io.Writer
	WritingFuncf func(io.Writer, string, ...interface{}) (int, error)
	WritingFunc  func(io.Writer, ...interface{}) (int, error)
}

func NewSimple(out io.Writer) *Simple {
	return &Simple{
		Out:          out,
		WritingFuncf: fmt.Fprintf,
		WritingFunc:  fmt.Fprint,
	}
}

func (s *Simple) Write(changedLinesWithCoverage domain.SourceLineCoverageReport) error {
	s.printf("Missed Instructions:\n")
	for _, r := range changedLinesWithCoverage {
		if r.MissedInstructionCount > 0 {
			s.printf("--- %v\n", lineDescription(r.SourceLine))
			s.printf("%v\n", r.LineValue)
		}
	}

	summaryLines := generateSummaryLines(changedLinesWithCoverage, func(linesWithDataCount int, linesWithoutDataCount int, covered int, missed int) []string {
		totalLines := linesWithDataCount + linesWithoutDataCount
		totalInstructions := covered + missed

		result := make([]string, 5)

		result[0] = fmt.Sprintf("Code Coverage Summary:\n")
		result[1] = fmt.Sprintf("Lines Without Coverage Data -> %.f%% (%d)\n", toPercent(safeDiv(float32(linesWithoutDataCount), float32(totalLines), 0)), linesWithoutDataCount)
		result[2] = fmt.Sprintf("Lines With Coverage Data    -> %.f%% (%d)\n", toPercent(safeDiv(float32(linesWithDataCount), float32(totalLines), 1)), linesWithDataCount)
		result[3] = fmt.Sprintf("Covered Instructions        -> %.f%% (%d)\n", toPercent(safeDiv(float32(covered), float32(totalInstructions), 1)), covered)
		result[4] = fmt.Sprintf("Missed Instructions         -> %.f%% (%d)\n", toPercent(safeDiv(float32(missed), float32(totalInstructions), 0)), missed)

		return result
	})

	s.print("\n")
	for _, line := range summaryLines {
		s.print(line)
	}

	return nil
}

func (s *Simple) GetName() string {
	return "simple stdout reporter"
}

func generateSummaryLines(changedLinesWithCoverage domain.SourceLineCoverageReport, formatter func(linesWithDataCount int, linesWithoutDataCount int, covered int, missed int) []string) []string {
	linesWithDataCount := changedLinesWithCoverage.TotalLinesWithData()
	linesWithoutDataCount := changedLinesWithCoverage.TotalLinesWithoutData()

	covered := changedLinesWithCoverage.TotalCoveredInstructions()
	missed := changedLinesWithCoverage.TotalMissedInstructions()

	return formatter(linesWithDataCount, linesWithoutDataCount, covered, missed)
}

func toPercent(decimal float32) float32 {
	return decimal * float32(100)
}

func safeDiv(top float32, bottom float32, safeDefault float32) float32 {
	if bottom == 0 {
		return safeDefault
	}

	return top / bottom
}

func (s *Simple) printf(str string, is ...interface{}) {
	if _, err := s.WritingFuncf(s.Out, str, is...); err != nil {
		log.Panic(err)
	}
}

func (s *Simple) print(str string) {
	if _, err := s.WritingFunc(s.Out, str); err != nil {
		log.Panic(err)
	}
}
