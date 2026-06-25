package reporter

import (
	"bytes"
	"fmt"
	"io"
	"sort"
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
	HTTPResponseOK      = 200
	HTTPResponseCreated = 201
)

// commentMarker is an HTML comment embedded at the top of the report. It renders
// invisibly on GitHub but lets a later run find the comment it posted earlier so
// it can update that one in place instead of posting a new comment every push.
const commentMarker = "<!-- pull-request-code-coverage:patch-coverage -->"

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

	existingID, findErr := s.findExistingCommentID()
	if findErr != nil {
		return findErr
	}

	// Update the comment from a previous run when we find one; otherwise post a
	// fresh comment. This keeps a single, always-current coverage comment on the
	// PR instead of a new one per push.
	if existingID != 0 {
		url := fmt.Sprintf("%v/repos/%v/%v/issues/comments/%v", s.baseURL(), s.owner, s.repo, existingID)
		return s.send("PATCH", url, body, HTTPResponseOK)
	}

	url := fmt.Sprintf("%v/repos/%v/%v/issues/%v/comments", s.baseURL(), s.owner, s.repo, s.pr)

	return s.send("POST", url, body, HTTPResponseCreated)
}

// baseURL returns the configured GitHub API root without a trailing slash.
func (s *GithubPullRequest) baseURL() string {
	return strings.TrimRight(s.apiBaseURL, "/")
}

