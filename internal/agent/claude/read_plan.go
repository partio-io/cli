package claude

import (
	"errors"
	"os"
	"path/filepath"
)

// ReadPlanFile reads the Claude Code plan file for the given slug.
// Returns ("", nil) if slug is empty or the file doesn't exist.
func ReadPlanFile(slug string) (string, error) {
	if slug == "" {
		return "", nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	path := filepath.Join(home, ".claude", "plans", slug+".md")
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", nil
		}
		return "", err
	}

	return string(data), nil
}
