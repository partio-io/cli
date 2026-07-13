package git

// CommitExists returns true if the given commit hash exists in the repository.
func CommitExists(hash string) bool {
	if hash == "" {
		return false
	}
	_, err := execGit("cat-file", "-e", hash+"^{commit}")
	return err == nil
}

// FindCommitByCheckpointID searches all reachable commits for one that has
// a "Partio-Checkpoint: <id>" trailer. Returns the commit hash, or an empty
// string if no matching commit is found.
func FindCommitByCheckpointID(id string) (string, error) {
	return execGit("log", "--all", "--grep=Partio-Checkpoint: "+id, "--format=%H", "-1")
}
