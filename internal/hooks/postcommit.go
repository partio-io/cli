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
	"github.com/partio-io/cli/internal/session"
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

	// Skip if this commit was already processed (e.g. duplicate hook invocations
	// during rebase, merge, or cherry-pick).
	partioDir := filepath.Join(repoRoot, config.PartioDir)
	cache := loadCommitCache(partioDir)
	if cache.contains(commitHash) {
		slog.Debug("post-commit: commit already processed, skipping", "commit", commitHash)
		return nil
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

	// Skip if this session is already fully condensed and ended — re-processing
	// it produces a redundant checkpoint with no new content.
	if sessionData != nil && sessionData.SessionID != "" {
		if shouldSkipSession(filepath.Join(repoRoot, config.PartioDir), sessionData.SessionID, sessionPath) {
			slog.Debug("post-commit: skipping already-condensed ended session", "session_id", sessionData.SessionID)
			return nil
		}
	}

	// Generate checkpoint ID and amend commit with trailers BEFORE writing
	// the checkpoint, so we capture the post-amend commit hash.
	cpID := checkpoint.NewID()

	trailers := map[string]string{
		"Partio-Checkpoint":  cpID,
		"Partio-Attribution": fmt.Sprintf("%d%% agent", attr.AgentPercent),
	}

	if err := git.AmendTrailers(trailers); err != nil {
		slog.Warn("could not add trailers to commit", "error", err)
	}

	// Get the post-amend commit hash (this is the hash that gets pushed)
	commitHash, err = git.CurrentCommit()
	if err != nil {
		return fmt.Errorf("getting post-amend commit: %w", err)
	}

	// Create checkpoint with the post-amend hash
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
		cp.PlanSlug = sessionData.PlanSlug
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

	if d, err := git.Diff(commitHash); err == nil {
		sessionFiles.Diff = d
	}

	if sessionData != nil {
		sessionFiles.Context = sessionData.Context
		sessionFiles.Prompt = sessionData.Prompt
		sessionFiles.Metadata.TotalTokens = sessionData.TotalTokens
		sessionFiles.Metadata.Duration = sessionData.Duration.String()
	}

	if sessionData != nil && sessionData.PlanSlug != "" {
		planContent, err := claude.ReadPlanFile(sessionData.PlanSlug)
		if err != nil {
			slog.Warn("could not read plan file", "slug", sessionData.PlanSlug, "error", err)
		} else {
			sessionFiles.Plan = planContent
		}
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

	// Mark the session as condensed so subsequent commits with the same session
	// are skipped. This is best-effort; failure is non-fatal.
	if sessionData != nil && sessionData.SessionID != "" {
		mgr := session.NewManager(filepath.Join(repoRoot, config.PartioDir))
		if markErr := mgr.MarkCondensed(sessionData.SessionID); markErr != nil {
			slog.Debug("could not mark session as condensed", "error", markErr)
		}
	}

	// Record the post-amend commit hash so duplicate hook invocations are no-ops.
	cache.add(commitHash)
	if saveErr := saveCommitCache(partioDir, cache); saveErr != nil {
		slog.Debug("could not save commit cache", "error", saveErr)
	}

	slog.Debug("checkpoint created", "id", cpID, "agent_pct", attr.AgentPercent)
	return nil
}
