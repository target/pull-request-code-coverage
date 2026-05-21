package mocks

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

const HTTPResponseCreated = 201

type CapturedRequest struct {
	req  *http.Request
	body []byte
}

func WithMockGithubAPI(doer func(mockServerURL string, requestAsserter GithubAPIRequestAsserter)) {
	asserter := &DefaultGithubAPIRequestAsserter{}

	ts := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			asserter.AddRequest(&CapturedRequest{
				req:  r,
				body: mustReadAll(r.Body),
			})

			w.WriteHeader(HTTPResponseCreated)
		}),
	)
	defer ts.Close()

	doer(ts.URL, asserter)
}

type GithubAPIRequestAsserter interface {
	AssertRequestWasMade(t *testing.T, path string, apikey string, body map[string]interface{})
	AssertRequestBodyContains(t *testing.T, path string, apikey string, substrings []string)
	AssertNoRequestsWereMade(t *testing.T)
}

type DefaultGithubAPIRequestAsserter struct {
	requests []*CapturedRequest
}

func (a *DefaultGithubAPIRequestAsserter) AssertNoRequestsWereMade(t *testing.T) {
	assert.Equal(t, 0, len(a.requests))
}

func (a *DefaultGithubAPIRequestAsserter) AssertRequestWasMade(t *testing.T, path string, apikey string, body map[string]interface{}) {
	for _, r := range a.requests {
		if r.req.URL.Path != path {
			continue
		}

		if r.req.Header.Get("Authorization") != "token "+apikey {
			continue
		}

		if r.req.Header.Get("Content-Type") != "application/json" {
			continue
		}

		var bodyData map[string]interface{}
		mustJSONUnmarshall(r.body, &bodyData)

		if assert.Equal(t, bodyData, body) {
			return
		}
	}

	assert.Fail(t, fmt.Sprintf("Request was not made for path=%v, apikey=%v, body=%v", path, apikey, body))
}

// AssertRequestBodyContains finds the request matching path/apikey and asserts
// its JSON "body" field contains every given substring. Use this (rather than an
// exact body match) for the PR comment, whose markdown formatting is expected to
// evolve — assert the meaningful content, not the exact bytes.
func (a *DefaultGithubAPIRequestAsserter) AssertRequestBodyContains(t *testing.T, path string, apikey string, substrings []string) {
	for _, r := range a.requests {
		if r.req.URL.Path != path {
			continue
		}

		if r.req.Header.Get("Authorization") != "token "+apikey {
			continue
		}

		if r.req.Header.Get("Content-Type") != "application/json" {
			continue
		}

		var bodyData map[string]interface{}
		mustJSONUnmarshall(r.body, &bodyData)
		body, _ := bodyData["body"].(string)

		for _, sub := range substrings {
			assert.Contains(t, body, sub)
		}

		return
	}

	assert.Fail(t, fmt.Sprintf("Request was not made for path=%v, apikey=%v", path, apikey))
}

func mustJSONUnmarshall(bytes []byte, result interface{}) {
	err := json.Unmarshal(bytes, result)

	if err != nil {
		panic(fmt.Sprintf("Unexpected error: %v", err))
	}
}

func mustReadAll(r io.Reader) []byte {
	result, err := ioutil.ReadAll(r)

	if err != nil {
		panic(fmt.Sprintf("Unexpected error: %v", err))
	}

	return result
}

func (a *DefaultGithubAPIRequestAsserter) AddRequest(request *CapturedRequest) {
	a.requests = append(a.requests, request)
}
