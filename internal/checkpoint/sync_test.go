package checkpoint

import (
	"os/exec"
	"strings"
	"testing"
)

// initTestRepo creates a temporary git repo with a checkpoint orphan branch
// and returns a Store pointing at it.
func initTestRepo(t *testing.T) *Store {
	t.Helper()
	dir := t.TempDir()

	run := func(args ...string) string {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		out, err := cmd.Output()
		if err != nil {
			t.Fatalf("git %v: %v", args, err)
		}
		return strings.TrimSpace(string(out))
	}

	run("init", "-b", "main")
	run("config", "user.email", "test@example.com")
	run("config", "user.name", "Test")

	// Create orphan checkpoint branch with empty tree
	emptyTree := run("mktree")
	initCommit := run("commit-tree", emptyTree, "-m", "init checkpoint branch")
	run("update-ref", "refs/heads/"+checkpointBranch, initCommit)

	return NewStore(dir)
}

// addCheckpointToStore writes a minimal checkpoint entry directly via git plumbing.
func addCheckpointToStore(t *testing.T, s *Store, id string) {
	t.Helper()
	shard := Shard(id)
	rest := Rest(id)

	metaHash, err := s.hashObject(`{"id":"` + id + `"}`)
	if err != nil {
		t.Fatalf("hashObject: %v", err)
	}

	cpTree, err := s.mktree([]treeEntry{
		{mode: "100644", typ: "blob", hash: metaHash, name: "metadata.json"},
	})
	if err != nil {
		t.Fatalf("mktree cp: %v", err)
	}

	currentTree, err := s.getCurrentTree()
	if err != nil {
		t.Fatalf("getCurrentTree: %v", err)
	}

	newRoot, err := s.addToTree(currentTree, shard, rest, cpTree)
	if err != nil {
		t.Fatalf("addToTree: %v", err)
	}

	parentCommit, err := s.git("rev-parse", checkpointBranch)
	if err != nil {
		t.Fatalf("rev-parse: %v", err)
	}

	commitHash, err := s.git("commit-tree", newRoot, "-p", parentCommit, "-m", "checkpoint: "+id)
	if err != nil {
		t.Fatalf("commit-tree: %v", err)
	}

	_, err = s.git("update-ref", "refs/heads/"+checkpointBranch, commitHash)
	if err != nil {
		t.Fatalf("update-ref: %v", err)
	}
}

func TestMergeTrees_DisjointShards(t *testing.T) {
	s := initTestRepo(t)

	// Add checkpoint with shard "ab" locally
	addCheckpointToStore(t, s, "ab1234567890")

	localTree, err := s.git("rev-parse", checkpointBranch+"^{tree}")
	if err != nil {
		t.Fatalf("rev-parse local tree: %v", err)
	}

	// Build a separate tree with shard "cd" to simulate a remote tree
	s2 := initTestRepo(t)
	addCheckpointToStore(t, s2, "cd1234567890")
	remoteTree, err := s2.git("rev-parse", checkpointBranch+"^{tree}")
	if err != nil {
		t.Fatalf("rev-parse remote tree: %v", err)
	}

	// Transplant the remote tree object into s's object store by fetching via pack
	// Instead, manually build a combined tree using s's merge logic
	// We need to create the remote shard tree in s's object store
	remoteShard, err := s2.git("rev-parse", checkpointBranch+"^{tree}:cd")
	if err != nil {
		t.Fatalf("rev-parse remote shard: %v", err)
	}

	// Add the remote shard tree hash to s's object store by re-creating it
	remoteMetaHash, err := s.hashObject(`{"id":"cd1234567890"}`)
	if err != nil {
		t.Fatalf("hashObject remote: %v", err)
	}
	_ = remoteShard

	remoteCpTree, err := s.mktree([]treeEntry{
		{mode: "100644", typ: "blob", hash: remoteMetaHash, name: "metadata.json"},
	})
	if err != nil {
		t.Fatalf("mktree remote cp: %v", err)
	}

	remoteShardTree, err := s.mktree([]treeEntry{
		{mode: "040000", typ: "tree", hash: remoteCpTree, name: "1234567890"},
	})
	if err != nil {
		t.Fatalf("mktree remote shard: %v", err)
	}
	_ = remoteTree

	// Build a simulated remote root tree in s's object store
	simulatedRemoteTree, err := s.mktree([]treeEntry{
		{mode: "040000", typ: "tree", hash: remoteShardTree, name: "cd"},
	})
	if err != nil {
		t.Fatalf("mktree simulated remote: %v", err)
	}

	merged, err := s.mergeTrees(localTree, simulatedRemoteTree)
	if err != nil {
		t.Fatalf("mergeTrees: %v", err)
	}

	// Merged tree should contain both "ab" and "cd" shards
	out, err := s.git("ls-tree", merged)
	if err != nil {
		t.Fatalf("ls-tree merged: %v", err)
	}

	if !strings.Contains(out, "ab") {
		t.Errorf("expected merged tree to contain shard 'ab', got:\n%s", out)
	}
	if !strings.Contains(out, "cd") {
		t.Errorf("expected merged tree to contain shard 'cd', got:\n%s", out)
	}
}

