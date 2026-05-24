package plugin

import (
	"bytes"

	"os"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/target/pull-request-code-coverage/internal/test/mocks"
)

func TestDefaultRunner_RunNotFoundProps(t *testing.T) {

	tts := []struct {
		foundProps   map[string]string
		missingProps []string
		expectedMsg  string
	}{
		{
			foundProps: map[string]string{
				"PARAMETER_SOURCE_DIRS":   "something",
				"PARAMETER_COVERAGE_TYPE": "something",
			},
			missingProps: []string{
				"PARAMETER_COVERAGE_FILE",
			},
			expectedMsg: "Missing property PARAMETER_COVERAGE_FILE",
		},
		{
			foundProps: map[string]string{
				"PARAMETER_SOURCE_DIRS": "something",
			},
			missingProps: []string{
				"PARAMETER_COVERAGE_TYPE",
			},
			expectedMsg: "Missing property PARAMETER_COVERAGE_TYPE",
		},
		{
			foundProps: map[string]string{},
			missingProps: []string{
				"PARAMETER_SOURCE_DIRS",
			},
			expectedMsg: "Missing property PARAMETER_SOURCE_DIRS",
		},
	}

	for idx, tt := range tts {
		t.Run(strconv.Itoa(idx), func(t *testing.T) {
			propGetter := mocks.NewMockPropertyGetter()

			for p, v := range tt.foundProps {
				propGetter.On("GetProperty", p).Return(v, true)
			}

			for _, p := range tt.missingProps {
				propGetter.On("GetProperty", p).Return("fff", false)
			}

			err := NewRunner().Run(propGetter.GetProperty, MustOpen(t, "../test/sample_unified.diff"), os.Stdout)
			assert.EqualError(t, err, tt.expectedMsg)

			propGetter.AssertExpectations(t)
		})
	}
}

func TestDefaultRunner_Run_GoExample_WithSourceDir(t *testing.T) {

	mocks.WithMockGithubAPI(func(mockServerURL string, requestAsserter mocks.GithubAPIRequestAsserter) {
		propGetter := mocks.NewMockPropertyGetter()

		propGetter.On("GetProperty", "PARAMETER_DEBUG").Return("false", true)
		propGetter.On("GetProperty", "PARAMETER_COVERAGE_FILE").Return("../test/example_go_coverage_with_source_dir.xml", true)
		propGetter.On("GetProperty", "PARAMETER_COVERAGE_TYPE").Return("cobertura", true)
		propGetter.On("GetProperty", "PARAMETER_MODULE").Return("", false)
		propGetter.On("GetProperty", "PARAMETER_SOURCE_DIRS").Return("/go/github.com/target/pull-request-code-coverage", true)
		propGetter.On("GetProperty", "PARAMETER_GH_API_KEY").Return("SOME_API_KEY", true)
		propGetter.On("GetProperty", "PARAMETER_GH_API_BASE_URL").Return(mockServerURL, true)
		propGetter.On("GetProperty", "PARAMETER_MODULE").Return("", false)
		propGetter.On("GetProperty", "BUILD_PULL_REQUEST_NUMBER").Return("123", true)
		propGetter.On("GetProperty", "REPOSITORY_ORG").Return("some_org", true)
		propGetter.On("GetProperty", "REPOSITORY_NAME").Return("some_repo", true)

		var buf bytes.Buffer

		err := NewRunner().Run(propGetter.GetProperty, MustOpen(t, "../test/example_go_unified.diff"), &buf)
		assert.NoError(t, err)

		assert.Equal(t, "──────────────────────────────────────────────────────────────\n 📊 Patch Coverage Report  —  changed lines only\n──────────────────────────────────────────────────────────────\n\n Diff coverage: 97% 🟢  —  177 of 182 changed instructions covered\n\n Summary\n   Covered instructions       97%  (177)\n   Missed instructions         3%  (5)\n   Tracked changed lines       8%  (182)\n   Untracked changed lines    92%  (2216)\n\n Note: \"lines\" are the source lines you changed; \"instructions\" are the\n executable units the coverage tool counts inside them (one line can hold\n several, e.g. JaCoCo bytecode), so the two counts differ.\n\n Coverage by file  (lowest coverage first)\n     0%     0 cov /   4 miss   main.go\n    96%    27 cov /   1 miss   internal/plugin/runner.go\n   100%    10 cov /   0 miss   internal/plugin/calculator/calculator.go\n   100%    29 cov /   0 miss   internal/plugin/coverage/jacoco/report.go\n   100%    19 cov /   0 miss   internal/plugin/domain/domain.go\n   100%    25 cov /   0 miss   internal/plugin/reporter/reporter.go\n   100%    64 cov /   0 miss   internal/plugin/sourcelines/unifieddiff/changed_source_loader.go\n   100%     3 cov /   0 miss   internal/test/mocks/property_getter.go\n   (25 file(s) with no measurable lines omitted)\n\n Uncovered lines (5)\n   - internal/plugin/runner.go:72\n         func GetCoverageReportLoader(coverageType string, sourceDir string) coverage.Loader {\n   - main.go:10\n         \terr := plugin.NewRunner().Run(os.LookupEnv, os.Stdin, os.Stdout)\n   - main.go:12\n         \tif err != nil {\n   - main.go:13\n         \t\tlog.WithFields(log.Fields{\n   - main.go:17\n         \t\tos.Exit(1)\n\n──────────────────────────────────────────────────────────────\n", buf.String())

		requestAsserter.AssertRequestWasMade(t, "/repos/some_org/some_repo/issues/123/comments", "SOME_API_KEY", map[string]interface{}{
			"body": "## 🛡️ Patch Coverage Report\n\n> Scope: **changed lines only** — the code this PR adds or edits, not whole files or the repo. It answers one thing — *did your tests run the code you just touched?*\n\n**Diff coverage:** `97%` 🟢 — `177` of `182` changed instructions covered\n\n| Metric | Value | |\n| :-- | --: | :-- |\n| 🟢 Covered instructions | `177` (97%) | changed code your tests executed |\n| 🔴 Missed instructions | `5` (3%) | changed code your tests never ran |\n| 📈 Tracked changed lines | `182` (8%) | lines the coverage tool could measure |\n| ⚪ Untracked changed lines | `2216` (92%) | comments, blanks, declarations |\n\n<sub>**Lines** = the source lines you changed. **Instructions** = the executable units the coverage tool counts inside those lines — one line can hold several (e.g. JaCoCo bytecode), so the two counts differ.</sub>\n\n### Coverage by file\n\n| File | Diff coverage | Covered / Missed |\n| :-- | :-: | :-: |\n| `main.go` | 🔴 0% | 0 / 4 |\n| `internal/plugin/runner.go` | 🟢 96% | 27 / 1 |\n| `internal/plugin/calculator/calculator.go` | 🟢 100% | 10 / 0 |\n| `internal/plugin/coverage/jacoco/report.go` | 🟢 100% | 29 / 0 |\n| `internal/plugin/domain/domain.go` | 🟢 100% | 19 / 0 |\n| `internal/plugin/reporter/reporter.go` | 🟢 100% | 25 / 0 |\n| `internal/plugin/sourcelines/unifieddiff/changed_source_loader.go` | 🟢 100% | 64 / 0 |\n| `internal/test/mocks/property_getter.go` | 🟢 100% | 3 / 0 |\n\n<sub>25 changed file(s) with no measurable lines (config, docs, generated, or test-only) omitted.</sub>\n\n\n<details><summary>🔍 Uncovered lines (5)</summary>\n\n```\n--- internal/plugin/runner.go:72\nfunc GetCoverageReportLoader(coverageType string, sourceDir string) coverage.Loader {\n--- main.go:10\n\terr := plugin.NewRunner().Run(os.LookupEnv, os.Stdin, os.Stdout)\n--- main.go:12\n\tif err != nil {\n--- main.go:13\n\t\tlog.WithFields(log.Fields{\n--- main.go:17\n\t\tos.Exit(1)\n```\n</details>\n\n<sub>🤖 Generated by <a href=\"https://github.com/target/pull-request-code-coverage\">pull-request-code-coverage</a> — coverage for changed lines only.</sub>\n",
		})

		propGetter.AssertExpectations(t)
	})
}

