package git

import (
	"testing"
)

func TestDiffNameOnly(t *testing.T) {
	t.Run("initial commit", func(t *testing.T) {
		dir, run := setupTestRepo(t)

		writeFile(t, dir, "hello.txt", "hello\n")
		run("git", "add", "hello.txt")
		run("git", "commit", "-m", "initial")
		hash := run("git", "rev-parse", "HEAD")

		t.Chdir(dir)

		files, err := DiffNameOnly(hash)
		if err != nil {
			t.Fatalf("DiffNameOnly() error: %v", err)
		}
		if len(files) != 1 || files[0] != "hello.txt" {
			t.Errorf("DiffNameOnly() = %v, want [hello.txt]", files)
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

		t.Chdir(dir)

		files, err := DiffNameOnly(hash)
		if err != nil {
			t.Fatalf("DiffNameOnly() error: %v", err)
		}
		if len(files) != 1 || files[0] != "b.txt" {
			t.Errorf("DiffNameOnly() = %v, want [b.txt]", files)
		}
	})

	t.Run("merge commit returns nil", func(t *testing.T) {
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

		files, err := DiffNameOnly(mergeHash)
		if err != nil {
			t.Fatalf("DiffNameOnly() error: %v", err)
		}
		if files != nil {
			t.Errorf("DiffNameOnly() on merge = %v, want nil", files)
		}
	})
}
