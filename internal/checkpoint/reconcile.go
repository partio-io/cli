package checkpoint

import (
	"fmt"
	"log/slog"
	"os/exec"
	"strings"

	"github.com/partio-io/cli/internal/git"
)

// ReconcileWithRemote fetches the remote checkpoint branch and merges it into
// the local checkpoint branch if they have diverged. Uses a union merge
// strategy: all checkpoints from both sides are preserved. If the same
// checkpoint ID appears on both sides, the local copy is kept (last-write-wins
// for local changes).
//
// Returns nil when reconciliation is not needed (no remote, no remote branch,
// or branches are already in sync). Returns a descriptive error if reconciliation
// fails in a way that may affect data integrity.
func ReconcileWithRemote(repoRoot, remote string) error {
	if !hasRemoteInDir(repoRoot, remote) {
		slog.Debug("no remote configured, skipping reconciliation", "remote", remote)
		return nil
	}

	if err := git.FetchCheckpointBranch(repoRoot, remote); err != nil {
		return fmt.Errorf("reconcile: %w", err)
	}

	s := NewStore(repoRoot)
	remoteRef := git.RemoteCheckpointRef(remote)

	remoteCommit, err := s.git("rev-parse", "--verify", remoteRef)
	if err != nil {
		slog.Debug("remote checkpoint branch not found after fetch", "ref", remoteRef)
		return nil
	}

	localCommit, err := s.git("rev-parse", "--verify", "refs/heads/"+checkpointBranch)
	if err != nil {
		// Local branch does not exist — initialise from remote.
		slog.Debug("local checkpoint branch missing, initialising from remote")
		_, err = s.git("update-ref", "refs/heads/"+checkpointBranch, remoteCommit)
		if err != nil {
			return fmt.Errorf("reconcile: initialising local branch from remote: %w", err)
		}
		return nil
	}

	if localCommit == remoteCommit {
		return nil
	}

	// Local already contains all remote commits — nothing to do.
	if git.IsAncestor(repoRoot, remoteCommit, localCommit) {
		slog.Debug("checkpoint branch up to date, no reconciliation needed")
		return nil
	}

	// Remote is strictly ahead of local — fast-forward.
	if git.IsAncestor(repoRoot, localCommit, remoteCommit) {
		slog.Debug("fast-forwarding local checkpoint branch to remote", "commit", remoteCommit)
		_, err = s.git("update-ref", "refs/heads/"+checkpointBranch, remoteCommit)
		if err != nil {
			return fmt.Errorf("reconcile: fast-forward: %w", err)
		}
		return nil
	}

	// Both branches have diverged — perform a union merge.
	slog.Debug("checkpoint branches have diverged, merging",
		"local", localCommit, "remote", remoteCommit)
	return s.mergeRemoteCheckpoints(localCommit, remoteCommit)
}

// mergeRemoteCheckpoints creates a union merge commit from localCommit and
// remoteCommit and updates the local checkpoint branch ref to the result.
func (s *Store) mergeRemoteCheckpoints(localCommit, remoteCommit string) error {
	localTree, err := s.git("rev-parse", localCommit+"^{tree}")
	if err != nil {
		return fmt.Errorf("reconcile: resolving local tree: %w", err)
	}

	remoteTree, err := s.git("rev-parse", remoteCommit+"^{tree}")
	if err != nil {
		return fmt.Errorf("reconcile: resolving remote tree: %w", err)
	}

	mergedTree, err := s.mergeCheckpointTrees(localTree, remoteTree)
	if err != nil {
		return fmt.Errorf("reconcile: merging trees: %w", err)
	}

	mergeCommit, err := s.git("commit-tree", mergedTree,
		"-p", localCommit,
		"-p", remoteCommit,
		"-m", "partio: reconcile checkpoint branch")
	if err != nil {
		return fmt.Errorf("reconcile: creating merge commit: %w", err)
	}

	if _, err = s.git("update-ref", "refs/heads/"+checkpointBranch, mergeCommit); err != nil {
		return fmt.Errorf("reconcile: updating branch ref: %w", err)
	}

	slog.Debug("checkpoint branches reconciled", "merge_commit", mergeCommit)
	return nil
}