func TestDefaultRunner_Run_GoExample(t *testing.T) {

	mocks.WithMockGithubAPI(func(mockServerURL string, requestAsserter mocks.GithubAPIRequestAsserter) {

		propGetter := mocks.NewMockPropertyGetter()

		propGetter.On("GetProperty", "PARAMETER_DEBUG").Return("false", true)
		propGetter.On("GetProperty", "PARAMETER_COVERAGE_FILE").Return("../test/example_go_coverage.xml", true)
		propGetter.On("GetProperty", "PARAMETER_COVERAGE_TYPE").Return("cobertura", true)
		propGetter.On("GetProperty", "PARAMETER_MODULE").Return("", false)
		propGetter.On("GetProperty", "PARAMETER_SOURCE_DIRS").Return("/go/github.com/target/pull-request-code-coverage", true)
		propGetter.On("GetProperty", "PARAMETER_GH_API_KEY").Return("SOME_API_KEY", true)
		propGetter.On("GetProperty", "PARAMETER_GH_API_BASE_URL").Return(mockServerURL, true)
		propGetter.On("GetProperty", "PARAMETER_MODULE").Return("", false)
		propGetter.On("GetProperty", "BUILD_PULL_REQUEST_NUMBER").Return("123", true)
		propGetter.On("GetProperty", "REPOSITORY_ORG").Return("some_org", true)
		propGetter.On("GetProperty", "REPOSITORY_NAME").Return("some_repo", true)

		var buf bytes.Buffer

		err := NewRunner().Run(propGetter.GetProperty, MustOpen(t, "../test/example_go_unified.diff"), &buf)
		assert.NoError(t, err)

		assert.Equal(t, "──────────────────────────────────────────────────────────────\n 📊 Patch Coverage Report  —  changed lines only\n──────────────────────────────────────────────────────────────\n\n Diff coverage: 97% 🟢  —  177 of 182 changed instructions covered\n\n Summary\n   Covered instructions       97%  (177)\n   Missed instructions         3%  (5)\n   Tracked changed lines       8%  (182)\n   Untracked changed lines    92%  (2216)\n\n Note: \"lines\" are the source lines you changed; \"instructions\" are the\n executable units the coverage tool counts inside them (one line can hold\n several, e.g. JaCoCo bytecode), so the two counts differ.\n\n Coverage by file  (lowest coverage first)\n     0%     0 cov /   4 miss   main.go\n    96%    27 cov /   1 miss   internal/plugin/runner.go\n   100%    10 cov /   0 miss   internal/plugin/calculator/calculator.go\n   100%    29 cov /   0 miss   internal/plugin/coverage/jacoco/report.go\n   100%    19 cov /   0 miss   internal/plugin/domain/domain.go\n   100%    25 cov /   0 miss   internal/plugin/reporter/reporter.go\n   100%    64 cov /   0 miss   internal/plugin/sourcelines/unifieddiff/changed_source_loader.go\n   100%     3 cov /   0 miss   internal/test/mocks/property_getter.go\n   (25 file(s) with no measurable lines omitted)\n\n Uncovered lines (5)\n   - internal/plugin/runner.go:72\n         func GetCoverageReportLoader(coverageType string, sourceDir string) coverage.Loader {\n   - main.go:10\n         \terr := plugin.NewRunner().Run(os.LookupEnv, os.Stdin, os.Stdout)\n   - main.go:12\n         \tif err != nil {\n   - main.go:13\n         \t\tlog.WithFields(log.Fields{\n   - main.go:17\n         \t\tos.Exit(1)\n\n──────────────────────────────────────────────────────────────\n", buf.String())

		requestAsserter.AssertRequestWasMade(t, "/repos/some_org/some_repo/issues/123/comments", "SOME_API_KEY", map[string]interface{}{
			"body": "## 🛡️ Patch Coverage Report\n\n> Scope: **changed lines only** — the code this PR adds or edits, not whole files or the repo. It answers one thing — *did your tests run the code you just touched?*\n\n**Diff coverage:** `97%` 🟢 — `177` of `182` changed instructions covered\n\n| Metric | Value | |\n| :-- | --: | :-- |\n| 🟢 Covered instructions | `177` (97%) | changed code your tests executed |\n| 🔴 Missed instructions | `5` (3%) | changed code your tests never ran |\n| 📈 Tracked changed lines | `182` (8%) | lines the coverage tool could measure |\n| ⚪ Untracked changed lines | `2216` (92%) | comments, blanks, declarations |\n\n<sub>**Lines** = the source lines you changed. **Instructions** = the executable units the coverage tool counts inside those lines — one line can hold several (e.g. JaCoCo bytecode), so the two counts differ.</sub>\n\n### Coverage by file\n\n| File | Diff coverage | Covered / Missed |\n| :-- | :-: | :-: |\n| `main.go` | 🔴 0% | 0 / 4 |\n| `internal/plugin/runner.go` | 🟢 96% | 27 / 1 |\n| `internal/plugin/calculator/calculator.go` | 🟢 100% | 10 / 0 |\n| `internal/plugin/coverage/jacoco/report.go` | 🟢 100% | 29 / 0 |\n| `internal/plugin/domain/domain.go` | 🟢 100% | 19 / 0 |\n| `internal/plugin/reporter/reporter.go` | 🟢 100% | 25 / 0 |\n| `internal/plugin/sourcelines/unifieddiff/changed_source_loader.go` | 🟢 100% | 64 / 0 |\n| `internal/test/mocks/property_getter.go` | 🟢 100% | 3 / 0 |\n\n<sub>25 changed file(s) with no measurable lines (config, docs, generated, or test-only) omitted.</sub>\n\n\n<details><summary>🔍 Uncovered lines (5)</summary>\n\n```\n--- internal/plugin/runner.go:72\nfunc GetCoverageReportLoader(coverageType string, sourceDir string) coverage.Loader {\n--- main.go:10\n\terr := plugin.NewRunner().Run(os.LookupEnv, os.Stdin, os.Stdout)\n--- main.go:12\n\tif err != nil {\n--- main.go:13\n\t\tlog.WithFields(log.Fields{\n--- main.go:17\n\t\tos.Exit(1)\n```\n</details>\n\n<sub>🤖 Generated by <a href=\"https://github.com/target/pull-request-code-coverage\">pull-request-code-coverage</a> — coverage for changed lines only.</sub>\n",
		})

		propGetter.AssertExpectations(t)
	})
}

