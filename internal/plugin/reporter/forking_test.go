package reporter

import (
	"testing"

	"git.target.com/target/pull-request-code-coverage/internal/plugin/domain"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestForking_Write(t *testing.T) {
	mockReporter := &MockReporter{}

	mockReporter.On("Write", mock.Anything).Return(errors.New("something bad happened"))

	reporter := NewForking([]Reporter{mockReporter})

	err := reporter.Write(domain.SourceLineCoverageReport{})

	assert.EqualError(t, err, "Unexpected errors occurred: [something bad happened]")
}
