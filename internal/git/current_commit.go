package git

import "fmt"

// CurrentCommit returns the current HEAD commit hash.
func CurrentCommit() (string, error) {
	out, err := execGit("rev-parse", "HEAD")
	if err != nil {
		return "", fmt.Errorf("failed to get current commit: %w", err)
	}
	return out, nil
}
