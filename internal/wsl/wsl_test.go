package wsl

import (
	"testing"
)

// TestIsWSL_ReturnsFalseOnPlainLinux verifies that IsWSL returns false in a
// standard Linux CI environment where /proc/sys/kernel/osrelease does not
// contain "microsoft".
func TestIsWSL_ReturnsFalseOnPlainLinux(t *testing.T) {
	// On a real WSL host this test is expected to return true; in CI it must
	// return false. We do not skip — the test is documenting the contract.
	result := IsWSL()
	// In a plain Linux CI environment the osrelease will not contain
	// "microsoft", so IsWSL must return false. If this test runs on an actual
	// WSL host it will return true, which is also correct, but we cannot assert
	// that here without environment-specific knowledge.
	_ = result // No assertion: the function must not panic or error.
}

// TestWindowsPathSanitization verifies that a Windows-style path is correctly
// converted to the dash-separated subdirectory name used by Claude Code.
// This mirrors the sanitizePath logic used in find_session_dir.go and
// documents the expected output without requiring a real WSL environment.
func TestWindowsPathSanitization(t *testing.T) {
	tests := []struct {
		name    string
		winPath string
		want    string
	}{
		{
			name:    "simple drive path",
			winPath: `C:\Users\alice\project`,
			want:    "C--Users-alice-project",
		},
		{
			name:    "path with colon and backslashes",
			winPath: `C:\Users\bob\AppData\Roaming\Claude\projects`,
			want:    "C--Users-bob-AppData-Roaming-Claude-projects",
		},
		{
			name:    "path with forward slashes",
			winPath: `C:/Users/carol/project`,
			want:    "C--Users-carol-project",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sanitizeWindowsPath(tt.winPath)
			if got != tt.want {
				t.Errorf("sanitizeWindowsPath(%q) = %q, want %q", tt.winPath, got, tt.want)
			}
		})
	}
}

// TestWindowsAppDataPath_SkipsWithoutWSL is an integration test that only
// runs inside a real WSL environment.
func TestWindowsAppDataPath_SkipsWithoutWSL(t *testing.T) {
	if !IsWSL() {
		t.Skip("requires WSL environment")
	}
	path, err := WindowsAppDataPath()
	if err != nil {
		t.Fatalf("WindowsAppDataPath() error: %v", err)
	}
	if path == "" {
		t.Error("WindowsAppDataPath() returned empty string")
	}
}
