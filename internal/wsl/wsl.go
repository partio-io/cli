// Package wsl provides WSL (Windows Subsystem for Linux) detection and
// Windows path resolution helpers for use by agent detection and session
// discovery code.
package wsl

import (
	"os"
	"os/exec"
	"strings"
)

// IsWSL returns true if Partio is running inside a WSL environment with
// Windows interop enabled. Both conditions must hold:
//  1. /proc/sys/kernel/osrelease contains "microsoft" (case-insensitive) —
//     present on both WSL1 and WSL2.
//  2. /proc/sys/fs/binfmt_misc/WSLInterop exists — present on WSL2 with
//     interop enabled; its absence means Windows executables cannot be called.
func IsWSL() bool {
	data, err := os.ReadFile("/proc/sys/kernel/osrelease")
	if err != nil {
		return false
	}
	if !strings.Contains(strings.ToLower(string(data)), "microsoft") {
		return false
	}
	_, err = os.Stat("/proc/sys/fs/binfmt_misc/WSLInterop")
	return err == nil
}

// sanitizeWindowsPath converts a Windows-format path to the dash-separated
// subdirectory name that Claude Code uses when keying sessions. It replaces
// backslashes, colons, and forward slashes with dashes, matching the
// sanitizePath logic used in the Linux session discovery code.
func sanitizeWindowsPath(winPath string) string {
	r := strings.NewReplacer(
		`\`, "-",
		`:`, "-",
		`/`, "-",
	)
	return r.Replace(winPath)
}

// WindowsAppDataPath returns the Windows %APPDATA% directory as a
// WSL-accessible mount path (e.g. /mnt/c/Users/username/AppData/Roaming).
//
// It runs cmd.exe to read %APPDATA%, strips the trailing carriage return and
// newline, then converts the Windows path to a Linux path via wslpath.
// Returns an error if either command fails; callers should treat any error as
// "Windows AppData path unavailable" and continue with Linux-only paths.
func WindowsAppDataPath() (string, error) {
	out, err := exec.Command("cmd.exe", "/C", "echo", "%APPDATA%").Output()
	if err != nil {
		return "", err
	}
	winPath := strings.TrimRight(string(out), "\r\n")

	linuxPath, err := exec.Command("wslpath", "-u", winPath).Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(linuxPath)), nil
}
