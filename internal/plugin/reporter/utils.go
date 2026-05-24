package reporter

import (
	"fmt"
	"strings"

	"github.com/target/pull-request-code-coverage/internal/plugin/domain"
)

func lineDescription(l domain.SourceLine) string {
	return fmt.Sprintf("%v:%v", filePath(l), l.LineNumber)
}

// filePath joins the non-empty path segments of a source line into a single
// file path (without the line number), used to group coverage by file.
func filePath(l domain.SourceLine) string {
	rawFileNameParts := []string{
		l.Module, l.SrcDir, l.Pkg, l.FileName,
	}

	var fileNameParts []string
	for _, part := range rawFileNameParts {
		if len(part) > 0 {
			fileNameParts = append(fileNameParts, part)
		}
	}

	return strings.Join(fileNameParts, "/")
}
