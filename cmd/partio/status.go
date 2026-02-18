package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/jcleira/ai-workflow-core/internal/config"
	"github.com/jcleira/ai-workflow-core/internal/git"
)

func newStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show partio status for the current repository",
		RunE:  runStatus,
	}
}

func runStatus(cmd *cobra.Command, args []string) error {
	repoRoot, err := git.RepoRoot()
	if err != nil {
		return fmt.Errorf("must be run inside a git repository")
	}

	// Check if enabled
	partioDir := filepath.Join(repoRoot, config.PartioDir)
	enabled := false
	if _, err := os.Stat(partioDir); err == nil {
		enabled = true
	}

	branch, _ := git.CurrentBranch()

	fmt.Printf("Repository: %s\n", repoRoot)
	fmt.Printf("Branch:     %s\n", branch)

	if !enabled {
		fmt.Println("Status:     not enabled (run 'partio enable' to set up)")
		return nil
	}

	fmt.Println("Status:     enabled")
	fmt.Printf("Strategy:   %s\n", cfg.Strategy)
	fmt.Printf("Agent:      %s\n", cfg.Agent)

	// Check hooks
	hooksDir, hooksErr := git.HooksDir(repoRoot)
	hooks := []string{"pre-commit", "post-commit", "pre-push"}
	allInstalled := true
	if hooksErr != nil {
		allInstalled = false
	} else {
		for _, h := range hooks {
			if _, err := os.Stat(filepath.Join(hooksDir, h)); err != nil {
				allInstalled = false
				break
			}
		}
	}

	if allInstalled {
		fmt.Println("Hooks:      installed")
	} else {
		fmt.Println("Hooks:      missing (run 'partio enable' to reinstall)")
	}

	// Check checkpoint branch
	_, err = git.ExecGit("rev-parse", "--verify", "partio/checkpoints/v1")
	if err == nil {
		fmt.Println("Checkpoints: branch exists")
	} else {
		fmt.Println("Checkpoints: branch missing")
	}

	return nil
}
