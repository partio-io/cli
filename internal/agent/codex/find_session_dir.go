package codex

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/partio-io/cli/internal/agent"
)

// FindSessionDir returns the Codex CLI session directory.
func (d *Detector) FindSessionDir(repoRoot string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("getting home directory: %w", err)
	}

	sessionDir := filepath.Join(home, ".codex", "sessions")
	if _, err := os.Stat(sessionDir); err != nil {
		return "", fmt.Errorf("no Codex session directory found: %w", err)
	}

	return sessionDir, nil
}

// FindLatestSession returns the most recent Codex session file for the given repo.
// It matches sessions by comparing the session_meta cwd against the repo root.
func (d *Detector) FindLatestSession(repoRoot string) (string, *agent.SessionData, error) {
	sessionDir, err := d.FindSessionDir(repoRoot)
	if err != nil {
		return "", nil, err
	}

	// Resolve symlinks for comparison
	absRepoRoot, err := filepath.EvalSymlinks(repoRoot)
	if err != nil {
		absRepoRoot = repoRoot
	}
	absRepoRoot, _ = filepath.Abs(absRepoRoot)

	// Collect all .jsonl files
	var files []string
	err = filepath.Walk(sessionDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // skip errors
		}
		if !info.IsDir() && strings.HasSuffix(path, ".jsonl") {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return "", nil, fmt.Errorf("walking session directory: %w", err)
	}

	// Sort descending by name (filenames contain timestamps)
	sort.Sort(sort.Reverse(sort.StringSlice(files)))

	// Find the most recent session that matches this repo
	for _, path := range files {
		cwd := PeekCWD(path)
		if cwd == "" {
			continue
		}

		absCWD, err := filepath.EvalSymlinks(cwd)
		if err != nil {
			absCWD = cwd
		}
		absCWD, _ = filepath.Abs(absCWD)

		// Check if the session cwd matches or is a parent of the repo root
		if absCWD == absRepoRoot || strings.HasPrefix(absRepoRoot, absCWD+"/") {
			data, parseErr := ParseJSONL(path)
			if parseErr != nil {
				continue
			}
			return path, data, nil
		}
	}

	return "", nil, fmt.Errorf("no Codex session found for repo %s", repoRoot)
}
