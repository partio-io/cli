package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/partio-io/cli/internal/checkpoint"
)

// syncForkFixture is a fork repository with checkpoints, an upstream
// repository, and a fake GitHub API describing pull requests on up/repo:
//
//	#1 from the fork contrib/repo
//	#3 from a deleted fork (null head repository)
//	#4 from up/repo itself (not a fork)
//	#5 from a fork with an empty clone URL
//	#6 with no head repository field at all
//
// Every pull request lists messages as its commits and reports
// commitCount as its total.
type syncForkFixture struct {
	fork, upstream string
	cloneURL       string
	messages       []string
	commitCount    int
}

var syncForkCommitsPath = regexp.MustCompile(`^/repos/up/repo/pulls/[0-9]+/commits$`)

func newSyncForkFixture(t *testing.T, forkIDs []string) *syncForkFixture {
	t.Helper()
	f := &syncForkFixture{fork: t.TempDir(), upstream: t.TempDir()}
	rewindInitTestRepo(t, f.fork)
	rewindInitTestRepo(t, f.upstream)
	for _, id := range forkIDs {
		writeSyncForkCheckpoint(t, f.fork, id, "fork")
	}
	f.cloneURL = "file://" + f.fork

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if syncForkCommitsPath.MatchString(r.URL.Path) {
			var commits []map[string]any
			if r.URL.Query().Get("page") == "1" {
				for i, m := range f.messages {
					commits = append(commits, map[string]any{"sha": fmt.Sprint(i), "commit": map[string]string{"message": m}})
				}
			}
			_ = json.NewEncoder(w).Encode(commits)
			return
		}
		head := map[string]any{"sha": "abc"}
		switch r.URL.Path {
		case "/repos/up/repo/pulls/1":
			head["repo"] = map[string]any{"full_name": "contrib/repo", "clone_url": f.cloneURL}
		case "/repos/up/repo/pulls/3":
			head["repo"] = nil
		case "/repos/up/repo/pulls/4":
			// The clone URL leads nowhere, so any fetch would fail.
			head["repo"] = map[string]any{"full_name": "up/repo", "clone_url": "file://" + f.upstream + "/missing"}
		case "/repos/up/repo/pulls/5":
			head["repo"] = map[string]any{"full_name": "contrib/repo", "clone_url": ""}
		case "/repos/up/repo/pulls/6":
		default:
			http.NotFound(w, r)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"commits": f.commitCount,
			"head":    head,
			"base":    map[string]any{"repo": map[string]any{"full_name": "up/repo"}},
		})
	}))
	t.Cleanup(srv.Close)

	t.Setenv("HOME", t.TempDir())
	t.Setenv("GITHUB_API_URL", srv.URL)
	t.Setenv("GH_TOKEN", "")
	t.Setenv("GITHUB_TOKEN", "")
	t.Chdir(f.upstream)
	return f
}

func writeSyncForkCheckpoint(t *testing.T, dir, id, prompt string) {
	t.Helper()
	s := checkpoint.NewStore(dir)
	cp := &checkpoint.Checkpoint{ID: id, Branch: "feature", CommitHash: "deadbeef", CreatedAt: time.Now(), Agent: "claude-code"}
	if err := s.Write(cp, &checkpoint.SessionFiles{Prompt: prompt}); err != nil {
		t.Fatalf("Store.Write: %v", err)
	}
}

func gitOut(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("git %v: %v", args, err)
	}
	return strings.TrimSpace(string(out))
}

func gitOutInput(t *testing.T, dir, stdin string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Stdin = strings.NewReader(stdin)
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("git %v: %v", args, err)
	}
	return strings.TrimSpace(string(out))
}

func runSyncForkCmd(t *testing.T, ref string) (string, error) {
	t.Helper()
	var err error
	out := captureOutput(t, func() {
		root := newRootCmd()
		root.SetArgs([]string{"sync-fork", ref})
		err = root.Execute()
	})
	return out, err
}

func assertNoSyncForkRefs(t *testing.T, dir string) {
	t.Helper()
	if refs := gitOut(t, dir, "for-each-ref", syncForkRefPrefix); refs != "" {
		t.Errorf("temporary refs left behind:\n%s", refs)
	}
}

