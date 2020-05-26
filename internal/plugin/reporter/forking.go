package reporter

import (
	"git.target.com/searchoss/pull-request-code-coverage/internal/plugin/domain"
	"github.com/pkg/errors"
)

type Forking struct {
	reporters []Reporter
}

func NewForking(reporters []Reporter) *Forking {
	return &Forking{
		reporters: reporters,
	}
}

func (s *Forking) Write(changedLinesWithCoverage domain.SourceLineCoverageReport) error {

	var errs []error

	for _, r := range s.reporters {
		if err := r.Write(changedLinesWithCoverage); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return errors.Errorf("Unexpected errors occurred: %v", errs)
	}

	return nil
}
