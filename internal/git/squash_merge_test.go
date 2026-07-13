package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// setupTestRepo initialises a temporary git repository with one commit and
// returns the directory path and the commit hash.
func setupTestRepo(t *testing.T) (dir string, commitHash string) {
	t.Helper()

	dir = t.TempDir()

	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		cmd.Env = append(os.Environ(),
			"GIT_AUTHOR_NAME=test",
			"GIT_AUTHOR_EMAIL=test@test.com",
			"GIT_COMMITTER_NAME=test",
			"GIT_COMMITTER_EMAIL=test@test.com",
		)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}

	run("init", "-b", "main")
	run("config", "user.email", "test@test.com")
	run("config", "user.name", "test")

	// Write a file and commit with a Partio-Checkpoint trailer.
	if err := os.WriteFile(filepath.Join(dir, "file.txt"), []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}
	run("add", ".")
	run("commit", "-m", "initial commit\n\nPartio-Checkpoint: abcdef123456")

	// Capture the commit hash.
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		t.Fatal(err)
	}

	hash := strings.TrimSpace(string(out))
	return dir, hash
}

func TestCommitExists(t *testing.T) {
	dir, hash := setupTestRepo(t)

	// Change working dir so execGit operates on the test repo.
	orig, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(orig) })

	if !CommitExists(hash) {
		t.Errorf("CommitExists(%q) = false, want true for existing commit", hash)
	}

	if CommitExists("0000000000000000000000000000000000000000") {
		t.Error("CommitExists(zero SHA) = true, want false for non-existent commit")
	}

	if CommitExists("") {
		t.Error("CommitExists(\"\") = true, want false")
	}
}

func TestFindCommitByCheckpointID(t *testing.T) {
	dir, hash := setupTestRepo(t)

	orig, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(orig) })

	got, err := FindCommitByCheckpointID("abcdef123456")
	if err != nil {
		t.Fatalf("FindCommitByCheckpointID: unexpected error: %v", err)
	}
	if got != hash {
		t.Errorf("FindCommitByCheckpointID(%q) = %q, want %q", "abcdef123456", got, hash)
	}

	// ID that does not appear in any commit.
	got2, _ := FindCommitByCheckpointID("000000000000")
	if got2 != "" {
		t.Errorf("FindCommitByCheckpointID(unknown ID) = %q, want empty", got2)
	}
}
