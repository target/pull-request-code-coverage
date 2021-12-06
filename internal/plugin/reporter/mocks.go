package reporter

import (
	"github.com/stretchr/testify/mock"
	"github.com/target/pull-request-code-coverage/internal/plugin/domain"
)

type MockReporter struct {
	mock.Mock
}

func (m *MockReporter) Write(d domain.SourceLineCoverageReport) error {
	args := m.Called(d)

	return args.Error(0)
}

func (m *MockReporter) GetName() string {
	args := m.Called()
	return args.String(0)

}
