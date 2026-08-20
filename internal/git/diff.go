package git

// Diff returns the unified diff for a specific commit.
func Diff(commitHash string) (string, error) {
	from, to, err := commitRange(commitHash)
	if err != nil {
		return "", err
	}

	return execGit("diff", from, to)
}
