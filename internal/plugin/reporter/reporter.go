package reporter

import (
	"git.target.com/searchoss/pull-request-code-coverage/internal/plugin/domain"
)

type Reporter interface {
	Write(domain.SourceLineCoverageReport) error

	GetName() string
}
