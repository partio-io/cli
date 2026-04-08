package hooks

import (
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/partio-io/cli/internal/agent"
	"github.com/partio-io/cli/internal/agent/claude"
	_ "github.com/partio-io/cli/internal/agent/codex"
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
	// Auto-detect running agents — check all registered detectors
	var detector agent.Detector
	var running bool

	active := agent.DetectActive()
	if len(active) > 0 {
		detector = active[0]
		running = true
		slog.Debug("auto-detected agent", "agent", detector.Name())
	}

	// Fall back to configured agent if none auto-detected
	if !running && cfg.Agent != "" {
		var err error
		detector, err = agent.NewDetector(cfg.Agent)
		if err != nil {
			slog.Warn("unknown agent", "agent", cfg.Agent, "error", err)
			detector = claude.New()
		}
	}

	if detector == nil {
		detector = claude.New()
	}

	// Check for condensed sessions (Claude-specific optimisation).
	if running {
		if cd, ok := detector.(*claude.Detector); ok {
			latestPath, pathErr := cd.FindLatestJSONLPath(repoRoot)
			if pathErr == nil {
				sid := claude.PeekSessionID(latestPath)
				if shouldSkipSession(filepath.Join(repoRoot, config.PartioDir), sid, latestPath) {
					slog.Debug("skipping already-condensed ended session", "session_id", sid)
					running = false
				}
			}
		}
	}

	// Find session data using the SessionParser interface (works for any agent).
	var sessionPath string
	if running {
		if sp, ok := detector.(agent.SessionParser); ok {
			path, _, findErr := sp.FindLatestSession(repoRoot)
			if findErr != nil {
				slog.Debug("agent running but no session found", "agent", detector.Name(), "error", findErr)
			} else {
				sessionPath = path
				slog.Debug("agent session detected", "agent", detector.Name(), "path", path)
			}
		}
	}

	branch, _ := git.CurrentBranch()
	commitHash, _ := git.CurrentCommit()

	// If the agent supports session parsing, require a session path.
	// Otherwise (no SessionParser), running is sufficient.
	agentActive := running
	if _, ok := detector.(agent.SessionParser); ok {
		agentActive = running && sessionPath != ""
	}

	state := preCommitState{
		AgentActive:   agentActive,
		AgentName:     detector.Name(),
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
