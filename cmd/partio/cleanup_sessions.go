package main

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/partio-io/cli/internal/config"
	"github.com/partio-io/cli/internal/git"
	"github.com/partio-io/cli/internal/session"
)

func newCleanupCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "cleanup",
		Short: "Clean up stale session files",
		Long: `Scan for session files left behind by crashed or improperly terminated agent
processes and transition them to ENDED state.

A session is considered stale when it is in ACTIVE or IDLE state, its state
file has not been updated within the configured threshold (default: 10 minutes),
and — when a process ID is recorded — the agent process is no longer alive.

The threshold is configurable via .partio/settings.json ("stale_session_threshold")
or the PARTIO_STALE_SESSION_THRESHOLD environment variable (e.g. "10m", "30m").`,
		RunE: runCleanup,
	}
}

func runCleanup(cmd *cobra.Command, args []string) error {
	repoRoot, err := git.RepoRoot()
	if err != nil {
		return fmt.Errorf("must be run inside a git repository")
	}

	partioDir := filepath.Join(repoRoot, config.PartioDir)
	mgr := session.NewManager(partioDir)

	result, err := mgr.CleanupStale(cfg.StaleSessionThreshold)
	if err != nil {
		return fmt.Errorf("cleanup failed: %w", err)
	}

	if result.Cleaned {
		fmt.Printf("Cleaned up stale session: %s (agent=%s)\n", result.Session.ID, result.Session.Agent)
	} else {
		fmt.Println("No stale sessions found.")
	}

	return nil
}
