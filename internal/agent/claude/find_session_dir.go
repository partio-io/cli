package claude

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/partio-io/cli/internal/wsl"
)

// FindSessionDir returns the Claude Code session directory for the given repo.
// Claude stores sessions at ~/.claude/projects/<sanitized-path>/
// where the path is the cwd where Claude was launched, which may be the
// immediate parent of the git repo root (e.g. in monorepos or worktrees).
//
// When both the repo root and its parent have matching directories, the one
// containing the most recently modified JSONL file is returned.
//
// In WSL with Windows interop enabled, it also searches for session files
// under the Windows user's AppData Claude projects directory.
func (d *Detector) FindSessionDir(repoRoot string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("getting home directory: %w", err)
	}

	projectsDir := filepath.Join(home, ".claude", "projects")

	// Resolve the Windows AppData path once if running in WSL.
	var windowsProjectsDir string
	if wsl.IsWSL() {
		appData, err := wsl.WindowsAppDataPath()
		if err != nil {
			slog.Warn("wsl: cannot resolve Windows AppData path; skipping Windows session search", "err", err)
		} else {
			windowsProjectsDir = filepath.Join(appData, "Claude", "projects")
		}
	}

	// Check the repo root and its immediate parent directory.
	var candidates []string
	dirs := []string{repoRoot, filepath.Dir(repoRoot)}
	for _, dir := range dirs {
		sanitized := sanitizePath(dir)
		sessionDir := filepath.Join(projectsDir, sanitized)
		if _, err := os.Stat(sessionDir); err == nil {
			candidates = append(candidates, sessionDir)
		}

		// In WSL, also check the Windows projects directory.
		if windowsProjectsDir != "" {
			winPath, err := exec.Command("wslpath", "-w", dir).Output()
			if err != nil {
				slog.Warn("wsl: wslpath failed for directory; skipping Windows candidate", "dir", dir, "err", err)
				continue
			}
			windowsDir := strings.TrimRight(string(winPath), "\r\n")
			sanitizedWin := sanitizeWindowsPath(windowsDir)
			winSessionDir := filepath.Join(windowsProjectsDir, sanitizedWin)
			if _, err := os.Stat(winSessionDir); err == nil {
				candidates = append(candidates, winSessionDir)
			}
		}
	}

	if len(candidates) == 0 {
		return "", fmt.Errorf("no Claude session directory found for %s", repoRoot)
	}

	if len(candidates) == 1 {
		return candidates[0], nil
	}

	// Pick the candidate with the most recently modified JSONL file.
	var bestDir string
	var bestTime time.Time
	for _, c := range candidates {
		entries, err := os.ReadDir(c)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".jsonl") {
				continue
			}
			info, err := e.Info()
			if err != nil {
				continue
			}
			if info.ModTime().After(bestTime) {
				bestTime = info.ModTime()
				bestDir = c
			}
		}
	}

	if bestDir == "" {
		// No JSONL files found in any candidate — fall back to first (closest to repo root).
		return candidates[0], nil
	}

	return bestDir, nil
}

// sanitizeWindowsPath converts a Windows absolute path to Claude's sanitized
// format. It replaces backslashes, colons, and forward slashes with hyphens.
// e.g. C:\Users\foo\project -> C--Users-foo-project
func sanitizeWindowsPath(p string) string {
	p = strings.ReplaceAll(p, `\`, "-")
	p = strings.ReplaceAll(p, ":", "-")
	p = strings.ReplaceAll(p, "/", "-")
	return p
}
