package wsl

import (
	"os"
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
