package codex

import (
	"fmt"
	"os"
	"path/filepath"
)

// FindSessionDir returns the Codex CLI session directory for the given repo.
// Codex stores its data at ~/.codex/.
func (d *Detector) FindSessionDir(repoRoot string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("getting home directory: %w", err)
	}

	sessionDir := filepath.Join(home, ".codex")
	if _, err := os.Stat(sessionDir); err != nil {
		return "", fmt.Errorf("no Codex session directory found: %w", err)
	}

	return sessionDir, nil
}
