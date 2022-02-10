package cobertura

import (
	"github.com/pkg/errors"

	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_silentlyCall(t *testing.T) {
	assert.Panics(t, func() {
		silentlyCall(func() error {
			return errors.New("anything")
		})
	})
}

func Test_Load_ReadAllFailes(t *testing.T) {
	l := NewReportLoader("anything")
	l.readAllFunc = func(io.Reader) ([]byte, error) {
		return nil, errors.New("anything")
	}

	_, e := l.Load("../../../test/example_go_coverage.xml")
	assert.EqualError(t, e, "Failed reading in all of coverage file ../../../test/example_go_coverage.xml: anything")
}

func Test_Load_BadXml(t *testing.T) {
	l := NewReportLoader("anything")
	_, e := l.Load("../../../test/jacocoTestReport.json")
	assert.EqualError(t, e, "Failed unmarshalling coverage file ../../../test/jacocoTestReport.json: EOF")
}

func Test_Load_NoFile(t *testing.T) {
	l := NewReportLoader("anything")
	_, e := l.Load("../../../test/anything.xml")
	assert.EqualError(t, e, "Could not open xml file ../../../test/anything.xml: open ../../../test/anything.xml: no such file or directory")
}
