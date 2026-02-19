package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/partio-io/cli/internal/git"
)

func newResetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "reset",
		Short: "Reset the checkpoint branch",
		Long:  `Deletes and recreates the partio/checkpoints/v1 branch. This removes all stored checkpoint data.`,
		RunE:  runReset,
	}
}

func runReset(cmd *cobra.Command, args []string) error {
	_, err := git.RepoRoot()
	if err != nil {
		return fmt.Errorf("must be run inside a git repository")
	}

	const branchName = "partio/checkpoints/v1"

	// Delete existing branch
	_, _ = git.ExecGit("branch", "-D", branchName)

	// Recreate
	if err := createCheckpointBranch(); err != nil {
		return fmt.Errorf("recreating checkpoint branch: %w", err)
	}

	fmt.Println("Checkpoint branch reset successfully.")
	return nil
}
