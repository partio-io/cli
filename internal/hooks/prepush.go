package hooks

import (
	"log/slog"

	"github.com/partio-io/cli/internal/config"
	"github.com/partio-io/cli/internal/git"
)

// PrePush runs pre-push hook logic.
func (r *Runner) PrePush() error {
	slog.Debug("pre-push hook running")
	return runPrePush(r.repoRoot, r.cfg)
}

func runPrePush(repoRoot string, cfg config.Config) error {
	if !cfg.StrategyOptions.PushSessions {
		slog.Debug("push_sessions disabled, skipping checkpoint push")
		return nil
	}

	if !git.HasRemote() {
		slog.Debug("no remote configured, skipping checkpoint push")
		return nil
	}

	if !git.BranchExists(git.CheckpointBranch) {
		slog.Debug("no checkpoint branch, skipping push")
		return nil
	}

	if err := git.PushBranch("origin", git.CheckpointBranch); err != nil {
		slog.Warn("could not push checkpoint branch", "error", err)
		// Don't fail the push
	}

	return nil
}
