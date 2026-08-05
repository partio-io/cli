package git

import "os/exec"

// CommitReachable reports whether the git object identified by sha exists and
// is reachable in the repository at repoDir.
// Returns (false, nil) when the object is not found (git exits non-zero).
// Returns (false, err) only on unexpected subprocess errors.
func CommitReachable(repoDir, sha string) (bool, error) {
	cmd := exec.Command("git", "cat-file", "-e", sha)
	cmd.Dir = repoDir
	err := cmd.Run()
	if err == nil {
		return true, nil
	}
	if _, ok := err.(*exec.ExitError); ok {
		return false, nil
	}
	return false, err
}
