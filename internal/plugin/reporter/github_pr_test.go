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

func sampleReport() domain.SourceLineCoverageReport {
	return domain.SourceLineCoverageReport{
		domain.SourceLineCoverage{
			CoverageData: domain.CoverageData{
				CoveredInstructionCount: 1,
			},
		},
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

	e := writer.Write(sampleReport())

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

	e := writer.Write(sampleReport())

	assert.EqualError(t, e, "Failed calling github: something bad happened")
}

func TestGithubPullRequest_Write_FailedListingComments_BadStatus(t *testing.T) {

	mockClient := &pluginhttp.MockClient{}
	request := httptest.NewRequest("GET", "http://anywhere", nil)

	mockClient.On("NewRequest", "GET", mock.Anything, mock.Anything).Return(request, nil)
	mockClient.On("Do", request).Return(&http.Response{StatusCode: 400, Body: io.NopCloser(strings.NewReader(""))}, nil)

	writer := NewGithubPullRequest("KEY", "https://api.github.com", "42", "some_owner", "some_repo", mockClient, &pluginjson.DefaultClient{})

	e := writer.Write(sampleReport())

	assert.EqualError(t, e, "Failed listing github comments: bad status code: 400")
}

func TestGithubPullRequest_Write_CreatesCommentWhenNoneExists(t *testing.T) {

	mockClient := &pluginhttp.MockClient{}
	listReq := httptest.NewRequest("GET", "http://list", nil)
	postReq := httptest.NewRequest("POST", "http://create", nil)

	mockClient.On("NewRequest", "GET", "https://api.github.com/repos/some_owner/some_repo/issues/42/comments?per_page=100", mock.Anything).Return(listReq, nil)
	mockClient.On("Do", listReq).Return(&http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader("[]"))}, nil)

	mockClient.On("NewRequest", "POST", "https://api.github.com/repos/some_owner/some_repo/issues/42/comments", mock.Anything).Return(postReq, nil)
	mockClient.On("Do", postReq).Return(&http.Response{StatusCode: 201, Body: io.NopCloser(strings.NewReader(""))}, nil)

	writer := NewGithubPullRequest("KEY", "https://api.github.com", "42", "some_owner", "some_repo", mockClient, &pluginjson.DefaultClient{})

	e := writer.Write(sampleReport())

	assert.NoError(t, e)
	mockClient.AssertExpectations(t)
}

func TestGithubPullRequest_Write_UpdatesExistingComment(t *testing.T) {

	mockClient := &pluginhttp.MockClient{}
	listReq := httptest.NewRequest("GET", "http://list", nil)
	patchReq := httptest.NewRequest("PATCH", "http://update", nil)

	existing := `[{"id": 7, "body": "stale"}, {"id": 99, "body": "old report ` + commentMarker + ` here"}]`

	mockClient.On("NewRequest", "GET", "https://api.github.com/repos/some_owner/some_repo/issues/42/comments?per_page=100", mock.Anything).Return(listReq, nil)
	mockClient.On("Do", listReq).Return(&http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(existing))}, nil)

	mockClient.On("NewRequest", "PATCH", "https://api.github.com/repos/some_owner/some_repo/issues/comments/99", mock.Anything).Return(patchReq, nil)
	mockClient.On("Do", patchReq).Return(&http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(""))}, nil)

	writer := NewGithubPullRequest("KEY", "https://api.github.com", "42", "some_owner", "some_repo", mockClient, &pluginjson.DefaultClient{})

	e := writer.Write(sampleReport())

	assert.NoError(t, e)
	mockClient.AssertExpectations(t)
}

func TestGithubPullRequest_Write_TrimsTrailingSlashFromEnterpriseURL(t *testing.T) {

	mockClient := &pluginhttp.MockClient{}
	listReq := httptest.NewRequest("GET", "http://list", nil)
	postReq := httptest.NewRequest("POST", "http://create", nil)

	mockClient.On("NewRequest", "GET", "https://git.target.com/api/v3/repos/some_owner/some_repo/issues/42/comments?per_page=100", mock.Anything).Return(listReq, nil)
	mockClient.On("Do", listReq).Return(&http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader("[]"))}, nil)

	mockClient.On("NewRequest", "POST", "https://git.target.com/api/v3/repos/some_owner/some_repo/issues/42/comments", mock.Anything).Return(postReq, nil)
	mockClient.On("Do", postReq).Return(&http.Response{StatusCode: 201, Body: io.NopCloser(strings.NewReader(""))}, nil)

	writer := NewGithubPullRequest("KEY", "https://git.target.com/api/v3/", "42", "some_owner", "some_repo", mockClient, &pluginjson.DefaultClient{})

	e := writer.Write(sampleReport())

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

	e := writer.Write(sampleReport())

	assert.EqualError(t, e, "Failed creating payload for github: Failed marshalling payload to json: something bad happened")
}
