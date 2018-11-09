package reporter

import (
	"fmt"
	"git.target.com/search-product-team/pull-request-code-coverage/internal/plugin/domain"
	"io"
	"log"
	"strings"
)

type Reporter interface {
	Write(domain.SourceLineCoverageReport)
}

type Simple struct {
	Out         io.Writer
	WritingFunc func(io.Writer, string, ...interface{}) (int, error)
}

func NewSimple(out io.Writer) *Simple {
	return &Simple{
		Out:         out,
		WritingFunc: fmt.Fprintf,
	}
}

func (s *Simple) Write(changedLinesWithCoverage domain.SourceLineCoverageReport) {
	s.printf("Missed Instructions:\n")
	for _, r := range changedLinesWithCoverage {
		if r.MissedInstructionCount > 0 {
			s.printf("--- %v\n", s.lineDescription(r.SourceLine))
			s.printf("%v\n", r.LineValue)
		}
	}

	linesWithDataCount := changedLinesWithCoverage.TotalLinesWithData()
	linesWithoutDataCount := changedLinesWithCoverage.TotalLinesWithoutData()
	totalLines := linesWithDataCount + linesWithoutDataCount

	covered := changedLinesWithCoverage.TotalCoveredInstructions()
	missed := changedLinesWithCoverage.TotalMissedInstructions()
	totalInstructions := covered + missed

	s.printf("\n")
	s.printf("Summary:\n")
	s.printf("Lines Without Coverage Data -> %.f%% (%d)\n", toPercent(safeDiv(float32(linesWithoutDataCount), float32(totalLines), 0)), linesWithoutDataCount)
	s.printf("Lines With Coverage Data    -> %.f%% (%d)\n", toPercent(safeDiv(float32(linesWithDataCount), float32(totalLines), 1)), linesWithDataCount)
	s.printf("Covered Instructions        -> %.f%% (%d)\n", toPercent(safeDiv(float32(covered), float32(totalInstructions), 1)), covered)
	s.printf("Missed Instructions         -> %.f%% (%d)\n", toPercent(safeDiv(float32(missed), float32(totalInstructions), 0)), missed)
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
	if _, err := s.WritingFunc(s.Out, str, is...); err != nil {
		log.Panic(err)
	}
}

func (s *Simple) lineDescription(l domain.SourceLine) string {
	rawFileNameParts := []string{
		l.Module, l.SrcDir, l.Pkg, l.FileName,
	}

	fileNameParts := []string{}
	for _, part := range rawFileNameParts {
		if len(part) > 0 {
			fileNameParts = append(fileNameParts, part)
		}
	}

	return fmt.Sprintf("%v:%v", strings.Join(fileNameParts, "/"), l.LineNumber)
}
