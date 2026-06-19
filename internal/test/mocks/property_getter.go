package mocks

import "github.com/stretchr/testify/mock"

type MockPropertyGetter struct {
	mock.Mock
}

func NewMockPropertyGetter() *MockPropertyGetter {
	return &MockPropertyGetter{}
}

func (m *MockPropertyGetter) GetProperty(s string) (string, bool) {
	// Properties a test did not explicitly stub resolve to ("", false), the same
	// as os.LookupEnv for an unset variable. This keeps each test focused on the
	// properties it cares about and means looking up a newer optional property
	// (e.g. PARAMETER_DIFF_SOURCE) does not panic in tests written before it
	// existed. Stubbed properties still go through testify so AssertExpectations
	// continues to verify them.
	if !m.hasExpectedCall(s) {
		return "", false
	}

	args := m.Called(s)
	return args.Get(0).(string), args.Bool(1)
}

func (m *MockPropertyGetter) hasExpectedCall(s string) bool {
	for _, c := range m.ExpectedCalls {
		if c.Method != "GetProperty" || len(c.Arguments) != 1 {
			continue
		}

		if name, ok := c.Arguments[0].(string); ok && name == s {
			return true
		}
	}

	return false
}
