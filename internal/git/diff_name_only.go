package git

import "strings"

// DiffNameOnly returns the list of file paths changed in a specific commit.
func DiffNameOnly(commitHash string) ([]string, error) {
	from, to, err := commitRange(commitHash)
	if err != nil {
		return nil, err
	}

	out, err := execGit("diff", "--name-only", from, to)
	if err != nil {
		return nil, err
	}
	if out == "" {
		return nil, nil
	}
	return strings.Split(out, "\n"), nil
}
