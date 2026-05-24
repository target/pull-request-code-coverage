// Package lcov reads LCOV coverage reports (the lcov.info format produced by
// JavaScript/TypeScript tooling such as Jest, nyc, Vitest and c8, all of which
// build on Istanbul). It matches files by their repo-relative path and tolerates
// the absolute SF: paths Istanbul commonly emits by also suffix-matching.
package lcov

import (
	"bufio"
	"io"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/target/pull-request-code-coverage/internal/plugin/coverage"
	"github.com/target/pull-request-code-coverage/internal/plugin/domain"
)

type DefaultLoader struct{}

func NewReportLoader() *DefaultLoader {
	return &DefaultLoader{}
}

func (l *DefaultLoader) Load(coverageFile string) (coverage.Report, error) {
	// nolint: gosec // coverageFile is a user-supplied report path; opening it is the intended behavior
	file, openErr := os.Open(coverageFile)
	if openErr != nil {
		return nil, errors.Wrapf(openErr, "Could not open lcov file %v", coverageFile)
	}

	defer silentlyCall(file.Close)

	report, parseErr := parse(file)
	if parseErr != nil {
		return nil, errors.Wrapf(parseErr, "Failed parsing lcov file %v", coverageFile)
	}

	return report, nil
}

func silentlyCall(c func() error) {
	if err := c(); err != nil {
		log.Panic(err)
	}
}

// Report holds, per source file, the hit count for each line that LCOV tracked.
type Report struct {
	order []string
	files map[string]map[int]int
}

// parse reads an LCOV stream into a Report. Only the SF: (source file) and DA:
// (line execution) records matter for changed-line coverage; everything else
// (functions, branches, summaries) is ignored.
func parse(r io.Reader) (*Report, error) {
	report := &Report{files: map[string]map[int]int{}}

	scanner := bufio.NewScanner(r)
	currentFile := ""

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		switch {
		case strings.HasPrefix(line, "SF:"):
			currentFile = normalizePath(strings.TrimPrefix(line, "SF:"))
			if _, seen := report.files[currentFile]; !seen {
				report.files[currentFile] = map[int]int{}
				report.order = append(report.order, currentFile)
			}

		case strings.HasPrefix(line, "DA:"):
			if currentFile == "" {
				continue
			}

			lineNumber, hits, daErr := parseDA(line)
			if daErr != nil {
				return nil, daErr
			}

			// LCOV can list a line more than once; keep the highest hit count.
			if existing, ok := report.files[currentFile][lineNumber]; !ok || hits > existing {
				report.files[currentFile][lineNumber] = hits
			}

		case line == "end_of_record":
			currentFile = ""
		}
	}

	if scanErr := scanner.Err(); scanErr != nil {
		return nil, errors.Wrap(scanErr, "Failed reading lcov data")
	}

	return report, nil
}

// parseDA parses a "DA:<line>,<hits>[,<checksum>]" record.
func parseDA(line string) (int, int, error) {
	parts := strings.Split(strings.TrimPrefix(line, "DA:"), ",")
	if len(parts) < 2 {
		return 0, 0, errors.Errorf("Invalid DA record %q", line)
	}

	lineNumber, lineErr := strconv.Atoi(strings.TrimSpace(parts[0]))
	if lineErr != nil {
		return 0, 0, errors.Wrapf(lineErr, "Invalid line number in DA record %q", line)
	}

	hits, hitsErr := strconv.Atoi(strings.TrimSpace(parts[1]))
	if hitsErr != nil {
		return 0, 0, errors.Wrapf(hitsErr, "Invalid hit count in DA record %q", line)
	}

	return lineNumber, hits, nil
}

func (r *Report) GetCoverageData(module string, sourceDir string, pkg string, fileName string, lineNumber int) (*domain.CoverageData, bool) {
	candidates := relativePathCandidates(module, sourceDir, pkg, fileName)

	for _, sf := range r.order {
		if !matchesAny(sf, candidates) {
			continue
		}

		hits, tracked := r.files[sf][lineNumber]
		if !tracked {
			continue
		}

		if hits > 0 {
			return &domain.CoverageData{CoveredInstructionCount: 1}, true
		}

		return &domain.CoverageData{MissedInstructionCount: 1}, true
	}

	return nil, false
}

// relativePathCandidates builds the repo-relative paths a changed line could
// appear under as an SF: entry. We try both the full path and the
// source-dir-stripped path, since reports can be written relative to the repo
// root or to a configured source dir.
func relativePathCandidates(module string, sourceDir string, pkg string, fileName string) []string {
	full := joinNonEmpty(module, sourceDir, pkg, fileName)
	withoutSourceDir := joinNonEmpty(module, pkg, fileName)

	if full == withoutSourceDir {
		return []string{full}
	}

	return []string{full, withoutSourceDir}
}

// matchesAny reports whether an SF: path is, or ends with, one of the candidate
// repo-relative paths. The suffix check handles the absolute paths Istanbul
// tooling typically writes (e.g. /home/runner/work/app/app/src/x.ts).
func matchesAny(sf string, candidates []string) bool {
	for _, candidate := range candidates {
		if sf == candidate || strings.HasSuffix(sf, "/"+candidate) {
			return true
		}
	}

	return false
}

func normalizePath(p string) string {
	return strings.TrimPrefix(strings.ReplaceAll(strings.TrimSpace(p), "\\", "/"), "./")
}

func joinNonEmpty(parts ...string) string {
	var nonEmpty []string
	for _, part := range parts {
		if len(part) > 0 {
			nonEmpty = append(nonEmpty, part)
		}
	}

	return strings.Join(nonEmpty, "/")
}
