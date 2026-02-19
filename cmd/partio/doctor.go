package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/partio-io/cli/internal/config"
	"github.com/partio-io/cli/internal/git"
)

func newDoctorCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Check partio installation health and fix common issues",
		RunE:  runDoctor,
	}
}

func runDoctor(cmd *cobra.Command, args []string) error {
	repoRoot, err := git.RepoRoot()
	if err != nil {
		return fmt.Errorf("must be run inside a git repository")
	}

	issues := 0

	// Check .partio/ directory
	partioDir := filepath.Join(repoRoot, config.PartioDir)
	if _, err := os.Stat(partioDir); os.IsNotExist(err) {
		fmt.Println("[WARN] .partio/ directory missing - run 'partio enable'")
		issues++
	} else {
		fmt.Println("[OK]   .partio/ directory exists")
	}

	// Check hooks
	hooksDir, hooksErr := git.HooksDir(repoRoot)
	hookNames := []string{"pre-commit", "post-commit", "pre-push"}
	for _, name := range hookNames {
		if hooksErr != nil {
			fmt.Printf("[WARN] %s hook: cannot resolve hooks directory\n", name)
			issues++
			continue
		}
		hookPath := filepath.Join(hooksDir, name)
		if _, err := os.Stat(hookPath); os.IsNotExist(err) {
			fmt.Printf("[WARN] %s hook missing\n", name)
			issues++
		} else {
			data, _ := os.ReadFile(hookPath)
			if strings.Contains(string(data), "partio _hook") {
				fmt.Printf("[OK]   %s hook installed\n", name)
			} else {
				fmt.Printf("[WARN] %s hook exists but not managed by partio\n", name)
				issues++
			}
		}
	}

	// Check checkpoint branch
	_, err = git.ExecGit("rev-parse", "--verify", "partio/checkpoints/v1")
	if err != nil {
		fmt.Println("[WARN] checkpoint branch missing")
		issues++
	} else {
		fmt.Println("[OK]   checkpoint branch exists")
	}

	// Check partio binary in PATH
	fmt.Println("[OK]   partio binary found (you're running it!)")

	if issues == 0 {
		fmt.Println("\nAll checks passed!")
	} else {
		fmt.Printf("\n%d issue(s) found. Run 'partio enable' to fix.\n", issues)
	}

	return nil
}
