// Package pythoncov reads coverage.py XML reports (the format produced by
// `coverage xml` and pytest-cov's --cov-report=xml). That XML follows the
// Cobertura schema, but unlike the cobertura loader this one matches purely on
// the repo-relative file path and ignores the absolute <source> root, so a
// typical pytest project run from the repo root works without pointing
// PARAMETER_SOURCE_DIRS at an absolute build path.
package pythoncov

import (
	"encoding/xml"
	"io"
	"log"
	"os"
	"strings"

	"github.com/pkg/errors"
	"github.com/target/pull-request-code-coverage/internal/plugin/coverage"
	"github.com/target/pull-request-code-coverage/internal/plugin/domain"
)

type DefaultLoader struct {
	readAllFunc func(io.Reader) ([]byte, error)
}

func NewReportLoader() *DefaultLoader {
	return &DefaultLoader{
		readAllFunc: io.ReadAll,
	}
}

func (l *DefaultLoader) Load(coverageFile string) (coverage.Report, error) {
	// nolint: gosec // coverageFile is a user-supplied report path; opening it is the intended behavior
	xmlFile, openFileErr := os.Open(coverageFile)
	if openFileErr != nil {
		return nil, errors.Wrapf(openFileErr, "Could not open xml file %v", coverageFile)
	}

	defer silentlyCall(xmlFile.Close)

	byteValue, readAllErr := l.readAllFunc(xmlFile)
	if readAllErr != nil {
		return nil, errors.Wrapf(readAllErr, "Failed reading in all of coverage file %v", coverageFile)
	}

	var coverageReport Report
	if err := xml.Unmarshal(byteValue, &coverageReport); err != nil {
		return nil, errors.Wrapf(err, "Failed unmarshalling coverage file %v", coverageFile)
	}

	return &coverageReport, nil
}

func silentlyCall(c func() error) {
	if err := c(); err != nil {
		log.Panic(err)
	}
}

type Report struct {
	XMLName xml.Name `xml:"coverage"`

	Packages []Package `xml:"packages>package"`
}

func (r *Report) GetCoverageData(module string, sourceDir string, pkg string, fileName string, lineNumber int) (*domain.CoverageData, bool) {
	candidates := relativePathCandidates(module, sourceDir, pkg, fileName)

	for _, p := range r.Packages {
		for _, c := range p.Classes {
			if !matchesAny(c.Filename, candidates) {
				continue
			}

			for _, l := range c.Lines {
				if l.Number == lineNumber {
					if l.Hits > 0 {
						return &domain.CoverageData{CoveredInstructionCount: 1}, true
					}

					return &domain.CoverageData{MissedInstructionCount: 1}, true
				}
			}
		}
	}

	return nil, false
}

// relativePathCandidates builds the repo-relative paths a changed line could
// appear under as a <class filename=...>. coverage.py writes filenames relative
// to its <source> root, which is commonly either the repo root (so the source
// dir is part of the path) or the configured source dir (so it is stripped) —
// we try both forms.
func relativePathCandidates(module string, sourceDir string, pkg string, fileName string) []string {
	full := joinNonEmpty(module, sourceDir, pkg, fileName)
	withoutSourceDir := joinNonEmpty(module, pkg, fileName)

	if full == withoutSourceDir {
		return []string{full}
	}

	return []string{full, withoutSourceDir}
}

func matchesAny(filename string, candidates []string) bool {
	for _, candidate := range candidates {
		if filename == candidate {
			return true
		}
	}

	return false
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

type Package struct {
	XMLName xml.Name `xml:"package"`

	Classes []Class `xml:"classes>class"`
}

type Class struct {
	XMLName xml.Name `xml:"class"`

	Filename string `xml:"filename,attr"`
	Lines    []Line `xml:"lines>line"`
}

type Line struct {
	XMLName xml.Name `xml:"line"`

	Number int `xml:"number,attr"`
	Hits   int `xml:"hits,attr"`
}
