package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/jcleira/ai-workflow-core/internal/config"
	"github.com/jcleira/ai-workflow-core/internal/git"
	githooks "github.com/jcleira/ai-workflow-core/internal/git/hooks"
)

func newDisableCmd() *cobra.Command {
	var removeData bool

	cmd := &cobra.Command{
		Use:   "disable",
		Short: "Disable partio in the current repository",
		Long:  `Removes git hooks installed by partio. By default preserves checkpoint data and config. Use --remove-data to also delete the .partio/ directory.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDisable(removeData)
		},
	}

	cmd.Flags().BoolVar(&removeData, "remove-data", false, "also remove .partio/ directory and config")

	return cmd
}

func runDisable(removeData bool) error {
	repoRoot, err := git.RepoRoot()
	if err != nil {
		return fmt.Errorf("must be run inside a git repository")
	}

	// Uninstall hooks
	if err := githooks.Uninstall(repoRoot); err != nil {
		return fmt.Errorf("uninstalling git hooks: %w", err)
	}

	fmt.Println("partio disabled.")
	fmt.Println("  - Removed git hooks (originals restored from backup if present)")

	if removeData {
		partioDir := filepath.Join(repoRoot, config.PartioDir)
		if err := os.RemoveAll(partioDir); err != nil {
			return fmt.Errorf("removing .partio directory: %w", err)
		}
		fmt.Println("  - Removed .partio/ directory")
	} else {
		fmt.Println("  - Checkpoint data and config preserved (use --remove-data to delete)")
	}

	return nil
}
