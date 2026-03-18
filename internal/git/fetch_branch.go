package git

import "fmt"

// FetchBranch fetches a remote branch using --filter=blob:none to skip blob objects.
// Only tree and commit objects are downloaded, which is sufficient for merging
// checkpoint entries where only the tree structure is needed.
func FetchBranch(remote, branch string) error {
	_, err := execGit("fetch", "--filter=blob:none", remote, branch)
	if err != nil {
		return fmt.Errorf("fetching %s from %s: %w", branch, remote, err)
	}
	return nil
}
