package wsl

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
)

// procVersionPath and osReleasePath are package-level variables so tests can
// redirect reads to temp files without a real WSL environment.
var (
	procVersionPath = "/proc/version"
	osReleasePath   = "/proc/sys/kernel/osrelease"
)

var (
	wslMu     sync.Mutex
	wslCached bool
	wslResult bool
)

// IsWSL reports whether the current process is running inside Windows Subsystem
// for Linux (WSL). The result is cached after the first call.
//
// Detection order:
//  1. Check WSL environment variables (WSL_DISTRO_NAME, WSL_INTEROP, WSLENV).
//     Any one present means WSL with Windows interop active.
//  2. Read /proc/version for the substring "microsoft" (case-insensitive).
//  3. Read /proc/sys/kernel/osrelease for the substring "microsoft" (case-insensitive).
func IsWSL() bool {
	wslMu.Lock()
	defer wslMu.Unlock()
	if !wslCached {
		wslResult = detectWSL()
		wslCached = true
	}
	return wslResult
}

// WindowsHomeDir returns the Windows user profile directory translated to a WSL
// Linux mount path (e.g. /mnt/c/Users/username). It first checks the
// PARTIO_WSL_WIN_HOME environment variable — if set, it is returned directly
// if it is already a Linux path, or converted via wslpath -u otherwise.
// Otherwise it runs cmd.exe /C echo %USERPROFILE% and translates the result
// with wslpath -u. Returns an error if the Windows directory cannot be resolved;
// callers should treat this as "Windows directory unavailable" and fall back to
// Linux-only logic.
func WindowsHomeDir() (string, error) {
	if override := os.Getenv("PARTIO_WSL_WIN_HOME"); override != "" {
		if strings.HasPrefix(override, "/") {
			return override, nil
		}
		out, err := exec.Command("wslpath", "-u", override).Output()
		if err != nil {
			return "", fmt.Errorf("wslpath -u %q: %w", override, err)
		}
		return strings.TrimSpace(string(out)), nil
	}

	out, err := exec.Command("cmd.exe", "/C", "echo", "%USERPROFILE%").Output()
	if err != nil {
		return "", fmt.Errorf("cmd.exe echo %%USERPROFILE%%: %w", err)
	}
	winPath := strings.TrimSpace(string(out))

	linuxPath, err := exec.Command("wslpath", "-u", winPath).Output()
	if err != nil {
		return "", fmt.Errorf("wslpath -u %q: %w", winPath, err)
	}
	return strings.TrimSpace(string(linuxPath)), nil
}

func detectWSL() bool {
	// Fast path: WSL sets one of these environment variables when interop is active.
	for _, env := range []string{"WSL_DISTRO_NAME", "WSL_INTEROP", "WSLENV"} {
		if os.Getenv(env) != "" {
			return true
		}
	}

	// Fallback: check proc files for "microsoft" substring.
	for _, path := range []string{procVersionPath, osReleasePath} {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		if strings.Contains(strings.ToLower(string(data)), "microsoft") {
			return true
		}
	}

	return false
}
