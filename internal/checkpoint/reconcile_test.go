package checkpoint

import (
	"os/exec"
	"strings"
	"testing"
	"time"
)

// initBareRepo creates a bare git repository in a temp directory.
func initBareRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	runIn(t, dir, "git", "init", "--bare")
	return dir
}

// addRemote adds a remote named "origin" pointing to bareDir in workDir.
func addRemote(t *testing.T, workDir, bareDir string) {
	t.Helper()
	runIn(t, workDir, "git", "remote", "add", "origin", bareDir)
}

// pushCheckpointBranch pushes the local checkpoint branch to origin.
func pushCheckpointBranch(t *testing.T, workDir string) {
	t.Helper()
	runIn(t, workDir, "git", "push", "origin",
		"refs/heads/"+checkpointBranch+":refs/heads/"+checkpointBranch)
}

// runIn runs a command in the given directory, failing the test on error.
func runIn(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("command %v in %s failed: %v\n%s", args, dir, err, out)
	}
}

// canReadCheckpoint returns true if the checkpoint's metadata.json is readable
// from the local checkpoint branch in the given repo.
func canReadCheckpoint(t *testing.T, dir, id string) bool {
	t.Helper()
	cmd := exec.Command("git", "show",
		checkpointBranch+":"+Shard(id)+"/"+Rest(id)+"/metadata.json")
	cmd.Dir = dir
	err := cmd.Run()
	return err == nil
}

// parentCount returns the number of parents of the HEAD commit on the
// checkpoint branch in the given repo.
func parentCount(t *testing.T, dir string) int {
	t.Helper()
	cmd := exec.Command("git", "rev-list", "--parents", "-n1", checkpointBranch)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("rev-list parents: %v", err)
	}
	// Format: "<commit> [<parent1> [<parent2> ...]]"
	fields := strings.Fields(strings.TrimSpace(string(out)))
	if len(fields) == 0 {
		return 0
	}
	return len(fields) - 1
}

func testCheckpoint(id string) *Checkpoint {
	return &Checkpoint{
		ID:          id,
		Branch:      "main",
		CommitHash:  "deadbeef" + id[:4],
		CreatedAt:   time.Now().Truncate(time.Second),
		Agent:       "claude-code",
		ContentHash: id,
	}
}

