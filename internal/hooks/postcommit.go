package hooks

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/partio-io/cli/internal/agent/claude"
	"github.com/partio-io/cli/internal/attribution"
	"github.com/partio-io/cli/internal/checkpoint"
	"github.com/partio-io/cli/internal/config"
	"github.com/partio-io/cli/internal/git"
)

// PostCommit runs post-commit hook logic.
func (r *Runner) PostCommit() error {
	slog.Debug("post-commit hook running")
	return runPostCommit(r.repoRoot, r.cfg)
}

func runPostCommit(repoRoot string, cfg config.Config) error {
	// Read pre-commit state
	stateFile := filepath.Join(repoRoot, config.PartioDir, "state", "pre-commit.json")
	data, err := os.ReadFile(stateFile)
	if err != nil {
		slog.Debug("no pre-commit state found, skipping checkpoint")
		return nil
	}
	// Remove immediately to prevent re-entry (amend triggers post-commit again)
	_ = os.Remove(stateFile)

	var state preCommitState
	if err := json.Unmarshal(data, &state); err != nil {
		return fmt.Errorf("parsing pre-commit state: %w", err)
	}

	if !state.AgentActive {
		slog.Debug("no agent was active, skipping checkpoint")
		return nil
	}

	// Get current commit hash
	commitHash, err := git.CurrentCommit()
	if err != nil {
		return fmt.Errorf("getting current commit: %w", err)
	}

	// Calculate attribution
	attr, err := attribution.Calculate(commitHash, state.AgentActive)
	if err != nil {
		slog.Warn("could not calculate attribution", "error", err)
		attr = &attribution.Result{AgentPercent: 100}
	}

	// Parse agent session data
	detector := claude.New()
	sessionPath, sessionData, err := detector.FindLatestSession(repoRoot)
	if err != nil {
		slog.Warn("could not read agent session", "error", err)
	}

	// Create checkpoint
	cpID := checkpoint.NewID()
	cp := &checkpoint.Checkpoint{
		ID:          cpID,
		CommitHash:  commitHash,
		Branch:      state.Branch,
		CreatedAt:   time.Now(),
		Agent:       cfg.Agent,
		AgentPct:    attr.AgentPercent,
		ContentHash: commitHash,
	}

	if sessionData != nil {
		cp.SessionID = sessionData.SessionID
	}

	// Prepare session files
	sessionFiles := &checkpoint.SessionFiles{
		ContentHash: commitHash,
		Context:     "",
		FullJSONL:   "",
		Metadata: checkpoint.SessionMetadata{
			Agent: cfg.Agent,
		},
		Prompt: "",
	}

	if sessionData != nil {
		sessionFiles.Context = sessionData.Context
		sessionFiles.Prompt = sessionData.Prompt
		sessionFiles.Metadata.TotalTokens = sessionData.TotalTokens
		sessionFiles.Metadata.Duration = sessionData.Duration.String()
	}

	if sessionPath != "" {
		rawJSONL, err := claude.ReadRawJSONL(sessionPath)
		if err == nil {
			sessionFiles.FullJSONL = string(rawJSONL)
		}
	}

	// Write checkpoint to orphan branch
	store := checkpoint.NewStore(repoRoot)
	if err := store.Write(cp, sessionFiles); err != nil {
		return fmt.Errorf("writing checkpoint: %w", err)
	}

	// Add trailers to commit
	trailers := map[string]string{
		"Partio-Checkpoint":  cpID,
		"Partio-Attribution": fmt.Sprintf("%d%% agent", attr.AgentPercent),
	}

	if err := git.AmendTrailers(trailers); err != nil {
		slog.Warn("could not add trailers to commit", "error", err)
	}

	slog.Debug("checkpoint created", "id", cpID, "agent_pct", attr.AgentPercent)
	return nil
}
