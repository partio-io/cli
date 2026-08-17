package git

import "strings"

// commitRange returns the two revision arguments that identify the content of
// commitHash. A commit with a parent compares against its first parent. A
// commit with no parent compares against git's empty tree.
//
// Git supplies the parent. An error from git propagates to the caller, so a
// commit that does not exist is not mistaken for a commit with no parent.
func commitRange(commitHash string) (string, string, error) {
	out, err := execGit("rev-list", "--parents", "-n", "1", commitHash)
	if err != nil {
		return "", "", err
	}

	// The first field is the commit itself. The fields after it are the
	// parents. No parent means this is the first commit of the repository.
	fields := strings.Fields(out)
	if len(fields) < 2 {
		return EmptyTree, commitHash, nil
	}

	return fields[1], commitHash, nil
}
