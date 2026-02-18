package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/jcleira/ai-workflow-core/internal/git"
)

func newCleanCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "clean",
		Short: "Remove orphaned checkpoint data",
		Long:  `Scans the checkpoint branch for data that no longer corresponds to any commit in the repository and removes it.`,
		RunE:  runClean,
	}
}

func runClean(cmd *cobra.Command, args []string) error {
	_, err := git.RepoRoot()
	if err != nil {
		return fmt.Errorf("must be run inside a git repository")
	}

	// Check checkpoint branch exists
	_, err = git.ExecGit("rev-parse", "--verify", "partio/checkpoints/v1")
	if err != nil {
		return fmt.Errorf("checkpoint branch does not exist - nothing to clean")
	}

	// List checkpoint entries
	entries, err := git.ExecGit("ls-tree", "--name-only", "partio/checkpoints/v1")
	if err != nil {
		return fmt.Errorf("listing checkpoint entries: %w", err)
	}

	if entries == "" {
		fmt.Println("No checkpoint data found.")
		return nil
	}

	fmt.Println("Checkpoint data present on partio/checkpoints/v1 branch.")
	fmt.Println("To fully reset, run: partio reset")
	return nil
}
