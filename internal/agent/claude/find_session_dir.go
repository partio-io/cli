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

// isWSLFunc reports whether the process is running in WSL.
// Overridable for testing without a real WSL environment.
var isWSLFunc = wsl.IsWSL

// windowsHomeDirFunc resolves the Windows user home directory when running in
// WSL. Overridable for testing without a real WSL environment.
var windowsHomeDirFunc = wsl.WindowsHomeDir

// wslpathToWindowsFunc converts a Linux path to a Windows-format path using
// wslpath -w. Overridable for testing without a real WSL environment.
var wslpathToWindowsFunc = func(linuxPath string) (string, error) {
	out, err := exec.Command("wslpath", "-w", linuxPath).Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// FindSessionDir returns the Claude Code session directory for the given repo.
// Claude stores sessions at ~/.claude/projects/<sanitized-path>/
// where the path is the cwd where Claude was launched, which may be the
// immediate parent of the git repo root (e.g. in monorepos or worktrees).
//
// When both the repo root and its parent have matching directories, the one
// containing the most recently modified JSONL file is returned.
//
// In WSL, Windows-filesystem candidates under %USERPROFILE%\.claude\projects\
// are also searched and included in the freshness comparison.
func (d *Detector) FindSessionDir(repoRoot string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("getting home directory: %w", err)
	}

	projectsDir := filepath.Join(home, ".claude", "projects")

	// Check the repo root and its immediate parent directory.
	var candidates []string
	dirs := []string{repoRoot, filepath.Dir(repoRoot)}
	for _, dir := range dirs {
		sanitized := sanitizePath(dir)
		sessionDir := filepath.Join(projectsDir, sanitized)
		if _, err := os.Stat(sessionDir); err == nil {
			candidates = append(candidates, sessionDir)
		}
	}

	// In WSL, also search the Windows user profile path.
	if isWSLFunc() {
		winHome, err := windowsHomeDirFunc()
		if err != nil {
			slog.Warn("failed to get Windows home directory for session search", "err", err)
		} else {
			winProjectsDir := filepath.Join(winHome, ".claude", "projects")
			for _, dir := range dirs {
				winPath, err := wslpathToWindowsFunc(dir)
				if err != nil {
					slog.Warn("wslpath -w failed", "dir", dir, "err", err)
					continue
				}
				sanitized := sanitizePath(winPath)
				sessionDir := filepath.Join(winProjectsDir, sanitized)
				if _, err := os.Stat(sessionDir); err == nil {
					candidates = append(candidates, sessionDir)
				}
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
