package checkpoint

import (
	"os/exec"
	"strings"
	"testing"
	"time"
)

// runCmd runs a git (or other) command in dir, fataling on failure.
func runCmd(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("command %v failed: %v\n%s", args, err, out)
	}
}

// runCmdOut runs a command and returns its trimmed stdout, fataling on failure.
func runCmdOut(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("command %v failed: %v\n%s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

// initWithRemote sets up localDir with an initial commit and a remote pointing
// to remoteDir (a bare repo). Returns remoteDir.
func initWithRemote(t *testing.T, localDir string) string {
	t.Helper()
	remoteDir := t.TempDir()

	runCmd(t, remoteDir, "git", "init", "--bare")

	runCmd(t, localDir, "git", "init")
	runCmd(t, localDir, "git", "config", "user.email", "test@example.com")
	runCmd(t, localDir, "git", "config", "user.name", "Test")
	runCmd(t, localDir, "git", "commit", "--allow-empty", "-m", "initial")
	runCmd(t, localDir, "git", "remote", "add", "origin", remoteDir)

	return remoteDir
}

// initDevice2 creates a second local repo that shares the same remote and has
// the checkpoint branch fetched from it. Call after localDir has already pushed
// the checkpoint branch to remote.
func initDevice2(t *testing.T, remoteDir string) string {
	t.Helper()
	dir := t.TempDir()

	runCmd(t, dir, "git", "init")
	runCmd(t, dir, "git", "config", "user.email", "test@example.com")
	runCmd(t, dir, "git", "config", "user.name", "Test")
	runCmd(t, dir, "git", "commit", "--allow-empty", "-m", "initial")
	runCmd(t, dir, "git", "remote", "add", "origin", remoteDir)

	// Fetch the checkpoint branch that localDir already pushed.
	runCmd(t, dir, "git", "fetch", "--no-tags", "origin",
		checkpointBranch+":refs/heads/"+checkpointBranch)

	return dir
}

// TestReconcileWithRemote_Diverged simulates two devices writing independent
// checkpoints to the same remote and verifies that ReconcileWithRemote merges
// both sets of checkpoints without losing any data.
func TestReconcileWithRemote_Diverged(t *testing.T) {
	localDir := t.TempDir()
	remoteDir := initWithRemote(t, localDir)

	now := time.Now().Truncate(time.Second)

	// Initialise the checkpoint orphan branch in localDir.
	initTestRepo(t, localDir)

	// Write two checkpoints on the local device.
	writeTestCheckpoint(t, localDir, &Checkpoint{
		ID: "aabbcc001100", Branch: "main", CommitHash: "deadbeef1",
		CreatedAt: now, Agent: "claude-code", ContentHash: "hash1",
	})
	writeTestCheckpoint(t, localDir, &Checkpoint{
		ID: "112233445566", Branch: "main", CommitHash: "deadbeef2",
		CreatedAt: now, Agent: "claude-code", ContentHash: "hash2",
	})

	// Push checkpoint branch to remote so device2 can base off it.
	runCmd(t, localDir, "git", "push", "origin", checkpointBranch)

	// Set up a second device and add a different checkpoint there.
	device2Dir := initDevice2(t, remoteDir)
	writeTestCheckpoint(t, device2Dir, &Checkpoint{
		ID: "ccddee001100", Branch: "feature", CommitHash: "deadbeef3",
		CreatedAt: now, Agent: "claude-code", ContentHash: "hash3",
	})

	// Push device2's checkpoint to remote — remote now has a checkpoint that
	// local does not (diverged from local's perspective).
	runCmd(t, device2Dir, "git", "push", "origin", checkpointBranch)

	// Local adds another checkpoint AFTER the push — ensures real divergence.
	writeTestCheckpoint(t, localDir, &Checkpoint{
		ID: "ddeeff001100", Branch: "main", CommitHash: "deadbeef4",
		CreatedAt: now, Agent: "claude-code", ContentHash: "hash4",
	})

	// At this point:
	//   local  has: aabbcc001100, 112233445566, ddeeff001100  (not pushed)
	//   remote has: aabbcc001100, 112233445566, ccddee001100  (from device2)
	// The branches have diverged.

	// Reconcile.
	if err := ReconcileWithRemote(localDir, "origin"); err != nil {
		t.Fatalf("ReconcileWithRemote() error = %v", err)
	}

	// Verify all four checkpoints are now in the local branch.
	wantIDs := []string{
		"aabbcc001100",
		"112233445566",
		"ccddee001100",
		"ddeeff001100",
	}

	for _, id := range wantIDs {
		shard := Shard(id)
		rest := Rest(id)
		out, err := runGitInDir(localDir, "show",
			checkpointBranch+":"+shard+"/"+rest+"/metadata.json")
		if err != nil {
			t.Errorf("checkpoint %s not found after reconciliation: %v", id, err)
		} else if out == "" {
			t.Errorf("checkpoint %s metadata is empty after reconciliation", id)
		}
	}
}

// TestReconcileWithRemote_NoRemote verifies that reconciliation is a no-op
// when the repository has no remote configured.
func TestReconcileWithRemote_NoRemote(t *testing.T) {
	dir := t.TempDir()
	initTestRepo(t, dir)

	// No remote — should be a clean no-op.
	if err := ReconcileWithRemote(dir, "origin"); err != nil {
		t.Fatalf("ReconcileWithRemote() with no remote: unexpected error %v", err)
	}
}

// TestReconcileWithRemote_FastForward verifies that when the remote is strictly
// ahead of local, the local branch is fast-forwarded without a merge commit.
func TestReconcileWithRemote_FastForward(t *testing.T) {
	localDir := t.TempDir()
	remoteDir := initWithRemote(t, localDir)

	now := time.Now().Truncate(time.Second)

	// Initialise checkpoint branch and push to remote.
	initTestRepo(t, localDir)
	writeTestCheckpoint(t, localDir, &Checkpoint{
		ID: "aabbcc001100", Branch: "main", CommitHash: "deadbeef1",
		CreatedAt: now, Agent: "claude-code", ContentHash: "hash1",
	})
	runCmd(t, localDir, "git", "push", "origin", checkpointBranch)

	// Add another checkpoint only on device2 and push.
	device2Dir := initDevice2(t, remoteDir)
	writeTestCheckpoint(t, device2Dir, &Checkpoint{
		ID: "ccddee001100", Branch: "feature", CommitHash: "deadbeef2",
		CreatedAt: now, Agent: "claude-code", ContentHash: "hash2",
	})
	runCmd(t, device2Dir, "git", "push", "origin", checkpointBranch)

	// Remote is ahead of local — reconcile should fast-forward.
	if err := ReconcileWithRemote(localDir, "origin"); err != nil {
		t.Fatalf("ReconcileWithRemote() error = %v", err)
	}

	// Both checkpoints should now be accessible locally.
	for _, id := range []string{"aabbcc001100", "ccddee001100"} {
		shard := Shard(id)
		rest := Rest(id)
		_, err := runGitInDir(localDir, "show",
			checkpointBranch+":"+shard+"/"+rest+"/metadata.json")
		if err != nil {
			t.Errorf("checkpoint %s not found after fast-forward: %v", id, err)
		}
	}
}

// TestReconcileWithRemote_AlreadyInSync verifies that reconciliation is a
// no-op when local and remote are at the same commit.
func TestReconcileWithRemote_AlreadyInSync(t *testing.T) {
	localDir := t.TempDir()
	initWithRemote(t, localDir)

	now := time.Now().Truncate(time.Second)

	initTestRepo(t, localDir)
	writeTestCheckpoint(t, localDir, &Checkpoint{
		ID: "aabbcc001100", Branch: "main", CommitHash: "deadbeef1",
		CreatedAt: now, Agent: "claude-code", ContentHash: "hash1",
	})
	runCmd(t, localDir, "git", "push", "origin", checkpointBranch)

	// Already in sync — should succeed without any changes.
	beforeRef := runCmdOut(t, localDir, "git", "rev-parse", "refs/heads/"+checkpointBranch)

	if err := ReconcileWithRemote(localDir, "origin"); err != nil {
		t.Fatalf("ReconcileWithRemote() error = %v", err)
	}

	afterRef := runCmdOut(t, localDir, "git", "rev-parse", "refs/heads/"+checkpointBranch)
	if beforeRef != afterRef {
		t.Errorf("ref changed unexpectedly: before=%s after=%s", beforeRef, afterRef)
	}
}

// runGitInDir runs a git command in dir and returns (stdout, error).
func runGitInDir(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	return strings.TrimSpace(string(out)), err
}