func TestDefaultRunner_Run_PythonExample(t *testing.T) {

	mocks.WithMockGithubAPI(func(mockServerURL string, requestAsserter mocks.GithubAPIRequestAsserter) {
		propGetter := mocks.NewMockPropertyGetter()

		propGetter.On("GetProperty", "PARAMETER_DEBUG").Return("false", true)
		propGetter.On("GetProperty", "PARAMETER_COVERAGE_FILE").Return("../test/example_python_coverage.xml", true)
		propGetter.On("GetProperty", "PARAMETER_COVERAGE_TYPE").Return("python", true)
		propGetter.On("GetProperty", "PARAMETER_MODULE").Return("", false)
		propGetter.On("GetProperty", "PARAMETER_SOURCE_DIRS").Return(".", true)
		propGetter.On("GetProperty", "PARAMETER_GH_API_KEY").Return("SOME_API_KEY", true)
		propGetter.On("GetProperty", "PARAMETER_GH_API_BASE_URL").Return(mockServerURL, true)
		propGetter.On("GetProperty", "BUILD_PULL_REQUEST_NUMBER").Return("123", true)
		propGetter.On("GetProperty", "REPOSITORY_ORG").Return("some_org", true)
		propGetter.On("GetProperty", "REPOSITORY_NAME").Return("some_repo", true)

		var buf bytes.Buffer

		err := NewRunner().Run(propGetter.GetProperty, MustOpen(t, "../test/example_python_unified.diff"), &buf)
		assert.NoError(t, err)

		assert.Equal(t, "──────────────────────────────────────────────────────────────\n 📊 Patch Coverage Report  —  changed lines only\n──────────────────────────────────────────────────────────────\n\n Diff coverage: 71% 🟡  —  5 of 7 changed instructions covered\n\n Summary\n   Covered instructions       71%  (5)\n   Missed instructions        29%  (2)\n   Tracked changed lines      78%  (7)\n   Untracked changed lines    22%  (2)\n\n Note: \"lines\" are the source lines you changed; \"instructions\" are the\n executable units the coverage tool counts inside them (one line can hold\n several, e.g. JaCoCo bytecode), so the two counts differ.\n\n Coverage by file  (lowest coverage first)\n    71%     5 cov /   2 miss   myapp/calculator.py\n\n Uncovered lines (2)\n   - myapp/calculator.py:6\n             return wrong_name\n   - myapp/calculator.py:9\n             return a / b\n\n──────────────────────────────────────────────────────────────\n", buf.String())

		requestAsserter.AssertRequestWasMade(t, "/repos/some_org/some_repo/issues/123/comments", "SOME_API_KEY", map[string]interface{}{
			"body": "## 🛡️ Patch Coverage Report\n\n> Scope: **changed lines only** — the code this PR adds or edits, not whole files or the repo. It answers one thing — *did your tests run the code you just touched?*\n\n**Diff coverage:** `71%` 🟡 — `5` of `7` changed instructions covered\n\n| Metric | Value | |\n| :-- | --: | :-- |\n| 🟢 Covered instructions | `5` (71%) | changed code your tests executed |\n| 🔴 Missed instructions | `2` (29%) | changed code your tests never ran |\n| 📈 Tracked changed lines | `7` (78%) | lines the coverage tool could measure |\n| ⚪ Untracked changed lines | `2` (22%) | comments, blanks, declarations |\n\n<sub>**Lines** = the source lines you changed. **Instructions** = the executable units the coverage tool counts inside those lines — one line can hold several (e.g. JaCoCo bytecode), so the two counts differ.</sub>\n\n### Coverage by file\n\n| File | Diff coverage | Covered / Missed |\n| :-- | :-: | :-: |\n| `myapp/calculator.py` | 🟡 71% | 5 / 2 |\n\n\n<details><summary>🔍 Uncovered lines (2)</summary>\n\n```\n--- myapp/calculator.py:6\n    return wrong_name\n--- myapp/calculator.py:9\n    return a / b\n```\n</details>\n\n<sub>🤖 Generated by <a href=\"https://github.com/target/pull-request-code-coverage\">pull-request-code-coverage</a> — coverage for changed lines only.</sub>\n",
		})

		propGetter.AssertExpectations(t)
	})
}

func TestDefaultRunner_Run(t *testing.T) {

	mocks.WithMockGithubAPI(func(mockServerURL string, requestAsserter mocks.GithubAPIRequestAsserter) {

		propGetter := mocks.NewMockPropertyGetter()

		propGetter.On("GetProperty", "PARAMETER_DEBUG").Return("false", true)
		propGetter.On("GetProperty", "PARAMETER_COVERAGE_FILE").Return("../test/jacocoTestReport.xml", true)
		propGetter.On("GetProperty", "PARAMETER_MODULE").Return("category-search", true)
		propGetter.On("GetProperty", "PARAMETER_COVERAGE_TYPE").Return("jacoco", true)
		propGetter.On("GetProperty", "PARAMETER_SOURCE_DIRS").Return("src/main/java", true)
		propGetter.On("GetProperty", "PARAMETER_GH_API_KEY").Return("SOME_API_KEY", true)
		propGetter.On("GetProperty", "PARAMETER_GH_API_BASE_URL").Return(mockServerURL, true)
		propGetter.On("GetProperty", "BUILD_PULL_REQUEST_NUMBER").Return("123", true)
		propGetter.On("GetProperty", "REPOSITORY_ORG").Return("some_org", true)
		propGetter.On("GetProperty", "REPOSITORY_NAME").Return("some_repo", true)
		var buf bytes.Buffer

		err := NewRunner().Run(propGetter.GetProperty, MustOpen(t, "../test/sample_unified.diff"), &buf)
		assert.NoError(t, err)

		assert.Equal(t, "──────────────────────────────────────────────────────────────\n 📊 Patch Coverage Report  —  changed lines only\n──────────────────────────────────────────────────────────────\n Modules: category-search\n\n Diff coverage: 73% 🟡  —  8 of 11 changed instructions covered\n\n Summary\n   Covered instructions       73%  (8)\n   Missed instructions        27%  (3)\n   Tracked changed lines      22%  (2)\n   Untracked changed lines    78%  (7)\n\n Note: \"lines\" are the source lines you changed; \"instructions\" are the\n executable units the coverage tool counts inside them (one line can hold\n several, e.g. JaCoCo bytecode), so the two counts differ.\n\n Coverage by file  (lowest coverage first)\n    73%     8 cov /   3 miss   category-search/src/main/java/com/tgt/CategorySearchApplication.java\n   (3 file(s) with no measurable lines omitted)\n\n Uncovered lines (1)\n   - category-search/src/main/java/com/tgt/CategorySearchApplication.java:52\n             System.out.print(\"Something\");\n\n──────────────────────────────────────────────────────────────\n", buf.String())

		requestAsserter.AssertRequestWasMade(t, "/repos/some_org/some_repo/issues/123/comments", "SOME_API_KEY", map[string]interface{}{
			"body": "## 🛡️ Patch Coverage Report\n\n> Scope: **changed lines only** — the code this PR adds or edits, not whole files or the repo. It answers one thing — *did your tests run the code you just touched?*\n\n*Modules:* category-search\n\n**Diff coverage:** `73%` 🟡 — `8` of `11` changed instructions covered\n\n| Metric | Value | |\n| :-- | --: | :-- |\n| 🟢 Covered instructions | `8` (73%) | changed code your tests executed |\n| 🔴 Missed instructions | `3` (27%) | changed code your tests never ran |\n| 📈 Tracked changed lines | `2` (22%) | lines the coverage tool could measure |\n| ⚪ Untracked changed lines | `7` (78%) | comments, blanks, declarations |\n\n<sub>**Lines** = the source lines you changed. **Instructions** = the executable units the coverage tool counts inside those lines — one line can hold several (e.g. JaCoCo bytecode), so the two counts differ.</sub>\n\n### Coverage by file\n\n| File | Diff coverage | Covered / Missed |\n| :-- | :-: | :-: |\n| `category-search/src/main/java/com/tgt/CategorySearchApplication.java` | 🟡 73% | 8 / 3 |\n\n<sub>3 changed file(s) with no measurable lines (config, docs, generated, or test-only) omitted.</sub>\n\n\n<details><summary>🔍 Uncovered lines (1)</summary>\n\n```\n--- category-search/src/main/java/com/tgt/CategorySearchApplication.java:52\n    System.out.print(\"Something\");\n```\n</details>\n\n<sub>🤖 Generated by <a href=\"https://github.com/target/pull-request-code-coverage\">pull-request-code-coverage</a> — coverage for changed lines only.</sub>\n",
		})

		propGetter.AssertExpectations(t)
	})
}

