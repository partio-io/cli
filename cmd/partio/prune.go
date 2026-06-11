package main

import (
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/spf13/cobra"

	"github.com/partio-io/cli/internal/checkpoint"
	"github.com/partio-io/cli/internal/git"
)

func newPruneCmd() *cobra.Command {
	var (
		olderThan string
		dryRun    bool
	)

	cmd := &cobra.Command{
		Use:   "prune",
		Short: "Delete old checkpoints",
		Long:  `Remove checkpoints older than a retention window from the partio/checkpoints/v1 branch. Never deletes the checkpoint linked to the current HEAD.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPrune(olderThan, dryRun)
		},
	}

	cmd.Flags().StringVar(&olderThan, "older-than", "90d", "retention window (e.g. 30d, 90d, 365d)")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "preview what would be deleted without making changes")

	return cmd
}

func runPrune(olderThan string, dryRun bool) error {
	repoRoot, err := git.RepoRoot()
	if err != nil {
		return fmt.Errorf("must be run inside a git repository")
	}

	duration, err := parseDuration(olderThan)
	if err != nil {
		return fmt.Errorf("invalid --older-than value %q: %w", olderThan, err)
	}

	currentCommit, err := git.CurrentCommit()
	if err != nil {
		return fmt.Errorf("getting current commit: %w", err)
	}

	store := checkpoint.NewStore(repoRoot)
	result, err := store.Prune(duration, currentCommit, dryRun)
	if err != nil {
		return fmt.Errorf("pruning checkpoints: %w", err)
	}

	if dryRun {
		fmt.Println("Dry run — no changes made.")
		fmt.Println()
	}

	if len(result.Removed) == 0 {
		fmt.Println("No checkpoints to prune.")
		return nil
	}

	verb := "Removed"
	if dryRun {
		verb = "Would remove"
	}

	for _, meta := range result.Removed {
		fmt.Printf("  %s %s (branch=%s, created=%s)\n", verb, meta.ID, meta.Branch, meta.CreatedAt)
	}
	fmt.Println()

	if dryRun {
		fmt.Printf("%d checkpoint(s) would be removed, %d kept.\n", len(result.Removed), len(result.Kept))
	} else {
		fmt.Printf("%d checkpoint(s) removed, %d kept.\n", len(result.Removed), len(result.Kept))
	}

	return nil
}

// parseDuration parses a duration string like "90d" into a time.Duration.
// Supports "d" for days as a suffix.
var durationPattern = regexp.MustCompile(`^(\d+)d$`)

func parseDuration(s string) (time.Duration, error) {
	matches := durationPattern.FindStringSubmatch(s)
	if matches == nil {
		// Try standard Go duration as fallback
		return time.ParseDuration(s)
	}

	days, err := strconv.Atoi(matches[1])
	if err != nil {
		return 0, err
	}
	return time.Duration(days) * 24 * time.Hour, nil
}
