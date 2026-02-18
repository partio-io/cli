package git

// DiffNumstat returns numstat for a specific commit.
func DiffNumstat(commitHash string) (string, error) {
	return execGit("diff", "--numstat", commitHash+"~1", commitHash)
}
