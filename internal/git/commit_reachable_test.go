package git

import (
	"os/exec"
	"strings"
	"testing"
)

func TestCommitReachable(t *testing.T) {
	dir := t.TempDir()

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
	run("git", "commit", "--allow-empty", "-m", "initial")

	sha := run("git", "rev-parse", "HEAD")

	tests := []struct {
		name        string
		sha         string
		wantReachable bool
	}{
		{
			name:        "existing commit is reachable",
			sha:         sha,
			wantReachable: true,
		},
		{
			name:        "unknown SHA is not reachable",
			sha:         "0000000000000000000000000000000000000000",
			wantReachable: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CommitReachable(dir, tt.sha)
			if err != nil {
				t.Fatalf("CommitReachable() unexpected error: %v", err)
			}
			if got != tt.wantReachable {
				t.Errorf("CommitReachable() = %v, want %v", got, tt.wantReachable)
			}
		})
	}
}
