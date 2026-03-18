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
	AgentActive   bool     `json:"agent_active"`
	SessionPath   string   `json:"session_path,omitempty"`
	PreCommitHash string   `json:"pre_commit_hash,omitempty"`
	Branch        string   `json:"branch"`
	AllAgentFiles []string `json:"all_agent_files,omitempty"` // staged + unstaged agent-modified files
	IsCarryForward bool    `json:"is_carry_forward,omitempty"` // activated by carry-forward
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

	agentActive := running && sessionPath != ""

	// Collect all agent-modified files (staged + unstaged) when agent is active.
	var allAgentFiles []string
	if agentActive {
		staged, _ := git.StagedFiles()
		unstaged, _ := git.UnstagedFiles()
		allAgentFiles = mergeFiles(staged, unstaged)
	}

	// Check carry-forward state: if a previous partial commit left pending files
	// that overlap with the current staged set, activate agent attribution.
	isCarryForward := false
	stateDir := filepath.Join(repoRoot, config.PartioDir, "state")
	if cf, err := loadCarryForward(stateDir); err == nil && cf != nil {
		staged, _ := git.StagedFiles()
		if activate, cfSessionPath := checkCarryForwardActivation(cf, staged); activate {
			if !agentActive {
				// Agent not running but carry-forward applies: use carry-forward session.
				agentActive = true
				sessionPath = cfSessionPath
				isCarryForward = true
				allAgentFiles = cf.PendingFiles
			} else {
				// Both agent active and carry-forward: merge file sets.
				isCarryForward = true
				allAgentFiles = mergeFiles(allAgentFiles, cf.PendingFiles)
			}
		}
	}

	state := preCommitState{
		AgentActive:    agentActive,
		SessionPath:    sessionPath,
		PreCommitHash:  commitHash,
		Branch:         branch,
		AllAgentFiles:  allAgentFiles,
		IsCarryForward: isCarryForward,
	}

	// Save state for post-commit
	if err := os.MkdirAll(stateDir, 0o755); err != nil {
		return err
	}

	data, err := json.Marshal(state)
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(stateDir, "pre-commit.json"), data, 0o644)
}
