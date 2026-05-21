package reporter

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/target/pull-request-code-coverage/internal/plugin/domain"
	"github.com/target/pull-request-code-coverage/internal/plugin/pluginhttp"
	"github.com/target/pull-request-code-coverage/internal/plugin/pluginjson"
)

func TestGithubPullRequest_commentsURL(t *testing.T) {

	tests := []struct {
		name       string
		apiBaseURL string
		expected   string
	}{
		{
			name:       "github enterprise host gets /api/v3 appended",
			apiBaseURL: "https://git.target.com",
			expected:   "https://git.target.com/api/v3/repos/some_org/some_repo/issues/123/comments",
		},
		{
			name:       "trailing slash is trimmed",
			apiBaseURL: "https://git.target.com/",
			expected:   "https://git.target.com/api/v3/repos/some_org/some_repo/issues/123/comments",
		},
		{
			name:       "base url already pointing at /api/v3 is not doubled",
			apiBaseURL: "https://git.target.com/api/v3",
			expected:   "https://git.target.com/api/v3/repos/some_org/some_repo/issues/123/comments",
		},
		{
			name:       "public github uses api.github.com without /api/v3",
			apiBaseURL: "https://api.github.com",
			expected:   "https://api.github.com/repos/some_org/some_repo/issues/123/comments",
		},
		{
			name:       "enterprise cloud data residency api host without /api/v3",
			apiBaseURL: "https://api.acme.ghe.com",
			expected:   "https://api.acme.ghe.com/repos/some_org/some_repo/issues/123/comments",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			writer := &GithubPullRequest{
				apiBaseURL: tt.apiBaseURL,
				owner:      "some_org",
				repo:       "some_repo",
				pr:         "123",
			}

			assert.Equal(t, tt.expected, writer.commentsURL())
		})
	}
}

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
	mockClient.On("Do", request).Return(&http.Response{StatusCode: 400}, nil)

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
