// Package gitdiff produces a unified diff by running git locally, reproducing
// the behaviour of the old scripts/start.sh that was removed in v1.0.1.
package gitdiff

import (
	"bytes"
	"io"
	"os/exec"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// Loader fetches the PR diff by running `git fetch` + `git diff` in the
// current working directory. git must be on $PATH and the remote "origin" must
// be reachable (credentials via .netrc, SSH key, or the git credential helper
// already configured in the calling environment).
type Loader struct {
	baseBranch string
	module     string
}

func NewLoader(baseBranch, module string) *Loader {
	return &Loader{baseBranch: baseBranch, module: module}
}

// Load fetches origin/<baseBranch> and returns a unified=0 diff against HEAD.
// When module is non-empty it is passed as a path filter to git diff, matching
// the behaviour of the old start.sh for Gradle multi-module projects.
func (l *Loader) Load() (io.Reader, error) {
	if l.baseBranch == "" {
		return nil, errors.New("base branch is required for PARAMETER_DIFF_SOURCE=git (set PARAMETER_BASE_BRANCH or VELA_PULL_REQUEST_TARGET)")
	}

	fetchArgs := []string{"fetch", "--no-tags", "origin", l.baseBranch}
	logrus.Infof("running git %v", fetchArgs)
	fetchCmd := exec.Command("git", fetchArgs...) //nolint:gosec // args are controlled internally, not user input
	if out, err := fetchCmd.CombinedOutput(); err != nil {
		return nil, errors.Wrapf(err, "git fetch failed: %s", string(out))
	}

	diffArgs := []string{"--no-pager", "diff", "--unified=0", "origin/" + l.baseBranch}
	if l.module != "" {
		diffArgs = append(diffArgs, l.module)
	}
	logrus.Infof("running git %v", diffArgs)
	diffCmd := exec.Command("git", diffArgs...) //nolint:gosec // args are controlled internally, not user input
	var stdout, stderr bytes.Buffer
	diffCmd.Stdout = &stdout
	diffCmd.Stderr = &stderr
	if err := diffCmd.Run(); err != nil {
		return nil, errors.Wrapf(err, "git diff failed: %s", stderr.String())
	}

	return &stdout, nil
}
