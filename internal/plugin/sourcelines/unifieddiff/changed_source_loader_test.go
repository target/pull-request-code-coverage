package unifieddiff

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// unified=0 is what the git-diff/stdin path produces: no context lines, every
// hunk line is an addition. This pins the original behavior.
func TestLoad_Unified0_NoContextLines(t *testing.T) {
	diff := strings.Join([]string{
		"diff --git a/foo.go b/foo.go",
		"--- a/foo.go",
		"+++ b/foo.go",
		"@@ -10,0 +11,2 @@ func foo() {",
		"+\ta := 1",
		"+\tb := 2",
	}, "\n")

	lines, err := NewChangedSourceLinesLoader("", []string{""}).Load(strings.NewReader(diff))

	assert.NoError(t, err)
	assert.Len(t, lines, 2)
	assert.Equal(t, 11, lines[0].LineNumber)
	assert.Equal(t, "\ta := 1", lines[0].LineValue)
	assert.Equal(t, 12, lines[1].LineNumber)
	assert.Equal(t, "\tb := 2", lines[1].LineValue)
}

// unified=3 is what the GitHub API returns: changed lines surrounded by context
// lines. Only the added lines should be recorded, and their line numbers must
// account for the context lines that precede them.
func TestLoad_Unified3_WithContextLines(t *testing.T) {
	diff := strings.Join([]string{
		"diff --git a/foo.go b/foo.go",
		"index 1234567..89abcde 100644",
		"--- a/foo.go",
		"+++ b/foo.go",
		"@@ -8,7 +8,8 @@ func foo() {",
		" \tline8",
		" \tline9",
		" \tline10",
		"-\told",
		"+\tnewA",
		"+\tnewB",
		" \tline12",
		" \tline13",
		" \tline14",
	}, "\n")

	lines, err := NewChangedSourceLinesLoader("", []string{""}).Load(strings.NewReader(diff))

	assert.NoError(t, err)
	assert.Len(t, lines, 2)
	// Hunk starts at new-file line 8; three context lines (8,9,10) precede the
	// additions, so the first added line is 11.
	assert.Equal(t, 11, lines[0].LineNumber)
	assert.Equal(t, "\tnewA", lines[0].LineValue)
	assert.Equal(t, 12, lines[1].LineNumber)
	assert.Equal(t, "\tnewB", lines[1].LineValue)
}

// A blank context line is emitted as a single space; it must still be counted so
// later line numbers stay correct.
func TestLoad_Unified3_BlankContextLine(t *testing.T) {
	diff := strings.Join([]string{
		"+++ b/foo.go",
		"@@ -1,3 +1,4 @@",
		" first",
		" ",
		"+added",
		" last",
	}, "\n")

	lines, err := NewChangedSourceLinesLoader("", []string{""}).Load(strings.NewReader(diff))

	assert.NoError(t, err)
	assert.Len(t, lines, 1)
	assert.Equal(t, 3, lines[0].LineNumber)
	assert.Equal(t, "added", lines[0].LineValue)
}