func TestDefaultRunner_Run_Vela(t *testing.T) {

	mocks.WithMockGithubAPI(func(mockServerURL string, requestAsserter mocks.GithubAPIRequestAsserter) {

		propGetter := mocks.NewMockPropertyGetter()

		propGetter.On("GetProperty", "PARAMETER_DEBUG").Return("false", true)
		propGetter.On("GetProperty", "PARAMETER_COVERAGE_FILE").Return("../test/jacocoTestReport.xml", true)
		propGetter.On("GetProperty", "PARAMETER_MODULE").Return("category-search", true)
		propGetter.On("GetProperty", "PARAMETER_COVERAGE_TYPE").Return("jacoco", true)
		propGetter.On("GetProperty", "PARAMETER_SOURCE_DIRS").Return("src/main/java", true)
		propGetter.On("GetProperty", "PARAMETER_GH_API_KEY").Return("SOME_API_KEY", true)
		propGetter.On("GetProperty", "PARAMETER_GH_API_BASE_URL").Return(mockServerURL, true)
		propGetter.On("GetProperty", "BUILD_PULL_REQUEST_NUMBER").Return("123", true)
		propGetter.On("GetProperty", "REPOSITORY_ORG").Return("some_org", true)
		propGetter.On("GetProperty", "REPOSITORY_NAME").Return("some_repo", true)

		var buf bytes.Buffer

		err := NewRunner().Run(propGetter.GetProperty, MustOpen(t, "../test/sample_unified.diff"), &buf)
		assert.NoError(t, err)

		assert.Equal(t, "──────────────────────────────────────────────────────────────\n 📊 Patch Coverage Report  —  changed lines only\n──────────────────────────────────────────────────────────────\n Modules: category-search\n\n Diff coverage: 73% 🟡  —  8 of 11 changed instructions covered\n\n Summary\n   Covered instructions       73%  (8)\n   Missed instructions        27%  (3)\n   Tracked changed lines      22%  (2)\n   Untracked changed lines    78%  (7)\n\n Note: \"lines\" are the source lines you changed; \"instructions\" are the\n executable units the coverage tool counts inside them (one line can hold\n several, e.g. JaCoCo bytecode), so the two counts differ.\n\n Coverage by file  (lowest coverage first)\n    73%     8 cov /   3 miss   category-search/src/main/java/com/tgt/CategorySearchApplication.java\n   (3 file(s) with no measurable lines omitted)\n\n Uncovered lines (1)\n   - category-search/src/main/java/com/tgt/CategorySearchApplication.java:52\n             System.out.print(\"Something\");\n\n──────────────────────────────────────────────────────────────\n", buf.String())

		requestAsserter.AssertRequestWasMade(t, "/repos/some_org/some_repo/issues/123/comments", "SOME_API_KEY", map[string]interface{}{
			"body": "## 🛡️ Patch Coverage Report\n\n> Scope: **changed lines only** — the code this PR adds or edits, not whole files or the repo. It answers one thing — *did your tests run the code you just touched?*\n\n*Modules:* category-search\n\n**Diff coverage:** `73%` 🟡 — `8` of `11` changed instructions covered\n\n| Metric | Value | |\n| :-- | --: | :-- |\n| 🟢 Covered instructions | `8` (73%) | changed code your tests executed |\n| 🔴 Missed instructions | `3` (27%) | changed code your tests never ran |\n| 📈 Tracked changed lines | `2` (22%) | lines the coverage tool could measure |\n| ⚪ Untracked changed lines | `7` (78%) | comments, blanks, declarations |\n\n<sub>**Lines** = the source lines you changed. **Instructions** = the executable units the coverage tool counts inside those lines — one line can hold several (e.g. JaCoCo bytecode), so the two counts differ.</sub>\n\n### Coverage by file\n\n| File | Diff coverage | Covered / Missed |\n| :-- | :-: | :-: |\n| `category-search/src/main/java/com/tgt/CategorySearchApplication.java` | 🟡 73% | 8 / 3 |\n\n<sub>3 changed file(s) with no measurable lines (config, docs, generated, or test-only) omitted.</sub>\n\n\n<details><summary>🔍 Uncovered lines (1)</summary>\n\n```\n--- category-search/src/main/java/com/tgt/CategorySearchApplication.java:52\n    System.out.print(\"Something\");\n```\n</details>\n\n<sub>🤖 Generated by <a href=\"https://github.com/target/pull-request-code-coverage\">pull-request-code-coverage</a> — coverage for changed lines only.</sub>\n",
		})

		propGetter.AssertExpectations(t)
	})
}

