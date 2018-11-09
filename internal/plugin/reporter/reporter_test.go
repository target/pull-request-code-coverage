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

	s.WritingFunc = func(io.Writer, string, ...interface{}) (int, error) {
		return 0, errors.New("anything")
	}

	assert.Panics(t, func() { s.printf("anything") })
}
