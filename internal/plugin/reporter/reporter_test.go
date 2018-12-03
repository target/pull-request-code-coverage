package reporter

import (
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"io"
	"os"
	"testing"
)

func TestSimple_printf_willPanicOnWritingErrors(t *testing.T) {
	s := NewSimple(os.Stdout)

	s.WritingFuncf = func(io.Writer, string, ...interface{}) (int, error) {
		return 0, errors.New("anything")
	}

	s.WritingFunc = func(io.Writer, ...interface{}) (int, error) {
		return 0, errors.New("anything2")
	}

	assert.Panics(t, func() { s.printf("anything") })
	assert.Panics(t, func() { s.print("anything2") })
}
