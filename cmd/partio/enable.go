package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/partio-io/cli/internal/config"
	"github.com/partio-io/cli/internal/git"
	githooks "github.com/partio-io/cli/internal/git/hooks"
)

func newEnableCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "enable",
		Short: "Enable partio in the current repository",
		Long:  `Sets up partio in the current git repository by creating the .partio/ config directory, installing git hooks, and creating the checkpoint orphan branch. If a remote checkpoint branch already exists (e.g., on a fresh clone), the local branch is created tracking the remote so existing checkpoint history is preserved.`,
		RunE:  runEnable,
	}
	cmd.Flags().Bool("absolute-path", false, "Install hooks using the absolute path to the partio binary (useful when partio is not on PATH in hook execution environments)")
	return cmd
}

func runEnable(cmd *cobra.Command, args []string) error {
	absolutePath, _ := cmd.Flags().GetBool("absolute-path")

	repoRoot, err := git.RepoRoot()
	if err != nil {
		return fmt.Errorf("must be run inside a git repository")
	}

	partioDir := filepath.Join(repoRoot, config.PartioDir)

	// Check if already enabled
	if _, err := os.Stat(partioDir); err == nil {
		fmt.Println("partio is already enabled in this repository.")
		return nil
	}

	// Create .partio/ directory
	if err := os.MkdirAll(partioDir, 0o755); err != nil {
		return fmt.Errorf("creating .partio directory: %w", err)
	}

	// Write default settings.json
	defaults := config.Defaults()
	data, err := json.MarshalIndent(defaults, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling default config: %w", err)
	}
	if err := os.WriteFile(filepath.Join(partioDir, "settings.json"), data, 0o644); err != nil {
		return fmt.Errorf("writing settings.json: %w", err)
	}

	// Add .partio/settings.local.json to .gitignore
	addToGitignore(repoRoot, ".partio/settings.local.json")

	// Install git hooks
	if absolutePath {
		exePath, err := os.Executable()
		if err != nil {
			return fmt.Errorf("resolving partio binary path: %w", err)
		}
		exePath, err = filepath.EvalSymlinks(exePath)
		if err != nil {
			return fmt.Errorf("resolving partio binary symlinks: %w", err)
		}
		if err := githooks.InstallAbsolute(repoRoot, exePath); err != nil {
			return fmt.Errorf("installing git hooks: %w", err)
		}
	} else if err := githooks.Install(repoRoot); err != nil {
		return fmt.Errorf("installing git hooks: %w", err)
	}

	// Create orphan checkpoint branch (or restore from remote)
	restored, err := createCheckpointBranch()
	if err != nil {
		slog.Warn("could not create checkpoint branch (may already exist)", "error", err)
	}

	fmt.Println("partio enabled successfully!")
	fmt.Println("  - Created .partio/ config directory")
	fmt.Println("  - Installed git hooks (pre-commit, post-commit, pre-push)")
	if restored {
		fmt.Println("  - Restored checkpoint history from remote branch")
	}
	fmt.Println("  - Ready to capture AI sessions on commit")
	return nil
}

func addToGitignore(repoRoot, entry string) {
	gitignore := filepath.Join(repoRoot, ".gitignore")

	// Read existing content
	existing, _ := os.ReadFile(gitignore)
	content := string(existing)

	// Check if already present
	for _, line := range filepath.SplitList(content) {
		if line == entry {
			return
		}
	}

	// Append
	f, err := os.OpenFile(gitignore, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		slog.Warn("could not update .gitignore", "error", err)
		return
	}
	defer func() { _ = f.Close() }()

	if len(existing) > 0 && existing[len(existing)-1] != '\n' {
		_, _ = f.WriteString("\n")
	}
	_, _ = f.WriteString(entry + "\n")
}

func createCheckpointBranch() (bool, error) {
	const branchName = "partio/checkpoints/v1"

	// Check if branch already exists
	_, err := git.ExecGit("rev-parse", "--verify", branchName)
	if err == nil {
		return false, nil // already exists
	}

	// Check if a remote tracking branch exists (e.g., fresh clone of an existing repo)
	if git.RemoteBranchExists(branchName) {
		_, err = git.ExecGit("branch", "--track", branchName, "origin/"+branchName)
		if err != nil {
			return false, fmt.Errorf("creating local branch from remote: %w", err)
		}
		return true, nil
	}

	// Create orphan branch with an empty initial commit using plumbing
	// 1. Create empty tree
	treeHash, err := git.ExecGit("hash-object", "-t", "tree", "/dev/null")
	if err != nil {
		// Alternative: write an empty tree
		treeHash, err = git.ExecGit("mktree", "--missing")
		if err != nil {
			return false, fmt.Errorf("creating empty tree: %w", err)
		}
	}

	// 2. Create commit with no parent
	commitHash, err := git.ExecGit("commit-tree", treeHash, "-m", "partio: initialize checkpoint storage")
	if err != nil {
		return false, fmt.Errorf("creating initial commit: %w", err)
	}

	// 3. Create the ref
	_, err = git.ExecGit("update-ref", "refs/heads/"+branchName, commitHash)
	if err != nil {
		return false, fmt.Errorf("creating branch ref: %w", err)
	}

	return false, nil
}
