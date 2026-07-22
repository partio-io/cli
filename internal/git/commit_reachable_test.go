package git

import (
	"os/exec"
	"strings"
	"testing"
)

func initRepo(t *testing.T) (string, string) {
	t.Helper()
	dir := t.TempDir()

	run := func(args ...string) string {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
		return string(out)
	}

	run("init", "-b", "main")
	run("config", "user.email", "test@example.com")
	run("config", "user.name", "Test")
	run("commit", "--allow-empty", "-m", "init")

	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("rev-parse HEAD: %v", err)
	}

	sha := strings.TrimSpace(string(out))
	return dir, sha
}

func TestCommitReachable(t *testing.T) {
	dir, existingSHA := initRepo(t)

	tests := []struct {
		name      string
		sha       string
		wantFound bool
	}{
		{
			name:      "existing commit is reachable",
			sha:       existingSHA,
			wantFound: true,
		},
		{
			name:      "unknown sha is not reachable",
			sha:       "0000000000000000000000000000000000000000",
			wantFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CommitReachable(dir, tt.sha)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.wantFound {
				t.Errorf("CommitReachable(%q) = %v, want %v", tt.sha, got, tt.wantFound)
			}
		})
	}
}
