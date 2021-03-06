package reporter

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/pkg/errors"
	"github.com/target/pull-request-code-coverage/internal/plugin/domain"
	"github.com/target/pull-request-code-coverage/internal/plugin/pluginhttp"
	"github.com/target/pull-request-code-coverage/internal/plugin/pluginjson"
)

type GithubPullRequest struct {
	apiKey     string
	apiBaseURL string
	pr         string
	owner      string
	repo       string
	httpClient pluginhttp.Client
	jsonClient pluginjson.Client
}

const (
	HTTPResponseCreated = 201
)

func NewGithubPullRequest(apiKey string, apiBaseURL string, pr string, owner string, repo string, httpClient pluginhttp.Client, jsonClient pluginjson.Client) *GithubPullRequest {
	return &GithubPullRequest{
		apiKey:     apiKey,
		apiBaseURL: apiBaseURL,
		pr:         pr,
		owner:      owner,
		repo:       repo,
		httpClient: httpClient,
		jsonClient: jsonClient,
	}
}

func (s *GithubPullRequest) Write(changedLinesWithCoverage domain.SourceLineCoverageReport) error {

	if changedLinesWithCoverage.TotalLinesWithData() == 0 {
		return nil
	}

	body, bodyErr := s.createCommentBody(changedLinesWithCoverage)

	if bodyErr != nil {
		return errors.Wrap(bodyErr, "Failed creating payload for github")
	}

	url := fmt.Sprintf("%v/api/v3/repos/%v/%v/issues/%v/comments", s.apiBaseURL, s.owner, s.repo, s.pr)

	req, newErr := s.httpClient.NewRequest(
		"POST",
		url,
		body)

	if newErr != nil {
		return errors.Wrap(newErr, "Failed creating request to github")
	}

	req.Header.Add("Authorization", "token "+s.apiKey)
	req.Header.Add("Content-Type", "application/json")

	resp, doErr := s.httpClient.Do(req)

	if doErr != nil {
		return errors.Wrap(doErr, "Failed calling github")
	}

	if resp.StatusCode != HTTPResponseCreated {
		return errors.Errorf("Failed calling github: bad status code: %v", resp.StatusCode)
	}

	return nil
}

func (s *GithubPullRequest) GetName() string {
	return "github pull request  reporter"
}

func (s *GithubPullRequest) createCommentBody(changedLinesWithCoverage domain.SourceLineCoverageReport) (io.Reader, error) {

	modules := collectModules(changedLinesWithCoverage)

	summaryLines := []string{}

	if len(modules) > 0 {
		summaryLines = append(summaryLines, fmt.Sprintf("*Modules: %v*\n\n", strings.Join(modules, ", ")))
	}
	var missedInstructions string

	for _, r := range changedLinesWithCoverage {
		if r.MissedInstructionCount > 0 {
			missedInstructions += fmt.Sprintf("--- %v\n", lineDescription(r.SourceLine))
			missedInstructions += fmt.Sprintf("%v\n", r.LineValue)
		}
	}

	summaryLines = append(summaryLines, generateSummaryLines(changedLinesWithCoverage, func(linesWithDataCount int, linesWithoutDataCount int, covered int, missed int) []string {
		totalLines := linesWithDataCount + linesWithoutDataCount
		totalInstructions := covered + missed

		result := make([]string, 5)

		result[0] = fmt.Sprintf("Code Coverage Summary:\n\n")
		result[1] = fmt.Sprintf("Lines Without Coverage Data -> %.f%% (%d)\n", toPercent(safeDiv(float32(linesWithoutDataCount), float32(totalLines), 0)), linesWithoutDataCount)
		result[2] = fmt.Sprintf("Lines With Coverage Data    -> %.f%% (%d)\n", toPercent(safeDiv(float32(linesWithDataCount), float32(totalLines), 1)), linesWithDataCount)
		result[3] = fmt.Sprintf("Covered Instructions        -> **%.f%%** (%d)\n", toPercent(safeDiv(float32(covered), float32(totalInstructions), 1)), covered)
		result[4] = fmt.Sprintf("Missed Instructions         -> %.f%% (%d)\n", toPercent(safeDiv(float32(missed), float32(totalInstructions), 0)), missed)

		return result
	})...)

	var summary string
	if missedInstructions == "" {
		summary = strings.Join(summaryLines, "")
	} else {

		summaryWithoutInstructions := strings.Join(summaryLines, "")
		summary = summaryWithoutInstructions + "\n<details><summary>Missed Instructions summary</summary>\n\n" + "```\n" + missedInstructions + "```" +
			"\n</details>"
	}

	data := map[string]string{
		"body": summary,
	}

	dataBytes, marshalErr := s.jsonClient.Marshal(data)

	if marshalErr != nil {
		return nil, errors.Wrap(marshalErr, "Failed marshalling payload to json")
	}

	return bytes.NewBuffer(dataBytes), nil
}

func collectModules(changedLinesWithCoverage domain.SourceLineCoverageReport) []string {
	collector := map[string]bool{}

	for _, line := range changedLinesWithCoverage {
		if line.Module != "" {
			collector[line.Module] = true
		}
	}

	keys := make([]string, 0, len(collector))
	for k := range collector {
		keys = append(keys, k)
	}

	return keys
}
