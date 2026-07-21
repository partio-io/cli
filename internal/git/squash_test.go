package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// runGitInDir runs a git command in dir and returns trimmed stdout. Fatals on error.
func runGitInDir(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(),
		"GIT_AUTHOR_NAME=test",
		"GIT_AUTHOR_EMAIL=test@test.com",
		"GIT_COMMITTER_NAME=test",
		"GIT_COMMITTER_EMAIL=test@test.com",
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

// initTestRepo creates a temporary git repo and returns its path.
func initTestRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	runGitInDir(t, dir, "init", "-b", "main")
	runGitInDir(t, dir, "config", "user.email", "test@test.com")
	runGitInDir(t, dir, "config", "user.name", "Test")
	return dir
}

// makeTestCommit creates a file and commits it. Returns the commit SHA.
func makeTestCommit(t *testing.T, dir, filename, content, msg string) string {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, filename), []byte(content), 0o644); err != nil {
		t.Fatalf("writing %s: %v", filename, err)
	}
	runGitInDir(t, dir, "add", ".")
	runGitInDir(t, dir, "commit", "-m", msg)
	return runGitInDir(t, dir, "rev-parse", "HEAD")
}

func TestCommitReachable(t *testing.T) {
	dir := initTestRepo(t)

	// Create an initial commit on the default branch
	mainSHA := makeTestCommit(t, dir, "main.txt", "main", "main commit")

	// Create a second branch, make a commit, then delete the branch
	runGitInDir(t, dir, "checkout", "-b", "temp-branch")
	tempSHA := makeTestCommit(t, dir, "temp.txt", "temp", "temp commit")
	runGitInDir(t, dir, "checkout", "-") // go back to previous branch
	runGitInDir(t, dir, "branch", "-D", "temp-branch")

	orig, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(orig) })

	if !CommitReachable(mainSHA) {
		t.Errorf("CommitReachable(%s) = false, want true for reachable commit", mainSHA[:8])
	}

	if CommitReachable(tempSHA) {
		t.Errorf("CommitReachable(%s) = true, want false for unreachable commit", tempSHA[:8])
	}

	const fakeSHA = "0000000000000000000000000000000000000000"
	if CommitReachable(fakeSHA) {
		t.Errorf("CommitReachable(nonexistent SHA) = true, want false")
	}
}

func TestFindSquashCommit(t *testing.T) {
	dir := initTestRepo(t)

	// Create initial commit on main
	makeTestCommit(t, dir, "base.txt", "base content", "initial commit")

	// Create a feature branch with one commit
	runGitInDir(t, dir, "checkout", "-b", "feature")
	makeTestCommit(t, dir, "feature.txt", "feature content", "feature work")
	featureSHA := runGitInDir(t, dir, "rev-parse", "HEAD")

	// Switch back and squash-merge the feature branch
	runGitInDir(t, dir, "checkout", "main")
	runGitInDir(t, dir, "merge", "--squash", "feature")
	runGitInDir(t, dir, "commit", "-m", "squash: merge feature")
	squashSHA := runGitInDir(t, dir, "rev-parse", "HEAD")

	// Delete feature branch so featureSHA is unreachable
	runGitInDir(t, dir, "branch", "-D", "feature")

	orig, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(orig) })

	got, err := FindSquashCommit(featureSHA)
	if err != nil {
		t.Fatalf("FindSquashCommit() error: %v", err)
	}
	if got != squashSHA {
		t.Errorf("FindSquashCommit() = %q, want %q", got, squashSHA)
	}
}

func TestFindSquashCommit_NotFound(t *testing.T) {
	dir := initTestRepo(t)

	// Create two unrelated commits with different trees
	sha1 := makeTestCommit(t, dir, "a.txt", "content a", "commit a")
	makeTestCommit(t, dir, "b.txt", "content b", "commit b")

	orig, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(orig) })

	// sha1 is still reachable and its tree is not replicated anywhere else
	got, err := FindSquashCommit(sha1)
	if err != nil {
		t.Fatalf("FindSquashCommit() error: %v", err)
	}
	if got != "" {
		t.Errorf("FindSquashCommit() = %q, want empty string for no match", got)
	}
}

func TestFindSquashCommit_InvalidSHA(t *testing.T) {
	dir := initTestRepo(t)
	makeTestCommit(t, dir, "a.txt", "content", "initial")

	orig, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(orig) })

	_, err := FindSquashCommit("0000000000000000000000000000000000000000")
	if err == nil {
		t.Error("FindSquashCommit(invalid SHA) expected error, got nil")
	}
}