func TestDefaultRunner_Run_2_Source_Dirs(t *testing.T) {

	mocks.WithMockGithubAPI(func(mockServerURL string, requestAsserter mocks.GithubAPIRequestAsserter) {

		propGetter := mocks.NewMockPropertyGetter()

		propGetter.On("GetProperty", "PARAMETER_DEBUG").Return("false", true)
		propGetter.On("GetProperty", "PARAMETER_COVERAGE_FILE").Return("../test/jacocoTestReport_2_source_dirs.xml", true)
		propGetter.On("GetProperty", "PARAMETER_MODULE").Return("category-search", true)
		propGetter.On("GetProperty", "PARAMETER_COVERAGE_TYPE").Return("jacoco", true)
		propGetter.On("GetProperty", "PARAMETER_SOURCE_DIRS").Return("src/main/java, src/main/kotlin", true)
		propGetter.On("GetProperty", "PARAMETER_GH_API_KEY").Return("SOME_API_KEY", true)
		propGetter.On("GetProperty", "PARAMETER_GH_API_BASE_URL").Return(mockServerURL, true)
		propGetter.On("GetProperty", "BUILD_PULL_REQUEST_NUMBER").Return("123", true)
		propGetter.On("GetProperty", "REPOSITORY_ORG").Return("some_org", true)
		propGetter.On("GetProperty", "REPOSITORY_NAME").Return("some_repo", true)

		var buf bytes.Buffer

		err := NewRunner().Run(propGetter.GetProperty, MustOpen(t, "../test/sample_unified_2_source_dirs.diff"), &buf)
		assert.NoError(t, err)

		assert.Equal(t, "──────────────────────────────────────────────────────────────\n 📊 Patch Coverage Report  —  changed lines only\n──────────────────────────────────────────────────────────────\n Modules: category-search\n\n Diff coverage: 88% 🟢  —  42 of 48 changed instructions covered\n\n Summary\n   Covered instructions       88%  (42)\n   Missed instructions        12%  (6)\n   Tracked changed lines      53%  (8)\n   Untracked changed lines    47%  (7)\n\n Note: \"lines\" are the source lines you changed; \"instructions\" are the\n executable units the coverage tool counts inside them (one line can hold\n several, e.g. JaCoCo bytecode), so the two counts differ.\n\n Coverage by file  (lowest coverage first)\n    73%     8 cov /   3 miss   category-search/src/main/java/com/tgt/CategorySearchApplication.java\n    92%    34 cov /   3 miss   category-search/src/main/kotlin/com/tgt/SomeOtherClass.kt\n   (3 file(s) with no measurable lines omitted)\n\n Uncovered lines (2)\n   - category-search/src/main/java/com/tgt/CategorySearchApplication.java:52\n             System.out.print(\"Something\");\n   - category-search/src/main/kotlin/com/tgt/SomeOtherClass.kt:12\n             System.out.print(\"Something2\");\n\n──────────────────────────────────────────────────────────────\n", buf.String())

		requestAsserter.AssertRequestWasMade(t, "/repos/some_org/some_repo/issues/123/comments", "SOME_API_KEY", map[string]interface{}{
			"body": "## 🛡️ Patch Coverage Report\n\n> Scope: **changed lines only** — the code this PR adds or edits, not whole files or the repo. It answers one thing — *did your tests run the code you just touched?*\n\n*Modules:* category-search\n\n**Diff coverage:** `88%` 🟢 — `42` of `48` changed instructions covered\n\n| Metric | Value | |\n| :-- | --: | :-- |\n| 🟢 Covered instructions | `42` (88%) | changed code your tests executed |\n| 🔴 Missed instructions | `6` (12%) | changed code your tests never ran |\n| 📈 Tracked changed lines | `8` (53%) | lines the coverage tool could measure |\n| ⚪ Untracked changed lines | `7` (47%) | comments, blanks, declarations |\n\n<sub>**Lines** = the source lines you changed. **Instructions** = the executable units the coverage tool counts inside those lines — one line can hold several (e.g. JaCoCo bytecode), so the two counts differ.</sub>\n\n### Coverage by file\n\n| File | Diff coverage | Covered / Missed |\n| :-- | :-: | :-: |\n| `category-search/src/main/java/com/tgt/CategorySearchApplication.java` | 🟡 73% | 8 / 3 |\n| `category-search/src/main/kotlin/com/tgt/SomeOtherClass.kt` | 🟢 92% | 34 / 3 |\n\n<sub>3 changed file(s) with no measurable lines (config, docs, generated, or test-only) omitted.</sub>\n\n\n<details><summary>🔍 Uncovered lines (2)</summary>\n\n```\n--- category-search/src/main/java/com/tgt/CategorySearchApplication.java:52\n    System.out.print(\"Something\");\n--- category-search/src/main/kotlin/com/tgt/SomeOtherClass.kt:12\n    System.out.print(\"Something2\");\n```\n</details>\n\n<sub>🤖 Generated by <a href=\"https://github.com/target/pull-request-code-coverage\">pull-request-code-coverage</a> — coverage for changed lines only.</sub>\n",
		})

		propGetter.AssertExpectations(t)
	})
}

func TestDefaultRunner_Run_2_Source_Dirs_Vela(t *testing.T) {

	mocks.WithMockGithubAPI(func(mockServerURL string, requestAsserter mocks.GithubAPIRequestAsserter) {

		propGetter := mocks.NewMockPropertyGetter()

		propGetter.On("GetProperty", "PARAMETER_DEBUG").Return("false", true)
		propGetter.On("GetProperty", "PARAMETER_COVERAGE_FILE").Return("../test/jacocoTestReport_2_source_dirs.xml", true)
		propGetter.On("GetProperty", "PARAMETER_MODULE").Return("category-search", true)
		propGetter.On("GetProperty", "PARAMETER_COVERAGE_TYPE").Return("jacoco", true)
		propGetter.On("GetProperty", "PARAMETER_SOURCE_DIRS").Return("src/main/java, src/main/kotlin", true)
		propGetter.On("GetProperty", "PARAMETER_GH_API_KEY").Return("SOME_API_KEY", true)
		propGetter.On("GetProperty", "PARAMETER_GH_API_BASE_URL").Return(mockServerURL, true)

		propGetter.On("GetProperty", "BUILD_PULL_REQUEST_NUMBER").Return("123", true)
		propGetter.On("GetProperty", "REPOSITORY_ORG").Return("some_org", true)
		propGetter.On("GetProperty", "REPOSITORY_NAME").Return("some_repo", true)

		var buf bytes.Buffer

		err := NewRunner().Run(propGetter.GetProperty, MustOpen(t, "../test/sample_unified_2_source_dirs.diff"), &buf)
		assert.NoError(t, err)

		assert.Equal(t, "──────────────────────────────────────────────────────────────\n 📊 Patch Coverage Report  —  changed lines only\n──────────────────────────────────────────────────────────────\n Modules: category-search\n\n Diff coverage: 88% 🟢  —  42 of 48 changed instructions covered\n\n Summary\n   Covered instructions       88%  (42)\n   Missed instructions        12%  (6)\n   Tracked changed lines      53%  (8)\n   Untracked changed lines    47%  (7)\n\n Note: \"lines\" are the source lines you changed; \"instructions\" are the\n executable units the coverage tool counts inside them (one line can hold\n several, e.g. JaCoCo bytecode), so the two counts differ.\n\n Coverage by file  (lowest coverage first)\n    73%     8 cov /   3 miss   category-search/src/main/java/com/tgt/CategorySearchApplication.java\n    92%    34 cov /   3 miss   category-search/src/main/kotlin/com/tgt/SomeOtherClass.kt\n   (3 file(s) with no measurable lines omitted)\n\n Uncovered lines (2)\n   - category-search/src/main/java/com/tgt/CategorySearchApplication.java:52\n             System.out.print(\"Something\");\n   - category-search/src/main/kotlin/com/tgt/SomeOtherClass.kt:12\n             System.out.print(\"Something2\");\n\n──────────────────────────────────────────────────────────────\n", buf.String())

		requestAsserter.AssertRequestWasMade(t, "/repos/some_org/some_repo/issues/123/comments", "SOME_API_KEY", map[string]interface{}{
			"body": "## 🛡️ Patch Coverage Report\n\n> Scope: **changed lines only** — the code this PR adds or edits, not whole files or the repo. It answers one thing — *did your tests run the code you just touched?*\n\n*Modules:* category-search\n\n**Diff coverage:** `88%` 🟢 — `42` of `48` changed instructions covered\n\n| Metric | Value | |\n| :-- | --: | :-- |\n| 🟢 Covered instructions | `42` (88%) | changed code your tests executed |\n| 🔴 Missed instructions | `6` (12%) | changed code your tests never ran |\n| 📈 Tracked changed lines | `8` (53%) | lines the coverage tool could measure |\n| ⚪ Untracked changed lines | `7` (47%) | comments, blanks, declarations |\n\n<sub>**Lines** = the source lines you changed. **Instructions** = the executable units the coverage tool counts inside those lines — one line can hold several (e.g. JaCoCo bytecode), so the two counts differ.</sub>\n\n### Coverage by file\n\n| File | Diff coverage | Covered / Missed |\n| :-- | :-: | :-: |\n| `category-search/src/main/java/com/tgt/CategorySearchApplication.java` | 🟡 73% | 8 / 3 |\n| `category-search/src/main/kotlin/com/tgt/SomeOtherClass.kt` | 🟢 92% | 34 / 3 |\n\n<sub>3 changed file(s) with no measurable lines (config, docs, generated, or test-only) omitted.</sub>\n\n\n<details><summary>🔍 Uncovered lines (2)</summary>\n\n```\n--- category-search/src/main/java/com/tgt/CategorySearchApplication.java:52\n    System.out.print(\"Something\");\n--- category-search/src/main/kotlin/com/tgt/SomeOtherClass.kt:12\n    System.out.print(\"Something2\");\n```\n</details>\n\n<sub>🤖 Generated by <a href=\"https://github.com/target/pull-request-code-coverage\">pull-request-code-coverage</a> — coverage for changed lines only.</sub>\n",
		})

		propGetter.AssertExpectations(t)
	})
}

