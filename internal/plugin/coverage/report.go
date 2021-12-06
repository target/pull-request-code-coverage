package coverage

import "git.target.com/target/pull-request-code-coverage/internal/plugin/domain"

type Loader interface {
	Load(coverageFile string) (Report, error)
}

type Report interface {
	GetCoverageData(module string, sourceDir string, pkg string, fileName string, lineNumber int) (*domain.CoverageData, bool)
}
