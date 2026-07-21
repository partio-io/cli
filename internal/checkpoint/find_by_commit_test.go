package checkpoint

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// runGit runs a git command in dir. Fatals on error.
func runGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

// initCheckpointRepo creates a temp git repo with the checkpoint orphan branch initialized.
func initCheckpointRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	runGit(t, dir, "init")
	runGit(t, dir, "config", "user.email", "test@test.com")
	runGit(t, dir, "config", "user.name", "Test")

	// Create an initial commit on the main branch so we have a valid repo
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("test"), 0o644); err != nil {
		t.Fatal(err)
	}
	runGit(t, dir, "add", ".")
	runGit(t, dir, "commit", "-m", "initial")

	// Create the orphan checkpoint branch (mirrors createCheckpointBranch in enable.go)
	treeHash := runGit(t, dir, "hash-object", "-t", "tree", "/dev/null")
	commitHash := runGit(t, dir, "commit-tree", treeHash, "-m", "partio: initialize checkpoint storage")
	runGit(t, dir, "update-ref", "refs/heads/partio/checkpoints/v1", commitHash)

	return dir
}

func TestFindByCommit(t *testing.T) {
	dir := initCheckpointRepo(t)
	t.Chdir(dir)

	targetCommit := "abcdef1234567890abcdef1234567890abcdef12"
	otherCommit := "1111111111111111111111111111111111111111"

	store := NewStore(dir)

	// Write a checkpoint for targetCommit
	cp := &Checkpoint{
		ID:          "aabbcc112233",
		CommitHash:  targetCommit,
		Branch:      "feature",
		CreatedAt:   time.Now(),
		Agent:       "claude-code",
		AgentPct:    100,
		ContentHash: "deadbeef",
	}
	sf := &SessionFiles{
		ContentHash: "deadbeef",
		Context:     "test context",
		Diff:        "test diff",
		FullJSONL:   `{"sessionId":"s1"}`,
		Prompt:      "test prompt",
		Plan:        "test plan",
	}
	if err := store.Write(cp, sf); err != nil {
		t.Fatalf("Write: %v", err)
	}

	// Write a checkpoint for a different commit
	cp2 := &Checkpoint{
		ID:          "ddeeff445566",
		CommitHash:  otherCommit,
		Branch:      "main",
		CreatedAt:   time.Now(),
		Agent:       "claude-code",
		AgentPct:    50,
		ContentHash: "cafebabe",
	}
	sf2 := &SessionFiles{
		ContentHash: "cafebabe",
		Context:     "other context",
		Diff:        "",
		FullJSONL:   `{"sessionId":"s2"}`,
		Prompt:      "other prompt",
		Plan:        "",
	}
	if err := store.Write(cp2, sf2); err != nil {
		t.Fatalf("Write: %v", err)
	}

	t.Run("finds checkpoint by commit SHA", func(t *testing.T) {
		ids, err := FindByCommit(targetCommit)
		if err != nil {
			t.Fatalf("FindByCommit() error: %v", err)
		}
		if len(ids) != 1 {
			t.Fatalf("FindByCommit() returned %d IDs, want 1: %v", len(ids), ids)
		}
		if ids[0] != cp.ID {
			t.Errorf("FindByCommit() = %q, want %q", ids[0], cp.ID)
		}
	})

	t.Run("does not return checkpoint for different commit", func(t *testing.T) {
		ids, err := FindByCommit(targetCommit)
		if err != nil {
			t.Fatalf("FindByCommit() error: %v", err)
		}
		for _, id := range ids {
			if id == cp2.ID {
				t.Errorf("FindByCommit(%s) returned %s which belongs to a different commit", targetCommit[:8], cp2.ID)
			}
		}
	})

	t.Run("returns empty for unknown commit", func(t *testing.T) {
		ids, err := FindByCommit("9999999999999999999999999999999999999999")
		if err != nil {
			t.Fatalf("FindByCommit() error: %v", err)
		}
		if len(ids) != 0 {
			t.Errorf("FindByCommit(unknown) = %v, want empty", ids)
		}
	})
}

func TestFindByCommit_NoBranch(t *testing.T) {
	dir := t.TempDir()
	runGit(t, dir, "init")
	runGit(t, dir, "config", "user.email", "test@test.com")
	runGit(t, dir, "config", "user.name", "Test")
	// No checkpoint branch created
	t.Chdir(dir)

	ids, err := FindByCommit("abc123")
	if err != nil {
		t.Fatalf("FindByCommit() with no branch error: %v", err)
	}
	if ids != nil {
		t.Errorf("FindByCommit() with no branch = %v, want nil", ids)
	}
}