func TestDefaultRunner_Run_NoChanges(t *testing.T) {

	mocks.WithMockGithubAPI(func(mockServerURL string, requestAsserter mocks.GithubAPIRequestAsserter) {

		propGetter := mocks.NewMockPropertyGetter()

		propGetter.On("GetProperty", "PARAMETER_DEBUG").Return("false", true)
		propGetter.On("GetProperty", "PARAMETER_COVERAGE_FILE").Return("../test/jacocoTestReportEmpty.xml", true)
		propGetter.On("GetProperty", "PARAMETER_MODULE").Return("category-search", true)
		propGetter.On("GetProperty", "PARAMETER_COVERAGE_TYPE").Return("jacoco", true)
		propGetter.On("GetProperty", "PARAMETER_SOURCE_DIRS").Return("src/main/java", true)
		propGetter.On("GetProperty", "PARAMETER_GH_API_KEY").Return("SOME_API_KEY", true)
		propGetter.On("GetProperty", "PARAMETER_GH_API_BASE_URL").Return(mockServerURL, true)
		propGetter.On("GetProperty", "BUILD_PULL_REQUEST_NUMBER").Return("123", true)
		propGetter.On("GetProperty", "REPOSITORY_ORG").Return("some_org", true)
		propGetter.On("GetProperty", "REPOSITORY_NAME").Return("some_repo", true)

		var buf bytes.Buffer

		err := NewRunner().Run(propGetter.GetProperty, MustOpen(t, "../test/sample_unified.diff"), &buf)
		assert.NoError(t, err)

		assert.Equal(t, "──────────────────────────────────────────────────────────────\n 📊 Patch Coverage Report  —  changed lines only\n──────────────────────────────────────────────────────────────\n Modules: category-search\n\n Diff coverage: 100% 🟢  —  0 of 0 changed instructions covered\n\n Summary\n   Covered instructions      100%  (0)\n   Missed instructions         0%  (0)\n   Tracked changed lines       0%  (0)\n   Untracked changed lines   100%  (9)\n\n Note: \"lines\" are the source lines you changed; \"instructions\" are the\n executable units the coverage tool counts inside them (one line can hold\n several, e.g. JaCoCo bytecode), so the two counts differ.\n\n Coverage by file  (lowest coverage first)\n   (no files with measurable lines)\n   (4 file(s) with no measurable lines omitted)\n\n Uncovered lines (0)\n   none 🎉\n\n──────────────────────────────────────────────────────────────\n", buf.String())

		requestAsserter.AssertNoRequestsWereMade(t)

		propGetter.AssertExpectations(t)
	})
}

func TestDefaultRunner_Run_NoChanges_Vela(t *testing.T) {

	mocks.WithMockGithubAPI(func(mockServerURL string, requestAsserter mocks.GithubAPIRequestAsserter) {

		propGetter := mocks.NewMockPropertyGetter()

		propGetter.On("GetProperty", "PARAMETER_DEBUG").Return("false", true)
		propGetter.On("GetProperty", "PARAMETER_COVERAGE_FILE").Return("../test/jacocoTestReportEmpty.xml", true)
		propGetter.On("GetProperty", "PARAMETER_MODULE").Return("category-search", true)
		propGetter.On("GetProperty", "PARAMETER_COVERAGE_TYPE").Return("jacoco", true)
		propGetter.On("GetProperty", "PARAMETER_SOURCE_DIRS").Return("src/main/java", true)
		propGetter.On("GetProperty", "PARAMETER_GH_API_KEY").Return("SOME_API_KEY", true)
		propGetter.On("GetProperty", "PARAMETER_GH_API_BASE_URL").Return(mockServerURL, true)

		propGetter.On("GetProperty", "BUILD_PULL_REQUEST_NUMBER").Return("123", true)
		propGetter.On("GetProperty", "REPOSITORY_ORG").Return("some_org", true)
		propGetter.On("GetProperty", "REPOSITORY_NAME").Return("some_repo", true)

		var buf bytes.Buffer

		err := NewRunner().Run(propGetter.GetProperty, MustOpen(t, "../test/sample_unified.diff"), &buf)
		assert.NoError(t, err)

		assert.Equal(t, "──────────────────────────────────────────────────────────────\n 📊 Patch Coverage Report  —  changed lines only\n──────────────────────────────────────────────────────────────\n Modules: category-search\n\n Diff coverage: 100% 🟢  —  0 of 0 changed instructions covered\n\n Summary\n   Covered instructions      100%  (0)\n   Missed instructions         0%  (0)\n   Tracked changed lines       0%  (0)\n   Untracked changed lines   100%  (9)\n\n Note: \"lines\" are the source lines you changed; \"instructions\" are the\n executable units the coverage tool counts inside them (one line can hold\n several, e.g. JaCoCo bytecode), so the two counts differ.\n\n Coverage by file  (lowest coverage first)\n   (no files with measurable lines)\n   (4 file(s) with no measurable lines omitted)\n\n Uncovered lines (0)\n   none 🎉\n\n──────────────────────────────────────────────────────────────\n", buf.String())

		requestAsserter.AssertNoRequestsWereMade(t)

		propGetter.AssertExpectations(t)
	})
}

