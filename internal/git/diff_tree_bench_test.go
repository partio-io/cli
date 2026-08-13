package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// BenchmarkDiffNumstat measures DiffNumstat performance on a repo with 1000+ files.
// Run with: go test -bench=BenchmarkDiffNumstat ./internal/git/
func BenchmarkDiffNumstat(b *testing.B) {
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

	writef := func(name, content string) {
		b.Helper()
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			b.Fatalf("WriteFile(%s): %v", name, err)
		}
	}

	run("git", "init")
	run("git", "config", "user.email", "bench@example.com")
	run("git", "config", "user.name", "Bench")

	// Create and commit 1000 files so the repo is large.
	for i := range 1000 {
		writef(fmt.Sprintf("file%04d.txt", i), fmt.Sprintf("content %d\n", i))
	}
	run("git", "add", ".")
	run("git", "commit", "-m", "initial commit with 1000 files")

	// Add one changed file; benchmark DiffNumstat against this commit.
	writef("changed.txt", "changed content\n")
	run("git", "add", "changed.txt")
	run("git", "commit", "-m", "add changed file")
	hash := run("git", "rev-parse", "HEAD")

	// Change working directory so execGit resolves the repo correctly.
	b.Chdir(dir)

	b.ResetTimer()
	for b.Loop() {
		_, err := DiffNumstat(hash)
		if err != nil {
			b.Fatalf("DiffNumstat() error: %v", err)
		}
	}
}
