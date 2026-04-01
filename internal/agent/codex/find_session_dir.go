package codex

import (
	"fmt"
	"os"
	"path/filepath"
)

// FindSessionDir returns the Codex CLI session directory for the given repo.
// Codex CLI stores sessions at ~/.codex/sessions/<sanitized-path>/.
func (d *Detector) FindSessionDir(repoRoot string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("getting home directory: %w", err)
	}

	sessionDir := filepath.Join(home, ".codex", "sessions", sanitizePath(repoRoot))
	if _, err := os.Stat(sessionDir); err != nil {
		return "", fmt.Errorf("no Codex session directory found for %s", repoRoot)
	}
	return sessionDir, nil
}
