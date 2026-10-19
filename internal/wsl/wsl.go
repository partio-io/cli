package wsl

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// IsWSL returns true when Partio is running inside a WSL environment with
// Windows interop enabled. Both conditions must hold:
//  1. /proc/sys/kernel/osrelease contains "microsoft" (case-insensitive).
//     This covers both WSL1 and WSL2.
//  2. /proc/sys/fs/binfmt_misc/WSLInterop exists. This file is present only
//     when Windows interop is enabled, which is required before calling any
//     Windows executables (tasklist.exe, cmd.exe, wslpath).
//
// A WSL environment with interop disabled is treated the same as plain Linux —
// IsWSL returns false so callers never attempt to invoke Windows executables.
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

// WindowsAppDataPath resolves the Windows %APPDATA% path as a WSL-accessible
// mount point. It runs cmd.exe to obtain the raw Windows path, then converts
// it to a Linux path via wslpath. Returns an error if either command fails.
// Callers should treat any error as "Windows path unavailable" and continue
// with Linux-only paths.
func WindowsAppDataPath() (string, error) {
	out, err := exec.Command("cmd.exe", "/C", "echo", "%APPDATA%").Output()
	if err != nil {
		return "", fmt.Errorf("cmd.exe failed: %w", err)
	}
	windowsPath := strings.TrimRight(string(out), "\r\n")
	if windowsPath == "" {
		return "", fmt.Errorf("cmd.exe returned empty APPDATA")
	}

	converted, err := exec.Command("wslpath", "-u", windowsPath).Output()
	if err != nil {
		return "", fmt.Errorf("wslpath failed: %w", err)
	}
	return strings.TrimRight(string(converted), "\r\n"), nil
}
