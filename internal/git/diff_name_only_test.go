package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// setupRepo initialises a bare git repo in dir with user config set.
func setupRepo(t *testing.T, dir string) func(args ...string) string {
	t.Helper()
	run := func(args ...string) string {
		t.Helper()
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = dir
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("command %v failed: %v\n%s", args, err, out)
		}
		return strings.TrimSpace(string(out))
	}
	run("git", "init")
	run("git", "config", "user.email", "test@example.com")
	run("git", "config", "user.name", "Test")
	return run
}

func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestDiffNameOnly(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(dir string, run func(...string) string) string // returns commit hash
		wantFiles []string
	}{
		{
			name: "initial commit lists all added files",
			setup: func(dir string, run func(...string) string) string {
				writeFile(t, dir, "a.go", "package a\n")
				writeFile(t, dir, "b.go", "package b\n")
				run("git", "add", ".")
				run("git", "commit", "-m", "initial")
				return run("git", "rev-parse", "HEAD")
			},
			wantFiles: []string{"a.go", "b.go"},
		},
		{
			name: "regular commit lists only changed files",
			setup: func(dir string, run func(...string) string) string {
				writeFile(t, dir, "a.go", "package a\n")
				writeFile(t, dir, "unchanged.go", "package u\n")
				run("git", "add", ".")
				run("git", "commit", "-m", "initial")
				writeFile(t, dir, "b.go", "package b\n")
				writeFile(t, dir, "a.go", "package a // updated\n")
				run("git", "add", ".")
				run("git", "commit", "-m", "add b, update a")
				return run("git", "rev-parse", "HEAD")
			},
			wantFiles: []string{"a.go", "b.go"},
		},
		{
			name: "merge commit with conflict resolution lists resolved file",
			setup: func(dir string, run func(...string) string) string {
				// Both branches modify shared.go differently, so the merge
				// resolution produces a file that differs from both parents.
				writeFile(t, dir, "shared.go", "package p\n// v0\n")
				run("git", "add", ".")
				run("git", "commit", "-m", "base")

				run("git", "checkout", "-b", "branch-a")
				writeFile(t, dir, "shared.go", "package p\n// branch-a\n")
				run("git", "add", ".")
				run("git", "commit", "-m", "branch a edits shared")

				run("git", "checkout", "-")
				writeFile(t, dir, "shared.go", "package p\n// trunk\n")
				run("git", "add", ".")
				run("git", "commit", "-m", "trunk edits shared")

				// Resolve conflict manually and commit the merge.
				cmd := exec.Command("git", "merge", "--no-ff", "branch-a", "--no-commit")
				cmd.Dir = dir
				_ = cmd.Run() // conflict expected
				writeFile(t, dir, "shared.go", "package p\n// merged\n")
				run("git", "add", "shared.go")
				run("git", "commit", "-m", "merge with resolution")
				return run("git", "rev-parse", "HEAD")
			},
			wantFiles: []string{"shared.go"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			run := setupRepo(t, dir)
			hash := tt.setup(dir, run)

			// Temporarily change to the repo dir so execGit works.
			orig, _ := os.Getwd()
			if err := os.Chdir(dir); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = os.Chdir(orig) })

			got, err := DiffNameOnly(hash)
			if err != nil {
				t.Fatalf("DiffNameOnly() error = %v", err)
			}

			sort.Strings(got)
			sort.Strings(tt.wantFiles)

			if strings.Join(got, ",") != strings.Join(tt.wantFiles, ",") {
				t.Errorf("DiffNameOnly() = %v, want %v", got, tt.wantFiles)
			}
		})
	}
}

func BenchmarkDiffNameOnly(b *testing.B) {
	dir := b.TempDir()

	run := func(args ...string) string {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = dir
		out, err := cmd.CombinedOutput()
		if err != nil {
			b.Fatalf("command %v failed: %v\n%s", args, err, out)
		}
		return strings.TrimSpace(string(out))
	}

	run("git", "init")
	run("git", "config", "user.email", "bench@example.com")
	run("git", "config", "user.name", "Bench")

	// Commit 1000+ files to establish the baseline tree size.
	const numFiles = 1000
	for i := 0; i < numFiles; i++ {
		path := filepath.Join(dir, fmt.Sprintf("file%04d.go", i))
		if err := os.WriteFile(path, []byte(fmt.Sprintf("package p%d\n", i)), 0o644); err != nil {
			b.Fatal(err)
		}
	}
	run("git", "add", ".")
	run("git", "commit", "-m", "initial: add 1000 files")

	// Second commit changes only a single file — this is the one we benchmark.
	if err := os.WriteFile(filepath.Join(dir, "file0001.go"), []byte("package changed\n"), 0o644); err != nil {
		b.Fatal(err)
	}
	run("git", "add", "file0001.go")
	run("git", "commit", "-m", "update one file")
	hash := run("git", "rev-parse", "HEAD")

	orig, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		b.Fatal(err)
	}
	b.Cleanup(func() { _ = os.Chdir(orig) })

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		files, err := DiffNameOnly(hash)
		if err != nil {
			b.Fatalf("DiffNameOnly() error = %v", err)
		}
		if len(files) != 1 {
			b.Fatalf("expected 1 file, got %d", len(files))
		}
	}
}