func TestDefaultRunner_RunNoCoverageData(t *testing.T) {
	propGetter := mocks.NewMockPropertyGetter()

	propGetter.On("GetProperty", "PARAMETER_DEBUG").Return("false", true)
	propGetter.On("GetProperty", "PARAMETER_COVERAGE_FILE").Return("../test/jacocoTestEmptyReport.xml", true)
	propGetter.On("GetProperty", "PARAMETER_MODULE").Return("category-search", true)
	propGetter.On("GetProperty", "PARAMETER_COVERAGE_TYPE").Return("jacoco", true)
	propGetter.On("GetProperty", "PARAMETER_SOURCE_DIRS").Return("src/main/java", true)
	propGetter.On("GetProperty", "PARAMETER_GH_API_KEY").Return("", false)
	propGetter.On("GetProperty", "PLUGIN_GH_API_KEY").Return("", false)
	propGetter.On("GetProperty", "PARAMETER_GH_API_BASE_URL").Return("", false)
	propGetter.On("GetProperty", "BUILD_PULL_REQUEST_NUMBER").Return("", false)
	propGetter.On("GetProperty", "REPOSITORY_ORG").Return("", false)
	propGetter.On("GetProperty", "REPOSITORY_NAME").Return("", false)

	var buf bytes.Buffer

	err := NewRunner().Run(propGetter.GetProperty, MustOpen(t, "../test/sample_unified.diff"), &buf)
	assert.NoError(t, err)

	assert.Equal(t, "──────────────────────────────────────────────────────────────\n 📊 Patch Coverage Report  —  changed lines only\n──────────────────────────────────────────────────────────────\n Modules: category-search\n\n Diff coverage: 100% 🟢  —  0 of 0 changed instructions covered\n\n Summary\n   Covered instructions      100%  (0)\n   Missed instructions         0%  (0)\n   Tracked changed lines       0%  (0)\n   Untracked changed lines   100%  (9)\n\n Note: \"lines\" are the source lines you changed; \"instructions\" are the\n executable units the coverage tool counts inside them (one line can hold\n several, e.g. JaCoCo bytecode), so the two counts differ.\n\n Coverage by file  (lowest coverage first)\n   (no files with measurable lines)\n   (4 file(s) with no measurable lines omitted)\n\n Uncovered lines (0)\n   none 🎉\n\n──────────────────────────────────────────────────────────────\n", buf.String())

	propGetter.AssertExpectations(t)
}

func MustOpen(t *testing.T, filename string) *os.File {
	f, err := os.Open(filename)

	if !assert.NoError(t, err) {
		t.FailNow()
	}

	return f
}

func TestDefaultRunner_RunErrOpeningCoverFile(t *testing.T) {
	propGetter := mocks.NewMockPropertyGetter()

	propGetter.On("GetProperty", "PARAMETER_COVERAGE_FILE").Return("../test/blahblah.xml", true)
	propGetter.On("GetProperty", "PARAMETER_MODULE").Return("category-search", true)
	propGetter.On("GetProperty", "PARAMETER_COVERAGE_TYPE").Return("jacoco", true)
	propGetter.On("GetProperty", "PARAMETER_SOURCE_DIRS").Return("anything", true)
	propGetter.On("GetProperty", "PARAMETER_GH_API_KEY").Return("", false)
	propGetter.On("GetProperty", "PLUGIN_GH_API_KEY").Return("", false)
	propGetter.On("GetProperty", "PARAMETER_GH_API_BASE_URL").Return("", false)
	propGetter.On("GetProperty", "BUILD_PULL_REQUEST_NUMBER").Return("", false)
	propGetter.On("GetProperty", "REPOSITORY_ORG").Return("", false)
	propGetter.On("GetProperty", "REPOSITORY_NAME").Return("", false)
	err := NewRunner().Run(propGetter.GetProperty, MustOpen(t, "../test/sample_unified.diff"), os.Stdout)
	assert.EqualError(t, err, "Failed loading coverage report: Could not open xml file ../test/blahblah.xml: open ../test/blahblah.xml: no such file or directory")

	propGetter.AssertExpectations(t)
}

func TestDefaultRunner_RunBadUnified_UnfinishedBlock(t *testing.T) {
	propGetter := mocks.NewMockPropertyGetter()

	propGetter.On("GetProperty", "PARAMETER_COVERAGE_FILE").Return("../test/jacocoTestReport.xml", true)
	propGetter.On("GetProperty", "PARAMETER_MODULE").Return("category-search", true)
	propGetter.On("GetProperty", "PARAMETER_COVERAGE_TYPE").Return("jacoco", true)
	propGetter.On("GetProperty", "PARAMETER_SOURCE_DIRS").Return("src/main/java", true)
	propGetter.On("GetProperty", "PARAMETER_GH_API_KEY").Return("", false)
	propGetter.On("GetProperty", "PARAMETER_GH_API_BASE_URL").Return("", false)
	propGetter.On("GetProperty", "PLUGIN_GH_API_KEY").Return("", false)
	propGetter.On("GetProperty", "BUILD_PULL_REQUEST_NUMBER").Return("", false)
	propGetter.On("GetProperty", "REPOSITORY_ORG").Return("", false)
	propGetter.On("GetProperty", "REPOSITORY_NAME").Return("", false)
	var buf bytes.Buffer

	err := NewRunner().Run(propGetter.GetProperty, MustOpen(t, "../test/unfinished_block_sample_unified.diff"), &buf)
	assert.EqualError(t, err, "Failed loading changed lines: Was not able to finish previous block category-search  test.something 1 1 2 +++ /dev/null")

	propGetter.AssertExpectations(t)
}

func TestDefaultRunner_RunBadUnified_UnfinishedBlock2(t *testing.T) {
	propGetter := mocks.NewMockPropertyGetter()

	propGetter.On("GetProperty", "PARAMETER_COVERAGE_FILE").Return("../test/jacocoTestReport.xml", true)
	propGetter.On("GetProperty", "PARAMETER_MODULE").Return("category-search", true)
	propGetter.On("GetProperty", "PARAMETER_COVERAGE_TYPE").Return("jacoco", true)
	propGetter.On("GetProperty", "PARAMETER_SOURCE_DIRS").Return("src/main/java", true)
	propGetter.On("GetProperty", "PARAMETER_GH_API_KEY").Return("", false)
	propGetter.On("GetProperty", "PARAMETER_GH_API_BASE_URL").Return("", false)
	propGetter.On("GetProperty", "PLUGIN_GH_API_KEY").Return("", false)
	propGetter.On("GetProperty", "BUILD_PULL_REQUEST_NUMBER").Return("", false)
	propGetter.On("GetProperty", "REPOSITORY_ORG").Return("", false)
	propGetter.On("GetProperty", "REPOSITORY_NAME").Return("", false)
	var buf bytes.Buffer

	err := NewRunner().Run(propGetter.GetProperty, MustOpen(t, "../test/unfinished_block_2_sample_unified.diff"), &buf)
	assert.EqualError(t, err, "Failed loading changed lines: Was not able to finish previous block category-search  test.something 1 1 2 @@ -1,3 +0,0 @@")

	propGetter.AssertExpectations(t)
}

func TestDefaultRunner_RunBadUnified_BadFilename(t *testing.T) {
	propGetter := mocks.NewMockPropertyGetter()

	propGetter.On("GetProperty", "PARAMETER_COVERAGE_FILE").Return("../test/jacocoTestReport.xml", true)
	propGetter.On("GetProperty", "PARAMETER_MODULE").Return("category-search", true)
	propGetter.On("GetProperty", "PARAMETER_COVERAGE_TYPE").Return("jacoco", true)
	propGetter.On("GetProperty", "PARAMETER_SOURCE_DIRS").Return("src/main/java", true)
	propGetter.On("GetProperty", "PARAMETER_GH_API_KEY").Return("", false)
	propGetter.On("GetProperty", "PARAMETER_GH_API_BASE_URL").Return("", false)
	propGetter.On("GetProperty", "PLUGIN_GH_API_KEY").Return("", false)
	propGetter.On("GetProperty", "BUILD_PULL_REQUEST_NUMBER").Return("", false)
	propGetter.On("GetProperty", "REPOSITORY_ORG").Return("", false)
	propGetter.On("GetProperty", "REPOSITORY_NAME").Return("", false)
	var buf bytes.Buffer

	err := NewRunner().Run(propGetter.GetProperty, MustOpen(t, "../test/badfilename_sample_unified.diff"), &buf)
	assert.EqualError(t, err, "Failed loading changed lines: Filename test.something is invalid with expected module category-search")

	propGetter.AssertExpectations(t)
}