func TestSyncFork(t *testing.T) {
	f := newSyncForkFixture(t, []string{"aabbccddeeff", "112233445566", "ffffffffffff"})
	f.messages = []string{
		"feat: one\n\nPartio-Checkpoint: aabbccddeeff\n",
		"feat: two\n\nPartio-Checkpoint: 112233445566\nPartio-Attribution: 100% agent\n",
		"feat: three\n\nPartio-Checkpoint: 999999999999\n",
	}
	forkRefsBefore := gitOut(t, f.fork, "show-ref")
	tipBefore := gitOut(t, f.upstream, "rev-parse", "partio/checkpoints/v1")

	out, err := runSyncForkCmd(t, "https://github.com/up/repo/pull/1/files")
	if err != nil {
		t.Fatalf("sync-fork: %v", err)
	}
	if !strings.Contains(out, "synced 2, already present 0, conflicts 0, missing 1") {
		t.Errorf("summary missing; got:\n%s", out)
	}

	for _, id := range []string{"aabbccddeeff", "112233445566"} {
		data, err := checkpoint.Read(id)
		if err != nil {
			t.Fatalf("checkpoint.Read(%s): %v", id, err)
		}
		if data.Metadata.ID != id || data.Prompt != "fork" {
			t.Errorf("checkpoint %s = %+v", id, data)
		}
	}
	if _, err := checkpoint.Read("ffffffffffff"); err == nil {
		t.Error("unreferenced fork checkpoint was synced")
	}

	if parents := gitOut(t, f.upstream, "log", "-1", "--format=%P", "partio/checkpoints/v1"); parents != tipBefore {
		t.Errorf("sync commit parents = %q, want %q", parents, tipBefore)
	}
	msg := gitOut(t, f.upstream, "log", "-1", "--format=%B", "partio/checkpoints/v1")
	for _, want := range []string{"sync-fork: up/repo#1 (2 checkpoints)", "contrib/repo", "aabbccddeeff", "112233445566"} {
		if !strings.Contains(msg, want) {
			t.Errorf("commit message missing %q:\n%s", want, msg)
		}
	}
	assertNoSyncForkRefs(t, f.upstream)

	// A second run finds everything present and writes nothing.
	tip := gitOut(t, f.upstream, "rev-parse", "partio/checkpoints/v1")
	out, err = runSyncForkCmd(t, "up/repo#1")
	if err != nil {
		t.Fatalf("second sync-fork: %v", err)
	}
	if !strings.Contains(out, "synced 0, already present 2, conflicts 0, missing 1") {
		t.Errorf("second summary; got:\n%s", out)
	}
	if got := gitOut(t, f.upstream, "rev-parse", "partio/checkpoints/v1"); got != tip {
		t.Error("second run wrote a commit")
	}
	assertNoSyncForkRefs(t, f.upstream)

	if after := gitOut(t, f.fork, "show-ref"); after != forkRefsBefore {
		t.Errorf("fork refs changed:\nbefore:\n%s\nafter:\n%s", forkRefsBefore, after)
	}
}

func TestSyncForkKeepsConflictingUpstreamCheckpoint(t *testing.T) {
	f := newSyncForkFixture(t, []string{"aabbccddeeff"})
	f.messages = []string{"feat: one\n\nPartio-Checkpoint: aabbccddeeff\n"}
	writeSyncForkCheckpoint(t, f.upstream, "aabbccddeeff", "upstream")
	tip := gitOut(t, f.upstream, "rev-parse", "partio/checkpoints/v1")

	out, err := runSyncForkCmd(t, "up/repo#1")
	if err != nil {
		t.Fatalf("sync-fork: %v", err)
	}
	if !strings.Contains(out, "synced 0, already present 0, conflicts 1, missing 0") {
		t.Errorf("summary; got:\n%s", out)
	}
	if got := gitOut(t, f.upstream, "rev-parse", "partio/checkpoints/v1"); got != tip {
		t.Error("conflict wrote a commit")
	}
	data, err := checkpoint.Read("aabbccddeeff")
	if err != nil {
		t.Fatal(err)
	}
	if data.Prompt != "upstream" {
		t.Errorf("prompt = %q, want the upstream copy", data.Prompt)
	}
	assertNoSyncForkRefs(t, f.upstream)
}

