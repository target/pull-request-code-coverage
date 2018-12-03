package reporter

import (
	"git.target.com/search-product-team/pull-request-code-coverage/internal/plugin/domain"
	"git.target.com/search-product-team/pull-request-code-coverage/internal/plugin/pluginhttp"
	"git.target.com/search-product-team/pull-request-code-coverage/internal/plugin/pluginjson"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGithubPullRequest_Write_FailedNewRequest(t *testing.T) {

	mockClient := &pluginhttp.MockClient{}
	mockClient.On("NewRequest", mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("something bad happened"))

	writer := &GithubPullRequest{
		apiBaseURL: "anything",
		httpClient: mockClient,
		jsonClient: &pluginjson.DefaultClient{},
	}

	e := writer.Write(domain.SourceLineCoverageReport{})

	assert.EqualError(t, e, "Failed creating request to github: something bad happened")
}

func TestGithubPullRequest_Write_FailedDo(t *testing.T) {

	mockClient := &pluginhttp.MockClient{}
	request := httptest.NewRequest("GET", "http://anywhere", nil)

	mockClient.On("NewRequest", mock.Anything, mock.Anything, mock.Anything).Return(request, nil)
	mockClient.On("Do", request).Return(nil, errors.New("something bad happened"))

	writer := &GithubPullRequest{
		apiBaseURL: "anything",
		httpClient: mockClient,
		jsonClient: &pluginjson.DefaultClient{},
	}

	e := writer.Write(domain.SourceLineCoverageReport{})

	assert.EqualError(t, e, "Failed calling github: something bad happened")
}

func TestGithubPullRequest_Write_FailedDo_BadStatus(t *testing.T) {

	mockClient := &pluginhttp.MockClient{}
	request := httptest.NewRequest("GET", "http://anywhere", nil)

	mockClient.On("NewRequest", mock.Anything, mock.Anything, mock.Anything).Return(request, nil)
	mockClient.On("Do", request).Return(&http.Response{StatusCode: 400}, nil)

	writer := &GithubPullRequest{
		apiBaseURL: "anything",
		httpClient: mockClient,
		jsonClient: &pluginjson.DefaultClient{},
	}

	e := writer.Write(domain.SourceLineCoverageReport{})

	assert.EqualError(t, e, "Failed calling github: bad status code: 400")
}

func TestGithubPullRequest_Write_FailedJsonMarshal(t *testing.T) {

	mockClient := &pluginjson.MockClient{}

	mockClient.On("Marshal", mock.Anything).Return(nil, errors.New("something bad happened"))

	writer := &GithubPullRequest{
		apiBaseURL: "anything",
		httpClient: &pluginhttp.DefaultClient{},
		jsonClient: mockClient,
	}

	e := writer.Write(domain.SourceLineCoverageReport{})

	assert.EqualError(t, e, "Failed creating payload for github: Failed marshalling payload to json: something bad happened")
}
