package checkpoint

import (
	"fmt"
	"os/exec"
	"testing"
	"time"
)

// newImportTestRepo creates a repository with one commit and an initialised
// checkpoint branch.
func newImportTestRepo(t *testing.T) *Store {
	t.Helper()
	dir := t.TempDir()
	for _, args := range [][]string{
		{"init"},
		{"config", "user.email", "test@example.com"},
		{"config", "user.name", "Test"},
		{"commit", "--allow-empty", "-m", "initial"},
	} {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	s := NewStore(dir)
	if err := s.InitBranch(); err != nil {
		t.Fatalf("InitBranch: %v", err)
	}
	return s
}

func mustGit(t *testing.T, s *Store, args ...string) string {
	t.Helper()
	out, err := s.git(args...)
	if err != nil {
		t.Fatalf("git %v: %v", args, err)
	}
	return out
}

func writeImportCheckpoint(t *testing.T, s *Store, id, prompt string) {
	t.Helper()
	cp := &Checkpoint{ID: id, Branch: "main", CommitHash: "deadbeef", CreatedAt: time.Now(), Agent: "claude-code"}
	if err := s.Write(cp, &SessionFiles{Prompt: prompt}); err != nil {
		t.Fatalf("Write: %v", err)
	}
}

// moveBranchToSource parks the checkpoint branch at refs/test/source and
// starts a fresh, empty checkpoint branch, as if the source were a fork.
func moveBranchToSource(t *testing.T, s *Store) string {
	t.Helper()
	const ref = "refs/test/source"
	mustGit(t, s, "update-ref", ref, "refs/heads/"+checkpointBranch)
	mustGit(t, s, "update-ref", "-d", "refs/heads/"+checkpointBranch)
	if err := s.InitBranch(); err != nil {
		t.Fatalf("InitBranch: %v", err)
	}
	return ref
}

// sourceRefWithTree builds a source ref whose root holds tree under id.
func sourceRefWithTree(t *testing.T, s *Store, id string, mode, typ, hash string) string {
	t.Helper()
	shard, err := s.mktree([]treeEntry{{mode: mode, typ: typ, hash: hash, name: Rest(id)}})
	if err != nil {
		t.Fatal(err)
	}
	root, err := s.mktree([]treeEntry{{mode: "040000", typ: "tree", hash: shard, name: Shard(id)}})
	if err != nil {
		t.Fatal(err)
	}
	commit := mustGit(t, s, "commit-tree", root, "-m", "source")
	mustGit(t, s, "update-ref", "refs/test/source", commit)
	return "refs/test/source"
}

func blob(t *testing.T, s *Store, content string) string {
	t.Helper()
	h, err := s.hashObject(content)
	if err != nil {
		t.Fatal(err)
	}
	return h
}

func tree(t *testing.T, s *Store, entries ...treeEntry) string {
	t.Helper()
	h, err := s.mktree(entries)
	if err != nil {
		t.Fatal(err)
	}
	return h
}

func testMessage(synced []string) string {
	return fmt.Sprintf("sync: %d", len(synced))
}

func outcomes(results []ImportResult) map[string]ImportOutcome {
	m := map[string]ImportOutcome{}
	for _, r := range results {
		m[r.ID] = r.Outcome
	}
	return m
}

func TestImportSyncsReferencedCheckpoints(t *testing.T) {
	s := newImportTestRepo(t)
	writeImportCheckpoint(t, s, "aabbccddeeff", "first")
	writeImportCheckpoint(t, s, "aa0000000001", "same shard")
	writeImportCheckpoint(t, s, "112233445566", "unreferenced")
	src := moveBranchToSource(t, s)
	parent := mustGit(t, s, "rev-parse", "refs/heads/"+checkpointBranch)

	results, err := s.Import(src, []string{"aabbccddeeff", "aa0000000001", "999999999999"}, testMessage)
	if err != nil {
		t.Fatalf("Import: %v", err)
	}
	got := outcomes(results)
	want := map[string]ImportOutcome{"aabbccddeeff": ImportSynced, "aa0000000001": ImportSynced, "999999999999": ImportMissing}
	if fmt.Sprint(got) != fmt.Sprint(want) {
		t.Fatalf("outcomes = %v, want %v", got, want)
	}

	for _, id := range []string{"aabbccddeeff", "aa0000000001"} {
		path := ":" + Shard(id) + "/" + Rest(id)
		if l, f := mustGit(t, s, "rev-parse", checkpointBranch+path), mustGit(t, s, "rev-parse", src+path); l != f {
			t.Errorf("%s tree = %s, want source tree %s", id, l, f)
		}
	}
	if _, err := s.git("rev-parse", "--verify", checkpointBranch+":11/2233445566"); err == nil {
		t.Error("unreferenced checkpoint was copied")
	}
	if names := mustGit(t, s, "ls-tree", "--name-only", checkpointBranch+":aa"); names != "0000000001\nbbccddeeff" {
		t.Errorf("shard aa entries = %q", names)
	}

	if parents := mustGit(t, s, "log", "-1", "--format=%P", checkpointBranch); parents != parent {
		t.Errorf("parents = %q, want only %q", parents, parent)
	}
	if msg := mustGit(t, s, "log", "-1", "--format=%B", checkpointBranch); msg != "sync: 2" {
		t.Errorf("message = %q", msg)
	}
}

func TestImportIsIdempotent(t *testing.T) {
	s := newImportTestRepo(t)
	writeImportCheckpoint(t, s, "aabbccddeeff", "first")
	src := moveBranchToSource(t, s)

	if _, err := s.Import(src, []string{"aabbccddeeff"}, testMessage); err != nil {
		t.Fatalf("first Import: %v", err)
	}
	tip := mustGit(t, s, "rev-parse", checkpointBranch)

	results, err := s.Import(src, []string{"aabbccddeeff"}, testMessage)
	if err != nil {
		t.Fatalf("second Import: %v", err)
	}
	if results[0].Outcome != ImportPresent {
		t.Errorf("outcome = %v, want present", results[0].Outcome)
	}
	if got := mustGit(t, s, "rev-parse", checkpointBranch); got != tip {
		t.Error("second Import wrote a commit")
	}
}

func TestImportKeepsConflictingLocalCopy(t *testing.T) {
	s := newImportTestRepo(t)
	writeImportCheckpoint(t, s, "aabbccddeeff", "fork version")
	src := moveBranchToSource(t, s)
	writeImportCheckpoint(t, s, "aabbccddeeff", "upstream version")
	tip := mustGit(t, s, "rev-parse", checkpointBranch)
	localTree := mustGit(t, s, "rev-parse", checkpointBranch+":aa/bbccddeeff")

	results, err := s.Import(src, []string{"aabbccddeeff"}, testMessage)
	if err != nil {
		t.Fatalf("Import: %v", err)
	}
	if results[0].Outcome != ImportConflict {
		t.Errorf("outcome = %v, want conflict", results[0].Outcome)
	}
	if got := mustGit(t, s, "rev-parse", checkpointBranch); got != tip {
		t.Error("conflict wrote a commit")
	}
	if got := mustGit(t, s, "rev-parse", checkpointBranch+":aa/bbccddeeff"); got != localTree {
		t.Error("local checkpoint was overwritten")
	}
}

func TestImportRejectsInvalidCheckpoints(t *testing.T) {
	const id = "aabbccddeeff"
	type built struct{ mode, typ, hash string }
	tests := []struct {
		name  string
		build func(t *testing.T, s *Store) built
	}{
		{"symlink entry", func(t *testing.T, s *Store) built {
			return built{"040000", "tree", tree(t, s,
				treeEntry{"100644", "blob", blob(t, s, `{"id":"`+id+`"}`), "metadata.json"},
				treeEntry{"120000", "blob", blob(t, s, "/etc/passwd"), "link"})}
		}},
		{"nested symlink", func(t *testing.T, s *Store) built {
			sub := tree(t, s, treeEntry{"120000", "blob", blob(t, s, "/etc/passwd"), "prompt.txt"})
			return built{"040000", "tree", tree(t, s,
				treeEntry{"100644", "blob", blob(t, s, `{"id":"`+id+`"}`), "metadata.json"},
				treeEntry{"040000", "tree", sub, "0"})}
		}},
		{"gitlink entry", func(t *testing.T, s *Store) built {
			return built{"040000", "tree", tree(t, s,
				treeEntry{"100644", "blob", blob(t, s, `{"id":"`+id+`"}`), "metadata.json"},
				treeEntry{"160000", "commit", mustGit(t, s, "rev-parse", "HEAD"), "sub"})}
		}},
		{"executable entry", func(t *testing.T, s *Store) built {
			return built{"040000", "tree", tree(t, s,
				treeEntry{"100644", "blob", blob(t, s, `{"id":"`+id+`"}`), "metadata.json"},
				treeEntry{"100755", "blob", blob(t, s, "#!/bin/sh"), "run.sh"})}
		}},
		{"missing metadata.json", func(t *testing.T, s *Store) built {
			return built{"040000", "tree", tree(t, s, treeEntry{"100644", "blob", blob(t, s, "x"), "prompt.txt"})}
		}},
		{"mismatched metadata ID", func(t *testing.T, s *Store) built {
			return built{"040000", "tree", tree(t, s, treeEntry{"100644", "blob", blob(t, s, `{"id":"112233445566"}`), "metadata.json"})}
		}},
		{"unparseable metadata", func(t *testing.T, s *Store) built {
			return built{"040000", "tree", tree(t, s, treeEntry{"100644", "blob", blob(t, s, "not json"), "metadata.json"})}
		}},
		{"checkpoint is a blob", func(t *testing.T, s *Store) built {
			return built{"100644", "blob", blob(t, s, `{"id":"`+id+`"}`)}
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newImportTestRepo(t)
			b := tt.build(t, s)
			src := sourceRefWithTree(t, s, id, b.mode, b.typ, b.hash)
			tip := mustGit(t, s, "rev-parse", checkpointBranch)

			results, err := s.Import(src, []string{id}, testMessage)
			if err != nil {
				t.Fatalf("Import: %v", err)
			}
			if results[0].Outcome != ImportRejected || results[0].Reason == "" {
				t.Errorf("result = %+v, want rejected with a reason", results[0])
			}
			if got := mustGit(t, s, "rev-parse", checkpointBranch); got != tip {
				t.Error("rejected checkpoint wrote a commit")
			}
		})
	}
}

func TestImportAcceptsMinimalValidTree(t *testing.T) {
	const id = "aabbccddeeff"
	s := newImportTestRepo(t)
	sub := tree(t, s, treeEntry{"100644", "blob", blob(t, s, "p"), "prompt.txt"})
	cp := tree(t, s,
		treeEntry{"100644", "blob", blob(t, s, `{"id":"`+id+`"}`), "metadata.json"},
		treeEntry{"040000", "tree", sub, "0"})
	src := sourceRefWithTree(t, s, id, "040000", "tree", cp)

	results, err := s.Import(src, []string{id}, testMessage)
	if err != nil {
		t.Fatalf("Import: %v", err)
	}
	if results[0].Outcome != ImportSynced {
		t.Fatalf("result = %+v, want synced", results[0])
	}
	if got := mustGit(t, s, "rev-parse", checkpointBranch+":aa/bbccddeeff"); got != cp {
		t.Errorf("tree = %s, want %s", got, cp)
	}
}

func TestImportBranchCreation(t *testing.T) {
	tests := []struct {
		name       string
		ids        []string
		wantBranch bool
	}{
		{name: "creates the branch when a checkpoint is synced", ids: []string{"aabbccddeeff"}, wantBranch: true},
		{name: "leaves no branch when nothing is synced", ids: []string{"999999999999"}, wantBranch: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newImportTestRepo(t)
			writeImportCheckpoint(t, s, "aabbccddeeff", "first")
			src := moveBranchToSource(t, s)
			mustGit(t, s, "update-ref", "-d", "refs/heads/"+checkpointBranch)

			if _, err := s.Import(src, tt.ids, testMessage); err != nil {
				t.Fatalf("Import: %v", err)
			}
			exists, err := s.branchExists()
			if err != nil {
				t.Fatal(err)
			}
			if exists != tt.wantBranch {
				t.Fatalf("branch exists = %v, want %v", exists, tt.wantBranch)
			}
			if !exists {
				return
			}
			log := mustGit(t, s, "log", "--format=%s", checkpointBranch)
			if want := "sync: 1\npartio: initialize checkpoint storage"; log != want {
				t.Errorf("log = %q, want %q", log, want)
			}
		})
	}
}

func TestInitBranch(t *testing.T) {
	s := newImportTestRepo(t)
	tip := mustGit(t, s, "rev-parse", checkpointBranch)
	if parents := mustGit(t, s, "log", "-1", "--format=%P", tip); parents != "" {
		t.Errorf("initial commit has parents %q", parents)
	}
	if msg := mustGit(t, s, "log", "-1", "--format=%s", tip); msg != "partio: initialize checkpoint storage" {
		t.Errorf("message = %q", msg)
	}
	if entries := mustGit(t, s, "ls-tree", tip); entries != "" {
		t.Errorf("initial tree not empty: %q", entries)
	}

	if err := s.InitBranch(); err != nil {
		t.Fatalf("second InitBranch: %v", err)
	}
	if got := mustGit(t, s, "rev-parse", checkpointBranch); got != tip {
		t.Error("InitBranch replaced an existing branch")
	}
}
