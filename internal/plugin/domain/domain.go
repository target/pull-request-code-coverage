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
