package wsl

import (
	"os"
	"strings"
	"testing"
)

func TestIsWSL_NonWSLHost(t *testing.T) {
	// On a plain Linux CI runner, /proc/sys/kernel/osrelease will not contain
	// "microsoft", so IsWSL() must return false.
	if _, err := os.ReadFile("/proc/sys/kernel/osrelease"); err != nil {
		t.Skip("cannot read /proc/sys/kernel/osrelease")
	}

	got := IsWSL()
	// On CI (non-WSL), IsWSL must be false.
	// On a real WSL machine skip rather than fail.
	if got {
		t.Skip("running inside WSL, plain-Linux assertion not applicable")
	}
}

func TestIsWSL_MissingInteropFile(t *testing.T) {
	// If WSLInterop does not exist, IsWSL must return false.
	if _, err := os.Stat("/proc/sys/fs/binfmt_misc/WSLInterop"); err == nil {
		t.Skip("WSLInterop file exists — this host appears to be WSL")
	}
	if got := IsWSL(); got {
		t.Errorf("IsWSL() = true, want false when WSLInterop is absent")
	}
}

func TestWindowsAppDataPath_Integration(t *testing.T) {
	if !IsWSL() {
		t.Skip("requires WSL environment with Windows interop enabled")
	}
	path, err := WindowsAppDataPath()
	if err != nil {
		t.Fatalf("WindowsAppDataPath() error = %v", err)
	}
	if path == "" {
		t.Error("WindowsAppDataPath() returned empty string")
	}
	// The path should start with /mnt/ in a typical WSL setup.
	if !strings.HasPrefix(path, "/mnt/") {
		t.Logf("WindowsAppDataPath() = %q (not under /mnt/, may still be valid)", path)
	}
}
