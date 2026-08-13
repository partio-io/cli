package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func setupTestRepo(t *testing.T) (dir string, run func(args ...string) string) {
	t.Helper()
	dir = t.TempDir()

	run = func(args ...string) string {
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

	return dir, run
}

func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
		t.Fatalf("writeFile(%s): %v", name, err)
	}
}

func TestDiffTree(t *testing.T) {
	t.Run("initial commit", func(t *testing.T) {
		dir, run := setupTestRepo(t)

		writeFile(t, dir, "hello.txt", "hello\n")
		run("git", "add", "hello.txt")
		run("git", "commit", "-m", "initial")
		hash := run("git", "rev-parse", "HEAD")

		lines, err := diffTree(dir, hash, "--name-only")
		if err != nil {
			t.Fatalf("diffTree() error: %v", err)
		}
		if len(lines) != 1 || lines[0] != "hello.txt" {
			t.Errorf("diffTree() = %v, want [hello.txt]", lines)
		}
	})

	t.Run("regular commit", func(t *testing.T) {
		dir, run := setupTestRepo(t)

		writeFile(t, dir, "a.txt", "first\n")
		run("git", "add", "a.txt")
		run("git", "commit", "-m", "initial")

		writeFile(t, dir, "b.txt", "second\n")
		run("git", "add", "b.txt")
		run("git", "commit", "-m", "add b")
		hash := run("git", "rev-parse", "HEAD")

		lines, err := diffTree(dir, hash, "--name-only")
		if err != nil {
			t.Fatalf("diffTree() error: %v", err)
		}
		if len(lines) != 1 || lines[0] != "b.txt" {
			t.Errorf("diffTree() = %v, want [b.txt]", lines)
		}
	})

	t.Run("merge commit returns empty slice", func(t *testing.T) {
		dir, run := setupTestRepo(t)

		writeFile(t, dir, "base.txt", "base\n")
		run("git", "add", "base.txt")
		run("git", "commit", "-m", "base")

		// create feature branch from current HEAD
		run("git", "checkout", "-b", "feature")
		writeFile(t, dir, "feature.txt", "feature\n")
		run("git", "add", "feature.txt")
		run("git", "commit", "-m", "feature commit")

		// return to the branch we started from
		run("git", "checkout", "-")

		// add a commit on the base branch to create a divergence
		writeFile(t, dir, "main_change.txt", "main\n")
		run("git", "add", "main_change.txt")
		run("git", "commit", "-m", "main commit")

		// merge feature (no -m flag in diff-tree means merge commits produce empty output)
		run("git", "merge", "--no-edit", "feature")
		mergeHash := run("git", "rev-parse", "HEAD")

		lines, err := diffTree(dir, mergeHash, "--name-only")
		if err != nil {
			t.Fatalf("diffTree() on merge commit error: %v", err)
		}
		// diff-tree without -m returns empty for merge commits — this is intentional
		if len(lines) != 0 {
			t.Errorf("diffTree() on merge commit = %v, want empty slice", lines)
		}
	})
}
