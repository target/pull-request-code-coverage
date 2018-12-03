package pluginhttp

import (
	"github.com/stretchr/testify/mock"
	"io"
	"net/http"
)

type MockClient struct {
	mock.Mock
}

func (m *MockClient) NewRequest(method string, url string, body io.Reader) (*http.Request, error) {
	args := m.Called(method, url, body)

	r := args.Get(0)
	e := args.Error(1)

	if r == nil {
		return nil, e
	}
	return r.(*http.Request), e
}

func (m *MockClient) Do(request *http.Request) (*http.Response, error) {
	args := m.Called(request)

	r := args.Get(0)
	e := args.Error(1)

	if r == nil {
		return nil, e
	}
	return r.(*http.Response), e
}
