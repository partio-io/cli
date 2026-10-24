package wsl

import (
	"os"
	"strings"
	"sync"
)

// procVersionPath and osReleasePath are package-level variables so tests can
// redirect them to temporary files without a real WSL environment.
var procVersionPath = "/proc/version"
var osReleasePath = "/proc/sys/kernel/osrelease"

var (
	once   sync.Once
	cached bool
)

// resetCache resets the sync.Once so the next call to IsWSL() re-detects.
// Only intended for use in tests.
func resetCache() {
	once = sync.Once{}
}

// IsWSL returns true if Partio is running inside a WSL environment with
// Windows interop active. The result is cached for the process lifetime.
//
// Detection uses a two-tier approach:
//  1. Fast path: check WSL env vars (WSL_DISTRO_NAME, WSL_INTEROP, WSLENV).
//  2. Fallback: read /proc/version then /proc/sys/kernel/osrelease for "microsoft".
func IsWSL() bool {
	once.Do(func() {
		cached = detect()
	})
	return cached
}

func detect() bool {
	// Fast path: WSL runtime sets these env vars when interop is active.
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
