package checkpoint

import "fmt"

// InitBranch creates the orphan checkpoint branch with a parentless
// "partio: initialize checkpoint storage" commit on an empty tree. It is a
// no-op when the branch already exists.
func (s *Store) InitBranch() error {
	exists, err := s.branchExists()
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	treeHash, err := s.git("hash-object", "-w", "-t", "tree", "/dev/null")
	if err != nil {
		return fmt.Errorf("creating empty tree: %w", err)
	}

	commitHash, err := s.git("commit-tree", treeHash, "-m", "partio: initialize checkpoint storage")
	if err != nil {
		return fmt.Errorf("creating initial commit: %w", err)
	}

	if _, err := s.git("update-ref", "refs/heads/"+checkpointBranch, commitHash); err != nil {
		return fmt.Errorf("creating branch ref: %w", err)
	}
	return nil
}

// branchExists reports whether the checkpoint branch exists. Only a clean
// "no such ref" answer means absent; any other failure is returned.
func (s *Store) branchExists() (bool, error) {
	out, err := s.git("for-each-ref", "--format=%(refname)", "refs/heads/"+checkpointBranch)
	if err != nil {
		return false, fmt.Errorf("looking up checkpoint branch: %w", err)
	}
	return out != "", nil
}
