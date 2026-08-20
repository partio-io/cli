package attribution

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// initRepo creates a git repository in a temporary directory, changes the
// process working directory into it, and returns the directory plus a git
// runner. The commit-diff functions read the process working directory, so
// the change of directory is required.
func initRepo(t *testing.T) (string, func(args ...string) string) {
	t.Helper()

	dir := t.TempDir()
	t.Chdir(dir)

	run := func(args ...string) string {
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

// writeLines writes a text file of the given line count into dir.
func writeLines(t *testing.T, dir, name string, lines int) {
	t.Helper()

	content := strings.Repeat("line\n", lines)
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
}

func TestCalculateFirstCommitTextFiles(t *testing.T) {
	dir, run := initRepo(t)

	writeLines(t, dir, "a.txt", 3)
	writeLines(t, dir, "b.txt", 2)
	run("git", "add", ".")
	run("git", "commit", "-m", "first")

	sha := run("git", "rev-parse", "HEAD")

	got, err := Calculate(sha, true)
	if err != nil {
		t.Fatalf("Calculate() unexpected error: %v", err)
	}
	if got.TotalLines != 5 {
		t.Errorf("TotalLines = %d, want 5", got.TotalLines)
	}
	if got.AgentLines != 5 {
		t.Errorf("AgentLines = %d, want 5", got.AgentLines)
	}
	if got.AgentPercent != 100 {
		t.Errorf("AgentPercent = %d, want 100", got.AgentPercent)
	}
}

func TestCalculateFirstCommitNoAgent(t *testing.T) {
	dir, run := initRepo(t)

	writeLines(t, dir, "a.txt", 3)
	writeLines(t, dir, "b.txt", 2)
	run("git", "add", ".")
	run("git", "commit", "-m", "first")

	sha := run("git", "rev-parse", "HEAD")

	got, err := Calculate(sha, false)
	if err != nil {
		t.Fatalf("Calculate() unexpected error: %v", err)
	}
	if got.TotalLines != 5 {
		t.Errorf("TotalLines = %d, want 5", got.TotalLines)
	}
	if got.HumanLines != 5 {
		t.Errorf("HumanLines = %d, want 5", got.HumanLines)
	}
	if got.AgentLines != 0 {
		t.Errorf("AgentLines = %d, want 0", got.AgentLines)
	}
	if got.AgentPercent != 0 {
		t.Errorf("AgentPercent = %d, want 0", got.AgentPercent)
	}
}

func TestCalculateFirstCommitBinaryFile(t *testing.T) {
	dir, run := initRepo(t)

	writeLines(t, dir, "a.txt", 4)
	// A NUL byte makes git treat the file as binary. Numstat then reports a
	// dash instead of a line count for it.
	if err := os.WriteFile(filepath.Join(dir, "logo.bin"), []byte{0x00, 0x01, 0x02, 0x00}, 0o644); err != nil {
		t.Fatalf("write logo.bin: %v", err)
	}
	run("git", "add", ".")
	run("git", "commit", "-m", "first")

	sha := run("git", "rev-parse", "HEAD")

	numstat := run("git", "show", "--numstat", "--format=", sha)
	if !strings.Contains(numstat, "-\t-\tlogo.bin") {
		t.Fatalf("git does not see logo.bin as binary, numstat:\n%s", numstat)
	}

	got, err := Calculate(sha, true)
	if err != nil {
		t.Fatalf("Calculate() unexpected error: %v", err)
	}
	if got.TotalLines != 4 {
		t.Errorf("TotalLines = %d, want 4", got.TotalLines)
	}
}

func TestCalculateFirstCommitEmpty(t *testing.T) {
	_, run := initRepo(t)

	run("git", "commit", "--allow-empty", "-m", "first")

	sha := run("git", "rev-parse", "HEAD")

	got, err := Calculate(sha, true)
	if err != nil {
		t.Fatalf("Calculate() unexpected error: %v", err)
	}
	if got.TotalLines != 0 {
		t.Errorf("TotalLines = %d, want 0", got.TotalLines)
	}
}

func TestCalculateUnresolvableCommit(t *testing.T) {
	dir, run := initRepo(t)

	writeLines(t, dir, "a.txt", 1)
	run("git", "add", ".")
	run("git", "commit", "-m", "first")

	// A well-formed hash that names no object in this repository. Git cannot
	// measure it, so the measurement failed. A failed measurement must not
	// look like an empty commit.
	got, err := Calculate("deadbeefdeadbeefdeadbeefdeadbeefdeadbeef", true)
	if err == nil {
		t.Fatalf("Calculate() error = nil, want an error; result = %+v", got)
	}
}

// addedLines sums the added column of a numstat block. It gives the tests the
// total that the comparison against the first parent produced before this
// slice, so a regression against that total is visible.
func addedLines(t *testing.T, numstat string) int {
	t.Helper()

	total := 0
	for _, line := range strings.Split(numstat, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		added, err := strconv.Atoi(fields[0])
		if err != nil {
			continue // binary file
		}
		total += added
	}

	return total
}

func TestCalculateRegularCommit(t *testing.T) {
	dir, run := initRepo(t)

	writeLines(t, dir, "a.txt", 1)
	run("git", "add", ".")
	run("git", "commit", "-m", "first")

	writeLines(t, dir, "b.txt", 6)
	run("git", "add", ".")
	run("git", "commit", "-m", "second")

	sha := run("git", "rev-parse", "HEAD")
	want := addedLines(t, run("git", "diff", "--numstat", sha+"~1", sha))
	if want != 6 {
		t.Fatalf("the comparison against the parent gives %d added lines, want 6", want)
	}

	got, err := Calculate(sha, true)
	if err != nil {
		t.Fatalf("Calculate() unexpected error: %v", err)
	}
	if got.TotalLines != want {
		t.Errorf("TotalLines = %d, want %d", got.TotalLines, want)
	}
}

func TestCalculateMergeCommit(t *testing.T) {
	dir, run := initRepo(t)

	writeLines(t, dir, "a.txt", 1)
	run("git", "add", ".")
	run("git", "commit", "-m", "first")

	base := run("git", "rev-parse", "--abbrev-ref", "HEAD")

	run("git", "checkout", "-b", "feature")
	writeLines(t, dir, "c.txt", 3)
	run("git", "add", ".")
	run("git", "commit", "-m", "feature work")

	run("git", "checkout", base)
	writeLines(t, dir, "b.txt", 2)
	run("git", "add", ".")
	run("git", "commit", "-m", "base work")

	run("git", "merge", "--no-ff", "-m", "merge feature", "feature")

	sha := run("git", "rev-parse", "HEAD")
	want := addedLines(t, run("git", "diff", "--numstat", sha+"~1", sha))
	if want != 3 {
		t.Fatalf("the comparison against the first parent gives %d added lines, want 3", want)
	}

	got, err := Calculate(sha, true)
	if err != nil {
		t.Fatalf("Calculate() unexpected error: %v", err)
	}
	// A merge commit must keep the lines it brings in against its first
	// parent. The --root flag makes this total zero, which is the regression
	// in pull request #653.
	if got.TotalLines != want {
		t.Errorf("TotalLines = %d, want %d", got.TotalLines, want)
	}
}
