package git

import "strings"

// StagedFiles returns the list of file paths staged for commit.
func StagedFiles() ([]string, error) {
	out, err := execGit("diff", "--cached", "--name-only")
	if err != nil {
		return nil, err
	}
	if out == "" {
		return nil, nil
	}
	return strings.Split(out, "\n"), nil
}

// CommittedFiles returns the list of file paths changed in a specific commit.
func CommittedFiles(commitHash string) ([]string, error) {
	out, err := execGit("diff", "--name-only", commitHash+"~1", commitHash)
	if err != nil {
		return nil, err
	}
	if out == "" {
		return nil, nil
	}
	return strings.Split(out, "\n"), nil
}

// UnstagedFiles returns the list of file paths with unstaged modifications.
func UnstagedFiles() ([]string, error) {
	out, err := execGit("diff", "--name-only")
	if err != nil {
		return nil, err
	}
	if out == "" {
		return nil, nil
	}
	return strings.Split(out, "\n"), nil
}
