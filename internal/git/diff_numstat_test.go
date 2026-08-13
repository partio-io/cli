package git

import (
	"strings"
	"testing"
)

func TestDiffNumstat(t *testing.T) {
	t.Run("initial commit", func(t *testing.T) {
		dir, run := setupTestRepo(t)

		writeFile(t, dir, "hello.txt", "hello\nworld\n")
		run("git", "add", "hello.txt")
		run("git", "commit", "-m", "initial")
		hash := run("git", "rev-parse", "HEAD")

		// Change cwd to the test repo so execGit-based helpers work.
		t.Chdir(dir)

		out, err := DiffNumstat(hash)
		if err != nil {
			t.Fatalf("DiffNumstat() error: %v", err)
		}
		// numstat format: "<added>\t<deleted>\t<path>"
		if !strings.Contains(out, "hello.txt") {
			t.Errorf("DiffNumstat() = %q, want output containing hello.txt", out)
		}
	})

	t.Run("regular commit", func(t *testing.T) {
		dir, run := setupTestRepo(t)

		writeFile(t, dir, "a.txt", "line1\n")
		run("git", "add", "a.txt")
		run("git", "commit", "-m", "initial")

		writeFile(t, dir, "b.txt", "line2\nline3\n")
		run("git", "add", "b.txt")
		run("git", "commit", "-m", "add b")
		hash := run("git", "rev-parse", "HEAD")

		t.Chdir(dir)

		out, err := DiffNumstat(hash)
		if err != nil {
			t.Fatalf("DiffNumstat() error: %v", err)
		}
		if !strings.Contains(out, "b.txt") {
			t.Errorf("DiffNumstat() = %q, want output containing b.txt", out)
		}
	})

	t.Run("merge commit returns empty", func(t *testing.T) {
		dir, run := setupTestRepo(t)

		writeFile(t, dir, "base.txt", "base\n")
		run("git", "add", "base.txt")
		run("git", "commit", "-m", "base")

		run("git", "checkout", "-b", "feature")
		writeFile(t, dir, "feature.txt", "feature\n")
		run("git", "add", "feature.txt")
		run("git", "commit", "-m", "feature commit")

		run("git", "checkout", "-")

		writeFile(t, dir, "main_change.txt", "main\n")
		run("git", "add", "main_change.txt")
		run("git", "commit", "-m", "main commit")

		run("git", "merge", "--no-edit", "feature")
		mergeHash := run("git", "rev-parse", "HEAD")

		t.Chdir(dir)

		out, err := DiffNumstat(mergeHash)
		if err != nil {
			t.Fatalf("DiffNumstat() error: %v", err)
		}
		if out != "" {
			t.Errorf("DiffNumstat() on merge = %q, want empty string", out)
		}
	})
}
