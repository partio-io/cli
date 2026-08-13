package git

import "strings"

// DiffNumstat returns numstat for a specific commit.
func DiffNumstat(commitHash string) (string, error) {
	lines, err := diffTree("", commitHash, "--numstat")
	if err != nil {
		return "", err
	}
	return strings.Join(lines, "\n"), nil
}
