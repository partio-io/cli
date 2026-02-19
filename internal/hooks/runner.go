package hooks

import (
	"github.com/partio-io/cli/internal/config"
	"github.com/partio-io/cli/internal/git"
)

// Runner executes hook logic.
type Runner struct {
	cfg      config.Config
	repoRoot string
}

// NewRunner creates a new hook runner.
func NewRunner(cfg config.Config) (*Runner, error) {
	repoRoot, err := git.RepoRoot()
	if err != nil {
		return nil, err
	}

	return &Runner{
		cfg:      cfg,
		repoRoot: repoRoot,
	}, nil
}
