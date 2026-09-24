package git

import (
	"fmt"
	"os/exec"
	"strings"
)

// FetchCheckpointBranch fetches the checkpoint branch from the named remote,
// storing it as refs/remotes/<remote>/partio/checkpoints/v1.
func FetchCheckpointBranch(repoRoot, remote string) error {
	remoteRef := "refs/remotes/" + remote + "/" + CheckpointBranch
	refspec := CheckpointBranch + ":" + remoteRef
	cmd := exec.Command("git", "fetch", "--no-tags", remote, refspec)
	cmd.Dir = repoRoot
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("fetching %s from %s: %w: %s",
			CheckpointBranch, remote, err, strings.TrimSpace(string(out)))
	}
	return nil
}

// RemoteCheckpointRef returns the local tracking ref for the checkpoint branch
// on the named remote.
func RemoteCheckpointRef(remote string) string {
	return "refs/remotes/" + remote + "/" + CheckpointBranch
}

// IsAncestor reports whether ancestor is a reachable ancestor of descendant
// in the repository at repoRoot. Returns false on any error.
func IsAncestor(repoRoot, ancestor, descendant string) bool {
	cmd := exec.Command("git", "merge-base", "--is-ancestor", ancestor, descendant)
	cmd.Dir = repoRoot
	return cmd.Run() == nil
}
