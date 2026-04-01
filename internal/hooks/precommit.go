package hooks

import (
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/partio-io/cli/internal/agent"
	"github.com/partio-io/cli/internal/agent/claude"
	"github.com/partio-io/cli/internal/agent/codex"
	"github.com/partio-io/cli/internal/config"
	"github.com/partio-io/cli/internal/git"
	"github.com/partio-io/cli/internal/session"
)

// preCommitState records the state captured during pre-commit for use by post-commit.
type preCommitState struct {
	AgentActive   bool   `json:"agent_active"`
	AgentName     string `json:"agent_name,omitempty"`
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
	detectors := []agent.Detector{claude.New(), codex.New()}

	// Find the first running agent detector.
	var activeDetector agent.Detector
	for _, d := range detectors {
		running, err := d.IsRunning()
		if err != nil {
			slog.Warn("could not detect agent process", "agent", d.Name(), "error", err)
			continue
		}
		if running {
			activeDetector = d
			break
		}
	}

	agentRunning := activeDetector != nil

	// For Claude: check whether the current session has already been captured.
	if agentRunning {
		if cd, ok := activeDetector.(*claude.Detector); ok {
			latestPath, pathErr := cd.FindLatestJSONLPath(repoRoot)
			if pathErr == nil {
				sid := claude.PeekSessionID(latestPath)
				if shouldSkipSession(filepath.Join(repoRoot, config.PartioDir), sid, latestPath) {
					slog.Debug("skipping already-condensed ended session", "session_id", sid)
					agentRunning = false
					activeDetector = nil
				}
			}
		}
	}

	// For Claude: find the latest session file.
	var sessionPath string
	if agentRunning {
		if cd, ok := activeDetector.(*claude.Detector); ok {
			path, _, err := cd.FindLatestSession(repoRoot)
			if err != nil {
				slog.Debug("agent running but no session found", "error", err)
			} else {
				sessionPath = path
				slog.Debug("agent session detected", "path", path)
			}
		}
	}

	// AgentActive is true when any agent is running. For Claude, require a
	// session path too (preserving existing behaviour). For other agents a
	// session path is not required.
	var agentActive bool
	var agentName string
	if agentRunning && activeDetector != nil {
		agentName = activeDetector.Name()
		if _, ok := activeDetector.(*claude.Detector); ok {
			agentActive = sessionPath != ""
		} else {
			agentActive = true
		}
	}

	branch, _ := git.CurrentBranch()
	commitHash, _ := git.CurrentCommit()

	state := preCommitState{
		AgentActive:   agentActive,
		AgentName:     agentName,
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
