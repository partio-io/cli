package hooks

import (
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/partio-io/cli/internal/agent/claude"
	"github.com/partio-io/cli/internal/config"
	"github.com/partio-io/cli/internal/git"
)

// preCommitState records the state captured during pre-commit for use by post-commit.
type preCommitState struct {
	AgentActive   bool   `json:"agent_active"`
	SessionPath   string `json:"session_path,omitempty"`
	PreCommitHash string `json:"pre_commit_hash,omitempty"`
	Branch        string `json:"branch"`
}

// PreCommit runs pre-commit hook logic.
func (r *Runner) PreCommit() error {
	slog.Debug("pre-commit hook running")
	return runPreCommit(r.repoRoot, r.cfg)
}

func runPreCommit(repoRoot string, cfg config.Config) error {
	detector := claude.New()

	// Detect if agent is running
	running, err := detector.IsRunning()
	if err != nil {
		slog.Warn("could not detect agent process", "error", err)
		running = false
	}

	var sessionPath string
	if running {
		path, _, err := detector.FindLatestSession(repoRoot)
		if err != nil {
			slog.Debug("agent running but no session found", "error", err)
		} else {
			sessionPath = path
			slog.Debug("agent session detected", "path", path)
		}
	}

	branch, _ := git.CurrentBranch()
	commitHash, _ := git.CurrentCommit()

	state := preCommitState{
		AgentActive:   running && sessionPath != "",
		SessionPath:   sessionPath,
		PreCommitHash: commitHash,
		Branch:        branch,
	}

	// Save state for post-commit
	stateDir := filepath.Join(repoRoot, config.PartioDir, "state")
	if err := os.MkdirAll(stateDir, 0o755); err != nil {
		return err
	}

	data, err := json.Marshal(state)
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(stateDir, "pre-commit.json"), data, 0o644)
}
