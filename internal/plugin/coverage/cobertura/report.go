package cobertura

import (
	"encoding/xml"
	"git.target.com/search-product-team/pull-request-code-coverage/internal/plugin/coverage"
	"git.target.com/search-product-team/pull-request-code-coverage/internal/plugin/domain"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
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

	if len(coverageReport.ReportSources.Sources) > 1 {
		return nil, errors.Errorf("Cannot process report with more than 1 source: %v", coverageReport.ReportSources.Sources)
	}

	if len(coverageReport.ReportSources.Sources) == 1 {
		coverageReport.classSourceDir = coverageReport.ReportSources.Sources[0]
	} else {
		coverageReport.classSourceDir = ""
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

	sourceDir      string
	classSourceDir string

	ReportSources    ReportSources    `xml:"sources"`
	PackageContainer PackageContainer `xml:"packages"`
}

func (r *Report) GetCoverageData(_ string, srcDir string, pkg string, fileName string, lineNumber int) (*domain.CoverageData, bool) {

	rawDesiredFileNameParts := []string{
		r.sourceDir, srcDir, pkg, fileName,
	}

	logrus.Debugf("looking for file %v", rawDesiredFileNameParts)

	desiredFileName := strings.Join(withoutEmpty(rawDesiredFileNameParts), "/")
	logrus.Debugf("looking for file %v", desiredFileName)

	for _, p := range r.PackageContainer.Packages {

		for _, c := range p.ClassesContainer.Classes {

			classFilename := strings.Join(withoutEmpty([]string{r.classSourceDir, c.Filename}), "/")

			matches := classFilename == desiredFileName
			logrus.Debugf("%v == %v => %v", classFilename, desiredFileName, matches)

			if matches {
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

func withoutEmpty(parts []string) []string {
	var result []string
	for _, part := range parts {
		if len(part) > 0 {
			result = append(result, part)
		}
	}
	return result
}

type ReportSources struct {
	XMLName xml.Name `xml:"sources"`

	Sources []string `xml:"source"`
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
