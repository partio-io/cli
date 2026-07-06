package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/partio-io/cli/internal/config"
	"github.com/partio-io/cli/internal/git"
	githooks "github.com/partio-io/cli/internal/git/hooks"
)

func newEnableCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "enable",
		Short: "Enable partio in the current repository",
		Long:  `Sets up partio in the current git repository by creating the .partio/ config directory, installing git hooks, and creating the checkpoint orphan branch.`,
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

	// Create .partio/ directory (no-op if it already exists)
	if err := os.MkdirAll(partioDir, 0o755); err != nil {
		return fmt.Errorf("creating .partio directory: %w", err)
	}

	// Ensure settings.json exists with enabled: true
	settingsPath := filepath.Join(partioDir, "settings.json")
	if err := ensureSettingsEnabled(settingsPath); err != nil {
		return fmt.Errorf("writing settings.json: %w", err)
	}

	// Add runtime files to .gitignore
	addToGitignore(repoRoot, ".partio/settings.local.json")
	addToGitignore(repoRoot, ".partio/sessions/")
	addToGitignore(repoRoot, ".partio/state/")

	// Install git hooks (reinstalls if missing or stale)
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

	// Create orphan checkpoint branch
	if err := createCheckpointBranch(); err != nil {
		slog.Warn("could not create checkpoint branch (may already exist)", "error", err)
	}

	fmt.Println("partio enabled successfully!")
	fmt.Println("  - Ensured .partio/ config directory exists")
	fmt.Println("  - Installed git hooks (pre-commit, post-commit, pre-push)")
	fmt.Println("  - Ready to capture AI sessions on commit")
	return nil
}

// ensureSettingsEnabled creates settings.json with defaults if it doesn't exist,
// or ensures enabled is set to true in existing settings.
func ensureSettingsEnabled(settingsPath string) error {
	existing, err := os.ReadFile(settingsPath)
	if err != nil {
		// File doesn't exist — write defaults
		defaults := config.Defaults()
		data, marshalErr := json.MarshalIndent(defaults, "", "  ")
		if marshalErr != nil {
			return fmt.Errorf("marshaling default config: %w", marshalErr)
		}
		return os.WriteFile(settingsPath, data, 0o644)
	}

	// File exists — parse and ensure enabled: true
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(existing, &raw); err != nil {
		// Corrupted JSON — overwrite with defaults
		defaults := config.Defaults()
		data, marshalErr := json.MarshalIndent(defaults, "", "  ")
		if marshalErr != nil {
			return fmt.Errorf("marshaling default config: %w", marshalErr)
		}
		return os.WriteFile(settingsPath, data, 0o644)
	}

	// Check if enabled is already true
	if v, ok := raw["enabled"]; ok {
		var enabled bool
		if json.Unmarshal(v, &enabled) == nil && enabled {
			return nil // already enabled, preserve existing config
		}
	}

	// Set enabled to true, preserving other settings
	raw["enabled"] = json.RawMessage("true")
	data, err := json.MarshalIndent(raw, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}
	return os.WriteFile(settingsPath, data, 0o644)
}

func addToGitignore(repoRoot, entry string) {
	gitignore := filepath.Join(repoRoot, ".gitignore")

	// Read existing content
	existing, _ := os.ReadFile(gitignore)
	content := string(existing)

	// Check if already present
	for _, line := range strings.Split(content, "\n") {
		if strings.TrimSpace(line) == entry {
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

func createCheckpointBranch() error {
	const branchName = "partio/checkpoints/v1"

	// Check if branch already exists
	_, err := git.ExecGit("rev-parse", "--verify", branchName)
	if err == nil {
		return nil // already exists
	}

	// Create orphan branch with an empty initial commit using plumbing
	// 1. Create empty tree
	treeHash, err := git.ExecGit("hash-object", "-t", "tree", "/dev/null")
	if err != nil {
		// Alternative: write an empty tree
		treeHash, err = git.ExecGit("mktree", "--missing")
		if err != nil {
			return fmt.Errorf("creating empty tree: %w", err)
		}
	}

	// 2. Create commit with no parent
	commitHash, err := git.ExecGit("commit-tree", treeHash, "-m", "partio: initialize checkpoint storage")
	if err != nil {
		return fmt.Errorf("creating initial commit: %w", err)
	}

	// 3. Create the ref
	_, err = git.ExecGit("update-ref", "refs/heads/"+branchName, commitHash)
	if err != nil {
		return fmt.Errorf("creating branch ref: %w", err)
	}

	return nil
}
