package reporter

import (
	"bytes"
	"fmt"
	"io"
	"net/url"
	"path/filepath"
	"sort"
	"strings"

	"github.com/pkg/errors"
	"github.com/target/pull-request-code-coverage/internal/plugin/domain"
	"github.com/target/pull-request-code-coverage/internal/plugin/pluginhttp"
	"github.com/target/pull-request-code-coverage/internal/plugin/pluginjson"
)

// DefaultGithubAPIBaseURL is the REST API root for public GitHub. It is used
// when no base URL is configured; GitHub Enterprise users override it via
// gh_api_base_url (e.g. https://git.target.com).
const DefaultGithubAPIBaseURL = "https://api.github.com"

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

	url := s.commentsURL()

	req, newErr := s.httpClient.NewRequest(
		"POST",
		url,
		body)

	if newErr != nil {
		return errors.Wrap(newErr, "Failed creating request to github")
	}

	req.Header.Add("Authorization", "token "+s.apiKey)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("User-Agent", "pull-request-code-coverage")

	resp, doErr := s.httpClient.Do(req)

	if doErr != nil {
		return errors.Wrap(doErr, "Failed calling github")
	}

	if resp.StatusCode != HTTPResponseCreated {
		return errors.Errorf("Failed calling github: bad status code: %v", resp.StatusCode)
	}

	return nil
}

// commentsURL builds the REST endpoint for posting an issue comment.
//
// Public GitHub (api.github.com) and Enterprise Cloud (api.*.ghe.com) serve the
// REST API at the host root, while Enterprise Server (e.g. git.target.com)
// serves it under /api/v3. We append /api/v3 only for the latter so existing
// host-only Enterprise configs keep working without change.
func (s *GithubPullRequest) commentsURL() string {
	base := strings.TrimRight(s.apiBaseURL, "/")

	if host := hostOf(base); !strings.HasPrefix(host, "api.") && !strings.HasSuffix(base, "/api/v3") {
		base += "/api/v3"
	}

	return fmt.Sprintf("%v/repos/%v/%v/issues/%v/comments", base, s.owner, s.repo, s.pr)
}

// hostOf returns the host component of a URL, or "" if it cannot be parsed.
func hostOf(rawURL string) string {
	if u, err := url.Parse(rawURL); err == nil {
		return u.Host
	}

	return ""
}

func (s *GithubPullRequest) GetName() string {
	return "github pull request  reporter"
}

// metricsLegend explains, in plain language, what each row of the metrics
// table means so readers of the PR comment don't have to guess.
const metricsLegend = "<details><summary>ℹ️ What do these metrics mean?</summary>\n\n" +
	"- **Covered instructions** — instructions (statements / bytecode) on the changed lines that were executed by your tests.\n" +
	"- **Missed instructions** — instructions on the changed lines that were **not** executed by any test.\n" +
	"- **Lines with coverage data** — changed lines the coverage tool tracks as executable code.\n" +
	"- **Lines without coverage data** — changed lines with no coverage information (comments, blank lines, declarations, etc.).\n\n" +
	"</details>\n"

func (s *GithubPullRequest) createCommentBody(changedLinesWithCoverage domain.SourceLineCoverageReport) (io.Reader, error) {

	var body strings.Builder

	body.WriteString("## 📊 Pull Request Code Coverage\n\n")
	body.WriteString("Coverage below is for **only the lines changed in this PR**.\n\n")

	if modules := collectModules(changedLinesWithCoverage); len(modules) > 0 {
		body.WriteString(fmt.Sprintf("**Modules:** %v\n\n", strings.Join(backtickEach(modules), ", ")))
	}

	body.WriteString(metricsTable(changedLinesWithCoverage))
	body.WriteString("\n")
	body.WriteString(metricsLegend)

	if details := missedInstructionsDetails(changedLinesWithCoverage); details != "" {
		body.WriteString("\n")
		body.WriteString(details)
	}

	data := map[string]string{
		"body": body.String(),
	}

	dataBytes, marshalErr := s.jsonClient.Marshal(data)

	if marshalErr != nil {
		return nil, errors.Wrap(marshalErr, "Failed marshalling payload to json")
	}

	return bytes.NewBuffer(dataBytes), nil
}

// metricsTable renders the coverage numbers as a GitHub-flavoured markdown table.
func metricsTable(changedLinesWithCoverage domain.SourceLineCoverageReport) string {
	rows := generateSummaryLines(changedLinesWithCoverage, func(linesWithDataCount int, linesWithoutDataCount int, covered int, missed int) []string {
		totalLines := linesWithDataCount + linesWithoutDataCount
		totalInstructions := covered + missed

		return []string{
			"| Metric | Coverage | Count |\n",
			"|:---|---:|---:|\n",
			fmt.Sprintf("| ✅ Covered instructions | **%.f%%** | %d |\n", toPercent(safeDiv(float32(covered), float32(totalInstructions), 1)), covered),
			fmt.Sprintf("| ❌ Missed instructions | %.f%% | %d |\n", toPercent(safeDiv(float32(missed), float32(totalInstructions), 0)), missed),
			fmt.Sprintf("| 📈 Lines with coverage data | %.f%% | %d |\n", toPercent(safeDiv(float32(linesWithDataCount), float32(totalLines), 1)), linesWithDataCount),
			fmt.Sprintf("| 📉 Lines without coverage data | %.f%% | %d |\n", toPercent(safeDiv(float32(linesWithoutDataCount), float32(totalLines), 0)), linesWithoutDataCount),
		}
	})

	return strings.Join(rows, "")
}

// missedInstructionsDetails renders a collapsible section listing each changed
// line that has missed instructions, with its source in a syntax-highlighted
// block. Returns "" when nothing was missed.
func missedInstructionsDetails(changedLinesWithCoverage domain.SourceLineCoverageReport) string {
	var missed []domain.SourceLineCoverage
	for _, r := range changedLinesWithCoverage {
		if r.MissedInstructionCount > 0 {
			missed = append(missed, r)
		}
	}

	if len(missed) == 0 {
		return ""
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("<details><summary>❌ Lines missing coverage (%d)</summary>\n\n", len(missed)))

	for _, r := range missed {
		b.WriteString(fmt.Sprintf("**`%v`**\n", lineDescription(r.SourceLine)))
		b.WriteString(fmt.Sprintf("```%v\n%v\n```\n\n", langForFile(r.FileName), r.LineValue))
	}

	b.WriteString("</details>\n")

	return b.String()
}

// langForFile maps a source file's extension to a markdown code-fence language
// so missed-line snippets get syntax highlighting. Unknown types render plain.
func langForFile(fileName string) string {
	switch strings.ToLower(filepath.Ext(fileName)) {
	case ".go":
		return "go"
	case ".java":
		return "java"
	case ".kt", ".kts":
		return "kotlin"
	case ".scala":
		return "scala"
	case ".py":
		return "python"
	case ".js":
		return "javascript"
	case ".ts":
		return "typescript"
	default:
		return ""
	}
}

func backtickEach(items []string) []string {
	out := make([]string, len(items))
	for i, item := range items {
		out[i] = "`" + item + "`"
	}

	return out
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

	sort.Strings(keys)

	return keys
}