func TestMergeTrees_OverlappingShard(t *testing.T) {
	s := initTestRepo(t)

	// Add two checkpoints with the same shard "ab" locally
	addCheckpointToStore(t, s, "ab1111111111")

	localTree, err := s.git("rev-parse", checkpointBranch+"^{tree}")
	if err != nil {
		t.Fatalf("rev-parse local tree: %v", err)
	}

	// Build a simulated remote tree with different checkpoint in same shard "ab"
	remoteMetaHash, err := s.hashObject(`{"id":"ab2222222222"}`)
	if err != nil {
		t.Fatalf("hashObject: %v", err)
	}

	remoteCpTree, err := s.mktree([]treeEntry{
		{mode: "100644", typ: "blob", hash: remoteMetaHash, name: "metadata.json"},
	})
	if err != nil {
		t.Fatalf("mktree: %v", err)
	}

	remoteShardTree, err := s.mktree([]treeEntry{
		{mode: "040000", typ: "tree", hash: remoteCpTree, name: "2222222222"},
	})
	if err != nil {
		t.Fatalf("mktree shard: %v", err)
	}

	simulatedRemoteTree, err := s.mktree([]treeEntry{
		{mode: "040000", typ: "tree", hash: remoteShardTree, name: "ab"},
	})
	if err != nil {
		t.Fatalf("mktree remote: %v", err)
	}

	merged, err := s.mergeTrees(localTree, simulatedRemoteTree)
	if err != nil {
		t.Fatalf("mergeTrees: %v", err)
	}

	// The merged shard should contain both checkpoint entries
	mergedShardTree, err := s.git("rev-parse", merged+":ab")
	if err != nil {
		t.Fatalf("rev-parse merged shard: %v", err)
	}

	out, err := s.git("ls-tree", mergedShardTree)
	if err != nil {
		t.Fatalf("ls-tree merged shard: %v", err)
	}

	if !strings.Contains(out, "1111111111") {
		t.Errorf("expected merged shard to contain local checkpoint, got:\n%s", out)
	}
	if !strings.Contains(out, "2222222222") {
		t.Errorf("expected merged shard to contain remote checkpoint, got:\n%s", out)
	}
}

func TestMergeTrees_LocalWinsOnConflict(t *testing.T) {
	s := initTestRepo(t)

	// Add checkpoint locally
	addCheckpointToStore(t, s, "ab1234567890")

	localTree, err := s.git("rev-parse", checkpointBranch+"^{tree}")
	if err != nil {
		t.Fatalf("rev-parse local tree: %v", err)
	}

	// Build a simulated remote tree with the same checkpoint ID but different content
	remoteMetaHash, err := s.hashObject(`{"id":"ab1234567890","extra":"remote"}`)
	if err != nil {
		t.Fatalf("hashObject: %v", err)
	}

	remoteCpTree, err := s.mktree([]treeEntry{
		{mode: "100644", typ: "blob", hash: remoteMetaHash, name: "metadata.json"},
	})
	if err != nil {
		t.Fatalf("mktree: %v", err)
	}

	remoteShardTree, err := s.mktree([]treeEntry{
		{mode: "040000", typ: "tree", hash: remoteCpTree, name: Rest("ab1234567890")},
	})
	if err != nil {
		t.Fatalf("mktree shard: %v", err)
	}

	simulatedRemoteTree, err := s.mktree([]treeEntry{
		{mode: "040000", typ: "tree", hash: remoteShardTree, name: "ab"},
	})
	if err != nil {
		t.Fatalf("mktree remote: %v", err)
	}

	merged, err := s.mergeTrees(localTree, simulatedRemoteTree)
	if err != nil {
		t.Fatalf("mergeTrees: %v", err)
	}

	// Local tree should be preserved (same hash as before merging)
	if merged != localTree {
		// The shard "ab" should still contain only one "34567890" entry (local wins)
		mergedShardTree, err := s.git("rev-parse", merged+":ab")
		if err != nil {
			t.Fatalf("rev-parse merged shard: %v", err)
		}

		localShardTree, err := s.git("rev-parse", localTree+":ab")
		if err != nil {
			t.Fatalf("rev-parse local shard: %v", err)
		}

		if mergedShardTree != localShardTree {
			t.Errorf("expected local shard tree to be preserved, got different tree")
		}
	}
}

func TestIsAncestor(t *testing.T) {
	s := initTestRepo(t)

	// Initial commit on checkpoint branch
	commit1, err := s.git("rev-parse", checkpointBranch)
	if err != nil {
		t.Fatalf("rev-parse: %v", err)
	}

	addCheckpointToStore(t, s, "ab1234567890")

	commit2, err := s.git("rev-parse", checkpointBranch)
	if err != nil {
		t.Fatalf("rev-parse: %v", err)
	}

	if !s.isAncestor(commit1, commit2) {
		t.Error("commit1 should be ancestor of commit2")
	}
	if s.isAncestor(commit2, commit1) {
		t.Error("commit2 should not be ancestor of commit1")
	}
}
