package reporter

import (
	"git.target.com/search-product-team/pull-request-code-coverage/internal/plugin/domain"
)

type Reporter interface {
	Write(domain.SourceLineCoverageReport) error
}
