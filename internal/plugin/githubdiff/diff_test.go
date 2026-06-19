package githubdiff

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/target/pull-request-code-coverage/internal/plugin/pluginhttp"
)

func TestLoader_Load_BuildsRequestAndReturnsDiff(t *testing.T) {
	mockClient := &pluginhttp.MockClient{}
	request := httptest.NewRequest("GET", "http://anywhere", nil)

	mockClient.
		On("NewRequest", "GET", "https://api.github.com/repos/some_org/some_repo/pulls/123", nil).
		Return(request, nil)
	mockClient.
		On("Do", request).
		Return(&http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader("THE DIFF"))}, nil)

	reader, err := NewLoader("SOME_API_KEY", "https://api.github.com", "123", "some_org", "some_repo", mockClient).Load()

	assert.NoError(t, err)

	body, _ := io.ReadAll(reader)
	assert.Equal(t, "THE DIFF", string(body))

	assert.Equal(t, "token SOME_API_KEY", request.Header.Get("Authorization"))
	assert.Equal(t, "application/vnd.github.v3.diff", request.Header.Get("Accept"))

	mockClient.AssertExpectations(t)
}

func TestLoader_Load_TrimsTrailingSlashOnBaseURL(t *testing.T) {
	mockClient := &pluginhttp.MockClient{}
	request := httptest.NewRequest("GET", "http://anywhere", nil)

	mockClient.
		On("NewRequest", "GET", "https://git.example.com/api/v3/repos/o/r/pulls/9", nil).
		Return(request, nil)
	mockClient.
		On("Do", request).
		Return(&http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(""))}, nil)

	_, err := NewLoader("k", "https://git.example.com/api/v3/", "9", "o", "r", mockClient).Load()

	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}

func TestLoader_Load_FailedNewRequest(t *testing.T) {
	mockClient := &pluginhttp.MockClient{}
	mockClient.On("NewRequest", mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("boom"))

	_, err := NewLoader("k", "https://api.github.com", "1", "o", "r", mockClient).Load()

	assert.EqualError(t, err, "Failed creating request to github: boom")
}

func TestLoader_Load_FailedDo(t *testing.T) {
	mockClient := &pluginhttp.MockClient{}
	request := httptest.NewRequest("GET", "http://anywhere", nil)

	mockClient.On("NewRequest", mock.Anything, mock.Anything, mock.Anything).Return(request, nil)
	mockClient.On("Do", request).Return(nil, errors.New("boom"))

	_, err := NewLoader("k", "https://api.github.com", "1", "o", "r", mockClient).Load()

	assert.EqualError(t, err, "Failed calling github: boom")
}

func TestLoader_Load_BadStatus(t *testing.T) {
	mockClient := &pluginhttp.MockClient{}
	request := httptest.NewRequest("GET", "http://anywhere", nil)

	mockClient.On("NewRequest", mock.Anything, mock.Anything, mock.Anything).Return(request, nil)
	mockClient.On("Do", request).Return(&http.Response{StatusCode: 404, Body: io.NopCloser(strings.NewReader(""))}, nil)

	_, err := NewLoader("k", "https://api.github.com", "1", "o", "r", mockClient).Load()

	assert.EqualError(t, err, "Failed calling github: bad status code: 404")
}
