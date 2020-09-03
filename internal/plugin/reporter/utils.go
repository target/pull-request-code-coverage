package reporter

import (
	"fmt"
	"strings"

	"git.target.com/searchoss/pull-request-code-coverage/internal/plugin/domain"
)

func lineDescription(l domain.SourceLine) string {
	rawFileNameParts := []string{
		l.Module, l.SrcDir, l.Pkg, l.FileName,
	}

	var fileNameParts []string
	for _, part := range rawFileNameParts {
		if len(part) > 0 {
			fileNameParts = append(fileNameParts, part)
		}
	}

	return fmt.Sprintf("%v:%v", strings.Join(fileNameParts, "/"), l.LineNumber)
}
