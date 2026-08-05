package main

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/partio-io/cli/internal/checkpoint"
)

// rewindInitTestRepo initialises a bare-minimum git repository in dir and
// creates the checkpoint orphan branch so that Store.Write can be called.
func rewindInitTestRepo(t *testing.T, dir string) {
	t.Helper()

	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("command %v failed: %v\n%s", args, err, out)
		}
	}

	run("git", "init")
	run("git", "config", "user.email", "test@example.com")
	run("git", "config", "user.name", "Test")
	run("git", "commit", "--allow-empty", "-m", "initial")

	// Create the checkpoint orphan branch using git plumbing.
	cmd := exec.Command("git", "hash-object", "-t", "tree", "/dev/null")
	cmd.Dir = dir
	treeHashBytes, err := cmd.Output()
	if err != nil {
		cmd2 := exec.Command("git", "mktree")
		cmd2.Dir = dir
		treeHashBytes, err = cmd2.Output()
		if err != nil {
			t.Fatalf("creating empty tree: %v", err)
		}
	}
	treeHash := strings.TrimSpace(string(treeHashBytes))

	commitCmd := exec.Command("git", "commit-tree", treeHash, "-m", "partio: initialize checkpoint storage")
	commitCmd.Dir = dir
	commitHashBytes, err := commitCmd.Output()
	if err != nil {
		t.Fatalf("creating orphan commit: %v", err)
	}
	commitHash := strings.TrimSpace(string(commitHashBytes))

	run("git", "update-ref", "refs/heads/partio/checkpoints/v1", commitHash)
}

// rewindWriteCheckpoint writes a checkpoint to the checkpoint branch in dir.
func rewindWriteCheckpoint(t *testing.T, dir string, cp *checkpoint.Checkpoint) {
	t.Helper()
	s := checkpoint.NewStore(dir)
	sf := &checkpoint.SessionFiles{
		ContentHash: "hash",
		Context:     "context",
		Metadata:    checkpoint.SessionMetadata{Agent: cp.Agent},
		Prompt:      "prompt",
	}
	if err := s.Write(cp, sf); err != nil {
		t.Fatalf("Store.Write: %v", err)
	}
}

// captureOutput redirects os.Stdout, runs f, and returns the captured output.
func captureOutput(t *testing.T, f func()) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	os.Stdout = w
	f()
	_ = w.Close()
	os.Stdout = old
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("reading stdout: %v", err)
	}
	return buf.String()
}

// repoHEAD returns the HEAD commit SHA of the git repo in dir.
func repoHEAD(t *testing.T, dir string) string {
	t.Helper()
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("git rev-parse HEAD: %v", err)
	}
	return strings.TrimSpace(string(out))
}

func TestRewindBranchFlagDefined(t *testing.T) {
	cmd := newRewindCmd()
	if cmd.Flags().Lookup("branch") == nil {
		t.Error("expected --branch flag to be defined on rewind command")
	}
}

func TestRunRewindListByBranch(t *testing.T) {
	now := time.Now().Truncate(time.Second)

	tests := []struct {
		name        string
		checkpoints []struct {
			id     string
			branch string
		}
		filterBranch    string
		wantInOutput    []string
		wantNotInOutput []string
	}{
		{
			name: "filters to matching branch only",
			checkpoints: []struct {
				id     string
				branch string
			}{
				{"aabbcc001100", "feature/foo"},
				{"112233445566", "main"},
			},
			filterBranch:    "feature/foo",
			wantInOutput:    []string{"aabbcc001100"},
			wantNotInOutput: []string{"112233445566"},
		},
		{
			name: "no match shows empty message",
			checkpoints: []struct {
				id     string
				branch string
			}{
				{"aabbcc001100", "feature/foo"},
			},
			filterBranch:    "nonexistent",
			wantInOutput:    []string{"No checkpoints found"},
			wantNotInOutput: []string{"aabbcc001100"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			rewindInitTestRepo(t, dir)

			for _, cp := range tt.checkpoints {
				rewindWriteCheckpoint(t, dir, &checkpoint.Checkpoint{
					ID:          cp.id,
					Branch:      cp.branch,
					CommitHash:  "deadbeef",
					CreatedAt:   now,
					Agent:       "claude-code",
					ContentHash: cp.id,
				})
			}

			t.Chdir(dir)

			var runErr error
			out := captureOutput(t, func() {
				runErr = runRewindList(tt.filterBranch)
			})

			if runErr != nil {
				t.Fatalf("runRewindList() error = %v", runErr)
			}
			for _, want := range tt.wantInOutput {
				if !strings.Contains(out, want) {
					t.Errorf("output missing %q; got:\n%s", want, out)
				}
			}
			for _, notWant := range tt.wantNotInOutput {
				if strings.Contains(out, notWant) {
					t.Errorf("output unexpectedly contains %q; got:\n%s", notWant, out)
				}
			}
		})
	}
}

func TestRunRewindTo_UnreachableCommit(t *testing.T) {
	dir := t.TempDir()
	rewindInitTestRepo(t, dir)

	// Write a checkpoint whose commit hash is all zeros — never reachable.
	cpID := "ff00001122aa"
	rewindWriteCheckpoint(t, dir, &checkpoint.Checkpoint{
		ID:          cpID,
		Branch:      "feature/squashed",
		CommitHash:  "0000000000000000000000000000000000000000",
		CreatedAt:   time.Now(),
		Agent:       "claude-code",
		ContentHash: "hash",
	})

	t.Chdir(dir)

	var runErr error
	out := captureOutput(t, func() {
		runErr = runRewindTo(cpID)
	})

	if runErr != nil {
		t.Fatalf("runRewindTo() returned error for unreachable commit: %v", runErr)
	}

	if !strings.Contains(out, "Warning:") {
		t.Errorf("expected warning for unreachable commit, got output:\n%s", out)
	}

	// Verify no rewind branch was created.
	checkBranch := exec.Command("git", "rev-parse", "--verify", "refs/heads/partio/rewind/"+cpID)
	checkBranch.Dir = dir
	if err := checkBranch.Run(); err == nil {
		t.Error("expected no rewind branch to be created for unreachable commit, but branch exists")
	}
}

func TestRunRewindTo_ReachableCommit(t *testing.T) {
	dir := t.TempDir()
	rewindInitTestRepo(t, dir)

	// Use the real HEAD commit so CommitReachable returns true.
	headSHA := repoHEAD(t, dir)

	cpID := "ee11223344ff"
	rewindWriteCheckpoint(t, dir, &checkpoint.Checkpoint{
		ID:          cpID,
		Branch:      "feature/normal",
		CommitHash:  headSHA,
		CreatedAt:   time.Now(),
		Agent:       "claude-code",
		ContentHash: "hash",
	})

	t.Chdir(dir)

	var runErr error
	out := captureOutput(t, func() {
		runErr = runRewindTo(cpID)
	})

	if runErr != nil {
		t.Fatalf("runRewindTo() returned error for reachable commit: %v", runErr)
	}

	if strings.Contains(out, "Warning:") {
		t.Errorf("unexpected warning for reachable commit; output:\n%s", out)
	}

	// Verify the rewind branch was created.
	checkBranch := exec.Command("git", "rev-parse", "--verify", "refs/heads/partio/rewind/"+cpID)
	checkBranch.Dir = dir
	if err := checkBranch.Run(); err != nil {
		t.Errorf("expected rewind branch to be created, got error: %v", err)
	}
}
