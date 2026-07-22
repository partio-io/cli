package checkpoint

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"testing"
)

// initTestRepo initialises a bare git repo in dir and returns a Store pointing
// at it.  It creates the orphan checkpoint branch and writes the given metadata
// entries so that FindByBranch can walk them.
func initTestRepo(t *testing.T, metas []Metadata) (string, *Store) {
	t.Helper()
	dir := t.TempDir()

	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}

	run("init", "-b", "main")
	run("config", "user.email", "test@example.com")
	run("config", "user.name", "Test")

	// Create an initial commit so we have a valid repo
	run("commit", "--allow-empty", "-m", "init")

	s := NewStore(dir)

	for _, meta := range metas {
		// Write each checkpoint to the orphan branch via the same git-plumbing
		// path that the real store uses: hash-object the JSON then build the
		// shard/rest/metadata.json tree.
		metaBytes, err := json.Marshal(meta)
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}

		shard := Shard(meta.ID)
		rest := Rest(meta.ID)

		// hash-object the metadata JSON
		metaHash, err := s.hashObject(string(metaBytes))
		if err != nil {
			t.Fatalf("hash-object metadata: %v", err)
		}

		// Build the checkpoint subtree (just metadata.json for the test)
		cpTree, err := s.mktree([]treeEntry{
			{mode: "100644", typ: "blob", hash: metaHash, name: "metadata.json"},
		})
		if err != nil {
			t.Fatalf("mktree checkpoint: %v", err)
		}

		// Build the shard tree
		shardTree, err := s.mktree([]treeEntry{
			{mode: "040000", typ: "tree", hash: cpTree, name: rest},
		})
		if err != nil {
			t.Fatalf("mktree shard: %v", err)
		}

		// Check if checkpoint branch already exists
		_, branchExists := s.git("rev-parse", "--verify", checkpointBranch)

		var rootTree string
		if branchExists == nil {
			// Merge the new shard into the existing root tree
			currentTree, err := s.getCurrentTree()
			if err != nil {
				t.Fatalf("getCurrentTree: %v", err)
			}
			rootTree, err = s.addToTree(currentTree, shard, rest, cpTree)
			if err != nil {
				t.Fatalf("addToTree: %v", err)
			}
		} else {
			// First checkpoint – root tree is just the one shard
			rootTree, err = s.mktree([]treeEntry{
				{mode: "040000", typ: "tree", hash: shardTree, name: shard},
			})
			if err != nil {
				t.Fatalf("mktree root: %v", err)
			}
		}

		// Create or update the checkpoint commit
		var commitArgs []string
		if branchExists == nil {
			parent, _ := s.git("rev-parse", checkpointBranch)
			commitArgs = []string{"commit-tree", rootTree, "-p", parent, "-m", "test checkpoint " + meta.ID}
		} else {
			commitArgs = []string{"commit-tree", rootTree, "-m", "test checkpoint " + meta.ID}
		}

		commitHash, err := s.git(commitArgs...)
		if err != nil {
			t.Fatalf("commit-tree: %v", err)
		}

		_, err = s.git("update-ref", "refs/heads/"+checkpointBranch, commitHash)
		if err != nil {
			t.Fatalf("update-ref: %v", err)
		}
	}

	return dir, s
}

func makeID(prefix string) string {
	// Pad to 12 hex chars for a valid checkpoint ID
	return fmt.Sprintf("%s%s", prefix, strings.Repeat("0", 12-len(prefix)))
}

func TestFindByBranch(t *testing.T) {
	id1 := makeID("aabbcc")
	id2 := makeID("ddeeff")
	id3 := makeID("112233")

	metas := []Metadata{
		{ID: id1, Branch: "feature/foo", CommitHash: "abc", CreatedAt: "2024-01-01T00:00:00Z"},
		{ID: id2, Branch: "feature/foo", CommitHash: "def", CreatedAt: "2024-01-02T00:00:00Z"},
		{ID: id3, Branch: "main", CommitHash: "ghi", CreatedAt: "2024-01-03T00:00:00Z"},
	}

	tests := []struct {
		name     string
		branch   string
		wantIDs  []string
		wantNone bool
	}{
		{
			name:    "single match",
			branch:  "main",
			wantIDs: []string{id3},
		},
		{
			name:    "multiple matches",
			branch:  "feature/foo",
			wantIDs: []string{id1, id2},
		},
		{
			name:     "no match",
			branch:   "nonexistent",
			wantNone: true,
		},
		{
			name:     "partial name does not match",
			branch:   "feature",
			wantNone: true,
		},
	}

	dir, _ := initTestRepo(t, metas)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FindByBranch(dir, tt.branch)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.wantNone {
				if len(got) != 0 {
					t.Errorf("expected no matches, got %d", len(got))
				}
				return
			}

			// Build a set of returned IDs for order-independent comparison
			gotIDs := make(map[string]bool)
			for _, m := range got {
				gotIDs[m.ID] = true
			}

			if len(got) != len(tt.wantIDs) {
				t.Errorf("expected %d matches, got %d", len(tt.wantIDs), len(got))
			}
			for _, wantID := range tt.wantIDs {
				if !gotIDs[wantID] {
					t.Errorf("expected ID %s in results, but not found", wantID)
				}
			}
		})
	}
}

func TestFindByBranchNoBranch(t *testing.T) {
	dir := t.TempDir()
	cmd := exec.Command("git", "init", "-b", "main")
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git init: %v\n%s", err, out)
	}

	got, err := FindByBranch(dir, "anything")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty result when no checkpoint branch, got %d", len(got))
	}
}
