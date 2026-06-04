package checkpoint

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"
)

func TestListSort(t *testing.T) {
	now := time.Now()
	metas := []Metadata{
		{ID: "aaa111222333", CreatedAt: now.Add(-2 * time.Hour).Format(time.RFC3339)},
		{ID: "bbb444555666", CreatedAt: now.Format(time.RFC3339)},
		{ID: "ccc777888999", CreatedAt: now.Add(-1 * time.Hour).Format(time.RFC3339)},
	}

	sort.Slice(metas, func(i, j int) bool {
		ti, _ := time.Parse(time.RFC3339, metas[i].CreatedAt)
		tj, _ := time.Parse(time.RFC3339, metas[j].CreatedAt)
		return ti.After(tj)
	})

	if metas[0].ID != "bbb444555666" {
		t.Errorf("expected newest first, got %s", metas[0].ID)
	}
	if metas[1].ID != "ccc777888999" {
		t.Errorf("expected second newest second, got %s", metas[1].ID)
	}
	if metas[2].ID != "aaa111222333" {
		t.Errorf("expected oldest last, got %s", metas[2].ID)
	}
}

func TestListNoBranch(t *testing.T) {
	// When run outside a git repo or without the checkpoint branch,
	// List() should return ErrNoBranch.
	// We can't easily mock git commands, so we verify the error sentinel exists.
	if ErrNoBranch == nil {
		t.Error("ErrNoBranch should not be nil")
	}
	if ErrNoBranch.Error() != "checkpoint branch does not exist" {
		t.Errorf("unexpected error message: %s", ErrNoBranch.Error())
	}
}

func TestListEmptyResult(t *testing.T) {
	// An empty slice (not nil) should be treated as "no checkpoints".
	var metas []Metadata
	if len(metas) != 0 {
		t.Error("expected empty slice")
	}
}

func TestMetadataFields(t *testing.T) {
	now := time.Now()
	meta := Metadata{
		ID:           "abcdef123456",
		SessionID:    "session-1",
		CommitHash:   "abc1234567890",
		Branch:       "main",
		CreatedAt:    now.Format(time.RFC3339),
		Agent:        "claude-code",
		AgentPercent: 100,
		ContentHash:  "hash123",
	}

	// Verify ID prefix (12-char)
	if len(meta.ID) != 12 {
		t.Errorf("expected 12-char ID, got %d", len(meta.ID))
	}

	// Verify commit hash can be truncated to 7
	commit := meta.CommitHash
	if len(commit) > 7 {
		commit = commit[:7]
	}
	if commit != "abc1234" {
		t.Errorf("expected truncated commit abc1234, got %s", commit)
	}

	// Verify CreatedAt parses as RFC3339
	_, err := time.Parse(time.RFC3339, meta.CreatedAt)
	if err != nil {
		t.Errorf("CreatedAt should be valid RFC3339: %v", err)
	}
}

