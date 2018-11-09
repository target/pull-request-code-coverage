package cobertura

import (
	"encoding/xml"
	"git.target.com/search-product-team/pull-request-code-coverage/internal/plugin/coverage"
	"git.target.com/search-product-team/pull-request-code-coverage/internal/plugin/domain"
	"github.com/pkg/errors"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

type DefaultLoader struct {
	sourceDir   string
	readAllFunc func(io.Reader) ([]byte, error)
}

func NewReportLoader(sourceDir string) *DefaultLoader {
	return &DefaultLoader{
		readAllFunc: ioutil.ReadAll,
		sourceDir:   sourceDir,
	}
}

func (l *DefaultLoader) Load(coverageFile string) (coverage.Report, error) {
	// nolint: gosec
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

	coverageReport.sourceDir = l.sourceDir

	return &coverageReport, nil
}

func silentlyCall(c func() error) {
	if err := c(); err != nil {
		log.Panic(err)
	}
}

type Report struct {
	XMLName xml.Name `xml:"coverage"`

	sourceDir        string
	ReportName       string           `xml:"name,attr"`
	PackageContainer PackageContainer `xml:"packages"`
}

func (r *Report) Name() string {
	return r.ReportName
}

func (r *Report) GetCoverageData(_ string, srcDir string, pkg string, fileName string, lineNumber int) (*domain.CoverageData, bool) {

	rawDesiredFileNameParts := []string{
		r.sourceDir, srcDir, pkg, fileName,
	}

	desiredFileNameParts := []string{}
	for _, part := range rawDesiredFileNameParts {
		if len(part) > 0 {
			desiredFileNameParts = append(desiredFileNameParts, part)
		}
	}

	desiredFileName := strings.Join(desiredFileNameParts, "/")

	for _, p := range r.PackageContainer.Packages {

		for _, c := range p.ClassesContainer.Classes {
			if c.Filename == desiredFileName {
				for _, l := range c.LinesContainer.Lines {
					if l.Number == lineNumber {

						if l.Hits > 0 {
							return &domain.CoverageData{
								MissedInstructionCount:  0,
								CoveredInstructionCount: 1,
							}, true
						}

						return &domain.CoverageData{
							MissedInstructionCount:  1,
							CoveredInstructionCount: 0,
						}, true
					}
				}
			}
		}
	}

	return nil, false
}

type PackageContainer struct {
	XMLName xml.Name `xml:"packages"`

	Packages []Package `xml:"package"`
}

type Package struct {
	XMLName xml.Name `xml:"package"`

	Name             string           `xml:"name,attr"`
	ClassesContainer ClassesContainer `xml:"classes"`
}

type ClassesContainer struct {
	XMLName xml.Name `xml:"classes"`

	Classes []Class `xml:"class"`
}

type Class struct {
	XMLName xml.Name `xml:"class"`

	Filename       string         `xml:"filename,attr"`
	LinesContainer LinesContainer `xml:"lines"`
}

type LinesContainer struct {
	XMLName xml.Name `xml:"lines"`

	Lines []Line `xml:"line"`
}

type Line struct {
	XMLName xml.Name `xml:"line"`

	Number int `xml:"number,attr"`
	Hits   int `xml:"hits,attr"`
}
