package hooks

import (
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/partio-io/cli/internal/agent/claude"
	"github.com/partio-io/cli/internal/config"
	"github.com/partio-io/cli/internal/git"
	"github.com/partio-io/cli/internal/session"
)

// claudeDetector is satisfied by *claude.Detector and allows
// Claude-specific session operations in hooks.
type claudeDetector interface {
	FindLatestJSONLPath(repoRoot string) (string, error)
}

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
	detector := resolveDetector(cfg.Agent)

	// Detect if agent is running
	running, err := detector.IsRunning()
	if err != nil {
		slog.Warn("could not detect agent process", "error", err)
		running = false
	}

	var sessionPath string

	// Claude-specific: check for already-condensed sessions and find JSONL path.
	if cd, ok := detector.(claudeDetector); running && ok {
		latestPath, pathErr := cd.FindLatestJSONLPath(repoRoot)
		if pathErr == nil {
			sid := claude.PeekSessionID(latestPath)
			if shouldSkipSession(filepath.Join(repoRoot, config.PartioDir), sid, latestPath) {
				slog.Debug("skipping already-condensed ended session", "session_id", sid)
				running = false
			}
		}

		if running {
			if cdFull, ok := detector.(*claude.Detector); ok {
				path, _, err := cdFull.FindLatestSession(repoRoot)
				if err != nil {
					slog.Debug("agent running but no session found", "error", err)
				} else {
					sessionPath = path
					slog.Debug("agent session detected", "path", path)
				}
			}
		}
	}

	branch, _ := git.CurrentBranch()
	commitHash, _ := git.CurrentCommit()

	// Claude requires a session path to be active; other agents are active when running.
	agentActive := running
	if _, ok := detector.(*claude.Detector); ok {
		agentActive = running && sessionPath != ""
	}

	state := preCommitState{
		AgentActive:   agentActive,
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
