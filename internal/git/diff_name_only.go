package git

import "strings"

// DiffNameOnly returns the list of file paths changed in a specific commit.
// It uses git diff-tree which reads directly from git objects, avoiding an
// O(N) tree walk. The --root flag ensures the initial commit is handled
// correctly by comparing against the empty tree.
func DiffNameOnly(commitHash string) ([]string, error) {
	out, err := execGit("diff-tree", "--no-commit-id", "-r", "-c", "--name-only", "--root", commitHash)
	if err != nil {
		return nil, err
	}
	if out == "" {
		return nil, nil
	}
	return strings.Split(out, "\n"), nil
}
