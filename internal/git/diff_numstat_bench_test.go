package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// BenchmarkCommitDiffStrategies measures the two ways to identify the content
// of a commit that the post-commit path can use:
//
//   - The two-commit "git diff" that production code runs today, through
//     DiffNumstat.
//   - The "git diff-tree --no-commit-id -r --root" plumbing call that issue
//     #30 proposed as the faster replacement.
//
// The diff-tree call is in this file only. Production code keeps one strategy.
// This benchmark measures the alternative. It does not adopt it.
//
// Each repository holds the number of tracked files in the benchmark name. The
// measured commit changes one file, which is what a post-commit hook sees in a
// real repository. Continuous integration does not gate on these numbers. A
// person reads them.
func BenchmarkCommitDiffStrategies(b *testing.B) {
	for _, fileCount := range []int{1000, 10000} {
		dir, commit := buildBenchRepo(b, fileCount)

		b.Run(fmt.Sprintf("current_git_diff/%d_files", fileCount), func(b *testing.B) {
			b.Chdir(dir)
			for b.Loop() {
				if _, err := DiffNumstat(commit); err != nil {
					b.Fatalf("DiffNumstat(%s): %v", commit, err)
				}
			}
		})

		// The two strategies differ in two ways at once: the git command, and
		// the number of git processes. DiffNumstat asks git for the parent
		// first, so it starts two processes. The diff-tree call starts one.
		// This third measurement times the diff command alone, with the two
		// revisions resolved before the loop. It isolates the git command, so
		// a reader can tell the two causes apart.
		b.Run(fmt.Sprintf("current_git_diff_command_only/%d_files", fileCount), func(b *testing.B) {
			b.Chdir(dir)
			from, to, err := commitRange(commit)
			if err != nil {
				b.Fatalf("commitRange(%s): %v", commit, err)
			}
			for b.Loop() {
				if _, err := execGit("diff", "--numstat", from, to); err != nil {
					b.Fatalf("git diff %s %s: %v", from, to, err)
				}
			}
		})

		b.Run(fmt.Sprintf("proposed_git_diff_tree/%d_files", fileCount), func(b *testing.B) {
			b.Chdir(dir)
			for b.Loop() {
				if _, err := benchDiffTree(commit); err != nil {
					b.Fatalf("benchDiffTree(%s): %v", commit, err)
				}
			}
		})
	}
}

// benchDiffTree runs the plumbing call that issue #30 proposed in place of the
// two-commit diff. It is the measured alternative. It is unexported, it is in
// this benchmark file, and no production file calls it.
//
// It mirrors execGit: it reads the process working directory and it takes
// stdout only, so the two strategies differ in the git command alone.
func benchDiffTree(commitHash string) (string, error) {
	cmd := exec.Command("git", "diff-tree", "--no-commit-id", "-r", "--root", "--numstat", commitHash)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// buildBenchRepo makes a repository with fileCount tracked files in the
// benchmark's temporary directory. It returns that directory and the hash of a
// second commit that changes one file. The setup is outside the timed loop.
func buildBenchRepo(b *testing.B, fileCount int) (string, string) {
	b.Helper()

	dir := b.TempDir()

	run := func(args ...string) string {
		b.Helper()
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = dir
		out, err := cmd.CombinedOutput()
		if err != nil {
			b.Fatalf("command %v failed: %v\n%s", args, err, out)
		}
		return strings.TrimSpace(string(out))
	}

	write := func(name, content string) {
		b.Helper()
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			b.Fatalf("write %s: %v", name, err)
		}
	}

	run("git", "init")
	run("git", "config", "user.email", "test@example.com")
	run("git", "config", "user.name", "Test")

	for i := range fileCount {
		write(fmt.Sprintf("file%05d.txt", i), "one\n")
	}
	run("git", "add", ".")
	run("git", "commit", "-m", "first")

	write("file00000.txt", "one\ntwo\n")
	run("git", "add", ".")
	run("git", "commit", "-m", "second")

	// The benchmark name states the file count. Check the repository agrees,
	// so the name cannot claim a size the repository does not have.
	if tracked := len(strings.Fields(run("git", "ls-files"))); tracked != fileCount {
		b.Fatalf("repository holds %d tracked files, want %d", tracked, fileCount)
	}

	return dir, run("git", "rev-parse", "HEAD")
}
