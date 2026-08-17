package git

import (
	"os/exec"
	"strings"
)

const CheckpointBranch = "partio/checkpoints/v1"

// EmptyTree is the object id of git's empty tree. Git always gives a tree
// with no entries this hash, so a diff against it shows the full content of
// a commit that has no parent.
const EmptyTree = "4b825dc642cb6eb9a060e54bf8d69288fbee4904"

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
