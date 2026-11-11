package wsl

import (
	"os"
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
