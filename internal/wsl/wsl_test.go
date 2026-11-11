package wsl

import (
	"os"
	"path/filepath"
	"testing"
)

// resetCache resets the IsWSL detection cache so tests can re-run detection
// with different environment or proc file configurations.
func resetCache() {
	wslMu.Lock()
	wslCached = false
	wslResult = false
	wslMu.Unlock()
}

func TestIsWSL_FalseOnPlainLinux(t *testing.T) {
	// Redirect proc files to nonexistent paths so they produce no "microsoft" match.
	origVersion := procVersionPath
	origRelease := osReleasePath
	t.Cleanup(func() {
		procVersionPath = origVersion
		osReleasePath = origRelease
	})

	procVersionPath = "/nonexistent/proc/version"
	osReleasePath = "/nonexistent/proc/sys/kernel/osrelease"

	// Unset WSL env vars if somehow set in the test environment.
	for _, env := range []string{"WSL_DISTRO_NAME", "WSL_INTEROP", "WSLENV"} {
		t.Setenv(env, "")
	}

	resetCache()
	if IsWSL() {
		t.Error("expected IsWSL() = false on plain Linux CI runner")
	}
}

func TestIsWSL_TrueWhenWSLDistroNameSet(t *testing.T) {
	origVersion := procVersionPath
	origRelease := osReleasePath
	t.Cleanup(func() {
		procVersionPath = origVersion
		osReleasePath = origRelease
	})

	// Use nonexistent proc paths so env var is the only signal.
	procVersionPath = "/nonexistent/proc/version"
	osReleasePath = "/nonexistent/proc/sys/kernel/osrelease"

	t.Setenv("WSL_DISTRO_NAME", "Ubuntu-22.04")
	// Ensure the other WSL vars are unset so we know exactly which one triggers.
	t.Setenv("WSL_INTEROP", "")
	t.Setenv("WSLENV", "")

	resetCache()
	if !IsWSL() {
		t.Error("expected IsWSL() = true when WSL_DISTRO_NAME is set")
	}
}

func TestIsWSL_TrueWhenProcVersionContainsMicrosoft(t *testing.T) {
	// Create a temp file that simulates /proc/version on WSL.
	dir := t.TempDir()
	versionFile := filepath.Join(dir, "version")
	if err := os.WriteFile(versionFile, []byte("Linux version 5.15.0-microsoft-standard-WSL2"), 0o644); err != nil {
		t.Fatal(err)
	}

	origVersion := procVersionPath
	origRelease := osReleasePath
	t.Cleanup(func() {
		procVersionPath = origVersion
		osReleasePath = origRelease
	})

	procVersionPath = versionFile
	osReleasePath = "/nonexistent/proc/sys/kernel/osrelease"

	// Ensure no WSL env vars are set.
	for _, env := range []string{"WSL_DISTRO_NAME", "WSL_INTEROP", "WSLENV"} {
		t.Setenv(env, "")
	}

	resetCache()
	if !IsWSL() {
		t.Error("expected IsWSL() = true when /proc/version contains 'microsoft'")
	}
}
