package reporter

import (
	"github.com/target/pull-request-code-coverage/internal/plugin/domain"
)

type Reporter interface {
	Write(domain.SourceLineCoverageReport) error

	GetName() string
}
