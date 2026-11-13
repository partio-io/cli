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
		Long:  `Sets up partio in the current git repository by creating the .partio/ config directory, installing git hooks, and creating the checkpoint orphan branch.

With --stealth, nothing in the working tree changes: the checkpoint branch is never pushed and commits get no Partio trailers. Both settings are written to the global config (~/.config/partio/settings.json), so they apply to every repository that does not override them. Undo stealth by editing that file.`,
		RunE:  runEnable,
	}
	cmd.Flags().Bool("absolute-path", false, "Install hooks using the absolute path to the partio binary (useful when partio is not on PATH in hook execution environments)")
	cmd.Flags().Bool("stealth", false, "Capture checkpoints locally only: write push_sessions=false and disable_trailers=true to the global config (~/.config/partio/settings.json), leave .partio/settings.json and .gitignore untouched, and hide .partio/ via git's info/exclude. The setting is global and applies to every repository without an override; undo it by editing ~/.config/partio/settings.json")
	return cmd
}

func runEnable(cmd *cobra.Command, args []string) error {
	absolutePath, flagErr := cmd.Flags().GetBool("absolute-path")
	if flagErr != nil {
		slog.Debug("could not read --absolute-path flag", "error", flagErr)
	}
	stealth, flagErr := cmd.Flags().GetBool("stealth")
	if flagErr != nil {
		slog.Debug("could not read --stealth flag", "error", flagErr)
	}

	repoRoot, err := git.RepoRoot()
	if err != nil {
		return fmt.Errorf("must be run inside a git repository")
	}

	// Stealth settings go to the global config only, which lives outside every
	// repository. Write them first so a malformed global file stops enable
	// before anything else is changed.
	var globalPath string
	if stealth {
		globalPath, err = config.SaveGlobalStrategyOptions(map[string]any{
			"push_sessions":    false,
			"disable_trailers": true,
		})
		if err != nil {
			return fmt.Errorf("writing global stealth settings: %w", err)
		}
	}

	partioDir := filepath.Join(repoRoot, config.PartioDir)

	// Create .partio/ directory (no-op if it already exists)
	if err := os.MkdirAll(partioDir, 0o755); err != nil {
		return fmt.Errorf("creating .partio directory: %w", err)
	}

	if stealth {
		// Leave .partio/settings.json and .gitignore alone so nothing shows up
		// as a change; enabled already defaults to true. Hide the whole
		// .partio/ directory through git's untracked, per-clone exclude file.
		addToGitExclude(".partio/")
	} else {
		// Ensure settings.json exists with enabled: true
		settingsPath := filepath.Join(partioDir, "settings.json")
		if err := ensureSettingsEnabled(settingsPath); err != nil {
			return fmt.Errorf("writing settings.json: %w", err)
		}

		// Add runtime files to .gitignore
		addToGitignore(repoRoot, ".partio/settings.local.json")
		addToGitignore(repoRoot, ".partio/sessions/")
		addToGitignore(repoRoot, ".partio/state/")
	}

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

	// Warn about external hook managers that may conflict with partio's hooks
	warnHookManagers(repoRoot)

	fmt.Println("partio enabled successfully!")
	fmt.Println("  - Ensured .partio/ config directory exists")
	fmt.Println("  - Installed git hooks (pre-commit, post-commit, pre-push)")
	fmt.Println("  - Ready to capture AI sessions on commit")
	if stealth {
		printStealthEnabled(globalPath)
		warnStealthOverrides(repoRoot)
	}
	return nil
}

// printStealthEnabled tells the user that stealth is on, where it is stored and
// how to find checkpoints now that commits carry no trailers. globalPath is
// the global settings file the stealth options were written to.
func printStealthEnabled(globalPath string) {
	fmt.Println("\nStealth mode active: checkpoints stay local; the checkpoint branch is not pushed and commits get no trailers.")
	fmt.Printf("  - Wrote push_sessions=false and disable_trailers=true to %s\n", globalPath)
	fmt.Println("    This applies to every repository without an override; undo it by editing that file.")
	fmt.Println("  - Hid .partio/ via git's info/exclude; no repository files were changed")
	fmt.Println("  - List checkpoints with `partio rewind --list` and inspect one with `partio explain <id>`")
}

// warnHookManagers detects external git hook managers and prints integration warnings.
func warnHookManagers(repoRoot string) {
	managers := githooks.DetectHookManagers(repoRoot)
	for _, m := range managers {
		fmt.Printf("\nWarning: %s detected.\n", m.Name)
		fmt.Println("  Partio's hooks must be called from within your hook manager's configuration.")
		fmt.Printf("  %s\n", m.Instructions)
	}
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
	if err := appendIgnoreLine(gitignore, entry); err != nil {
		slog.Warn("could not update .gitignore", "entry", entry, "error", err)
	}
}

// addToGitExclude adds entry to the file git resolves for info/exclude. That
// file is never tracked, and since git 2.5 it lives in the common git dir, so
// one entry covers every linked worktree of the clone. A failure is a warning:
// enable continues without it.
func addToGitExclude(entry string) {
	excludePath, err := git.ExecGit("rev-parse", "--git-path", "info/exclude")
	if err != nil {
		slog.Warn("could not locate git info/exclude", "error", err)
		return
	}
	// --git-path may print a path relative to the current directory.
	if abs, absErr := filepath.Abs(excludePath); absErr == nil {
		excludePath = abs
	}
	if err := os.MkdirAll(filepath.Dir(excludePath), 0o755); err != nil {
		slog.Warn("could not update git info/exclude", "path", excludePath, "error", err)
		return
	}
	if err := appendIgnoreLine(excludePath, entry); err != nil {
		slog.Warn("could not update git info/exclude", "path", excludePath, "entry", entry, "error", err)
	}
}

// appendIgnoreLine appends entry as its own line to the ignore file at path,
// creating the file if needed. It does nothing if the line is already there.
func appendIgnoreLine(path, entry string) error {
	// Read existing content; a missing file is expected (created below).
	existing, readErr := os.ReadFile(path)
	if readErr != nil && !os.IsNotExist(readErr) {
		slog.Debug("could not read ignore file", "path", path, "error", readErr)
	}

	// Check if already present
	for _, line := range strings.Split(string(existing), "\n") {
		if strings.TrimSpace(line) == entry {
			return nil
		}
	}

	// Append
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}

	if len(existing) > 0 && existing[len(existing)-1] != '\n' {
		entry = "\n" + entry
	}
	_, writeErr := f.WriteString(entry + "\n")
	closeErr := f.Close()
	if writeErr != nil {
		return writeErr
	}
	return closeErr
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
