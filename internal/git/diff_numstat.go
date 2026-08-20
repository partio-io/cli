package git

// DiffNumstat returns numstat for a specific commit.
func DiffNumstat(commitHash string) (string, error) {
	from, to, err := commitRange(commitHash)
	if err != nil {
		return "", err
	}

	return execGit("diff", "--numstat", from, to)
}
