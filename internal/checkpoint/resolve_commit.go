package checkpoint

import "github.com/partio-io/cli/internal/git"

// ResolveCommitHash returns the best available commit hash for a checkpoint.
//
// It first checks whether the original commit still exists in the repository.
// If it does not (e.g. the branch was squash-merged), it searches the git log
// for a commit that references this checkpoint ID via the Partio-Checkpoint
// trailer. This allows resume and rewind to function even after a squash merge
// collapses the original commits into a new one.
//
// If no commit can be found via the trailer search either, the original hash
// is returned as-is so callers can produce a meaningful error message.
func ResolveCommitHash(id, originalHash string) string {
	if git.CommitExists(originalHash) {
		return originalHash
	}
	// Squash-merge scenario: the original commit no longer exists. Search the
	// log for a commit whose message contains the Partio-Checkpoint trailer.
	found, err := git.FindCommitByCheckpointID(id)
	if err != nil || found == "" {
		return originalHash
	}
	return found
}
