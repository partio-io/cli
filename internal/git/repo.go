package git

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// RepoRoot returns the top-level directory of the current git repository.
func RepoRoot() (string, error) {
	out, err := execGit("rev-parse", "--show-toplevel")
	if err != nil {
		return "", fmt.Errorf("not a git repository: %w", err)
	}
	return out, nil
}

// HooksDir returns the path to the git hooks directory, handling worktrees.
// In worktrees, hooks live in the common git dir (shared across worktrees),
// not the per-worktree git dir.
// dir specifies the working directory for the git command; if empty, the
// current working directory is used.
func HooksDir(dir string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "--git-common-dir")
	if dir != "" {
		cmd.Dir = dir
	}
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("resolving git common directory: %w", err)
	}
	gitDir := strings.TrimSpace(string(out))
	if !filepath.IsAbs(gitDir) {
		if dir != "" {
			gitDir = filepath.Join(dir, gitDir)
		} else {
			root, err := RepoRoot()
			if err != nil {
				return "", err
			}
			gitDir = filepath.Join(root, gitDir)
		}
	}
	return filepath.Join(gitDir, "hooks"), nil
}

// IsRepo returns true if the current directory is inside a git repository.
func IsRepo() bool {
	_, err := RepoRoot()
	return err == nil
}
