package git

import (
	"fmt"
	"strings"
)

// CommitReachable reports whether the commit exists and is reachable from any local branch.
func CommitReachable(sha string) bool {
	if !CommitExists(sha) {
		return false
	}
	out, err := execGit("branch", "--contains", sha)
	if err != nil {
		return false
	}
	return strings.TrimSpace(out) != ""
}

// FindSquashCommit searches the recent git history (up to 200 commits from HEAD) for a commit
// whose tree matches the tree of commitSHA, indicating it may have been squash-merged.
// Returns the squash commit SHA, or an empty string if not found.
func FindSquashCommit(commitSHA string) (string, error) {
	origTree, err := execGit("rev-parse", commitSHA+"^{tree}")
	if err != nil {
		return "", fmt.Errorf("cannot resolve commit %s: %w", commitSHA, err)
	}

	log, err := execGit("log", "--format=%H %T", "-n", "200")
	if err != nil {
		return "", fmt.Errorf("reading git log: %w", err)
	}

	for _, line := range strings.Split(strings.TrimSpace(log), "\n") {
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) == 2 && parts[1] == origTree && parts[0] != commitSHA {
			return parts[0], nil
		}
	}

	return "", nil
}
