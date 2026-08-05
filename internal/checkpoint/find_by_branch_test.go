package checkpoint

import (
	"os/exec"
	"strings"
	"testing"
	"time"
)

// initTestRepo initialises a bare-minimum git repository in dir and creates
// the checkpoint orphan branch so that Store.Write can be called immediately.
func initTestRepo(t *testing.T, dir string) {
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

	// Create an initial commit so the repo has a HEAD.
	run("git", "commit", "--allow-empty", "-m", "initial")

	// Create the checkpoint orphan branch using git plumbing.
	cmd := exec.Command("git", "hash-object", "-t", "tree", "/dev/null")
	cmd.Dir = dir
	treeHashBytes, err := cmd.Output()
	if err != nil {
		// Fallback: mktree with empty input.
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

	run("git", "update-ref", "refs/heads/"+checkpointBranch, commitHash)
}

// writeTestCheckpoint writes a single checkpoint into dir's checkpoint branch.
func writeTestCheckpoint(t *testing.T, dir string, cp *Checkpoint) {
	t.Helper()
	s := NewStore(dir)
	sf := &SessionFiles{
		ContentHash: cp.CommitHash,
		Context:     "ctx",
		Diff:        "",
		FullJSONL:   "",
		Metadata:    SessionMetadata{Agent: cp.Agent},
		Plan:        "",
		Prompt:      "prompt",
	}
	if err := s.Write(cp, sf); err != nil {
		t.Fatalf("Store.Write: %v", err)
	}
}

func TestFindByBranch(t *testing.T) {
	now := time.Now().Truncate(time.Second)

	tests := []struct {
		name       string
		checkpoints []struct {
			id     string
			branch string
		}
		queryBranch string
		wantIDs     []string // expected IDs in any order
	}{
		{
			name: "single match",
			checkpoints: []struct {
				id     string
				branch string
			}{
				{"aabbcc001100", "feature/foo"},
				{"112233445566", "main"},
			},
			queryBranch: "feature/foo",
			wantIDs:     []string{"aabbcc001100"},
		},
		{
			name: "multiple matches",
			checkpoints: []struct {
				id     string
				branch string
			}{
				{"aa0000000001", "feature/bar"},
				{"bb0000000002", "feature/bar"},
				{"cc0000000003", "other"},
			},
			queryBranch: "feature/bar",
			wantIDs:     []string{"aa0000000001", "bb0000000002"},
		},
		{
			name: "no match",
			checkpoints: []struct {
				id     string
				branch string
			}{
				{"dd0000000001", "main"},
			},
			queryBranch: "nonexistent",
			wantIDs:     []string{},
		},
		{
			name: "partial branch name does not match",
			checkpoints: []struct {
				id     string
				branch string
			}{
				{"ee0000000001", "feature/foo"},
			},
			queryBranch: "feature/f", // prefix only — must NOT match
			wantIDs:     []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			initTestRepo(t, dir)

			for _, cp := range tt.checkpoints {
				writeTestCheckpoint(t, dir, &Checkpoint{
					ID:          cp.id,
					Branch:      cp.branch,
					CommitHash:  "deadbeef",
					CreatedAt:   now,
					Agent:       "claude-code",
					ContentHash: cp.id,
				})
			}

			got, err := FindByBranch(dir, tt.queryBranch)
			if err != nil {
				t.Fatalf("FindByBranch() error = %v", err)
			}

			if len(got) != len(tt.wantIDs) {
				t.Fatalf("FindByBranch() returned %d results, want %d", len(got), len(tt.wantIDs))
			}

			gotIDs := make(map[string]bool, len(got))
			for _, cp := range got {
				gotIDs[cp.ID] = true
			}
			for _, wantID := range tt.wantIDs {
				if !gotIDs[wantID] {
					t.Errorf("FindByBranch() missing expected ID %q", wantID)
				}
			}
		})
	}
}

func TestFindByBranchNoBranch(t *testing.T) {
	dir := t.TempDir()

	// Init a repo WITHOUT the checkpoint branch.
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

	got, err := FindByBranch(dir, "any-branch")
	if err != nil {
		t.Fatalf("FindByBranch() with no checkpoint branch: unexpected error %v", err)
	}
	if len(got) != 0 {
		t.Errorf("FindByBranch() with no checkpoint branch: got %d results, want 0", len(got))
	}
}
