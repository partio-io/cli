package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestCommitRange(t *testing.T) {
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

	write := func(name, content string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}

	run("git", "init")
	run("git", "config", "user.email", "test@example.com")
	run("git", "config", "user.name", "Test")

	write("a.txt", "one\n")
	run("git", "add", ".")
	run("git", "commit", "-m", "first")
	first := run("git", "rev-parse", "HEAD")

	write("a.txt", "one\ntwo\n")
	run("git", "add", ".")
	run("git", "commit", "-m", "second")
	second := run("git", "rev-parse", "HEAD")

	tests := []struct {
		name     string
		commit   string
		wantFrom string
		wantTo   string
		wantErr  bool
	}{
		{
			name:     "commit with no parent compares against the empty tree",
			commit:   first,
			wantFrom: EmptyTree,
			wantTo:   first,
		},
		{
			name:     "commit with a parent compares against that parent",
			commit:   second,
			wantFrom: first,
			wantTo:   second,
		},
		{
			name:    "an error that is not the absence of a parent propagates",
			commit:  "0000000000000000000000000000000000000000",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			from, to, err := commitRange(tt.commit)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("commitRange(%s) = (%q, %q, nil), want an error", tt.commit, from, to)
				}
				return
			}
			if err != nil {
				t.Fatalf("commitRange() unexpected error: %v", err)
			}
			if from != tt.wantFrom {
				t.Errorf("from = %q, want %q", from, tt.wantFrom)
			}
			if to != tt.wantTo {
				t.Errorf("to = %q, want %q", to, tt.wantTo)
			}
		})
	}
}
