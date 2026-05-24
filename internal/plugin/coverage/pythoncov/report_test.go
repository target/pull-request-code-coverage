package pythoncov

import (
	"io"
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

func Test_Load_ReadAllFails(t *testing.T) {
	l := NewReportLoader()
	l.readAllFunc = func(io.Reader) ([]byte, error) {
		return nil, errors.New("anything")
	}

	_, e := l.Load("../../../test/example_python_coverage.xml")
	assert.EqualError(t, e, "Failed reading in all of coverage file ../../../test/example_python_coverage.xml: anything")
}

func Test_Load_BadXml(t *testing.T) {
	l := NewReportLoader()
	_, e := l.Load("../../../test/jacocoTestReport.json")
	assert.EqualError(t, e, "Failed unmarshalling coverage file ../../../test/jacocoTestReport.json: EOF")
}

func Test_Load_NoFile(t *testing.T) {
	l := NewReportLoader()
	_, e := l.Load("../../../test/anything.xml")
	assert.EqualError(t, e, "Could not open xml file ../../../test/anything.xml: open ../../../test/anything.xml: no such file or directory")
}

func loadReport(t *testing.T) *Report {
	r, e := NewReportLoader().Load("../../../test/example_python_coverage.xml")
	assert.NoError(t, e)

	report, ok := r.(*Report)
	assert.True(t, ok)

	return report
}

func Test_GetCoverageData_CoveredLine(t *testing.T) {
	data, found := loadReport(t).GetCoverageData("", "", "myapp", "calculator.py", 1)

	assert.True(t, found)
	assert.Equal(t, 1, data.CoveredInstructionCount)
	assert.Equal(t, 0, data.MissedInstructionCount)
}

func Test_GetCoverageData_MissedLine(t *testing.T) {
	data, found := loadReport(t).GetCoverageData("", "", "myapp", "calculator.py", 6)

	assert.True(t, found)
	assert.Equal(t, 0, data.CoveredInstructionCount)
	assert.Equal(t, 1, data.MissedInstructionCount)
}

func Test_GetCoverageData_UntrackedLineNotFound(t *testing.T) {
	_, found := loadReport(t).GetCoverageData("", "", "myapp", "calculator.py", 3)

	assert.False(t, found)
}

func Test_GetCoverageData_UnknownFileNotFound(t *testing.T) {
	_, found := loadReport(t).GetCoverageData("", "", "other", "missing.py", 1)

	assert.False(t, found)
}

func Test_GetCoverageData_MatchesWhenSourceDirStripped(t *testing.T) {
	// coverage.py wrote filenames relative to a "src"-like <source> root, so the
	// report path ("myapp/calculator.py") omits the source dir the diff carries.
	data, found := loadReport(t).GetCoverageData("", "src", "myapp", "calculator.py", 1)

	assert.True(t, found)
	assert.Equal(t, 1, data.CoveredInstructionCount)
}
