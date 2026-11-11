package claude

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/partio-io/cli/internal/wsl"
)

// overrideWSL sets isWSLFunc to return the given value for the duration of the
// test, restoring the original on cleanup.
func overrideWSL(t *testing.T, inWSL bool) {
	t.Helper()
	orig := isWSLFunc
	t.Cleanup(func() { isWSLFunc = orig })
	isWSLFunc = func() bool { return inWSL }
}

func TestFindSessionDir(t *testing.T) {
	tests := []struct {
		name string
		// dirs maps sanitized directory names to JSONL filenames and their age
		// relative to "now". Positive means older (subtracted from now).
		dirs    map[string][]fileAge
		wantIdx int // index into dirs iteration isn't stable, so we use wantDir name
		wantDir string
		wantErr bool
	}{
		{
			name: "single candidate returns it",
			dirs: map[string][]fileAge{
				"child": {{name: "session.jsonl", age: time.Hour}},
			},
			wantDir: "child",
		},
		{
			name: "picks directory with freshest JSONL",
			dirs: map[string][]fileAge{
				"child":  {{name: "old.jsonl", age: 24 * time.Hour}},
				"parent": {{name: "fresh.jsonl", age: time.Minute}},
			},
			wantDir: "parent",
		},
		{
			name: "multiple files per directory — picks freshest overall",
			dirs: map[string][]fileAge{
				"child": {
					{name: "a.jsonl", age: 48 * time.Hour},
					{name: "b.jsonl", age: 2 * time.Hour},
				},
				"parent": {
					{name: "c.jsonl", age: 24 * time.Hour},
				},
			},
			wantDir: "child", // b.jsonl at 2h is fresher than c.jsonl at 24h
		},
		{
			name: "no JSONL files falls back to first candidate",
			dirs: map[string][]fileAge{
				"child":  {},
				"parent": {},
			},
			wantDir: "child",
		},
		{
			name:    "no candidates returns error",
			dirs:    map[string][]fileAge{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Build a fake filesystem:
			//   tmpDir/home/.claude/projects/<dir>/  — session directories
			//   tmpDir/repo/child/                   — fake repo root
			//   tmpDir/repo/                         — parent directory
			tmpDir := t.TempDir()
			home := filepath.Join(tmpDir, "home")
			projectsDir := filepath.Join(home, ".claude", "projects")

			// We simulate two path levels: tmpDir/repo/child (repoRoot)
			// and tmpDir/repo (parent). Their sanitized forms are computed
			// by sanitizePath, so we create directories matching those.
			repoParent := filepath.Join(tmpDir, "repo")
			repoRoot := filepath.Join(repoParent, "child")
			if err := os.MkdirAll(repoRoot, 0o755); err != nil {
				t.Fatal(err)
			}

			// Map logical names to actual paths
			pathMap := map[string]string{
				"child":  sanitizePath(repoRoot),
				"parent": sanitizePath(repoParent),
			}

			now := time.Now()
			for dirName, files := range tt.dirs {
				sanitized, ok := pathMap[dirName]
				if !ok {
					t.Fatalf("unknown dir name %q", dirName)
				}
				sessionDir := filepath.Join(projectsDir, sanitized)
				if err := os.MkdirAll(sessionDir, 0o755); err != nil {
					t.Fatal(err)
				}
				for _, f := range files {
					p := filepath.Join(sessionDir, f.name)
					if err := os.WriteFile(p, []byte("{}"), 0o644); err != nil {
						t.Fatal(err)
					}
					modTime := now.Add(-f.age)
					if err := os.Chtimes(p, modTime, modTime); err != nil {
						t.Fatal(err)
					}
				}
			}

			// Override HOME so FindSessionDir looks in our temp directory.
			t.Setenv("HOME", home)

			d := &Detector{}
			got, err := d.FindSessionDir(repoRoot)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			wantSanitized := pathMap[tt.wantDir]
			want := filepath.Join(projectsDir, wantSanitized)
			if got != want {
				t.Errorf("got %s, want %s", got, want)
			}
		})
	}
}

type fileAge struct {
	name string
	age  time.Duration
}

