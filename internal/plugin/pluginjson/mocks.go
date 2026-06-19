package pluginjson

import (
	"github.com/stretchr/testify/mock"
)

type MockClient struct {
	mock.Mock
}

func (m *MockClient) Marshal(data interface{}) ([]byte, error) {
	args := m.Called(data)

	r := args.Get(0)
	e := args.Error(1)

	if r == nil {
		return nil, e
	}
	return r.([]byte), e
}

func (m *MockClient) Unmarshal(data []byte, v interface{}) error {
	args := m.Called(data, v)

	return args.Error(0)
}
