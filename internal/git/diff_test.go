package git

import (
	"strings"
	"testing"
)

// TestDiffFirstCommit is the tracer bullet for slice 03. A first commit has no
// parent, so the old revision expression `HEAD~1` made git fail and the hook
// stored a checkpoint with no diff at all.
func TestDiffFirstCommit(t *testing.T) {
	run, write := newDiffNameOnlyRepo(t)

	write("a.txt", "one\n")
	run("git", "add", ".")
	run("git", "commit", "-m", "first")
	first := run("git", "rev-parse", "HEAD")

	got, err := Diff(first)
	if err != nil {
		t.Fatalf("Diff() unexpected error: %v", err)
	}
	if !strings.Contains(got, "a.txt") {
		t.Errorf("Diff() = %q, want it to name a.txt", got)
	}
	if !strings.Contains(got, "+one") {
		t.Errorf("Diff() = %q, want it to hold the added content", got)
	}
}

// TestDiffUnchangedForCommitsWithAParent holds the behaviour of every commit
// that has a parent. Such a commit still compares against its first parent, so
// the result must equal the old `<commit>~1 <commit>` expression exactly.
func TestDiffUnchangedForCommitsWithAParent(t *testing.T) {
	run, write := newDiffNameOnlyRepo(t)

	write("a.txt", "one\n")
	run("git", "add", ".")
	run("git", "commit", "-m", "first")
	first := run("git", "rev-parse", "HEAD")

	write("a.txt", "one\nthree\n")
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
		name         string
		commit       string
		wantContains string
		wantEmpty    bool
	}{
		{
			name:         "a regular commit returns the diff it returns today",
			commit:       second,
			wantContains: "+three",
		},
		{
			// Pull request #653 added the --root flag, which makes git return
			// nothing at all for a merge commit. This row fails if that
			// mistake returns.
			name:         "a merge commit returns the diff against its first parent",
			commit:       merge,
			wantContains: "d.txt",
		},
		{
			name:      "an empty commit returns an empty result and no error",
			commit:    empty,
			wantEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Diff(tt.commit)
			if err != nil {
				t.Fatalf("Diff() unexpected error: %v", err)
			}
			want := run("git", "diff", tt.commit+"~1", tt.commit)
			if strings.TrimSpace(got) != want {
				t.Errorf("Diff() = %q, want %q", got, want)
			}
			if tt.wantContains != "" && !strings.Contains(got, tt.wantContains) {
				t.Errorf("Diff() = %q, want it to hold %q", got, tt.wantContains)
			}
			if tt.wantEmpty && strings.TrimSpace(got) != "" {
				t.Errorf("Diff() = %q, want an empty result", got)
			}
		})
	}
}
