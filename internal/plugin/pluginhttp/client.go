package pluginhttp

import (
	"io"
	"net/http"
)

type Client interface {
	NewRequest(method string, url string, body io.Reader) (*http.Request, error)
	Do(request *http.Request) (*http.Response, error)
}

type DefaultClient struct{}

func (c *DefaultClient) NewRequest(method string, url string, body io.Reader) (*http.Request, error) {
	return http.NewRequest(method, url, body)
}

func (c *DefaultClient) Do(request *http.Request) (*http.Response, error) {
	return http.DefaultClient.Do(request)
}