func TestSyncForkCreatesMissingBranch(t *testing.T) {
	f := newSyncForkFixture(t, []string{"aabbccddeeff"})
	f.messages = []string{"feat: one\n\nPartio-Checkpoint: aabbccddeeff\n"}
	gitOut(t, f.upstream, "update-ref", "-d", "refs/heads/partio/checkpoints/v1")

	if _, err := runSyncForkCmd(t, "up/repo#1"); err != nil {
		t.Fatalf("sync-fork: %v", err)
	}
	if _, err := checkpoint.Read("aabbccddeeff"); err != nil {
		t.Fatalf("checkpoint.Read: %v", err)
	}
	if log := gitOut(t, f.upstream, "log", "--format=%s", "partio/checkpoints/v1"); log != "sync-fork: up/repo#1 (1 checkpoints)\npartio: initialize checkpoint storage" {
		t.Errorf("log = %q", log)
	}
}

func TestSyncForkFailures(t *testing.T) {
	tests := []struct {
		name    string
		ref     string
		setup   func(t *testing.T, f *syncForkFixture)
		wantErr string
	}{
		{name: "unparseable reference", ref: "not-a-pr", wantErr: "invalid pull request reference"},
		{name: "pull request not found", ref: "up/repo#2", wantErr: "404"},
		{
			name:    "fork is unreachable",
			ref:     "up/repo#1",
			wantErr: "checking partio/checkpoints/v1",
			setup: func(t *testing.T, f *syncForkFixture) {
				f.cloneURL = "file://" + t.TempDir() + "/missing"
			},
		},
		{
			// ls-remote sees the branch, but the fetch cannot send the
			// tip's tree: a real fetch error, not an absent branch.
			name:    "fetch fails on a corrupt fork",
			ref:     "up/repo#1",
			wantErr: "fetching",
			setup: func(t *testing.T, f *syncForkFixture) {
				tree := gitOut(t, f.fork, "rev-parse", "partio/checkpoints/v1^{tree}")
				if err := os.Remove(filepath.Join(f.fork, ".git", "objects", tree[:2], tree[2:])); err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			// Fails after the fork was fetched, so cleanup must run on the
			// error path. git refuses to point a branch at a blob, so the
			// loose ref file is written directly.
			name:    "corrupt local checkpoint branch",
			ref:     "up/repo#1",
			wantErr: "importing checkpoints",
			setup: func(t *testing.T, f *syncForkFixture) {
				blob := gitOutInput(t, f.upstream, "not a commit", "hash-object", "-w", "--stdin")
				refFile := filepath.Join(f.upstream, ".git", "refs", "heads", "partio", "checkpoints", "v1")
				if err := os.WriteFile(refFile, []byte(blob+"\n"), 0o644); err != nil {
					t.Fatal(err)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newSyncForkFixture(t, []string{"aabbccddeeff"})
			f.messages = []string{"feat: one\n\nPartio-Checkpoint: aabbccddeeff\n"}
			if tt.setup != nil {
				tt.setup(t, f)
			}
			tip := gitOut(t, f.upstream, "rev-parse", "refs/heads/partio/checkpoints/v1")

			_, err := runSyncForkCmd(t, tt.ref)
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("error = %v, want one containing %q", err, tt.wantErr)
			}
			if got := gitOut(t, f.upstream, "rev-parse", "refs/heads/partio/checkpoints/v1"); got != tip {
				t.Error("failed run moved the checkpoint branch")
			}
			assertNoSyncForkRefs(t, f.upstream)
		})
	}
}

func TestSyncForkNothingToDo(t *testing.T) {
	tests := []struct {
		name       string
		ref        string
		messages   []string
		setup      func(t *testing.T, f *syncForkFixture)
		wantReason string
	}{
		{name: "same-repo pull request", ref: "up/repo#4", wantReason: "is not from a fork"},
		{name: "deleted fork", ref: "up/repo#3", wantReason: "fork was deleted"},
		{name: "empty clone URL", ref: "up/repo#5", wantReason: "fork was deleted"},
		{name: "missing head repository", ref: "up/repo#6", wantReason: "fork was deleted"},
		{
			name:       "fork has no checkpoint branch",
			ref:        "up/repo#1",
			wantReason: "fork contrib/repo has no partio/checkpoints/v1 branch",
			setup: func(t *testing.T, f *syncForkFixture) {
				gitOut(t, f.fork, "update-ref", "-d", "refs/heads/partio/checkpoints/v1")
			},
		},
		{
			name:       "no trailers",
			ref:        "up/repo#1",
			messages:   []string{"feat: one\n\nSigned-off-by: A <a@example.com>\n"},
			wantReason: "carries a valid Partio-Checkpoint trailer",
		},
		{
			name:       "only invalid trailers",
			ref:        "up/repo#1",
			messages:   []string{"feat: one\n\nPartio-Checkpoint: ../../etc\n"},
			wantReason: "carries a valid Partio-Checkpoint trailer",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newSyncForkFixture(t, []string{"aabbccddeeff"})
			f.messages = []string{"feat: one\n\nPartio-Checkpoint: aabbccddeeff\n"}
			if tt.messages != nil {
				f.messages = tt.messages
			}
			if tt.setup != nil {
				tt.setup(t, f)
			}
			tip := gitOut(t, f.upstream, "rev-parse", "refs/heads/partio/checkpoints/v1")

			out, err := runSyncForkCmd(t, tt.ref)
			if err != nil {
				t.Fatalf("sync-fork: %v", err)
			}
			out = strings.TrimSpace(out)
			if strings.Count(out, "\n") != 0 || !strings.HasPrefix(out, "nothing to sync: ") || !strings.Contains(out, tt.wantReason) {
				t.Errorf("output = %q, want one line containing %q", out, tt.wantReason)
			}
			if got := gitOut(t, f.upstream, "rev-parse", "refs/heads/partio/checkpoints/v1"); got != tip {
				t.Error("nothing-to-do run moved the checkpoint branch")
			}
			assertNoSyncForkRefs(t, f.upstream)
		})
	}
}

