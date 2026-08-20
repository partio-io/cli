package hooks

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/partio-io/cli/internal/checkpoint"
	"github.com/partio-io/cli/internal/config"
	"github.com/partio-io/cli/internal/git"
)

// newPostCommitRepo makes an empty git repository in a temporary directory,
// changes into it, and returns the repository root and a command runner.
func newPostCommitRepo(t *testing.T) (dir string, run func(args ...string) string) {
	t.Helper()

	dir = t.TempDir()
	t.Chdir(dir)

	run = func(args ...string) string {
		t.Helper()
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = dir
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("command %v failed: %v\n%s", args, err, out)
		}
		return strings.TrimSpace(string(out))
	}

	run("git", "init")
	run("git", "config", "user.email", "test@example.com")
	run("git", "config", "user.name", "Test")

	return dir, run
}

// enableCheckpointBranch makes the orphan branch that the checkpoint store
// writes to. In production `partio enable` does this.
func enableCheckpointBranch(t *testing.T, run func(args ...string) string) {
	t.Helper()

	tree := run("git", "hash-object", "-t", "tree", "/dev/null")
	commit := run("git", "commit-tree", tree, "-m", "partio: initialize checkpoint storage")
	run("git", "update-ref", "refs/heads/"+git.CheckpointBranch, commit)
}

// TestPostCommitFirstCommitStoresDiff proves the end of the slice: the
// checkpoint written for the first commit of a repository holds that commit's
// diff. Before this slice git.Diff failed on a root commit, the hook swallowed
// the error in an `err == nil` check, and the checkpoint held no code at all.
func TestPostCommitFirstCommitStoresDiff(t *testing.T) {
	dir, run := newPostCommitRepo(t)
	enableCheckpointBranch(t, run)

	if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("one\ntwo\n"), 0o644); err != nil {
		t.Fatalf("write a.txt: %v", err)
	}
	run("git", "add", ".")
	run("git", "commit", "-m", "first")

	writePreCommitState(t, dir, preCommitState{
		AgentActive: true,
		AgentName:   "claude-code",
		Branch:      run("git", "rev-parse", "--abbrev-ref", "HEAD"),
	})

	if err := runPostCommit(dir, config.Config{Agent: "claude-code"}); err != nil {
		t.Fatalf("runPostCommit() unexpected error: %v", err)
	}

	id := checkpointIDFromTrailers(t, run("git", "log", "-1", "--format=%B"))
	data, err := checkpoint.Read(id)
	if err != nil {
		t.Fatalf("checkpoint.Read(%s): %v", id, err)
	}

	if strings.TrimSpace(data.Diff) == "" {
		t.Fatal("checkpoint holds an empty diff, want the diff of the first commit")
	}
	if !strings.Contains(data.Diff, "a.txt") {
		t.Errorf("checkpoint diff = %q, want it to name a.txt", data.Diff)
	}
	if !strings.Contains(data.Diff, "+one") {
		t.Errorf("checkpoint diff = %q, want it to hold the added content", data.Diff)
	}
}

// captureLogs sends the default logger to a buffer for the length of the test
// and returns the buffer. The hook reports a failed measurement through the
// default logger, so the buffer is where the warning becomes visible.
func captureLogs(t *testing.T) *bytes.Buffer {
	t.Helper()

	buf := &bytes.Buffer{}
	previous := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(buf, &slog.HandlerOptions{Level: slog.LevelWarn})))
	t.Cleanup(func() { slog.SetDefault(previous) })

	return buf
}

// breakCommitObject deletes the object file of the given commit. Git then reads
// the reference but cannot read the commit, so it cannot measure it. This is
// the failure the corrupted empty tree constant produced in the field: `fatal:
// bad object`.
func breakCommitObject(t *testing.T, repoRoot, sha string) {
	t.Helper()

	object := filepath.Join(repoRoot, ".git", "objects", sha[:2], sha[2:])
	if err := os.Remove(object); err != nil {
		t.Fatalf("remove commit object %s: %v", sha, err)
	}
}

// TestPostCommitWarnsWhenAttributionFails proves the point of the slice: a
// measurement that git cannot make is reported. Before this slice the
// attribution function returned an empty result and a nil error, so this
// warning branch of the hook never ran and the defect stayed silent.
func TestPostCommitWarnsWhenAttributionFails(t *testing.T) {
	dir, run := newPostCommitRepo(t)
	enableCheckpointBranch(t, run)

	if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("one\ntwo\n"), 0o644); err != nil {
		t.Fatalf("write a.txt: %v", err)
	}
	run("git", "add", ".")
	run("git", "commit", "-m", "first")

	branch := run("git", "rev-parse", "--abbrev-ref", "HEAD")
	breakCommitObject(t, dir, run("git", "rev-parse", "HEAD"))

	writePreCommitState(t, dir, preCommitState{
		AgentActive: true,
		AgentName:   "claude-code",
		Branch:      branch,
	})

	logs := captureLogs(t)

	// The hook must not fail the git operation when a measurement fails.
	if err := runPostCommit(dir, config.Config{Agent: "claude-code"}); err != nil {
		t.Fatalf("runPostCommit() error = %v, want nil; a failed measurement must not fail the git operation", err)
	}

	if !strings.Contains(logs.String(), "could not calculate attribution") {
		t.Errorf("the hook wrote no attribution warning, logs:\n%s", logs.String())
	}
}

func writePreCommitState(t *testing.T, repoRoot string, state preCommitState) {
	t.Helper()

	stateDir := filepath.Join(repoRoot, config.PartioDir, "state")
	if err := os.MkdirAll(stateDir, 0o755); err != nil {
		t.Fatalf("make state dir: %v", err)
	}
	data, err := json.Marshal(state)
	if err != nil {
		t.Fatalf("marshal pre-commit state: %v", err)
	}
	if err := os.WriteFile(filepath.Join(stateDir, "pre-commit.json"), data, 0o644); err != nil {
		t.Fatalf("write pre-commit state: %v", err)
	}
}

func checkpointIDFromTrailers(t *testing.T, message string) string {
	t.Helper()

	const prefix = "Partio-Checkpoint:"
	for line := range strings.SplitSeq(message, "\n") {
		if after, ok := strings.CutPrefix(strings.TrimSpace(line), prefix); ok {
			return strings.TrimSpace(after)
		}
	}
	t.Fatalf("commit message holds no %s trailer:\n%s", prefix, message)
	return ""
}
