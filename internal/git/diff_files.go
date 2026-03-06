package git

import "strings"

// DiffFiles returns the list of file paths modified in a commit.
func DiffFiles(commitHash string) ([]string, error) {
	out, err := execGit("diff", "--name-only", commitHash+"~1", commitHash)
	if err != nil {
		// Retry against the empty tree for first commits (no parent).
		out, err = execGit("diff", "--name-only", "4b825dc642cb6eb9a060e54bf899d69f82cf7ee2", commitHash)
		if err != nil {
			return nil, err
		}
	}
	if out == "" {
		return nil, nil
	}
	var files []string
	for _, line := range strings.Split(out, "\n") {
		if line != "" {
			files = append(files, line)
		}
	}
	return files, nil
}
