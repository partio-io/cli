package claude

import (
	"fmt"
	"os"
	"path/filepath"
)

// FindSessionDir returns the Claude Code session directory for the given repo.
// Claude stores sessions at ~/.claude/projects/<sanitized-path>/
// where the path is the cwd where Claude was launched, which may be a parent
// of the git repo root (e.g. in monorepos or worktrees).
func (d *Detector) FindSessionDir(repoRoot string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("getting home directory: %w", err)
	}

	projectsDir := filepath.Join(home, ".claude", "projects")

	// Try the repo root first, then walk up to parent directories.
	// Claude Code keys sessions to the cwd where it was launched, which
	// may be a parent of the git repo root.
	dir := repoRoot
	for {
		sanitized := sanitizePath(dir)
		sessionDir := filepath.Join(projectsDir, sanitized)
		if _, err := os.Stat(sessionDir); err == nil {
			return sessionDir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return "", fmt.Errorf("no Claude session directory found for %s", repoRoot)
}
