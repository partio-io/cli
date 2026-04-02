package hooks

import (
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/partio-io/cli/internal/agent/claude"
	"github.com/partio-io/cli/internal/agent/codex"
	"github.com/partio-io/cli/internal/config"
	"github.com/partio-io/cli/internal/git"
	"github.com/partio-io/cli/internal/session"
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
	claudeDetector := claude.New()

	// Detect if Claude Code is running (with full session capture support).
	running, err := claudeDetector.IsRunning()
	if err != nil {
		slog.Warn("could not detect agent process", "error", err)
		running = false
	}

	if running {
		// Quick check: find the latest JSONL path without a full parse and see if
		// we have already captured this session in a fully-condensed ended state.
		// This avoids the expensive JSONL parse for stale sessions.
		latestPath, pathErr := claudeDetector.FindLatestJSONLPath(repoRoot)
		if pathErr == nil {
			sid := claude.PeekSessionID(latestPath)
			if shouldSkipSession(filepath.Join(repoRoot, config.PartioDir), sid, latestPath) {
				slog.Debug("skipping already-condensed ended session", "session_id", sid)
				running = false
			}
		}
	}

	// If Claude Code is not running, check for Codex CLI.
	if !running {
		codexDetector := codex.New()
		codexRunning, codexErr := codexDetector.IsRunning()
		if codexErr != nil {
			slog.Warn("could not detect codex process", "error", codexErr)
		} else if codexRunning {
			running = true
			slog.Debug("codex agent detected")
		}
	}

	var sessionPath string
	if running {
		path, _, err := claudeDetector.FindLatestSession(repoRoot)
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

// shouldSkipSession returns true when the Partio session state shows that
// sessionID has already been fully captured (condensed + ended) and the JSONL
// at sessionPath has not been modified since the capture time. Both ENDED and
// Condensed must be set; IDLE or ACTIVE sessions are never skipped.
func shouldSkipSession(partioDir, sessionID, sessionPath string) bool {
	if sessionID == "" || sessionPath == "" {
		return false
	}

	mgr := session.NewManager(partioDir)
	sess, err := mgr.Current()
	if err != nil || sess == nil {
		return false
	}

	if sess.State != session.StateEnded || !sess.Condensed || sess.CapturedSessionID != sessionID {
		return false
	}

	// Allow skip only if the JSONL hasn't been modified since we captured it.
	// If new messages were added, the modification time will be after CapturedAt.
	if !sess.CapturedAt.IsZero() {
		info, statErr := os.Stat(sessionPath)
		if statErr == nil && info.ModTime().After(sess.CapturedAt) {
			return false
		}
	}

	return true
}
