package git

// BranchExists checks if a branch exists.
func BranchExists(name string) bool {
	_, err := execGit("rev-parse", "--verify", "refs/heads/"+name)
	return err == nil
}

// RemoteBranchExists checks if a remote tracking branch exists for origin.
func RemoteBranchExists(name string) bool {
	_, err := execGit("rev-parse", "--verify", "refs/remotes/origin/"+name)
	return err == nil
}