// send issues a write request (POST/PATCH) carrying the comment payload and
// verifies the response status.
func (s *GithubPullRequest) send(method string, url string, body io.Reader, wantStatus int) error {
	req, newErr := s.httpClient.NewRequest(method, url, body)

	if newErr != nil {
		return errors.Wrap(newErr, "Failed creating request to github")
	}

	req.Header.Add("Authorization", "token "+s.apiKey)
	req.Header.Add("Content-Type", "application/json")

	resp, doErr := s.httpClient.Do(req)

	if doErr != nil {
		return errors.Wrap(doErr, "Failed calling github")
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != wantStatus {
		return errors.Errorf("Failed calling github: bad status code: %v", resp.StatusCode)
	}

	return nil
}

// findExistingCommentID looks for a coverage comment this plugin posted on an
// earlier run, identified by the hidden commentMarker. It returns 0 when none is
// found. Only the first page of comments is checked (per_page=100), which covers
// any realistic PR. The GET only needs read access, so it also works on fork PRs
// even though the follow-up write may not.
func (s *GithubPullRequest) findExistingCommentID() (int64, error) {
	url := fmt.Sprintf("%v/repos/%v/%v/issues/%v/comments?per_page=100", s.baseURL(), s.owner, s.repo, s.pr)

	req, newErr := s.httpClient.NewRequest("GET", url, nil)
	if newErr != nil {
		return 0, errors.Wrap(newErr, "Failed creating request to github")
	}

	req.Header.Add("Authorization", "token "+s.apiKey)

	resp, doErr := s.httpClient.Do(req)
	if doErr != nil {
		return 0, errors.Wrap(doErr, "Failed calling github")
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != HTTPResponseOK {
		return 0, errors.Errorf("Failed listing github comments: bad status code: %v", resp.StatusCode)
	}

	respBody, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return 0, errors.Wrap(readErr, "Failed reading github comments response")
	}

	var comments []struct {
		ID   int64  `json:"id"`
		Body string `json:"body"`
	}

	if unmarshalErr := s.jsonClient.Unmarshal(respBody, &comments); unmarshalErr != nil {
		return 0, errors.Wrap(unmarshalErr, "Failed parsing github comments response")
	}

	for _, c := range comments {
		if strings.Contains(c.Body, commentMarker) {
			return c.ID, nil
		}
	}

	return 0, nil
}

func (s *GithubPullRequest) GetName() string {
	return "github pull request  reporter"
}

func (s *GithubPullRequest) createCommentBody(changedLinesWithCoverage domain.SourceLineCoverageReport) (io.Reader, error) {

	modules := collectModules(changedLinesWithCoverage)

	covered := changedLinesWithCoverage.TotalCoveredInstructions()
	missed := changedLinesWithCoverage.TotalMissedInstructions()
	linesWithData := changedLinesWithCoverage.TotalLinesWithData()
	linesWithoutData := changedLinesWithCoverage.TotalLinesWithoutData()

	totalInstructions := covered + missed
	totalLines := linesWithData + linesWithoutData

	coveredPct := toPercent(safeDiv(float32(covered), float32(totalInstructions), 1))
	missedPct := toPercent(safeDiv(float32(missed), float32(totalInstructions), 0))
	withDataPct := toPercent(safeDiv(float32(linesWithData), float32(totalLines), 1))
	withoutDataPct := toPercent(safeDiv(float32(linesWithoutData), float32(totalLines), 0))

	var b strings.Builder

	// Hidden marker so a later run can find and update this comment in place.
	b.WriteString(commentMarker + "\n")
	fmt.Fprintf(&b, "## 🛡️ Patch Coverage Report — `%.f%%` %v\n\n",
		coveredPct, coverageStatusEmoji(coveredPct))
	b.WriteString("> Scope: **changed lines only** — the code this PR adds or edits, not whole files or the repo. ")
	b.WriteString("It answers one thing — *did your tests run the code you just touched?*\n\n")

	if len(modules) > 0 {
		fmt.Fprintf(&b, "*Modules:* %v\n\n", strings.Join(modules, ", "))
	}

	fmt.Fprintf(&b, "**Diff coverage:** `%.f%%` %v — `%d` of `%d` changed instructions covered\n\n",
		coveredPct, coverageStatusEmoji(coveredPct), covered, totalInstructions)

	b.WriteString("| Metric | Value | |\n")
	b.WriteString("| :-- | --: | :-- |\n")
	fmt.Fprintf(&b, "| 🟢 Covered instructions | `%d` (%.f%%) | changed code your tests executed |\n", covered, coveredPct)
	fmt.Fprintf(&b, "| 🔴 Missed instructions | `%d` (%.f%%) | changed code your tests never ran |\n", missed, missedPct)
	fmt.Fprintf(&b, "| 📈 Tracked changed lines | `%d` (%.f%%) | lines the coverage tool could measure |\n", linesWithData, withDataPct)
	fmt.Fprintf(&b, "| ⚪ Untracked changed lines | `%d` (%.f%%) | comments, blanks, declarations |\n", linesWithoutData, withoutDataPct)
	b.WriteString("\n")
	b.WriteString("<sub>**Lines** = the source lines you changed. **Instructions** = the executable units the coverage tool counts inside those lines — one line can hold several (e.g. JaCoCo bytecode), so the two counts differ.</sub>\n\n")

	b.WriteString(fileCoverageSection(changedLinesWithCoverage))
	b.WriteString(missedInstructionsSection(changedLinesWithCoverage))
	b.WriteString("\n<sub>🤖 Generated by <a href=\"https://github.com/target/pull-request-code-coverage\">pull-request-code-coverage</a> — coverage for changed lines only.</sub>\n")

	data := map[string]string{
		"body": b.String(),
	}

	dataBytes, marshalErr := s.jsonClient.Marshal(data)

	if marshalErr != nil {
		return nil, errors.Wrap(marshalErr, "Failed marshalling payload to json")
	}

	return bytes.NewBuffer(dataBytes), nil
}

// fileCoverage holds the aggregated changed-line coverage for a single file.
type fileCoverage struct {
	path             string
	covered          int
	missed           int
	linesWithData    int
	linesWithoutData int
}

// fileCoverageSection renders a per-file breakdown of changed-line coverage,
// worst-covered files first so the riskiest changes surface at the top.
func fileCoverageSection(changedLinesWithCoverage domain.SourceLineCoverageReport) string {
	files := collectFileCoverage(changedLinesWithCoverage)

	if len(files) == 0 {
		return ""
	}

	// Files whose only changed lines carry no coverage data (config, docs,
	// generated, test-only) would just be a wall of "n/a" rows, so keep them
	// out of the table and summarise the count underneath instead.
	unmeasured := 0

	var b strings.Builder

	b.WriteString("### Coverage by file\n\n")
	b.WriteString("| File | Diff coverage | Covered / Missed |\n")
	b.WriteString("| :-- | :-: | :-: |\n")

	for _, f := range files {
		instructions := f.covered + f.missed

		if instructions == 0 {
			unmeasured++
			continue
		}

		pct := toPercent(safeDiv(float32(f.covered), float32(instructions), 1))
		fmt.Fprintf(&b, "| `%v` | %v %.f%% | %d / %d |\n",
			f.path, coverageStatusEmoji(pct), pct, f.covered, f.missed)
	}

	b.WriteString("\n")

	if unmeasured > 0 {
		fmt.Fprintf(&b, "<sub>%d changed file(s) with no measurable lines (config, docs, generated, or test-only) omitted.</sub>\n\n", unmeasured)
	}

	return b.String()
}

// collectFileCoverage aggregates changed-line coverage per file, sorted with the
// lowest-covered files first (files with no measurable instructions last) so a
// reviewer sees the riskiest files at the top of the table.
func collectFileCoverage(changedLinesWithCoverage domain.SourceLineCoverageReport) []fileCoverage {
	order := []string{}
	byPath := map[string]*fileCoverage{}

	for _, line := range changedLinesWithCoverage {
		path := filePath(line.SourceLine)

		fc, ok := byPath[path]
		if !ok {
			fc = &fileCoverage{path: path}
			byPath[path] = fc
			order = append(order, path)
		}

		fc.covered += line.CoveredInstructionCount
		fc.missed += line.MissedInstructionCount

		if line.HasData() {
			fc.linesWithData++
		} else {
			fc.linesWithoutData++
		}
	}

	result := make([]fileCoverage, 0, len(order))
	for _, path := range order {
		result = append(result, *byPath[path])
	}

	sort.SliceStable(result, func(i, j int) bool {
		return fileCoveragePct(result[i]) < fileCoveragePct(result[j])
	})

	return result
}

// fileCoveragePct returns a file's changed-line coverage as a sortable value;
// files with no measurable instructions sort last (returned above 100%).
func fileCoveragePct(f fileCoverage) float32 {
	instructions := f.covered + f.missed
	if instructions == 0 {
		return 101
	}

	return toPercent(safeDiv(float32(f.covered), float32(instructions), 1))
}

// missedInstructionsSection renders a collapsible block listing each changed
// line that was not executed by tests. Returns "" when nothing was missed.
func missedInstructionsSection(changedLinesWithCoverage domain.SourceLineCoverageReport) string {
	var missedInstructions string
	missedLineCount := 0

	for _, r := range changedLinesWithCoverage {
		if r.MissedInstructionCount > 0 {
			missedLineCount++
			missedInstructions += fmt.Sprintf("--- %v\n", lineDescription(r.SourceLine))
			missedInstructions += fmt.Sprintf("%v\n", r.LineValue)
		}
	}

	if missedInstructions == "" {
		return ""
	}

	return fmt.Sprintf("\n<details><summary>🔍 Uncovered lines (%d)</summary>\n\n", missedLineCount) +
		"```\n" + missedInstructions + "```" + "\n</details>\n"
}

// coverageStatusEmoji maps a covered-instruction percentage to a traffic-light
// status icon, so the headline reads at a glance.
func coverageStatusEmoji(coveredPct float32) string {
	switch {
	case coveredPct >= 80:
		return "🟢"
	case coveredPct >= 50:
		return "🟡"
	default:
		return "🔴"
	}
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
