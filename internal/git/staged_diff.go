package git

// StagedDiff returns the diff of staged changes.
func StagedDiff() (string, error) {
	return execGit("diff", "--cached", "--numstat")
}
