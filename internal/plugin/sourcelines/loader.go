package sourcelines

import (
	"os"

	"git.target.com/target/pull-request-code-coverage/internal/plugin/domain"
)

type Loader interface {
	Load(inFile *os.File) ([]domain.SourceLine, error)
}
