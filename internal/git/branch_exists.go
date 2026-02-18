package git

// BranchExists checks if a branch exists.
func BranchExists(name string) bool {
	_, err := execGit("rev-parse", "--verify", "refs/heads/"+name)
	return err == nil
}
