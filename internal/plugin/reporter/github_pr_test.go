package reporter

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/target/pull-request-code-coverage/internal/plugin/domain"
	"github.com/target/pull-request-code-coverage/internal/plugin/pluginhttp"
	"github.com/target/pull-request-code-coverage/internal/plugin/pluginjson"
)

func TestGithubPullRequest_Write_FailedNewRequest(t *testing.T) {

	mockClient := &pluginhttp.MockClient{}
	mockClient.On("NewRequest", mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("something bad happened"))

	writer := &GithubPullRequest{
		apiBaseURL: "anything",
		httpClient: mockClient,
		jsonClient: &pluginjson.DefaultClient{},
	}

	e := writer.Write(domain.SourceLineCoverageReport{
		domain.SourceLineCoverage{
			CoverageData: domain.CoverageData{
				CoveredInstructionCount: 1,
			},
		},
	})

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

	e := writer.Write(domain.SourceLineCoverageReport{
		domain.SourceLineCoverage{
			CoverageData: domain.CoverageData{
				CoveredInstructionCount: 1,
			},
		},
	})

	assert.EqualError(t, e, "Failed calling github: something bad happened")
}

func TestGithubPullRequest_Write_FailedDo_BadStatus(t *testing.T) {

	mockClient := &pluginhttp.MockClient{}
	request := httptest.NewRequest("GET", "http://anywhere", nil)

	mockClient.On("NewRequest", mock.Anything, mock.Anything, mock.Anything).Return(request, nil)
	mockClient.On("Do", request).Return(&http.Response{StatusCode: 400, Body: io.NopCloser(strings.NewReader(""))}, nil)

	writer := &GithubPullRequest{
		apiBaseURL: "anything",
		httpClient: mockClient,
		jsonClient: &pluginjson.DefaultClient{},
	}

	e := writer.Write(domain.SourceLineCoverageReport{
		domain.SourceLineCoverage{
			CoverageData: domain.CoverageData{
				CoveredInstructionCount: 1,
			},
		},
	})

	assert.EqualError(t, e, "Failed calling github: bad status code: 400")
}

func TestGithubPullRequest_Write_BuildsPublicGithubURL(t *testing.T) {

	mockClient := &pluginhttp.MockClient{}
	request := httptest.NewRequest("POST", "http://anywhere", nil)

	mockClient.On("NewRequest", "POST", "https://api.github.com/repos/some_owner/some_repo/issues/42/comments", mock.Anything).Return(request, nil)
	mockClient.On("Do", request).Return(&http.Response{StatusCode: 201, Body: io.NopCloser(strings.NewReader(""))}, nil)

	writer := NewGithubPullRequest("KEY", "https://api.github.com", "42", "some_owner", "some_repo", mockClient, &pluginjson.DefaultClient{})

	e := writer.Write(domain.SourceLineCoverageReport{
		domain.SourceLineCoverage{
			CoverageData: domain.CoverageData{
				CoveredInstructionCount: 1,
			},
		},
	})

	assert.NoError(t, e)
	mockClient.AssertExpectations(t)
}

func TestGithubPullRequest_Write_TrimsTrailingSlashFromEnterpriseURL(t *testing.T) {

	mockClient := &pluginhttp.MockClient{}
	request := httptest.NewRequest("POST", "http://anywhere", nil)

	mockClient.On("NewRequest", "POST", "https://git.target.com/api/v3/repos/some_owner/some_repo/issues/42/comments", mock.Anything).Return(request, nil)
	mockClient.On("Do", request).Return(&http.Response{StatusCode: 201, Body: io.NopCloser(strings.NewReader(""))}, nil)

	writer := NewGithubPullRequest("KEY", "https://git.target.com/api/v3/", "42", "some_owner", "some_repo", mockClient, &pluginjson.DefaultClient{})

	e := writer.Write(domain.SourceLineCoverageReport{
		domain.SourceLineCoverage{
			CoverageData: domain.CoverageData{
				CoveredInstructionCount: 1,
			},
		},
	})

	assert.NoError(t, e)
	mockClient.AssertExpectations(t)
}

func TestGithubPullRequest_Write_FailedJsonMarshal(t *testing.T) {

	mockClient := &pluginjson.MockClient{}

	mockClient.On("Marshal", mock.Anything).Return(nil, errors.New("something bad happened"))

	writer := &GithubPullRequest{
		apiBaseURL: "anything",
		httpClient: &pluginhttp.DefaultClient{},
		jsonClient: mockClient,
	}

	e := writer.Write(domain.SourceLineCoverageReport{
		domain.SourceLineCoverage{
			CoverageData: domain.CoverageData{
				CoveredInstructionCount: 1,
			},
		},
	})

	assert.EqualError(t, e, "Failed creating payload for github: Failed marshalling payload to json: something bad happened")
}