func TestStore_Reconcile(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(t *testing.T) (workDir string)
		wantErr    bool
		checkAfter func(t *testing.T, workDir string)
	}{
		{
			name: "no remote: returns nil without error",
			setup: func(t *testing.T) string {
				dir := t.TempDir()
				initTestRepo(t, dir)
				return dir
			},
			wantErr: false,
			checkAfter: func(t *testing.T, workDir string) {
				// No remote-tracking ref should exist.
				cmd := exec.Command("git", "rev-parse",
					"refs/remotes/origin/"+checkpointBranch)
				cmd.Dir = workDir
				if err := cmd.Run(); err == nil {
					t.Error("remote-tracking ref should not exist when there is no remote")
				}
			},
		},
		{
			name: "remote branch absent: returns nil without error",
			setup: func(t *testing.T) string {
				bareDir := initBareRepo(t)
				workDir := t.TempDir()
				initTestRepo(t, workDir)
				addRemote(t, workDir, bareDir)
				// Do NOT push the checkpoint branch — remote has nothing.
				return workDir
			},
			wantErr: false,
		},
		{
			name: "already equal: returns nil and creates no new commit",
			setup: func(t *testing.T) string {
				bareDir := initBareRepo(t)
				workDir := t.TempDir()
				initTestRepo(t, workDir)
				addRemote(t, workDir, bareDir)
				writeTestCheckpoint(t, workDir, testCheckpoint("aa0011223344"))
				// Push so local and remote are in sync.
				pushCheckpointBranch(t, workDir)
				// Fetch so the remote-tracking ref is up to date.
				runIn(t, workDir, "git", "fetch", "--no-tags", "origin",
					"refs/heads/"+checkpointBranch+":refs/remotes/origin/"+checkpointBranch)
				return workDir
			},
			wantErr: false,
			checkAfter: func(t *testing.T, workDir string) {
				// No merge commit should have been created; parent count unchanged (1).
				if n := parentCount(t, workDir); n != 1 {
					t.Errorf("parent count = %d, want 1 (no merge commit expected)", n)
				}
			},
		},
		{
			name: "local absent remote present: fast-forwards local to remote tip",
			setup: func(t *testing.T) string {
				// Machine A: write a checkpoint and push it to the bare remote.
				bareDir := initBareRepo(t)
				machineA := t.TempDir()
				initTestRepo(t, machineA)
				addRemote(t, machineA, bareDir)
				writeTestCheckpoint(t, machineA, testCheckpoint("aa0000000001"))
				pushCheckpointBranch(t, machineA)

				// Machine B: only a regular git init, no checkpoint branch yet.
				machineB := t.TempDir()
				runIn(t, machineB, "git", "init")
				runIn(t, machineB, "git", "config", "user.email", "test@example.com")
				runIn(t, machineB, "git", "config", "user.name", "Test")
				runIn(t, machineB, "git", "commit", "--allow-empty", "-m", "initial")
				addRemote(t, machineB, bareDir)
				return machineB
			},
			wantErr: false,
			checkAfter: func(t *testing.T, workDir string) {
				// After reconcile, machine B should have machine A's checkpoint.
				if !canReadCheckpoint(t, workDir, "aa0000000001") {
					t.Error("expected checkpoint aa0000000001 to be readable after fast-forward")
				}
				// Should be a regular commit (1 parent), not a merge commit.
				if n := parentCount(t, workDir); n != 1 {
					t.Errorf("parent count = %d, want 1 (fast-forward should not create merge commit)", n)
				}
			},
		},
		{
			name: "both diverged: creates merge commit with union of checkpoints",
			setup: func(t *testing.T) string {
				bareDir := initBareRepo(t)

				// Machine A: write checkpoint aa, push to remote.
				machineA := t.TempDir()
				initTestRepo(t, machineA)
				addRemote(t, machineA, bareDir)
				writeTestCheckpoint(t, machineA, testCheckpoint("aa0000000001"))
				pushCheckpointBranch(t, machineA)

				// Machine B: independent init, write checkpoint bb — diverged from A.
				machineB := t.TempDir()
				initTestRepo(t, machineB)
				addRemote(t, machineB, bareDir)
				writeTestCheckpoint(t, machineB, testCheckpoint("bb0000000002"))
				return machineB
			},
			wantErr: false,
			checkAfter: func(t *testing.T, workDir string) {
				if !canReadCheckpoint(t, workDir, "aa0000000001") {
					t.Error("expected remote checkpoint aa0000000001 to be readable after reconcile")
				}
				if !canReadCheckpoint(t, workDir, "bb0000000002") {
					t.Error("expected local checkpoint bb0000000002 to be readable after reconcile")
				}
				// Merge commit must have exactly two parents.
				if n := parentCount(t, workDir); n != 2 {
					t.Errorf("parent count = %d, want 2 (merge commit)", n)
				}
			},
		},
		{
			name: "fetch failure unreachable remote: returns non-nil error",
			setup: func(t *testing.T) string {
				workDir := t.TempDir()
				initTestRepo(t, workDir)
				// Point origin at a path that doesn't exist.
				runIn(t, workDir, "git", "remote", "add", "origin",
					"/nonexistent/path/that/does/not/exist")
				return workDir
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			workDir := tt.setup(t)
			s := NewStore(workDir)

			err := s.Reconcile()
			if (err != nil) != tt.wantErr {
				t.Errorf("Reconcile() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.checkAfter != nil {
				tt.checkAfter(t, workDir)
			}
		})
	}
}

// TestStore_Reconcile_DivergedBranches is a dedicated integration test that
// simulates real multi-device divergence at the git-object level and verifies
// that Reconcile produces a merge commit containing all checkpoints.
func TestStore_Reconcile_DivergedBranches(t *testing.T) {
	bareDir := initBareRepo(t)

	// --- Machine A ---
	machineA := t.TempDir()
	initTestRepo(t, machineA)
	addRemote(t, machineA, bareDir)
	writeTestCheckpoint(t, machineA, &Checkpoint{
		ID:          "aabbcc001100",
		Branch:      "main",
		CommitHash:  "cafebabe",
		CreatedAt:   time.Now().Truncate(time.Second),
		Agent:       "claude-code",
		ContentHash: "aabbcc001100",
	})
	pushCheckpointBranch(t, machineA)

	// --- Machine B (diverged) ---
	machineB := t.TempDir()
	initTestRepo(t, machineB)
	addRemote(t, machineB, bareDir)
	writeTestCheckpoint(t, machineB, &Checkpoint{
		ID:          "112233445566",
		Branch:      "feature/x",
		CommitHash:  "deadbeef",
		CreatedAt:   time.Now().Truncate(time.Second),
		Agent:       "claude-code",
		ContentHash: "112233445566",
	})

	// Reconcile on machine B should merge A's checkpoint into B's branch.
	storeB := NewStore(machineB)
	if err := storeB.Reconcile(); err != nil {
		t.Fatalf("Reconcile() unexpected error: %v", err)
	}

	// Both checkpoints must be readable from machine B.
	if !canReadCheckpoint(t, machineB, "aabbcc001100") {
		t.Error("machine A's checkpoint aabbcc001100 not readable on machine B after reconcile")
	}
	if !canReadCheckpoint(t, machineB, "112233445566") {
		t.Error("machine B's checkpoint 112233445566 not readable after reconcile")
	}

	// The reconciled commit must have two parents.
	if n := parentCount(t, machineB); n != 2 {
		t.Errorf("expected merge commit with 2 parents, got %d", n)
	}
}