// TestFindSessionDir_WindowsCandidate tests that FindSessionDir finds a Windows
// filesystem candidate when running in a simulated WSL environment. It overrides
// windowsHomeDirFunc and wslpathToWindowsFunc to avoid real WSL dependencies.
func TestFindSessionDir_WindowsCandidate(t *testing.T) {
	overrideWSL(t, true)

	// The wsl package caches its result; we need a fresh detection.
	// We use the internal wsl_test resetCache indirectly by ensuring the test
	// runs in a subprocess-like context. Since we cannot call wsl.resetCache()
	// from an external package, we skip if IsWSL() already returned false
	// (meaning the cache was populated before our t.Setenv took effect).
	//
	// In practice, go test runs each test file's tests sequentially and
	// t.Setenv takes effect before IsWSL() is called below.

	tmpDir := t.TempDir()

	// Fake Linux home (no Linux candidates exist).
	linuxHome := filepath.Join(tmpDir, "linuxhome")
	if err := os.MkdirAll(filepath.Join(linuxHome, ".claude", "projects"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", linuxHome)

	// Fake Windows home dir (returned by our injected windowsHomeDirFunc).
	winHome := filepath.Join(tmpDir, "winhome")

	// Fake repo root.
	repoRoot := filepath.Join(tmpDir, "repo", "child")
	if err := os.MkdirAll(repoRoot, 0o755); err != nil {
		t.Fatal(err)
	}

	// The injected wslpath -w conversion returns a fake Windows path.
	fakeWinPath := `C:\fake\repo\child`
	expectedSanitized := sanitizePath(fakeWinPath) // "C--fake-repo-child"

	// Create the matching Windows session directory with a JSONL file.
	winSessionDir := filepath.Join(winHome, ".claude", "projects", expectedSanitized)
	if err := os.MkdirAll(winSessionDir, 0o755); err != nil {
		t.Fatal(err)
	}
	jsonlPath := filepath.Join(winSessionDir, "session.jsonl")
	if err := os.WriteFile(jsonlPath, []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}

	origWinHomeFunc := windowsHomeDirFunc
	origWslpathFunc := wslpathToWindowsFunc
	t.Cleanup(func() {
		windowsHomeDirFunc = origWinHomeFunc
		wslpathToWindowsFunc = origWslpathFunc
	})

	windowsHomeDirFunc = func() (string, error) { return winHome, nil }
	wslpathToWindowsFunc = func(linuxPath string) (string, error) {
		if linuxPath == repoRoot {
			return fakeWinPath, nil
		}
		// Parent: return a different path that has no matching directory.
		return `C:\fake\repo`, nil
	}

	d := &Detector{}
	got, err := d.FindSessionDir(repoRoot)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != winSessionDir {
		t.Errorf("got %s, want %s", got, winSessionDir)
	}
}

// TestFindSessionDir_WindowsCandidateFresher tests that the Windows candidate
// is preferred over a Linux candidate when its JSONL file is newer.
func TestFindSessionDir_WindowsCandidateFresher(t *testing.T) {
	overrideWSL(t, true)

	tmpDir := t.TempDir()

	// Fake Linux home with a Linux candidate (older JSONL).
	linuxHome := filepath.Join(tmpDir, "linuxhome")
	repoRoot := filepath.Join(tmpDir, "repo", "child")
	if err := os.MkdirAll(repoRoot, 0o755); err != nil {
		t.Fatal(err)
	}

	linuxSessionDir := filepath.Join(linuxHome, ".claude", "projects", sanitizePath(repoRoot))
	if err := os.MkdirAll(linuxSessionDir, 0o755); err != nil {
		t.Fatal(err)
	}
	oldFile := filepath.Join(linuxSessionDir, "old.jsonl")
	if err := os.WriteFile(oldFile, []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	oldTime := time.Now().Add(-24 * time.Hour)
	if err := os.Chtimes(oldFile, oldTime, oldTime); err != nil {
		t.Fatal(err)
	}

	t.Setenv("HOME", linuxHome)

	// Fake Windows home with a Windows candidate (newer JSONL).
	winHome := filepath.Join(tmpDir, "winhome")
	fakeWinPath := `C:\fake\repo\child`
	winSessionDir := filepath.Join(winHome, ".claude", "projects", sanitizePath(fakeWinPath))
	if err := os.MkdirAll(winSessionDir, 0o755); err != nil {
		t.Fatal(err)
	}
	newFile := filepath.Join(winSessionDir, "fresh.jsonl")
	if err := os.WriteFile(newFile, []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	// Keep fresh file at current time (default).

	origWinHomeFunc := windowsHomeDirFunc
	origWslpathFunc := wslpathToWindowsFunc
	t.Cleanup(func() {
		windowsHomeDirFunc = origWinHomeFunc
		wslpathToWindowsFunc = origWslpathFunc
	})

	windowsHomeDirFunc = func() (string, error) { return winHome, nil }
	wslpathToWindowsFunc = func(linuxPath string) (string, error) {
		if linuxPath == repoRoot {
			return fakeWinPath, nil
		}
		return `C:\fake\repo`, nil
	}

	d := &Detector{}
	got, err := d.FindSessionDir(repoRoot)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != winSessionDir {
		t.Errorf("got %s (want Windows candidate %s)", got, winSessionDir)
	}
}

// TestFindSessionDir_NoWSLSearch tests that Windows paths are not searched
// when not running in WSL.
func TestFindSessionDir_NoWSLSearch(t *testing.T) {
	overrideWSL(t, false)

	tmpDir := t.TempDir()
	repoRoot := filepath.Join(tmpDir, "repo", "child")
	if err := os.MkdirAll(repoRoot, 0o755); err != nil {
		t.Fatal(err)
	}

	// Linux home with a Linux candidate.
	linuxHome := filepath.Join(tmpDir, "home")
	linuxSessionDir := filepath.Join(linuxHome, ".claude", "projects", sanitizePath(repoRoot))
	if err := os.MkdirAll(linuxSessionDir, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", linuxHome)

	called := false
	origWinHomeFunc := windowsHomeDirFunc
	t.Cleanup(func() { windowsHomeDirFunc = origWinHomeFunc })
	windowsHomeDirFunc = func() (string, error) {
		called = true
		return "", nil
	}

	d := &Detector{}
	got, err := d.FindSessionDir(repoRoot)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != linuxSessionDir {
		t.Errorf("got %s, want %s", got, linuxSessionDir)
	}
	if called {
		t.Error("windowsHomeDirFunc was called when not in WSL")
	}
}

// TestWindowsHomeDir_EnvOverrideLinuxPath tests that PARTIO_WSL_WIN_HOME with a
// Linux path is returned as-is without calling wslpath.
func TestWindowsHomeDir_EnvOverrideLinuxPath(t *testing.T) {
	t.Setenv("PARTIO_WSL_WIN_HOME", "/mnt/c/Users/testuser")
	got, err := wsl.WindowsHomeDir()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "/mnt/c/Users/testuser" {
		t.Errorf("got %q, want %q", got, "/mnt/c/Users/testuser")
	}
}

// TestWindowsHomeDir_CmdExeUnavailable tests that WindowsHomeDir returns an
// error (not a panic) when cmd.exe is unavailable on plain Linux.
func TestWindowsHomeDir_CmdExeUnavailable(t *testing.T) {
	t.Setenv("PARTIO_WSL_WIN_HOME", "")
	_, err := wsl.WindowsHomeDir()
	if err == nil {
		t.Skip("cmd.exe appears to be available — running in a real WSL environment")
	}
	// Error is expected on plain Linux — just verify we got one without panicking.
}

// TestWindowsHomeDir_Integration tests WindowsHomeDir on a real WSL environment.
func TestWindowsHomeDir_Integration(t *testing.T) {
	if os.Getenv("WSL_DISTRO_NAME") == "" && os.Getenv("WSL_INTEROP") == "" && os.Getenv("WSLENV") == "" {
		t.Skip("requires WSL environment")
	}
	t.Setenv("PARTIO_WSL_WIN_HOME", "")
	got, err := wsl.WindowsHomeDir()
	if err != nil {
		t.Fatalf("WindowsHomeDir() error: %v", err)
	}
	if got == "" {
		t.Error("WindowsHomeDir() returned empty string")
	}
	if !filepath.IsAbs(got) {
		t.Errorf("WindowsHomeDir() = %q, want absolute path", got)
	}
}
