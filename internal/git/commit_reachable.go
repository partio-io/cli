package git

import "os/exec"

// CommitReachable reports whether the given SHA refers to a reachable git
// object. It returns (true, nil) when the object exists, (false, nil) when it
// does not, and (false, err) only when the subprocess itself fails in an
// unexpected way.
func CommitReachable(repoDir, sha string) (bool, error) {
	cmd := exec.Command("git", "cat-file", "-e", sha)
	cmd.Dir = repoDir
	err := cmd.Run()
	if err != nil {
		// exit status 1 means the object does not exist – not an error for our
		// purposes.
		if exitErr, ok := err.(*exec.ExitError); ok {
			_ = exitErr
			return false, nil
		}
		return false, err
	}
	return true, nil
}
