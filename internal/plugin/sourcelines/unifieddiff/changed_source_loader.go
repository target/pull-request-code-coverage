package unifieddiff

import (
	"bufio"
	"io"
	"regexp"

	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/target/pull-request-code-coverage/internal/plugin/domain"
)

type Loader struct {
	Module     string
	SourceDirs []string
}

func NewChangedSourceLinesLoader(module string, sourceDirs []string) *Loader {
	return &Loader{
		Module:     module,
		SourceDirs: sourceDirs,
	}
}

func (l *Loader) Load(inReader io.Reader) ([]domain.SourceLine, error) {

	scanner := bufio.NewScanner(inReader)

	lines := []string{}
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	return getChangedLinesFromUnifiedDiff(lines, l.Module, l.SourceDirs)
}

var changedFileLine = regexp.MustCompile("^[+][+][+][ ]b?[/](.*)")
var changedLineCounts = regexp.MustCompile("^[@][@][ ][-].*?[ ][+](.*?)[ ][@][@].*")
var addedLine = regexp.MustCompile("^[+].*")
var emptyStr = ""

// nolint: gocyclo
func getChangedLinesFromUnifiedDiff(unifiedDiffLines []string, module string, sourceDirs []string) ([]domain.SourceLine, error) {

	result := []domain.SourceLine{}

	var currentModule *string
	var currentSourceDir *string
	var currentPkg *string
	var currentFilename *string
	currentLineOffset := -1
	currentRelativeLine := 0
	linesLeftInBlock := -1

	for _, line := range unifiedDiffLines {
		if matches := changedFileLine.FindStringSubmatch(line); len(matches) > 0 {

			if linesLeftInBlock > 0 {
				return nil, errors.Errorf("Was not able to finish previous block %v %v %v %v %v %v %v", *currentModule, *currentPkg, *currentFilename, currentLineOffset, currentRelativeLine, linesLeftInBlock, line)
			}

			currentModule = &emptyStr
			workingPkg := emptyStr
			currentFilename = &(matches[1])
			currentSourceDir = &emptyStr

			fileNameParts := strings.Split(*currentFilename, "/")

			if len(module) > 0 && *currentFilename != "dev/null" {

				if len(fileNameParts) < 2 || fileNameParts[0] != module {
					return nil, errors.Errorf("Filename %v is invalid with expected module %v", *currentFilename, module)
				}

				currentModule = &(fileNameParts[0])
				fileNameParts = fileNameParts[1:]
			}

			if len(fileNameParts) == 1 {
				currentFilename = &(fileNameParts[0])
			} else {
				pkg := strings.Join(fileNameParts[:len(fileNameParts)-1], "/")
				workingPkg = pkg
				currentFilename = &(fileNameParts[len(fileNameParts)-1])
			}

			if sourceDir, found := findSourcePrefix(workingPkg, sourceDirs); found {
				sourceDirLessPkg := workingPkg[len(sourceDir+"/"):]
				currentPkg = &sourceDirLessPkg
				currentSourceDir = &sourceDir
			} else {
				currentPkg = &workingPkg
			}

			currentLineOffset = -1
			currentRelativeLine = 0
			linesLeftInBlock = -1
		} else if matches := changedLineCounts.FindStringSubmatch(line); len(matches) > 0 {

			if linesLeftInBlock > 0 {
				return nil, errors.Errorf("Was not able to finish previous block %v %v %v %v %v %v %v", *currentModule, *currentPkg, *currentFilename, currentLineOffset, currentRelativeLine, linesLeftInBlock, line)
			}

			rawLineOffsetData := matches[1]
			rawLineOffsetDatas := strings.Split(rawLineOffsetData, ",")

			var currentLineOffsetErr error
			currentLineOffset, currentLineOffsetErr = strconv.Atoi(rawLineOffsetDatas[0])
			if currentLineOffsetErr != nil {
				return nil, errors.Wrapf(currentLineOffsetErr, "Invalid line offset in line %v", line)
			}

			currentRelativeLine = 0

			if len(rawLineOffsetDatas) > 1 {
				var linesLeftInBlockErr error
				linesLeftInBlock, linesLeftInBlockErr = strconv.Atoi(rawLineOffsetDatas[1])
				if linesLeftInBlockErr != nil {
					return nil, errors.Wrapf(linesLeftInBlockErr, "Invalid line offset in line %v", line)
				}
			} else {
				linesLeftInBlock = 1
			}

		} else if addedLine.MatchString(line) {

			if linesLeftInBlock <= 0 {
				return nil, errors.Errorf("Finished previous block early %v %v %v %v %v %v %v", *currentModule, *currentPkg, *currentFilename, currentLineOffset, currentRelativeLine, linesLeftInBlock, line)
			}

			result = append(result, domain.SourceLine{
				LineValue:  line[1:],
				LineNumber: currentLineOffset + currentRelativeLine,
				FileName:   *currentFilename,
				Pkg:        *currentPkg,
				SrcDir:     *currentSourceDir,
				Module:     *currentModule,
			})

			currentRelativeLine++
			linesLeftInBlock--
		}
	}

	return result, nil
}

func findSourcePrefix(workingPkg string, sourceDirs []string) (string, bool) {
	for _, sourceDir := range sourceDirs {
		if strings.HasPrefix(workingPkg, sourceDir+"/") {
			return sourceDir, true
		}
	}

	return "", false
}
