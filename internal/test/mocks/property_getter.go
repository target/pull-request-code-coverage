package mocks

import "github.com/stretchr/testify/mock"

type MockPropertyGetter struct {
	mock.Mock
}

func NewMockPropertyGetter() *MockPropertyGetter {
	return &MockPropertyGetter{}
}

func (m *MockPropertyGetter) GetProperty(s string) (string, bool) {
	args := m.Called(s)
	return args.Get(0).(string), args.Bool(1)
}
