package git

import (
	"os/exec"
	"strings"
)

const CheckpointBranch = "partio/checkpoints/v1"

// execGit runs a git command and returns trimmed stdout.
func execGit(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// ExecGit runs a git command and returns trimmed stdout (exported).
func ExecGit(args ...string) (string, error) {
	return execGit(args...)
}
