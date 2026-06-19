package plugin

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/target/pull-request-code-coverage/internal/test/mocks"
)

// When PARAMETER_DIFF_SOURCE is unset the runner must behave exactly as before:
// read the diff from the reader it was given (stdin). Covered implicitly by the
// existing golden tests; this asserts the explicit "stdin" value is equivalent.
func TestDefaultRunner_Run_DiffSourceStdin_Explicit(t *testing.T) {
	propGetter := mocks.NewMockPropertyGetter()

	propGetter.On("GetProperty", "PARAMETER_COVERAGE_FILE").Return("../test/jacocoTestReport.xml", true)
	propGetter.On("GetProperty", "PARAMETER_COVERAGE_TYPE").Return("jacoco", true)
	propGetter.On("GetProperty", "PARAMETER_SOURCE_DIRS").Return("src/main/java", true)
	propGetter.On("GetProperty", "PARAMETER_MODULE").Return("category-search", true)
	propGetter.On("GetProperty", "PARAMETER_DIFF_SOURCE").Return("stdin", true)

	var buf bytes.Buffer

	err := NewRunner().Run(propGetter.GetProperty, MustOpen(t, "../test/sample_unified.diff"), &buf)
	assert.NoError(t, err)
	assert.Contains(t, buf.String(), "Patch Coverage Report")

	propGetter.AssertExpectations(t)
}

func TestDefaultRunner_Run_DiffSourceGithub_MissingCreds(t *testing.T) {
	propGetter := mocks.NewMockPropertyGetter()

	propGetter.On("GetProperty", "PARAMETER_COVERAGE_FILE").Return("../test/jacocoTestReport.xml", true)
	propGetter.On("GetProperty", "PARAMETER_COVERAGE_TYPE").Return("jacoco", true)
	propGetter.On("GetProperty", "PARAMETER_SOURCE_DIRS").Return("src/main/java", true)
	propGetter.On("GetProperty", "PARAMETER_DIFF_SOURCE").Return("github", true)

	err := NewRunner().Run(propGetter.GetProperty, strings.NewReader(""), os.Stdout)
	assert.EqualError(t, err, "PARAMETER_DIFF_SOURCE=github requires a GitHub API key (PARAMETER_GH_API_KEY), BUILD_PULL_REQUEST_NUMBER, REPOSITORY_ORG and REPOSITORY_NAME")

	propGetter.AssertExpectations(t)
}

func TestDefaultRunner_Run_DiffSourceUnknown(t *testing.T) {
	propGetter := mocks.NewMockPropertyGetter()

	propGetter.On("GetProperty", "PARAMETER_COVERAGE_FILE").Return("../test/jacocoTestReport.xml", true)
	propGetter.On("GetProperty", "PARAMETER_COVERAGE_TYPE").Return("jacoco", true)
	propGetter.On("GetProperty", "PARAMETER_SOURCE_DIRS").Return("src/main/java", true)
	propGetter.On("GetProperty", "PARAMETER_DIFF_SOURCE").Return("banana", true)

	err := NewRunner().Run(propGetter.GetProperty, strings.NewReader(""), os.Stdout)
	assert.EqualError(t, err, "Unknown PARAMETER_DIFF_SOURCE \"banana\" (expected \"stdin\" or \"github\")")

	propGetter.AssertExpectations(t)
}

