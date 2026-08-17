package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// newDiffNameOnlyRepo makes an empty git repository in a temporary directory,
// changes into it, and returns a command runner and a file writer for it.
func newDiffNameOnlyRepo(t *testing.T) (run func(args ...string) string, write func(name, content string)) {
	t.Helper()

	dir := t.TempDir()
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

	write = func(name, content string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}

	run("git", "init")
	run("git", "config", "user.email", "test@example.com")
	run("git", "config", "user.name", "Test")

	return run, write
}

func TestDiffNameOnly(t *testing.T) {
	run, write := newDiffNameOnlyRepo(t)

	write("a.txt", "one\n")
	write("b.txt", "two\n")
	run("git", "add", ".")
	run("git", "commit", "-m", "first")
	first := run("git", "rev-parse", "HEAD")

	write("a.txt", "one\nthree\n")
	write("c.txt", "four\n")
	run("git", "add", ".")
	run("git", "commit", "-m", "second")
	second := run("git", "rev-parse", "HEAD")

	// A merge commit brings in the branch file. It has a parent, so it takes
	// the unchanged path and is compared against that first parent.
	run("git", "checkout", "-b", "side", first)
	write("d.txt", "five\n")
	run("git", "add", ".")
	run("git", "commit", "-m", "side")
	run("git", "checkout", "-")
	run("git", "merge", "--no-ff", "-m", "merge", "side")
	merge := run("git", "rev-parse", "HEAD")

	run("git", "commit", "--allow-empty", "-m", "empty")
	empty := run("git", "rev-parse", "HEAD")

	tests := []struct {
		name   string
		commit string
		want   []string
	}{
		{
			name:   "a first commit returns every path it adds",
			commit: first,
			want:   []string{"a.txt", "b.txt"},
		},
		{
			name:   "a regular commit returns the paths it changes",
			commit: second,
			want:   []string{"a.txt", "c.txt"},
		},
		{
			name:   "a merge commit returns the paths it brings in against its first parent",
			commit: merge,
			want:   []string{"d.txt"},
		},
		{
			name:   "an empty commit returns an empty result and no error",
			commit: empty,
			want:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DiffNameOnly(tt.commit)
			if err != nil {
				t.Fatalf("DiffNameOnly() unexpected error: %v", err)
			}
			if !equalPaths(got, tt.want) {
				t.Errorf("DiffNameOnly() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestDiffNameOnlyBinaryFirstCommit needs its own repository, because the
// binary file must land in the first commit.
func TestDiffNameOnlyBinaryFirstCommit(t *testing.T) {
	run, write := newDiffNameOnlyRepo(t)

	write("a.txt", "one\n")
	write("logo.png", string([]byte{0x00, 0x01, 0x02, 0x00, 0xff}))
	run("git", "add", ".")
	run("git", "commit", "-m", "first")
	first := run("git", "rev-parse", "HEAD")

	got, err := DiffNameOnly(first)
	if err != nil {
		t.Fatalf("DiffNameOnly() unexpected error: %v", err)
	}
	want := []string{"a.txt", "logo.png"}
	if !equalPaths(got, want) {
		t.Errorf("DiffNameOnly() = %v, want %v", got, want)
	}
}

func equalPaths(got, want []string) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}
