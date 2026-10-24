package wsl

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsWSL_PlainLinux(t *testing.T) {
	// On a plain Linux CI runner, no WSL env vars are set and neither proc file
	// contains "microsoft" — IsWSL() must return false.
	t.Setenv("WSL_DISTRO_NAME", "")
	t.Setenv("WSL_INTEROP", "")
	t.Setenv("WSLENV", "")

	tmp := t.TempDir()
	procVersionPath = filepath.Join(tmp, "version")
	osReleasePath = filepath.Join(tmp, "osrelease")
	if err := os.WriteFile(procVersionPath, []byte("Linux version 6.1.0 (gcc)"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(osReleasePath, []byte("6.1.0-49-amd64"), 0o644); err != nil {
		t.Fatal(err)
	}
	resetCache()

	if IsWSL() {
		t.Error("expected IsWSL() = false on plain Linux, got true")
	}
}

func TestIsWSL_EnvVar(t *testing.T) {
	// Any one WSL env var should make IsWSL() return true.
	for _, env := range []string{"WSL_DISTRO_NAME", "WSL_INTEROP", "WSLENV"} {
		t.Run(env, func(t *testing.T) {
			t.Setenv("WSL_DISTRO_NAME", "")
			t.Setenv("WSL_INTEROP", "")
			t.Setenv("WSLENV", "")
			t.Setenv(env, "Ubuntu-22.04")

			tmp := t.TempDir()
			procVersionPath = filepath.Join(tmp, "version")
			osReleasePath = filepath.Join(tmp, "osrelease")
			if err := os.WriteFile(procVersionPath, []byte("Linux version 5.15.0"), 0o644); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(osReleasePath, []byte("5.15.0"), 0o644); err != nil {
				t.Fatal(err)
			}
			resetCache()

			if !IsWSL() {
				t.Errorf("expected IsWSL() = true when %s is set, got false", env)
			}
		})
	}
}

func TestIsWSL_ProcVersion(t *testing.T) {
	// IsWSL() returns true when /proc/version contains "microsoft".
	t.Setenv("WSL_DISTRO_NAME", "")
	t.Setenv("WSL_INTEROP", "")
	t.Setenv("WSLENV", "")

	tmp := t.TempDir()
	procVersionPath = filepath.Join(tmp, "version")
	osReleasePath = filepath.Join(tmp, "osrelease")
	if err := os.WriteFile(procVersionPath, []byte("Linux version 5.15.90.1-microsoft-standard-WSL2"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(osReleasePath, []byte("5.15.90.1"), 0o644); err != nil {
		t.Fatal(err)
	}
	resetCache()

	if !IsWSL() {
		t.Error("expected IsWSL() = true when procVersion contains 'microsoft', got false")
	}
}

func TestIsWSL_OSRelease(t *testing.T) {
	// IsWSL() returns true when osrelease contains "microsoft" but procVersion doesn't.
	t.Setenv("WSL_DISTRO_NAME", "")
	t.Setenv("WSL_INTEROP", "")
	t.Setenv("WSLENV", "")

	tmp := t.TempDir()
	procVersionPath = filepath.Join(tmp, "version")
	osReleasePath = filepath.Join(tmp, "osrelease")
	if err := os.WriteFile(procVersionPath, []byte("Linux version 5.15.0 (gcc)"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(osReleasePath, []byte("5.15.90.1-microsoft-standard-WSL2"), 0o644); err != nil {
		t.Fatal(err)
	}
	resetCache()

	if !IsWSL() {
		t.Error("expected IsWSL() = true when osrelease contains 'microsoft', got false")
	}
}

func TestWindowsHomeDir_RequiresWSL(t *testing.T) {
	t.Skip("requires WSL environment")
}
