package git

import (
	"os/exec"
	"strings"
)

// diffTree invokes git diff-tree with the standard plumbing flags
// (--no-commit-id -r --root) plus any extra args against the given commit hash.
// It returns the output as a slice of non-empty lines.
//
// Empty output (e.g. merge commits, which have multiple parents) is returned as
// an empty slice, not an error. Merge commits intentionally produce no output
// because they introduce no original content; callers must not add -m or --cc
// to change this behaviour without deliberate design consideration.
//
// repoPath sets the working directory for the subprocess; pass "" to use the
// process's current working directory (consistent with execGit).
func diffTree(repoPath, commitHash string, extraArgs ...string) ([]string, error) {
	args := append([]string{"diff-tree", "--no-commit-id", "-r", "--root"}, extraArgs...)
	args = append(args, commitHash)

	cmd := exec.Command("git", args...)
	if repoPath != "" {
		cmd.Dir = repoPath
	}

	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	raw := strings.TrimSpace(string(out))
	if raw == "" {
		return []string{}, nil
	}

	return strings.Split(raw, "\n"), nil
}
