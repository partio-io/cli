package git

import "fmt"

// CurrentBranch returns the current branch name.
func CurrentBranch() (string, error) {
	out, err := execGit("rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", fmt.Errorf("failed to get current branch: %w", err)
	}
	return out, nil
}
