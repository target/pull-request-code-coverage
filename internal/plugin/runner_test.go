package plugin

import (
	"bytes"
	"git.target.com/search-product-team/pull-request-code-coverage/internal/test/mocks"
	"github.com/stretchr/testify/assert"
	"os"
	"strconv"
	"testing"
)

func TestDefaultRunner_RunNotFoundProps(t *testing.T) {

	tts := []struct {
		foundProps   map[string]string
		missingProps []string
		expectedMsg  string
	}{
		{
			foundProps: map[string]string{
				"PLUGIN_SOURCE_DIRS":   "something",
				"PLUGIN_COVERAGE_TYPE": "something",
			},
			missingProps: []string{
				"PARAMETER_COVERAGE_FILE",
				"PLUGIN_COVERAGE_FILE",
			},
			expectedMsg: "Missing property PARAMETER_COVERAGE_FILE",
		},
		{
			foundProps: map[string]string{
				"PLUGIN_SOURCE_DIRS": "something",
			},
			missingProps: []string{
				"PLUGIN_COVERAGE_TYPE",
				"PARAMETER_COVERAGE_TYPE",
			},
			expectedMsg: "Missing property PARAMETER_COVERAGE_TYPE",
		},
		{
			foundProps: map[string]string{},
			missingProps: []string{
				"PLUGIN_SOURCE_DIRS",
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

		propGetter.On("GetProperty", "PLUGIN_DEBUG").Return("false", true)
		propGetter.On("GetProperty", "PLUGIN_COVERAGE_FILE").Return("../test/example_go_coverage_with_source_dir.xml", true)
		propGetter.On("GetProperty", "PLUGIN_COVERAGE_TYPE").Return("cobertura", true)
		propGetter.On("GetProperty", "PLUGIN_MODULE").Return("", false)
		propGetter.On("GetProperty", "PLUGIN_SOURCE_DIRS").Return("/go/git.target.com/search-product-team/pull-request-code-coverage", true)
		propGetter.On("GetProperty", "PLUGIN_GH_API_KEY").Return("SOME_API_KEY", true)
		propGetter.On("GetProperty", "PLUGIN_GH_API_BASE_URL").Return(mockServerURL, true)
		propGetter.On("GetProperty", "PARAMETER_MODULE").Return("", false)
		propGetter.On("GetProperty", "DRONE_PULL_REQUEST").Return("123", true)
		propGetter.On("GetProperty", "DRONE_REPO_OWNER").Return("some_org", true)
		propGetter.On("GetProperty", "DRONE_REPO_NAME").Return("some_repo", true)

		var buf bytes.Buffer

		err := NewRunner().Run(propGetter.GetProperty, MustOpen(t, "../test/example_go_unified.diff"), &buf)
		assert.NoError(t, err)

		assert.Equal(t, `Missed Instructions:
--- internal/plugin/runner.go:72
func GetCoverageReportLoader(coverageType string, sourceDir string) coverage.Loader {
--- main.go:10
	err := plugin.NewRunner().Run(os.LookupEnv, os.Stdin, os.Stdout)
--- main.go:12
	if err != nil {
--- main.go:13
		log.WithFields(log.Fields{
--- main.go:17
		os.Exit(1)

Code Coverage Summary:
Lines Without Coverage Data -> 92% (2216)
Lines With Coverage Data    -> 8% (182)
Covered Instructions        -> 97% (177)
Missed Instructions         -> 3% (5)
`, buf.String())

		requestAsserter.AssertRequestWasMade(t, "/api/v3/repos/some_org/some_repo/issues/123/comments", "SOME_API_KEY", map[string]interface{}{
			"body": `Code Coverage Summary:

Lines With Coverage Data    -> 8% (182)
Covered Instructions        -> **97%** (177)
`,
		})

		propGetter.AssertExpectations(t)
	})
}

func TestDefaultRunner_Run_GoExample(t *testing.T) {

	mocks.WithMockGithubAPI(func(mockServerURL string, requestAsserter mocks.GithubAPIRequestAsserter) {
		propGetter := mocks.NewMockPropertyGetter()

		propGetter.On("GetProperty", "PLUGIN_DEBUG").Return("false", true)
		propGetter.On("GetProperty", "PLUGIN_COVERAGE_FILE").Return("../test/example_go_coverage.xml", true)
		propGetter.On("GetProperty", "PLUGIN_COVERAGE_TYPE").Return("cobertura", true)
		propGetter.On("GetProperty", "PLUGIN_MODULE").Return("", false)
		propGetter.On("GetProperty", "PLUGIN_SOURCE_DIRS").Return("/go/git.target.com/search-product-team/pull-request-code-coverage", true)
		propGetter.On("GetProperty", "PLUGIN_GH_API_KEY").Return("SOME_API_KEY", true)
		propGetter.On("GetProperty", "PLUGIN_GH_API_BASE_URL").Return(mockServerURL, true)
		propGetter.On("GetProperty", "PARAMETER_MODULE").Return("", false)
		propGetter.On("GetProperty", "DRONE_PULL_REQUEST").Return("123", true)
		propGetter.On("GetProperty", "DRONE_REPO_OWNER").Return("some_org", true)
		propGetter.On("GetProperty", "DRONE_REPO_NAME").Return("some_repo", true)

		var buf bytes.Buffer

		err := NewRunner().Run(propGetter.GetProperty, MustOpen(t, "../test/example_go_unified.diff"), &buf)
		assert.NoError(t, err)

		assert.Equal(t, `Missed Instructions:
--- internal/plugin/runner.go:72
func GetCoverageReportLoader(coverageType string, sourceDir string) coverage.Loader {
--- main.go:10
	err := plugin.NewRunner().Run(os.LookupEnv, os.Stdin, os.Stdout)
--- main.go:12
	if err != nil {
--- main.go:13
		log.WithFields(log.Fields{
--- main.go:17
		os.Exit(1)

Code Coverage Summary:
Lines Without Coverage Data -> 92% (2216)
Lines With Coverage Data    -> 8% (182)
Covered Instructions        -> 97% (177)
Missed Instructions         -> 3% (5)
`, buf.String())

		requestAsserter.AssertRequestWasMade(t, "/api/v3/repos/some_org/some_repo/issues/123/comments", "SOME_API_KEY", map[string]interface{}{
			"body": `Code Coverage Summary:

Lines With Coverage Data    -> 8% (182)
Covered Instructions        -> **97%** (177)
`,
		})

		propGetter.AssertExpectations(t)
	})
}

func TestDefaultRunner_Run(t *testing.T) {

	mocks.WithMockGithubAPI(func(mockServerURL string, requestAsserter mocks.GithubAPIRequestAsserter) {

		propGetter := mocks.NewMockPropertyGetter()

		propGetter.On("GetProperty", "PLUGIN_DEBUG").Return("false", true)
		propGetter.On("GetProperty", "PLUGIN_COVERAGE_FILE").Return("../test/jacocoTestReport.xml", true)
		propGetter.On("GetProperty", "PLUGIN_MODULE").Return("category-search", true)
		propGetter.On("GetProperty", "PLUGIN_COVERAGE_TYPE").Return("jacoco", true)
		propGetter.On("GetProperty", "PLUGIN_SOURCE_DIRS").Return("src/main/java", true)
		propGetter.On("GetProperty", "PLUGIN_GH_API_KEY").Return("SOME_API_KEY", true)
		propGetter.On("GetProperty", "PLUGIN_GH_API_BASE_URL").Return(mockServerURL, true)
		propGetter.On("GetProperty", "DRONE_PULL_REQUEST").Return("123", true)
		propGetter.On("GetProperty", "DRONE_REPO_OWNER").Return("some_org", true)
		propGetter.On("GetProperty", "DRONE_REPO_NAME").Return("some_repo", true)
		var buf bytes.Buffer

		err := NewRunner().Run(propGetter.GetProperty, MustOpen(t, "../test/sample_unified.diff"), &buf)
		assert.NoError(t, err)

		assert.Equal(t, `Missed Instructions:
--- category-search/src/main/java/com/tgt/CategorySearchApplication.java:52
    System.out.print("Soemthing");

Code Coverage Summary:
Lines Without Coverage Data -> 78% (7)
Lines With Coverage Data    -> 22% (2)
Covered Instructions        -> 73% (8)
Missed Instructions         -> 27% (3)
`, buf.String())

		requestAsserter.AssertRequestWasMade(t, "/api/v3/repos/some_org/some_repo/issues/123/comments", "SOME_API_KEY", map[string]interface{}{
			"body": `*Modules: category-search*

Code Coverage Summary:

Lines With Coverage Data    -> 22% (2)
Covered Instructions        -> **73%** (8)
`,
		})

		propGetter.AssertExpectations(t)
	})
}

func TestDefaultRunner_Run_Vela(t *testing.T) {

	mocks.WithMockGithubAPI(func(mockServerURL string, requestAsserter mocks.GithubAPIRequestAsserter) {

		propGetter := mocks.NewMockPropertyGetter()

		propGetter.On("GetProperty", "PLUGIN_DEBUG").Return("", false)
		propGetter.On("GetProperty", "PLUGIN_COVERAGE_FILE").Return("", false)
		propGetter.On("GetProperty", "PLUGIN_MODULE").Return("", false)
		propGetter.On("GetProperty", "PLUGIN_COVERAGE_TYPE").Return("", false)
		propGetter.On("GetProperty", "PLUGIN_SOURCE_DIRS").Return("", false)
		propGetter.On("GetProperty", "PLUGIN_GH_API_KEY").Return("", false)
		propGetter.On("GetProperty", "PLUGIN_GH_API_BASE_URL").Return("", false)
		propGetter.On("GetProperty", "PARAMETER_DEBUG").Return("false", true)
		propGetter.On("GetProperty", "PARAMETER_COVERAGE_FILE").Return("../test/jacocoTestReport.xml", true)
		propGetter.On("GetProperty", "PARAMETER_MODULE").Return("category-search", true)
		propGetter.On("GetProperty", "PARAMETER_COVERAGE_TYPE").Return("jacoco", true)
		propGetter.On("GetProperty", "PARAMETER_SOURCE_DIRS").Return("src/main/java", true)
		propGetter.On("GetProperty", "PARAMETER_GH_API_KEY").Return("SOME_API_KEY", true)
		propGetter.On("GetProperty", "PARAMETER_GH_API_BASE_URL").Return(mockServerURL, true)
		propGetter.On("GetProperty", "DRONE_PULL_REQUEST").Return("", false)
		propGetter.On("GetProperty", "BUILD_PULL_REQUEST_NUMBER").Return("123", true)
		propGetter.On("GetProperty", "DRONE_REPO_OWNER").Return("", false)
		propGetter.On("GetProperty", "REPOSITORY_ORG").Return("some_org", true)
		propGetter.On("GetProperty", "DRONE_REPO_NAME").Return("", false)
		propGetter.On("GetProperty", "REPOSITORY_NAME").Return("some_repo", true)

		var buf bytes.Buffer

		err := NewRunner().Run(propGetter.GetProperty, MustOpen(t, "../test/sample_unified.diff"), &buf)
		assert.NoError(t, err)

		assert.Equal(t, `Missed Instructions:
--- category-search/src/main/java/com/tgt/CategorySearchApplication.java:52
    System.out.print("Soemthing");

Code Coverage Summary:
Lines Without Coverage Data -> 78% (7)
Lines With Coverage Data    -> 22% (2)
Covered Instructions        -> 73% (8)
Missed Instructions         -> 27% (3)
`, buf.String())

		requestAsserter.AssertRequestWasMade(t, "/api/v3/repos/some_org/some_repo/issues/123/comments", "SOME_API_KEY", map[string]interface{}{
			"body": `*Modules: category-search*

Code Coverage Summary:

Lines With Coverage Data    -> 22% (2)
Covered Instructions        -> **73%** (8)
`,
		})

		propGetter.AssertExpectations(t)
	})
}

func TestDefaultRunner_Run_2_Source_Dirs(t *testing.T) {

	mocks.WithMockGithubAPI(func(mockServerURL string, requestAsserter mocks.GithubAPIRequestAsserter) {

		propGetter := mocks.NewMockPropertyGetter()

		propGetter.On("GetProperty", "PLUGIN_DEBUG").Return("false", true)
		propGetter.On("GetProperty", "PLUGIN_COVERAGE_FILE").Return("../test/jacocoTestReport_2_source_dirs.xml", true)
		propGetter.On("GetProperty", "PLUGIN_MODULE").Return("category-search", true)
		propGetter.On("GetProperty", "PLUGIN_COVERAGE_TYPE").Return("jacoco", true)
		propGetter.On("GetProperty", "PLUGIN_SOURCE_DIRS").Return("src/main/java, src/main/kotlin", true)
		propGetter.On("GetProperty", "PLUGIN_GH_API_KEY").Return("SOME_API_KEY", true)
		propGetter.On("GetProperty", "PLUGIN_GH_API_BASE_URL").Return(mockServerURL, true)
		propGetter.On("GetProperty", "DRONE_PULL_REQUEST").Return("123", true)
		propGetter.On("GetProperty", "DRONE_REPO_OWNER").Return("some_org", true)
		propGetter.On("GetProperty", "DRONE_REPO_NAME").Return("some_repo", true)

		var buf bytes.Buffer

		err := NewRunner().Run(propGetter.GetProperty, MustOpen(t, "../test/sample_unified_2_source_dirs.diff"), &buf)
		assert.NoError(t, err)

		assert.Equal(t, `Missed Instructions:
--- category-search/src/main/java/com/tgt/CategorySearchApplication.java:52
    System.out.print("Soemthing");
--- category-search/src/main/kotlin/com/tgt/SomeOtherClass.kt:12
    System.out.print("Soemthing2");

Code Coverage Summary:
Lines Without Coverage Data -> 47% (7)
Lines With Coverage Data    -> 53% (8)
Covered Instructions        -> 88% (42)
Missed Instructions         -> 12% (6)
`, buf.String())

		requestAsserter.AssertRequestWasMade(t, "/api/v3/repos/some_org/some_repo/issues/123/comments", "SOME_API_KEY", map[string]interface{}{
			"body": `*Modules: category-search*

Code Coverage Summary:

Lines With Coverage Data    -> 53% (8)
Covered Instructions        -> **88%** (42)
`,
		})

		propGetter.AssertExpectations(t)
	})
}

func TestDefaultRunner_Run_2_Source_Dirs_Vela(t *testing.T) {

	mocks.WithMockGithubAPI(func(mockServerURL string, requestAsserter mocks.GithubAPIRequestAsserter) {

		propGetter := mocks.NewMockPropertyGetter()

		propGetter.On("GetProperty", "PLUGIN_DEBUG").Return("", false)
		propGetter.On("GetProperty", "PLUGIN_COVERAGE_FILE").Return("", false)
		propGetter.On("GetProperty", "PLUGIN_MODULE").Return("", false)
		propGetter.On("GetProperty", "PLUGIN_COVERAGE_TYPE").Return("", false)
		propGetter.On("GetProperty", "PLUGIN_SOURCE_DIRS").Return("", false)
		propGetter.On("GetProperty", "PLUGIN_GH_API_KEY").Return("", false)
		propGetter.On("GetProperty", "PLUGIN_GH_API_BASE_URL").Return("", false)

		propGetter.On("GetProperty", "PARAMETER_DEBUG").Return("false", true)
		propGetter.On("GetProperty", "PARAMETER_COVERAGE_FILE").Return("../test/jacocoTestReport_2_source_dirs.xml", true)
		propGetter.On("GetProperty", "PARAMETER_MODULE").Return("category-search", true)
		propGetter.On("GetProperty", "PARAMETER_COVERAGE_TYPE").Return("jacoco", true)
		propGetter.On("GetProperty", "PARAMETER_SOURCE_DIRS").Return("src/main/java, src/main/kotlin", true)
		propGetter.On("GetProperty", "PARAMETER_GH_API_KEY").Return("SOME_API_KEY", true)
		propGetter.On("GetProperty", "PARAMETER_GH_API_BASE_URL").Return(mockServerURL, true)

		propGetter.On("GetProperty", "DRONE_PULL_REQUEST").Return("", false)
		propGetter.On("GetProperty", "BUILD_PULL_REQUEST_NUMBER").Return("123", true)
		propGetter.On("GetProperty", "DRONE_REPO_OWNER").Return("", false)
		propGetter.On("GetProperty", "REPOSITORY_ORG").Return("some_org", true)
		propGetter.On("GetProperty", "DRONE_REPO_NAME").Return("", false)
		propGetter.On("GetProperty", "REPOSITORY_NAME").Return("some_repo", true)

		var buf bytes.Buffer

		err := NewRunner().Run(propGetter.GetProperty, MustOpen(t, "../test/sample_unified_2_source_dirs.diff"), &buf)
		assert.NoError(t, err)

		assert.Equal(t, `Missed Instructions:
--- category-search/src/main/java/com/tgt/CategorySearchApplication.java:52
    System.out.print("Soemthing");
--- category-search/src/main/kotlin/com/tgt/SomeOtherClass.kt:12
    System.out.print("Soemthing2");

Code Coverage Summary:
Lines Without Coverage Data -> 47% (7)
Lines With Coverage Data    -> 53% (8)
Covered Instructions        -> 88% (42)
Missed Instructions         -> 12% (6)
`, buf.String())

		requestAsserter.AssertRequestWasMade(t, "/api/v3/repos/some_org/some_repo/issues/123/comments", "SOME_API_KEY", map[string]interface{}{
			"body": `*Modules: category-search*

Code Coverage Summary:

Lines With Coverage Data    -> 53% (8)
Covered Instructions        -> **88%** (42)
`,
		})

		propGetter.AssertExpectations(t)
	})
}

func TestDefaultRunner_Run_NoChanges(t *testing.T) {

	mocks.WithMockGithubAPI(func(mockServerURL string, requestAsserter mocks.GithubAPIRequestAsserter) {

		propGetter := mocks.NewMockPropertyGetter()

		propGetter.On("GetProperty", "PLUGIN_DEBUG").Return("false", true)
		propGetter.On("GetProperty", "PLUGIN_COVERAGE_FILE").Return("../test/jacocoTestReportEmpty.xml", true)
		propGetter.On("GetProperty", "PLUGIN_MODULE").Return("category-search", true)
		propGetter.On("GetProperty", "PLUGIN_COVERAGE_TYPE").Return("jacoco", true)
		propGetter.On("GetProperty", "PLUGIN_SOURCE_DIRS").Return("src/main/java", true)
		propGetter.On("GetProperty", "PLUGIN_GH_API_KEY").Return("SOME_API_KEY", true)
		propGetter.On("GetProperty", "PLUGIN_GH_API_BASE_URL").Return(mockServerURL, true)
		propGetter.On("GetProperty", "DRONE_PULL_REQUEST").Return("123", true)
		propGetter.On("GetProperty", "DRONE_REPO_OWNER").Return("some_org", true)
		propGetter.On("GetProperty", "DRONE_REPO_NAME").Return("some_repo", true)

		var buf bytes.Buffer

		err := NewRunner().Run(propGetter.GetProperty, MustOpen(t, "../test/sample_unified.diff"), &buf)
		assert.NoError(t, err)

		assert.Equal(t, `Missed Instructions:

Code Coverage Summary:
Lines Without Coverage Data -> 100% (9)
Lines With Coverage Data    -> 0% (0)
Covered Instructions        -> 100% (0)
Missed Instructions         -> 0% (0)
`, buf.String())

		requestAsserter.AssertNoRequestsWereMade(t)

		propGetter.AssertExpectations(t)
	})
}

func TestDefaultRunner_Run_NoChanges_Vela(t *testing.T) {

	mocks.WithMockGithubAPI(func(mockServerURL string, requestAsserter mocks.GithubAPIRequestAsserter) {

		propGetter := mocks.NewMockPropertyGetter()

		propGetter.On("GetProperty", "PLUGIN_DEBUG").Return("", false)
		propGetter.On("GetProperty", "PLUGIN_COVERAGE_FILE").Return("", false)
		propGetter.On("GetProperty", "PLUGIN_MODULE").Return("", false)
		propGetter.On("GetProperty", "PLUGIN_COVERAGE_TYPE").Return("", false)
		propGetter.On("GetProperty", "PLUGIN_SOURCE_DIRS").Return("", false)
		propGetter.On("GetProperty", "PLUGIN_GH_API_KEY").Return("", false)
		propGetter.On("GetProperty", "PLUGIN_GH_API_BASE_URL").Return("", false)

		propGetter.On("GetProperty", "PARAMETER_DEBUG").Return("false", true)
		propGetter.On("GetProperty", "PARAMETER_COVERAGE_FILE").Return("../test/jacocoTestReportEmpty.xml", true)
		propGetter.On("GetProperty", "PARAMETER_MODULE").Return("category-search", true)
		propGetter.On("GetProperty", "PARAMETER_COVERAGE_TYPE").Return("jacoco", true)
		propGetter.On("GetProperty", "PARAMETER_SOURCE_DIRS").Return("src/main/java", true)
		propGetter.On("GetProperty", "PARAMETER_GH_API_KEY").Return("SOME_API_KEY", true)
		propGetter.On("GetProperty", "PARAMETER_GH_API_BASE_URL").Return(mockServerURL, true)

		propGetter.On("GetProperty", "DRONE_PULL_REQUEST").Return("", false)
		propGetter.On("GetProperty", "BUILD_PULL_REQUEST_NUMBER").Return("123", true)
		propGetter.On("GetProperty", "DRONE_REPO_OWNER").Return("", false)
		propGetter.On("GetProperty", "REPOSITORY_ORG").Return("some_org", true)
		propGetter.On("GetProperty", "DRONE_REPO_NAME").Return("", false)
		propGetter.On("GetProperty", "REPOSITORY_NAME").Return("some_repo", true)

		var buf bytes.Buffer

		err := NewRunner().Run(propGetter.GetProperty, MustOpen(t, "../test/sample_unified.diff"), &buf)
		assert.NoError(t, err)

		assert.Equal(t, `Missed Instructions:

Code Coverage Summary:
Lines Without Coverage Data -> 100% (9)
Lines With Coverage Data    -> 0% (0)
Covered Instructions        -> 100% (0)
Missed Instructions         -> 0% (0)
`, buf.String())

		requestAsserter.AssertNoRequestsWereMade(t)

		propGetter.AssertExpectations(t)
	})
}

func TestDefaultRunner_RunNoCoverageData(t *testing.T) {
	propGetter := mocks.NewMockPropertyGetter()

	propGetter.On("GetProperty", "PLUGIN_DEBUG").Return("false", true)
	propGetter.On("GetProperty", "PLUGIN_COVERAGE_FILE").Return("../test/jacocoTestEmptyReport.xml", true)
	propGetter.On("GetProperty", "PLUGIN_MODULE").Return("category-search", true)
	propGetter.On("GetProperty", "PLUGIN_COVERAGE_TYPE").Return("jacoco", true)
	propGetter.On("GetProperty", "PLUGIN_SOURCE_DIRS").Return("src/main/java", true)
	propGetter.On("GetProperty", "PLUGIN_GH_API_KEY").Return("", false)
	propGetter.On("GetProperty", "PARAMETER_GH_API_KEY").Return("", false)
	propGetter.On("GetProperty", "PARAMETER_GH_API_BASE_URL").Return("", false)
	propGetter.On("GetProperty", "PLUGIN_GH_API_BASE_URL").Return("", false)
	propGetter.On("GetProperty", "DRONE_PULL_REQUEST").Return("", false)
	propGetter.On("GetProperty", "DRONE_REPO_OWNER").Return("", false)
	propGetter.On("GetProperty", "DRONE_REPO_NAME").Return("", false)
	propGetter.On("GetProperty", "BUILD_PULL_REQUEST_NUMBER").Return("", false)
	propGetter.On("GetProperty", "REPOSITORY_ORG").Return("", false)
	propGetter.On("GetProperty", "REPOSITORY_NAME").Return("", false)

	var buf bytes.Buffer

	err := NewRunner().Run(propGetter.GetProperty, MustOpen(t, "../test/sample_unified.diff"), &buf)
	assert.NoError(t, err)

	assert.Equal(t, `Missed Instructions:

Code Coverage Summary:
Lines Without Coverage Data -> 100% (9)
Lines With Coverage Data    -> 0% (0)
Covered Instructions        -> 100% (0)
Missed Instructions         -> 0% (0)
`, buf.String())

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

	propGetter.On("GetProperty", "PLUGIN_COVERAGE_FILE").Return("../test/blahblah.xml", true)
	propGetter.On("GetProperty", "PLUGIN_MODULE").Return("category-search", true)
	propGetter.On("GetProperty", "PLUGIN_COVERAGE_TYPE").Return("jacoco", true)
	propGetter.On("GetProperty", "PLUGIN_SOURCE_DIRS").Return("anything", true)
	propGetter.On("GetProperty", "PLUGIN_GH_API_KEY").Return("", false)
	propGetter.On("GetProperty", "PLUGIN_GH_API_BASE_URL").Return("", false)
	propGetter.On("GetProperty", "PARAMETER_GH_API_KEY").Return("", false)
	propGetter.On("GetProperty", "PARAMETER_GH_API_BASE_URL").Return("", false)
	propGetter.On("GetProperty", "DRONE_PULL_REQUEST").Return("", false)
	propGetter.On("GetProperty", "DRONE_REPO_OWNER").Return("", false)
	propGetter.On("GetProperty", "DRONE_REPO_NAME").Return("", false)
	propGetter.On("GetProperty", "BUILD_PULL_REQUEST_NUMBER").Return("", false)
	propGetter.On("GetProperty", "REPOSITORY_ORG").Return("", false)
	propGetter.On("GetProperty", "REPOSITORY_NAME").Return("", false)
	err := NewRunner().Run(propGetter.GetProperty, MustOpen(t, "../test/sample_unified.diff"), os.Stdout)
	assert.EqualError(t, err, "Failed loading coverage report: Could not open xml file ../test/blahblah.xml: open ../test/blahblah.xml: no such file or directory")

	propGetter.AssertExpectations(t)
}

func TestDefaultRunner_RunBadUnified_UnfinishedBlock(t *testing.T) {
	propGetter := mocks.NewMockPropertyGetter()

	propGetter.On("GetProperty", "PLUGIN_COVERAGE_FILE").Return("../test/jacocoTestReport.xml", true)
	propGetter.On("GetProperty", "PLUGIN_MODULE").Return("category-search", true)
	propGetter.On("GetProperty", "PLUGIN_COVERAGE_TYPE").Return("jacoco", true)
	propGetter.On("GetProperty", "PLUGIN_SOURCE_DIRS").Return("src/main/java", true)
	propGetter.On("GetProperty", "PLUGIN_GH_API_KEY").Return("", false)
	propGetter.On("GetProperty", "PLUGIN_GH_API_BASE_URL").Return("", false)
	propGetter.On("GetProperty", "PARAMETER_GH_API_KEY").Return("", false)
	propGetter.On("GetProperty", "PARAMETER_GH_API_BASE_URL").Return("", false)
	propGetter.On("GetProperty", "DRONE_PULL_REQUEST").Return("", false)
	propGetter.On("GetProperty", "DRONE_REPO_OWNER").Return("", false)
	propGetter.On("GetProperty", "DRONE_REPO_NAME").Return("", false)
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

	propGetter.On("GetProperty", "PLUGIN_COVERAGE_FILE").Return("../test/jacocoTestReport.xml", true)
	propGetter.On("GetProperty", "PLUGIN_MODULE").Return("category-search", true)
	propGetter.On("GetProperty", "PLUGIN_COVERAGE_TYPE").Return("jacoco", true)
	propGetter.On("GetProperty", "PLUGIN_SOURCE_DIRS").Return("src/main/java", true)
	propGetter.On("GetProperty", "PLUGIN_GH_API_KEY").Return("", false)
	propGetter.On("GetProperty", "PLUGIN_GH_API_BASE_URL").Return("", false)
	propGetter.On("GetProperty", "PARAMETER_GH_API_KEY").Return("", false)
	propGetter.On("GetProperty", "PARAMETER_GH_API_BASE_URL").Return("", false)
	propGetter.On("GetProperty", "DRONE_PULL_REQUEST").Return("", false)
	propGetter.On("GetProperty", "DRONE_REPO_OWNER").Return("", false)
	propGetter.On("GetProperty", "DRONE_REPO_NAME").Return("", false)
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

	propGetter.On("GetProperty", "PLUGIN_COVERAGE_FILE").Return("../test/jacocoTestReport.xml", true)
	propGetter.On("GetProperty", "PLUGIN_MODULE").Return("category-search", true)
	propGetter.On("GetProperty", "PLUGIN_COVERAGE_TYPE").Return("jacoco", true)
	propGetter.On("GetProperty", "PLUGIN_SOURCE_DIRS").Return("src/main/java", true)
	propGetter.On("GetProperty", "PLUGIN_GH_API_KEY").Return("", false)
	propGetter.On("GetProperty", "PLUGIN_GH_API_BASE_URL").Return("", false)
	propGetter.On("GetProperty", "PARAMETER_GH_API_KEY").Return("", false)
	propGetter.On("GetProperty", "PARAMETER_GH_API_BASE_URL").Return("", false)
	propGetter.On("GetProperty", "DRONE_PULL_REQUEST").Return("", false)
	propGetter.On("GetProperty", "DRONE_REPO_OWNER").Return("", false)
	propGetter.On("GetProperty", "DRONE_REPO_NAME").Return("", false)
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

	propGetter.On("GetProperty", "PLUGIN_COVERAGE_FILE").Return("../test/jacocoTestReport.xml", true)
	propGetter.On("GetProperty", "PLUGIN_MODULE").Return("category-search", true)
	propGetter.On("GetProperty", "PLUGIN_COVERAGE_TYPE").Return("jacoco", true)
	propGetter.On("GetProperty", "PLUGIN_SOURCE_DIRS").Return("src/main/java", true)
	propGetter.On("GetProperty", "PLUGIN_GH_API_KEY").Return("", false)
	propGetter.On("GetProperty", "PLUGIN_GH_API_BASE_URL").Return("", false)
	propGetter.On("GetProperty", "PARAMETER_GH_API_KEY").Return("", false)
	propGetter.On("GetProperty", "PARAMETER_GH_API_BASE_URL").Return("", false)
	propGetter.On("GetProperty", "DRONE_PULL_REQUEST").Return("", false)
	propGetter.On("GetProperty", "DRONE_REPO_OWNER").Return("", false)
	propGetter.On("GetProperty", "DRONE_REPO_NAME").Return("", false)
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

	propGetter.On("GetProperty", "PLUGIN_COVERAGE_FILE").Return("../test/jacocoTestReport.xml", true)
	propGetter.On("GetProperty", "PLUGIN_MODULE").Return("category-search", true)
	propGetter.On("GetProperty", "PLUGIN_COVERAGE_TYPE").Return("jacoco", true)
	propGetter.On("GetProperty", "PLUGIN_SOURCE_DIRS").Return("src/main/java", true)
	propGetter.On("GetProperty", "PLUGIN_GH_API_KEY").Return("", false)
	propGetter.On("GetProperty", "PLUGIN_GH_API_BASE_URL").Return("", false)
	propGetter.On("GetProperty", "PARAMETER_GH_API_KEY").Return("", false)
	propGetter.On("GetProperty", "PARAMETER_GH_API_BASE_URL").Return("", false)
	propGetter.On("GetProperty", "DRONE_PULL_REQUEST").Return("", false)
	propGetter.On("GetProperty", "DRONE_REPO_OWNER").Return("", false)
	propGetter.On("GetProperty", "DRONE_REPO_NAME").Return("", false)
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

	propGetter.On("GetProperty", "PLUGIN_COVERAGE_FILE").Return("../test/jacocoTestReport.xml", true)
	propGetter.On("GetProperty", "PLUGIN_MODULE").Return("category-search", true)
	propGetter.On("GetProperty", "PLUGIN_COVERAGE_TYPE").Return("jacoco", true)
	propGetter.On("GetProperty", "PLUGIN_SOURCE_DIRS").Return("src/main/java", true)
	propGetter.On("GetProperty", "PLUGIN_GH_API_KEY").Return("", false)
	propGetter.On("GetProperty", "PLUGIN_GH_API_BASE_URL").Return("", false)
	propGetter.On("GetProperty", "PARAMETER_GH_API_KEY").Return("", false)
	propGetter.On("GetProperty", "PARAMETER_GH_API_BASE_URL").Return("", false)
	propGetter.On("GetProperty", "DRONE_PULL_REQUEST").Return("", false)
	propGetter.On("GetProperty", "DRONE_REPO_OWNER").Return("", false)
	propGetter.On("GetProperty", "DRONE_REPO_NAME").Return("", false)
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

	propGetter.On("GetProperty", "PLUGIN_COVERAGE_FILE").Return("../test/jacocoTestReport.json", true)
	propGetter.On("GetProperty", "PLUGIN_MODULE").Return("category-search", true)
	propGetter.On("GetProperty", "PLUGIN_COVERAGE_TYPE").Return("jacoco", true)
	propGetter.On("GetProperty", "PLUGIN_SOURCE_DIRS").Return("src/main/java", true)
	propGetter.On("GetProperty", "PLUGIN_GH_API_KEY").Return("", false)
	propGetter.On("GetProperty", "PLUGIN_GH_API_BASE_URL").Return("", false)
	propGetter.On("GetProperty", "PARAMETER_GH_API_KEY").Return("", false)
	propGetter.On("GetProperty", "PARAMETER_GH_API_BASE_URL").Return("", false)
	propGetter.On("GetProperty", "DRONE_PULL_REQUEST").Return("", false)
	propGetter.On("GetProperty", "DRONE_REPO_OWNER").Return("", false)
	propGetter.On("GetProperty", "DRONE_REPO_NAME").Return("", false)
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

	propGetter.On("GetProperty", "PLUGIN_COVERAGE_TYPE").Return("cobertura", true)
	propGetter.On("GetProperty", "PLUGIN_SOURCE_DIRS").Return("src/main/java,src/main/kotlin", true)

	var buf bytes.Buffer

	err := NewRunner().Run(propGetter.GetProperty, MustOpen(t, "../test/sample_unified.diff"), &buf)
	assert.EqualError(t, err, "Failed opening coverage loader: Currently not supporting multiple source dirs with cobertura coverage report type.")

	propGetter.AssertExpectations(t)
}
