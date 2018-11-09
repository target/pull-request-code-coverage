package sourcelines

import (
	"git.target.com/search-product-team/pull-request-code-coverage/internal/plugin/domain"
	"os"
)

type Loader interface {
	Load(inFile *os.File) ([]domain.SourceLine, error)
}
