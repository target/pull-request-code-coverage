package domain

type SourceLineCoverageReport []SourceLineCoverage

func (r SourceLineCoverageReport) TotalLinesWithData() int {
	result := 0

	for _, l := range r {
		if l.HasData() {
			result++
		}
	}

	return result
}

func (r SourceLineCoverageReport) TotalLinesWithoutData() int {
	result := 0

	for _, l := range r {
		if !l.HasData() {
			result++
		}
	}

	return result
}

func (r SourceLineCoverageReport) TotalCoveredInstructions() int {
	result := 0

	for _, l := range r {
		result += l.CoveredInstructionCount
	}

	return result
}

func (r SourceLineCoverageReport) TotalMissedInstructions() int {
	result := 0

	for _, l := range r {
		result += l.MissedInstructionCount
	}

	return result
}

// DiffCoveragePercent returns the headline diff-coverage figure: covered
// instructions / (covered + missed) * 100. When the PR changed no measurable
// instructions it returns 100, matching the reporters which treat "nothing to
// measure" as fully covered, so such PRs are never failed by the coverage gate.
func (r SourceLineCoverageReport) DiffCoveragePercent() float64 {
	covered := r.TotalCoveredInstructions()
	total := covered + r.TotalMissedInstructions()

	if total == 0 {
		return 100
	}

	return float64(covered) / float64(total) * 100
}

type SourceLineCoverage struct {
	SourceLine
	CoverageData
}

func (slc SourceLineCoverage) HasData() bool {
	return (slc.MissedInstructionCount + slc.CoveredInstructionCount) > 0
}

type CoverageData struct {
	MissedInstructionCount  int
	CoveredInstructionCount int
}

type SourceLine struct {
	Module     string
	SrcDir     string
	Pkg        string
	FileName   string
	LineNumber int
	LineValue  string
}
