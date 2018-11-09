package jacoco

import (
	"encoding/xml"
	"git.target.com/search-product-team/pull-request-code-coverage/internal/plugin/coverage"
	"git.target.com/search-product-team/pull-request-code-coverage/internal/plugin/domain"
	"github.com/pkg/errors"
	"io"
	"io/ioutil"
	"log"
	"os"
)

type DefaultLoader struct {
	readAllFunc func(io.Reader) ([]byte, error)
}

func NewReportLoader() *DefaultLoader {
	return &DefaultLoader{
		readAllFunc: ioutil.ReadAll,
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

	return &coverageReport, nil
}

func silentlyCall(c func() error) {
	if err := c(); err != nil {
		log.Panic(err)
	}
}

type Report struct {
	XMLName xml.Name `xml:"report"`

	ReportName string    `xml:"name,attr"`
	Packages   []Package `xml:"package"`
}

func (r *Report) Name() string {
	return r.ReportName
}

func (r *Report) GetCoverageData(_ string, _ string, pkg string, fileName string, lineNumber int) (*domain.CoverageData, bool) {
	for _, p := range r.Packages {
		if p.RelativeName == pkg {
			return p.GetCoverageData(fileName, lineNumber)
		}
	}

	return nil, false
}

type Package struct {
	XMLName xml.Name `xml:"package"`

	RelativeName string       `xml:"name,attr"`
	SourceFiles  []SourceFile `xml:"sourcefile"`
}

func (p Package) GetCoverageData(fileName string, lineNumber int) (*domain.CoverageData, bool) {
	for _, f := range p.SourceFiles {
		if f.RelativeName == fileName {
			return f.GetCoverageData(lineNumber)
		}
	}

	return nil, false
}

type SourceFile struct {
	XMLName xml.Name `xml:"sourcefile"`

	RelativeName string `xml:"name,attr"`
	Lines        []Line `xml:"line"`
}

func (f SourceFile) GetCoverageData(lineNumber int) (*domain.CoverageData, bool) {
	for _, l := range f.Lines {
		if l.LineNumber == lineNumber {
			coverageData := l.CoverageData.toDomain()
			return &coverageData, true
		}
	}

	return nil, false
}

type Line struct {
	XMLName xml.Name `xml:"line"`

	CoverageData
	LineNumber int `xml:"nr,attr"`
}

type CoverageData struct {
	MissedInstructionCount  int `xml:"mi,attr"`
	CoveredInstructionCount int `xml:"ci,attr"`
}

func (data CoverageData) toDomain() domain.CoverageData {
	return domain.CoverageData{
		MissedInstructionCount:  data.MissedInstructionCount,
		CoveredInstructionCount: data.CoveredInstructionCount,
	}
}
