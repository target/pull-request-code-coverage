package lcov

import (
	"strings"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func Test_silentlyCall(t *testing.T) {
	assert.Panics(t, func() {
		silentlyCall(func() error {
			return errors.New("anything")
		})
	})
}

func Test_Load_NoFile(t *testing.T) {
	_, e := NewReportLoader().Load("../../../test/anything.info")
	assert.EqualError(t, e, "Could not open lcov file ../../../test/anything.info: open ../../../test/anything.info: no such file or directory")
}

func Test_Load_BadDARecord(t *testing.T) {
	_, e := NewReportLoader().Load("../../../test/example_lcov_bad.info")
	assert.EqualError(t, e, "Failed parsing lcov file ../../../test/example_lcov_bad.info: Invalid line number in DA record \"DA:notanumber,1\": strconv.Atoi: parsing \"notanumber\": invalid syntax")
}

func loadReport(t *testing.T) *Report {
	r, e := NewReportLoader().Load("../../../test/example_lcov.info")
	assert.NoError(t, e)

	report, ok := r.(*Report)
	assert.True(t, ok)

	return report
}

func Test_GetCoverageData_CoveredLine_SuffixMatchesAbsolutePath(t *testing.T) {
	// SF: in the fixture is absolute, so this exercises suffix matching.
	data, found := loadReport(t).GetCoverageData("", "", "src", "calculator.ts", 1)

	assert.True(t, found)
	assert.Equal(t, 1, data.CoveredInstructionCount)
	assert.Equal(t, 0, data.MissedInstructionCount)
}

func Test_GetCoverageData_MissedLine(t *testing.T) {
	data, found := loadReport(t).GetCoverageData("", "", "src", "calculator.ts", 6)

	assert.True(t, found)
	assert.Equal(t, 0, data.CoveredInstructionCount)
	assert.Equal(t, 1, data.MissedInstructionCount)
}

func Test_GetCoverageData_UntrackedLineNotFound(t *testing.T) {
	_, found := loadReport(t).GetCoverageData("", "", "src", "calculator.ts", 3)

	assert.False(t, found)
}

func Test_GetCoverageData_UnknownFileNotFound(t *testing.T) {
	_, found := loadReport(t).GetCoverageData("", "", "other", "missing.ts", 1)

	assert.False(t, found)
}

func Test_GetCoverageData_ExactRelativePathMatch(t *testing.T) {
	report, e := parse(strings.NewReader("SF:src/app.ts\nDA:10,1\nDA:11,0\nend_of_record\n"))
	assert.NoError(t, e)

	covered, foundCovered := report.GetCoverageData("", "", "src", "app.ts", 10)
	assert.True(t, foundCovered)
	assert.Equal(t, 1, covered.CoveredInstructionCount)

	missed, foundMissed := report.GetCoverageData("", "", "src", "app.ts", 11)
	assert.True(t, foundMissed)
	assert.Equal(t, 1, missed.MissedInstructionCount)
}

func Test_GetCoverageData_MatchesWhenSourceDirStripped(t *testing.T) {
	// Report path omits the "src" source dir the diff carries.
	report, e := parse(strings.NewReader("SF:app.ts\nDA:1,1\nend_of_record\n"))
	assert.NoError(t, e)

	data, found := report.GetCoverageData("", "src", "", "app.ts", 1)
	assert.True(t, found)
	assert.Equal(t, 1, data.CoveredInstructionCount)
}
