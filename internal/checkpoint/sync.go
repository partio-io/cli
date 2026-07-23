package checkpoint

import (
	"fmt"
	"strings"

	"github.com/partio-io/cli/internal/git"
)

// SyncWithRemote fetches the remote checkpoint branch using --filter=blob:none,
// merges remote checkpoint entries into the local branch by unioning tree entries,
// and updates the local ref. Returns true if the remote had the branch, false if not.
func (s *Store) SyncWithRemote(remote string) (bool, error) {
	if err := git.FetchBranch(remote, checkpointBranch); err != nil {
		return false, fmt.Errorf("fetching remote checkpoint branch: %w", err)
	}

	remoteCommit, err := s.git("rev-parse", "FETCH_HEAD")
	if err != nil {
		return false, fmt.Errorf("resolving FETCH_HEAD: %w", err)
	}

	localCommit, err := s.git("rev-parse", checkpointBranch)
	if err != nil {
		return false, fmt.Errorf("resolving local checkpoint branch: %w", err)
	}

	if remoteCommit == localCommit {
		return true, nil
	}

	// If remote is already an ancestor of local, local is ahead — nothing to merge.
	if s.isAncestor(remoteCommit, localCommit) {
		return true, nil
	}

	localTree, err := s.git("rev-parse", checkpointBranch+"^{tree}")
	if err != nil {
		return false, fmt.Errorf("getting local tree: %w", err)
	}

	remoteTree, err := s.git("rev-parse", "FETCH_HEAD^{tree}")
	if err != nil {
		return false, fmt.Errorf("getting remote tree: %w", err)
	}

	mergedTree, err := s.mergeTrees(localTree, remoteTree)
	if err != nil {
		return false, fmt.Errorf("merging checkpoint trees: %w", err)
	}

	commitHash, err := s.git("commit-tree", mergedTree,
		"-p", localCommit,
		"-p", remoteCommit,
		"-m", "sync: merge remote checkpoint entries",
	)
	if err != nil {
		return false, fmt.Errorf("creating merge commit: %w", err)
	}

	_, err = s.git("update-ref", "refs/heads/"+checkpointBranch, commitHash)
	if err != nil {
		return false, fmt.Errorf("updating checkpoint branch ref: %w", err)
	}

	return true, nil
}

// isAncestor returns true if candidate is an ancestor of descendant.
func (s *Store) isAncestor(candidate, descendant string) bool {
	_, err := s.git("merge-base", "--is-ancestor", candidate, descendant)
	return err == nil
}

// mergeTrees merges two checkpoint root trees by unioning shard entries.
// Within each shard, checkpoint entries (identified by UUID suffix) are unioned.
// Local entries take precedence when the same UUID exists in both.
func (s *Store) mergeTrees(localTree, remoteTree string) (string, error) {
	localShards, err := s.parseTree(localTree)
	if err != nil {
		return "", fmt.Errorf("parsing local tree: %w", err)
	}

	remoteShards, err := s.parseTree(remoteTree)
	if err != nil {
		return "", fmt.Errorf("parsing remote tree: %w", err)
	}

	for name, remoteEntry := range remoteShards {
		localEntry, exists := localShards[name]
		if !exists {
			localShards[name] = remoteEntry
		} else {
			mergedShardTree, err := s.mergeShardTrees(localEntry.hash, remoteEntry.hash)
			if err != nil {
				return "", fmt.Errorf("merging shard %s: %w", name, err)
			}
			localShards[name] = treeEntry{
				mode: localEntry.mode,
				typ:  localEntry.typ,
				hash: mergedShardTree,
				name: localEntry.name,
			}
		}
	}

	var entries []treeEntry
	for _, e := range localShards {
		entries = append(entries, e)
	}
	return s.mktree(entries)
}

// mergeShardTrees unions checkpoint entries within a shard tree.
// Local entries take precedence when the same checkpoint UUID exists in both.
func (s *Store) mergeShardTrees(localShardTree, remoteShardTree string) (string, error) {
	localEntries, err := s.parseTree(localShardTree)
	if err != nil {
		return "", err
	}
	remoteEntries, err := s.parseTree(remoteShardTree)
	if err != nil {
		return "", err
	}

	for name, e := range remoteEntries {
		if _, exists := localEntries[name]; !exists {
			localEntries[name] = e
		}
	}

	var entries []treeEntry
	for _, e := range localEntries {
		entries = append(entries, e)
	}
	return s.mktree(entries)
}

// parseTree reads a git tree object and returns a map of name -> treeEntry.
func (s *Store) parseTree(tree string) (map[string]treeEntry, error) {
	out, _ := s.git("ls-tree", tree)
	result := make(map[string]treeEntry)
	if out == "" {
		return result, nil
	}
	for _, line := range strings.Split(out, "\n") {
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		tabParts := strings.SplitN(line, "\t", 2)
		name := ""
		if len(tabParts) >= 2 {
			name = tabParts[1]
		}
		if len(parts) >= 3 {
			result[name] = treeEntry{
				mode: parts[0],
				typ:  parts[1],
				hash: parts[2],
				name: name,
			}
		}
	}
	return result, nil
}
