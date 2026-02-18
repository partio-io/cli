package git

import "fmt"

// IsWorktree returns true if the current directory is a git worktree (not the main repo).
func IsWorktree() (bool, error) {
	gitDir, err := execGit("rev-parse", "--git-dir")
	if err != nil {
		return false, fmt.Errorf("not a git repository: %w", err)
	}
	commonDir, err := execGit("rev-parse", "--git-common-dir")
	if err != nil {
		return false, fmt.Errorf("failed to get common dir: %w", err)
	}
	return gitDir != commonDir, nil
}