func TestSyncForkSummaryWarnings(t *testing.T) {
	tests := []struct {
		name        string
		origin      string
		commitCount int
		want        []string
		notWant     []string
	}{
		{
			name:    "origin matches base",
			origin:  "git@github.com:up/repo.git",
			want:    []string{"base repository: up/repo\n"},
			notWant: []string{"origin is", "not scanned"},
		},
		{
			name:   "origin differs from base",
			origin: "https://github.com/me/repo.git",
			want:   []string{"base repository: up/repo (origin is me/repo)"},
		},
		{
			name:        "exactly 250 commits",
			commitCount: 250,
			notWant:     []string{"not scanned"},
		},
		{
			name:        "more than 250 commits",
			commitCount: 300,
			want:        []string{"not scanned: commits 251-300"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newSyncForkFixture(t, []string{"aabbccddeeff"})
			f.messages = []string{"feat: one\n\nPartio-Checkpoint: aabbccddeeff\n"}
			f.commitCount = tt.commitCount
			if tt.origin != "" {
				gitOut(t, f.upstream, "remote", "add", "origin", tt.origin)
			}

			out, err := runSyncForkCmd(t, "up/repo#1")
			if err != nil {
				t.Fatalf("sync-fork: %v", err)
			}
			want := append([]string{"synced 1, already present 0, conflicts 0, missing 0"}, tt.want...)
			for _, w := range want {
				if !strings.Contains(out, w) {
					t.Errorf("output missing %q:\n%s", w, out)
				}
			}
			for _, w := range tt.notWant {
				if strings.Contains(out, w) {
					t.Errorf("output unexpectedly contains %q:\n%s", w, out)
				}
			}
			// A warning never stops the sync.
			if _, err := checkpoint.Read("aabbccddeeff"); err != nil {
				t.Errorf("checkpoint.Read: %v", err)
			}
		})
	}
}
