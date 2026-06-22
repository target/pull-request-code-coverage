// Package githubdiff fetches a pull request's unified diff directly from the
// GitHub REST API, so the plugin can determine what a PR changed without a git
// checkout. It is an alternative to reading the diff piped in on stdin.
package githubdiff

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/pkg/errors"
	"github.com/target/pull-request-code-coverage/internal/plugin/pluginhttp"
)

const httpResponseOK = 200

// Loader retrieves the diff for a single pull request from the GitHub API.
type Loader struct {
	apiKey     string
	apiBaseURL string
	pr         string
	owner      string
	repo       string
	httpClient pluginhttp.Client
}

func NewLoader(apiKey string, apiBaseURL string, pr string, owner string, repo string, httpClient pluginhttp.Client) *Loader {
	return &Loader{
		apiKey:     apiKey,
		apiBaseURL: apiBaseURL,
		pr:         pr,
		owner:      owner,
		repo:       repo,
		httpClient: httpClient,
	}
}

// Load requests the pull request diff using the `application/vnd.github.v3.diff`
// media type. GitHub returns the same unified diff it shows reviewers — computed
// against the merge base and carrying context lines — which the unified-diff
// parser handles. The whole response is read into memory and returned as a
// reader so the caller can treat it exactly like the stdin diff.
func (l *Loader) Load() (io.Reader, error) {
	url := fmt.Sprintf("%v/repos/%v/%v/pulls/%v", strings.TrimRight(l.apiBaseURL, "/"), l.owner, l.repo, l.pr)

	req, newErr := l.httpClient.NewRequest("GET", url, nil)
	if newErr != nil {
		return nil, errors.Wrap(newErr, "Failed creating request to github")
	}

	req.Header.Add("Authorization", "token "+l.apiKey)
	req.Header.Add("Accept", "application/vnd.github.v3.diff")

	resp, doErr := l.httpClient.Do(req)
	if doErr != nil {
		return nil, errors.Wrap(doErr, "Failed calling github")
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != httpResponseOK {
		return nil, errors.Errorf("Failed calling github: bad status code: %v", resp.StatusCode)
	}

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return nil, errors.Wrap(readErr, "Failed reading diff response from github")
	}

	return bytes.NewReader(body), nil
}