// mergeCheckpointTrees merges two checkpoint root trees using a union strategy.
// Shards present on only one side are kept as-is. Shards on both sides have
// their checkpoint entries merged. Local wins if the same checkpoint ID appears
// in both (last-write-wins for local changes).
func (s *Store) mergeCheckpointTrees(localTree, remoteTree string) (string, error) {
	localShards, err := s.listTreeEntries(localTree)
	if err != nil {
		return "", fmt.Errorf("listing local checkpoint tree: %w", err)
	}

	remoteShards, err := s.listTreeEntries(remoteTree)
	if err != nil {
		return "", fmt.Errorf("listing remote checkpoint tree: %w", err)
	}

	localIdx := make(map[string]treeEntry, len(localShards))
	for _, e := range localShards {
		localIdx[e.name] = e
	}
	remoteIdx := make(map[string]treeEntry, len(remoteShards))
	for _, e := range remoteShards {
		remoteIdx[e.name] = e
	}

	// Collect all shard names from both sides.
	names := make(map[string]struct{})
	for _, e := range localShards {
		names[e.name] = struct{}{}
	}
	for _, e := range remoteShards {
		names[e.name] = struct{}{}
	}

	var rootEntries []treeEntry
	for name := range names {
		le, inLocal := localIdx[name]
		re, inRemote := remoteIdx[name]

		switch {
		case inLocal && !inRemote:
			rootEntries = append(rootEntries, le)
		case !inLocal && inRemote:
			rootEntries = append(rootEntries, re)
		default:
			// Present on both sides — merge shard trees.
			mergedShard, mergeErr := s.mergeShardTrees(le.hash, re.hash)
			if mergeErr != nil {
				return "", fmt.Errorf("merging shard %s: %w", name, mergeErr)
			}
			rootEntries = append(rootEntries, treeEntry{
				mode: "040000",
				typ:  "tree",
				hash: mergedShard,
				name: name,
			})
		}
	}

	return s.mktree(rootEntries)
}

// mergeShardTrees merges two shard-level trees with a union strategy.
// Local wins when the same checkpoint ID appears on both sides.
func (s *Store) mergeShardTrees(localHash, remoteHash string) (string, error) {
	remoteEntries, err := s.listTreeEntries(remoteHash)
	if err != nil {
		return "", fmt.Errorf("listing remote shard: %w", err)
	}

	localEntries, err := s.listTreeEntries(localHash)
	if err != nil {
		return "", fmt.Errorf("listing local shard: %w", err)
	}

	merged := make(map[string]treeEntry)
	for _, e := range remoteEntries {
		merged[e.name] = e
	}
	for _, e := range localEntries {
		merged[e.name] = e // local wins on conflict
	}

	var entries []treeEntry
	for _, e := range merged {
		entries = append(entries, e)
	}

	return s.mktree(entries)
}

// listTreeEntries returns the immediate children of a tree object.
// Returns nil (not an error) when the tree is empty.
func (s *Store) listTreeEntries(treeHash string) ([]treeEntry, error) {
	out, err := s.git("ls-tree", treeHash)
	if err != nil {
		return nil, err
	}
	return parseLsTreeEntries(out), nil
}

// parseLsTreeEntries parses "git ls-tree" output into treeEntry values.
// Format per line: <mode> SP <type> SP <hash> TAB <name>
func parseLsTreeEntries(out string) []treeEntry {
	var entries []treeEntry
	for _, line := range strings.Split(out, "\n") {
		if line == "" {
			continue
		}
		tabIdx := strings.IndexByte(line, '\t')
		if tabIdx < 0 {
			continue
		}
		name := line[tabIdx+1:]
		fields := strings.Fields(line[:tabIdx])
		if len(fields) < 3 {
			continue
		}
		entries = append(entries, treeEntry{
			mode: fields[0],
			typ:  fields[1],
			hash: fields[2],
			name: name,
		})
	}
	return entries
}

// hasRemoteInDir checks whether the repository at repoRoot has a remote with
// the given name.
func hasRemoteInDir(repoRoot, remote string) bool {
	cmd := exec.Command("git", "remote", "get-url", remote)
	cmd.Dir = repoRoot
	return cmd.Run() == nil
}