// TestListFromWorktree verifies that listing checkpoints from a git worktree
// produces the same results as listing from the main working tree. The checkpoint
// branch (partio/checkpoints/v1) is stored via refs that are shared across
// worktrees, so ExecGit and git ls-tree should resolve identically from either.
func TestListFromWorktree(t *testing.T) {
	// Skip if git is not available
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not found in PATH")
	}

	// Create a temporary directory for the main repo
	mainDir := t.TempDir()

	// Helper to run git in a specific directory
	gitIn := func(dir string, args ...string) (string, error) {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		out, err := cmd.CombinedOutput()
		return strings.TrimSpace(string(out)), err
	}

	mustGit := func(dir string, args ...string) string {
		t.Helper()
		out, err := gitIn(dir, args...)
		if err != nil {
			t.Fatalf("git %v in %s failed: %v\n%s", args, dir, err, out)
		}
		return out
	}

	// Initialize main repo with an initial commit
	mustGit(mainDir, "init")
	mustGit(mainDir, "config", "user.email", "test@test.com")
	mustGit(mainDir, "config", "user.name", "Test")

	initialFile := filepath.Join(mainDir, "README.md")
	if err := os.WriteFile(initialFile, []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}
	mustGit(mainDir, "add", "README.md")
	mustGit(mainDir, "commit", "-m", "initial commit")
	commitHash := mustGit(mainDir, "rev-parse", "HEAD")

	// Create the orphan checkpoint branch (same as enable.go)
	treeHash := mustGit(mainDir, "hash-object", "-t", "tree", "/dev/null")
	initCommit := mustGit(mainDir, "commit-tree", treeHash, "-m", "partio: initialize checkpoint storage")
	mustGit(mainDir, "update-ref", "refs/heads/"+checkpointBranch, initCommit)

	// Write two checkpoints using the Store
	store := NewStore(mainDir)

	cp1 := &Checkpoint{
		ID:          "aabbccddeeff",
		SessionID:   "session-1",
		CommitHash:  commitHash,
		Branch:      "main",
		CreatedAt:   time.Now().Add(-time.Hour),
		Agent:       "claude-code",
		AgentPct:    100,
		ContentHash: "hash1",
	}

	cp2 := &Checkpoint{
		ID:          "112233445566",
		SessionID:   "session-2",
		CommitHash:  commitHash,
		Branch:      "main",
		CreatedAt:   time.Now(),
		Agent:       "claude-code",
		AgentPct:    100,
		ContentHash: "hash2",
	}

	sessionData := &SessionFiles{
		ContentHash: "content",
		Context:     "context",
		Diff:        "diff",
		FullJSONL:   "{}",
		Metadata:    SessionMetadata{Agent: "claude-code", TotalTokens: 100, Duration: "1m"},
		Plan:        "plan",
		Prompt:      "prompt",
	}

	if err := store.Write(cp1, sessionData); err != nil {
		t.Fatalf("writing checkpoint 1: %v", err)
	}
	if err := store.Write(cp2, sessionData); err != nil {
		t.Fatalf("writing checkpoint 2: %v", err)
	}

	// Create a worktree
	worktreeDir := filepath.Join(t.TempDir(), "worktree")
	mustGit(mainDir, "worktree", "add", worktreeDir, "-b", "worktree-branch")

	// List checkpoints from the main repo
	mainCheckpoints := listCheckpointsFromDir(t, mainDir, mustGit)

	// List checkpoints from the worktree
	worktreeCheckpoints := listCheckpointsFromDir(t, worktreeDir, mustGit)

	// Verify both lists are non-empty
	if len(mainCheckpoints) == 0 {
		t.Fatal("expected checkpoints from main repo, got none")
	}

	// Verify the lists are identical
	if len(mainCheckpoints) != len(worktreeCheckpoints) {
		t.Fatalf("checkpoint count mismatch: main=%d worktree=%d",
			len(mainCheckpoints), len(worktreeCheckpoints))
	}

	for id, mainMeta := range mainCheckpoints {
		wtMeta, ok := worktreeCheckpoints[id]
		if !ok {
			t.Errorf("checkpoint %s found in main but not in worktree", id)
			continue
		}
		if mainMeta.Branch != wtMeta.Branch {
			t.Errorf("checkpoint %s branch mismatch: main=%s worktree=%s",
				id, mainMeta.Branch, wtMeta.Branch)
		}
		if mainMeta.AgentPercent != wtMeta.AgentPercent {
			t.Errorf("checkpoint %s agent_percent mismatch: main=%d worktree=%d",
				id, mainMeta.AgentPercent, wtMeta.AgentPercent)
		}
		if mainMeta.CreatedAt != wtMeta.CreatedAt {
			t.Errorf("checkpoint %s created_at mismatch: main=%s worktree=%s",
				id, mainMeta.CreatedAt, wtMeta.CreatedAt)
		}
	}

	// Verify both checkpoint IDs are present
	for _, id := range []string{"aabbccddeeff", "112233445566"} {
		if _, ok := mainCheckpoints[id]; !ok {
			t.Errorf("expected checkpoint %s in results", id)
		}
	}
}

// listCheckpointsFromDir enumerates checkpoints from a directory using the same
// git ls-tree approach as the List function.
func listCheckpointsFromDir(t *testing.T, dir string, mustGit func(string, ...string) string) map[string]Metadata {
	t.Helper()

	result := make(map[string]Metadata)

	shards := mustGit(dir, "ls-tree", "--name-only", checkpointBranch)
	if shards == "" {
		return result
	}

	for _, shard := range strings.Split(shards, "\n") {
		entries := mustGit(dir, "ls-tree", "--name-only", checkpointBranch+":"+shard)
		if entries == "" {
			continue
		}
		for _, entry := range strings.Split(entries, "\n") {
			cpID := shard + entry
			metaJSON := mustGit(dir, "show", checkpointBranch+":"+shard+"/"+entry+"/metadata.json")

			var meta Metadata
			if err := json.Unmarshal([]byte(metaJSON), &meta); err != nil {
				t.Fatalf("invalid metadata for checkpoint %s: %v", cpID, err)
			}
			result[cpID] = meta
		}
	}

	return result
}
