package git

import (
	"fmt"
	"os/exec"
)

// FetchBranch fetches a remote branch into the local remote-tracking ref
// refs/remotes/<remote>/<branch>. If the remote does not have the named
// branch, FetchBranch returns nil — a missing remote branch is not an error.
// If the remote is unreachable, the error is returned to the caller.
func FetchBranch(remote, branch string) error {
	// ls-remote --exit-code exits with code 2 when no matching refs are found.
	// This lets us distinguish "branch absent from remote" (benign) from
	// "remote unreachable" (real error) without parsing error message strings.
	_, err := execGit("ls-remote", "--exit-code", remote, "refs/heads/"+branch)
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 2 {
			// Branch does not exist on the remote — not an error.
			return nil
		}
		return fmt.Errorf("fetching %s from %s: %w", branch, remote, err)
	}

	_, err = execGit("fetch", "--no-tags", remote,
		"refs/heads/"+branch+":refs/remotes/"+remote+"/"+branch)
	if err != nil {
		return fmt.Errorf("fetching %s from %s: %w", branch, remote, err)
	}
	return nil
}
