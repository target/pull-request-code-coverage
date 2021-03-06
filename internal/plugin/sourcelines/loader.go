package sourcelines

import (
	"os"

	"github.com/target/pull-request-code-coverage/internal/plugin/domain"
)

type Loader interface {
	Load(inFile *os.File) ([]domain.SourceLine, error)
}
