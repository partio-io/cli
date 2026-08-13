package git

// DiffNameOnly returns the list of file paths changed in a specific commit.
func DiffNameOnly(commitHash string) ([]string, error) {
	lines, err := diffTree("", commitHash, "--name-only")
	if err != nil {
		return nil, err
	}
	if len(lines) == 0 {
		return nil, nil
	}
	return lines, nil
}