func TestDefaultRunner_RunBadUnified_ExtraLinesBlock(t *testing.T) {
	propGetter := mocks.NewMockPropertyGetter()

	propGetter.On("GetProperty", "PARAMETER_COVERAGE_FILE").Return("../test/jacocoTestReport.xml", true)
	propGetter.On("GetProperty", "PARAMETER_MODULE").Return("category-search", true)
	propGetter.On("GetProperty", "PARAMETER_COVERAGE_TYPE").Return("jacoco", true)
	propGetter.On("GetProperty", "PARAMETER_SOURCE_DIRS").Return("src/main/java", true)
	propGetter.On("GetProperty", "PARAMETER_GH_API_KEY").Return("", false)
	propGetter.On("GetProperty", "PARAMETER_GH_API_BASE_URL").Return("", false)
	propGetter.On("GetProperty", "PLUGIN_GH_API_KEY").Return("", false)
	propGetter.On("GetProperty", "BUILD_PULL_REQUEST_NUMBER").Return("", false)
	propGetter.On("GetProperty", "REPOSITORY_ORG").Return("", false)
	propGetter.On("GetProperty", "REPOSITORY_NAME").Return("", false)
	var buf bytes.Buffer

	err := NewRunner().Run(propGetter.GetProperty, MustOpen(t, "../test/extralines_block_sample_unified.diff"), &buf)
	assert.EqualError(t, err, "Failed loading changed lines: Finished previous block early category-search  test.something 1 3 0 +extra")

	propGetter.AssertExpectations(t)
}

func TestDefaultRunner_RunBadUnified_AlphaInAtBlock(t *testing.T) {
	propGetter := mocks.NewMockPropertyGetter()

	propGetter.On("GetProperty", "PARAMETER_COVERAGE_FILE").Return("../test/jacocoTestReport.xml", true)
	propGetter.On("GetProperty", "PARAMETER_MODULE").Return("category-search", true)
	propGetter.On("GetProperty", "PARAMETER_COVERAGE_TYPE").Return("jacoco", true)
	propGetter.On("GetProperty", "PARAMETER_SOURCE_DIRS").Return("src/main/java", true)
	propGetter.On("GetProperty", "PARAMETER_GH_API_KEY").Return("", false)
	propGetter.On("GetProperty", "PARAMETER_GH_API_BASE_URL").Return("", false)
	propGetter.On("GetProperty", "PLUGIN_GH_API_KEY").Return("", false)
	propGetter.On("GetProperty", "BUILD_PULL_REQUEST_NUMBER").Return("", false)
	propGetter.On("GetProperty", "REPOSITORY_ORG").Return("", false)
	propGetter.On("GetProperty", "REPOSITORY_NAME").Return("", false)

	var buf bytes.Buffer

	err := NewRunner().Run(propGetter.GetProperty, MustOpen(t, "../test/bad_at_nums_sample_unified.diff"), &buf)
	assert.EqualError(t, err, "Failed loading changed lines: Invalid line offset in line @@ -0,0 +a,3 @@: strconv.Atoi: parsing \"a\": invalid syntax")

	propGetter.AssertExpectations(t)
}

func TestDefaultRunner_RunBadUnified_AlphaInAtBlock2(t *testing.T) {
	propGetter := mocks.NewMockPropertyGetter()

	propGetter.On("GetProperty", "PARAMETER_COVERAGE_FILE").Return("../test/jacocoTestReport.xml", true)
	propGetter.On("GetProperty", "PARAMETER_MODULE").Return("category-search", true)
	propGetter.On("GetProperty", "PARAMETER_COVERAGE_TYPE").Return("jacoco", true)
	propGetter.On("GetProperty", "PARAMETER_SOURCE_DIRS").Return("src/main/java", true)
	propGetter.On("GetProperty", "PARAMETER_GH_API_KEY").Return("", false)
	propGetter.On("GetProperty", "PARAMETER_GH_API_BASE_URL").Return("", false)
	propGetter.On("GetProperty", "PLUGIN_GH_API_KEY").Return("", false)
	propGetter.On("GetProperty", "BUILD_PULL_REQUEST_NUMBER").Return("", false)
	propGetter.On("GetProperty", "REPOSITORY_ORG").Return("", false)
	propGetter.On("GetProperty", "REPOSITORY_NAME").Return("", false)

	var buf bytes.Buffer

	err := NewRunner().Run(propGetter.GetProperty, MustOpen(t, "../test/bad_at_nums_2_sample_unified.diff"), &buf)
	assert.EqualError(t, err, "Failed loading changed lines: Invalid line offset in line @@ -0,0 +1,a @@: strconv.Atoi: parsing \"a\": invalid syntax")

	propGetter.AssertExpectations(t)
}

func TestDefaultRunner_RunCoverageNotXml(t *testing.T) {
	propGetter := mocks.NewMockPropertyGetter()

	propGetter.On("GetProperty", "PARAMETER_COVERAGE_FILE").Return("../test/jacocoTestReport.json", true)
	propGetter.On("GetProperty", "PARAMETER_MODULE").Return("category-search", true)
	propGetter.On("GetProperty", "PARAMETER_COVERAGE_TYPE").Return("jacoco", true)
	propGetter.On("GetProperty", "PARAMETER_SOURCE_DIRS").Return("src/main/java", true)
	propGetter.On("GetProperty", "PARAMETER_GH_API_KEY").Return("", false)
	propGetter.On("GetProperty", "PARAMETER_GH_API_BASE_URL").Return("", false)
	propGetter.On("GetProperty", "PLUGIN_GH_API_KEY").Return("", false)
	propGetter.On("GetProperty", "BUILD_PULL_REQUEST_NUMBER").Return("", false)
	propGetter.On("GetProperty", "REPOSITORY_ORG").Return("", false)
	propGetter.On("GetProperty", "REPOSITORY_NAME").Return("", false)

	var buf bytes.Buffer

	err := NewRunner().Run(propGetter.GetProperty, MustOpen(t, "../test/sample_unified.diff"), &buf)
	assert.EqualError(t, err, "Failed loading coverage report: Failed unmarshalling coverage file ../test/jacocoTestReport.json: EOF")

	propGetter.AssertExpectations(t)
}

func TestDefaultRunner_Run_2SourceDirsCobertura(t *testing.T) {
	propGetter := mocks.NewMockPropertyGetter()

	propGetter.On("GetProperty", "PARAMETER_COVERAGE_TYPE").Return("cobertura", true)
	propGetter.On("GetProperty", "PARAMETER_SOURCE_DIRS").Return("src/main/java,src/main/kotlin", true)

	var buf bytes.Buffer

	err := NewRunner().Run(propGetter.GetProperty, MustOpen(t, "../test/sample_unified.diff"), &buf)
	assert.EqualError(t, err, "Failed opening coverage loader: Currently not supporting multiple source dirs with cobertura coverage report type.")

	propGetter.AssertExpectations(t)
}
