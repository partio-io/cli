package checkpoint

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// newTestRepo initialises a temporary git repo and returns its path along
// with the hash of the single initial commit (which carries a
// Partio-Checkpoint trailer so FindCommitByCheckpointID can locate it).
func newTestRepo(t *testing.T, checkpointID string) (dir, hash string) {
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

	if err := os.WriteFile(filepath.Join(dir, "file.txt"), []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}
	run("add", ".")
	run("commit", "-m", "feat: initial\n\nPartio-Checkpoint: "+checkpointID)

	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		t.Fatal(err)
	}
	return dir, strings.TrimSpace(string(out))
}

func TestResolveCommitHash_existingCommit(t *testing.T) {
	dir, hash := newTestRepo(t, "abcdef123456")

	orig, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(orig) })

	// When the original commit exists, ResolveCommitHash should return it unchanged.
	got := ResolveCommitHash("abcdef123456", hash)
	if got != hash {
		t.Errorf("ResolveCommitHash with existing commit = %q, want %q", got, hash)
	}
}

func TestResolveCommitHash_squashMerged(t *testing.T) {
	const id = "abcdef123456"
	dir, hash := newTestRepo(t, id)

	orig, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(orig) })

	// Simulate a squash-merge: the "original" commit SHA does not exist but the
	// commit in the repo carries the checkpoint trailer.
	fakeOriginal := "deadbeefdeadbeefdeadbeefdeadbeefdeadbeef"
	got := ResolveCommitHash(id, fakeOriginal)
	if got != hash {
		t.Errorf("ResolveCommitHash (squash-merged) = %q, want %q (squash commit)", got, hash)
	}
}

func TestResolveCommitHash_noFallback(t *testing.T) {
	dir, _ := newTestRepo(t, "abcdef123456")

	orig, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(orig) })

	// Neither the original commit exists nor the trailer ID matches anything.
	const unknownID = "000000000000"
	fakeOriginal := "deadbeefdeadbeefdeadbeefdeadbeefdeadbeef"
	got := ResolveCommitHash(unknownID, fakeOriginal)
	// Should fall back to the original (unchanged) so callers can give a useful error.
	if got != fakeOriginal {
		t.Errorf("ResolveCommitHash (no fallback) = %q, want %q (original)", got, fakeOriginal)
	}
}
