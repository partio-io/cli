package checkpoint

import (
	"testing"
	"time"
)

// TestStore_Write_AutoReconcile verifies that Store.Write automatically
// reconciles the remote branch before writing, so the resulting branch
// contains both the remote checkpoint (from machine A) and the new
// local checkpoint (from machine B).
func TestStore_Write_AutoReconcile(t *testing.T) {
	bareDir := initBareRepo(t)

	// Machine A: write a checkpoint and push it to the remote.
	machineA := t.TempDir()
	initTestRepo(t, machineA)
	addRemote(t, machineA, bareDir)
	cpA := &Checkpoint{
		ID:          "aa1111111111",
		Branch:      "main",
		CommitHash:  "cafebabe",
		CreatedAt:   time.Now().Truncate(time.Second),
		Agent:       "claude-code",
		ContentHash: "aa1111111111",
	}
	writeTestCheckpoint(t, machineA, cpA)
	pushCheckpointBranch(t, machineA)

	// Machine B: independent init, no knowledge of machine A's checkpoint.
	machineB := t.TempDir()
	initTestRepo(t, machineB)
	addRemote(t, machineB, bareDir)

	// Machine B writes its own checkpoint — this should trigger auto-reconcile
	// before appending, pulling in machine A's checkpoint from remote.
	cpB := &Checkpoint{
		ID:          "bb2222222222",
		Branch:      "feature/x",
		CommitHash:  "deadbeef",
		CreatedAt:   time.Now().Truncate(time.Second),
		Agent:       "claude-code",
		ContentHash: "bb2222222222",
	}
	storeB := NewStore(machineB)
	sf := &SessionFiles{
		ContentHash: cpB.CommitHash,
		Context:     "ctx",
		Diff:        "",
		FullJSONL:   "",
		Metadata:    SessionMetadata{Agent: cpB.Agent},
		Plan:        "",
		Prompt:      "prompt",
	}
	if err := storeB.Write(cpB, sf); err != nil {
		t.Fatalf("Store.Write on machine B: %v", err)
	}

	// Both checkpoints must be present on machine B's checkpoint branch.
	if !canReadCheckpoint(t, machineB, cpA.ID) {
		t.Errorf("machine A's checkpoint %s not found on machine B after Write", cpA.ID)
	}
	if !canReadCheckpoint(t, machineB, cpB.ID) {
		t.Errorf("machine B's checkpoint %s not found after Write", cpB.ID)
	}
}
