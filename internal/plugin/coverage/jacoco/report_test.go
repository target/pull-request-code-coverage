package jacoco

import (
	"github.com/pkg/errors"

	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_silentlyCall(t *testing.T) {
	assert.Panics(t, func() {
		silentlyCall(func() error {
			return errors.New("anuthing")
		})
	})
}

func Test_Load_ReadAllFailes(t *testing.T) {
	l := NewReportLoader()
	l.readAllFunc = func(io.Reader) ([]byte, error) {
		return nil, errors.New("anything")
	}

	_, e := l.Load("../../../test/jacocoTestReport.xml")
	assert.EqualError(t, e, "Failed reading in all of coverage file ../../../test/jacocoTestReport.xml: anything")
}