// End-to-end: with PARAMETER_DIFF_SOURCE=github the runner fetches the diff from
// the GitHub API instead of stdin (an empty reader here) and produces the same
// report. The mock server serves the same diff fixture for the PR-diff GET and
// accepts the PR-comment POST, so the output matches the stdin golden exactly.
func TestDefaultRunner_Run_DiffSourceGithub_FetchesDiff(t *testing.T) {
	diff, readErr := os.ReadFile("../test/example_go_unified.diff")
	assert.NoError(t, readErr)

	var diffRequests int

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/repos/some_org/some_repo/pulls/123" {
			diffRequests++
			assert.Equal(t, "application/vnd.github.v3.diff", r.Header.Get("Accept"))
			assert.Equal(t, "token SOME_API_KEY", r.Header.Get("Authorization"))
			w.WriteHeader(200)
			_, _ = w.Write(diff)
			return
		}

		// The PR-comment POST from the github reporter.
		w.WriteHeader(201)
	}))
	defer ts.Close()

	propGetter := mocks.NewMockPropertyGetter()

	propGetter.On("GetProperty", "PARAMETER_COVERAGE_FILE").Return("../test/example_go_coverage_with_source_dir.xml", true)
	propGetter.On("GetProperty", "PARAMETER_COVERAGE_TYPE").Return("cobertura", true)
	propGetter.On("GetProperty", "PARAMETER_SOURCE_DIRS").Return("/go/github.com/target/pull-request-code-coverage", true)
	propGetter.On("GetProperty", "PARAMETER_GH_API_KEY").Return("SOME_API_KEY", true)
	propGetter.On("GetProperty", "PARAMETER_GH_API_BASE_URL").Return(ts.URL, true)
	propGetter.On("GetProperty", "BUILD_PULL_REQUEST_NUMBER").Return("123", true)
	propGetter.On("GetProperty", "REPOSITORY_ORG").Return("some_org", true)
	propGetter.On("GetProperty", "REPOSITORY_NAME").Return("some_repo", true)
	propGetter.On("GetProperty", "PARAMETER_DIFF_SOURCE").Return("github", true)

	var buf bytes.Buffer

	// stdin is intentionally empty — the diff must come from the GitHub API.
	err := NewRunner().Run(propGetter.GetProperty, strings.NewReader(""), &buf)
	assert.NoError(t, err)

	assert.Equal(t, 1, diffRequests, "expected exactly one PR-diff fetch")
	assert.Equal(t, "──────────────────────────────────────────────────────────────\n 📊 Patch Coverage Report  —  changed lines only\n──────────────────────────────────────────────────────────────\n\n Diff coverage: 97% 🟢  —  177 of 182 changed instructions covered\n\n Summary\n   Covered instructions       97%  (177)\n   Missed instructions         3%  (5)\n   Tracked changed lines       8%  (182)\n   Untracked changed lines    92%  (2216)\n\n Note: \"lines\" are the source lines you changed; \"instructions\" are the\n executable units the coverage tool counts inside them (one line can hold\n several, e.g. JaCoCo bytecode), so the two counts differ.\n\n Coverage by file  (lowest coverage first)\n     0%     0 cov /   4 miss   main.go\n    96%    27 cov /   1 miss   internal/plugin/runner.go\n   100%    10 cov /   0 miss   internal/plugin/calculator/calculator.go\n   100%    29 cov /   0 miss   internal/plugin/coverage/jacoco/report.go\n   100%    19 cov /   0 miss   internal/plugin/domain/domain.go\n   100%    25 cov /   0 miss   internal/plugin/reporter/reporter.go\n   100%    64 cov /   0 miss   internal/plugin/sourcelines/unifieddiff/changed_source_loader.go\n   100%     3 cov /   0 miss   internal/test/mocks/property_getter.go\n   (25 file(s) with no measurable lines omitted)\n\n Uncovered lines (5)\n   - internal/plugin/runner.go:72\n         func GetCoverageReportLoader(coverageType string, sourceDir string) coverage.Loader {\n   - main.go:10\n         \terr := plugin.NewRunner().Run(os.LookupEnv, os.Stdin, os.Stdout)\n   - main.go:12\n         \tif err != nil {\n   - main.go:13\n         \t\tlog.WithFields(log.Fields{\n   - main.go:17\n         \t\tos.Exit(1)\n\n──────────────────────────────────────────────────────────────\n", buf.String())

	propGetter.AssertExpectations(t)
}
