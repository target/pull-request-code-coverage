package reporter

import (
	"git.target.com/search-product-team/pull-request-code-coverage/internal/plugin/domain"
	"github.com/stretchr/testify/mock"
)

type MockReporter struct {
	mock.Mock
}

func (m *MockReporter) Write(d domain.SourceLineCoverageReport) error {
	args := m.Called(d)

	return args.Error(0)
}
