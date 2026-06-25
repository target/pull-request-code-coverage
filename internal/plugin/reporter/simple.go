package reporter

import (
	"fmt"

	"io"
	"log"
	"strings"

	"github.com/target/pull-request-code-coverage/internal/plugin/domain"
)

const consoleRule = "──────────────────────────────────────────────────────────────"

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
	covered := changedLinesWithCoverage.TotalCoveredInstructions()
	missed := changedLinesWithCoverage.TotalMissedInstructions()
	linesWithData := changedLinesWithCoverage.TotalLinesWithData()
	linesWithoutData := changedLinesWithCoverage.TotalLinesWithoutData()

	totalInstructions := covered + missed
	totalLines := linesWithData + linesWithoutData

	coveredPct := toPercent(safeDiv(float32(covered), float32(totalInstructions), 1))
	missedPct := toPercent(safeDiv(float32(missed), float32(totalInstructions), 0))
	withDataPct := toPercent(safeDiv(float32(linesWithData), float32(totalLines), 1))
	withoutDataPct := toPercent(safeDiv(float32(linesWithoutData), float32(totalLines), 0))

	// Build the whole report first and emit it in one write so it stays a
	// contiguous block in the CI console instead of interleaving with logs.
	var b strings.Builder

	b.WriteString(consoleRule + "\n")
	fmt.Fprintf(&b, " 📊 Patch Coverage Report — %.f%% %s — changed lines only\n",
		coveredPct, coverageStatusEmoji(coveredPct))
	b.WriteString(consoleRule + "\n")

	if modules := collectModules(changedLinesWithCoverage); len(modules) > 0 {
		fmt.Fprintf(&b, " Modules: %s\n", strings.Join(modules, ", "))
	}

	fmt.Fprintf(&b, "\n Diff coverage: %.f%% %s  —  %d of %d changed instructions covered\n\n",
		coveredPct, coverageStatusEmoji(coveredPct), covered, totalInstructions)

	b.WriteString(" Summary\n")
	fmt.Fprintf(&b, "   %-26s%3.f%%  (%d)\n", "Covered instructions", coveredPct, covered)
	fmt.Fprintf(&b, "   %-26s%3.f%%  (%d)\n", "Missed instructions", missedPct, missed)
	fmt.Fprintf(&b, "   %-26s%3.f%%  (%d)\n", "Tracked changed lines", withDataPct, linesWithData)
	fmt.Fprintf(&b, "   %-26s%3.f%%  (%d)\n", "Untracked changed lines", withoutDataPct, linesWithoutData)

	b.WriteString("\n Note: \"lines\" are the source lines you changed; \"instructions\" are the\n")
	b.WriteString(" executable units the coverage tool counts inside them (one line can hold\n")
	b.WriteString(" several, e.g. JaCoCo bytecode), so the two counts differ.\n\n")

	b.WriteString(fileCoverageConsoleSection(changedLinesWithCoverage))
	b.WriteString(uncoveredLinesConsoleSection(changedLinesWithCoverage))

	b.WriteString(consoleRule + "\n")

	s.print(b.String())

	return nil
}

func (s *Simple) GetName() string {
	return "simple stdout reporter"
}

// fileCoverageConsoleSection renders the per-file breakdown for the console,
// lowest-covered first, omitting files with no measurable lines.
func fileCoverageConsoleSection(changedLinesWithCoverage domain.SourceLineCoverageReport) string {
	files := collectFileCoverage(changedLinesWithCoverage)

	var b strings.Builder
	b.WriteString(" Coverage by file  (lowest coverage first)\n")

	measured := 0
	unmeasured := 0

	for _, f := range files {
		instructions := f.covered + f.missed

		if instructions == 0 {
			unmeasured++
			continue
		}

		measured++
		pct := toPercent(safeDiv(float32(f.covered), float32(instructions), 1))
		fmt.Fprintf(&b, "   %3.f%%   %3d cov / %3d miss   %s\n", pct, f.covered, f.missed, f.path)
	}

	if measured == 0 {
		b.WriteString("   (no files with measurable lines)\n")
	}

	if unmeasured > 0 {
		fmt.Fprintf(&b, "   (%d file(s) with no measurable lines omitted)\n", unmeasured)
	}

	b.WriteString("\n")

	return b.String()
}

// uncoveredLinesConsoleSection lists each changed line that tests never ran.
func uncoveredLinesConsoleSection(changedLinesWithCoverage domain.SourceLineCoverageReport) string {
	var rows []string

	for _, r := range changedLinesWithCoverage {
		if r.MissedInstructionCount > 0 {
			rows = append(rows, fmt.Sprintf("   - %s\n         %s\n", lineDescription(r.SourceLine), r.LineValue))
		}
	}

	var b strings.Builder
	fmt.Fprintf(&b, " Uncovered lines (%d)\n", len(rows))

	if len(rows) == 0 {
		b.WriteString("   none 🎉\n")
	}

	for _, row := range rows {
		b.WriteString(row)
	}

	b.WriteString("\n")

	return b.String()
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
