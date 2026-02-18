package git

import "fmt"

// PushBranch pushes a branch to the remote.
func PushBranch(remote, branch string) error {
	_, err := execGit("push", "--no-verify", remote, branch)
	if err != nil {
		return fmt.Errorf("pushing %s to %s: %w", branch, remote, err)
	}
	return nil
}
