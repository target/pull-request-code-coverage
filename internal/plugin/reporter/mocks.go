package reporter

import (
	"git.target.com/target/pull-request-code-coverage/internal/plugin/domain"
	"github.com/stretchr/testify/mock"
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
