package git

// Diff returns the unified diff for a specific commit.
func Diff(commitHash string) (string, error) {
	return execGit("diff", commitHash+"~1", commitHash)
}
