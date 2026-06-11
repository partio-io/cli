package checkpoint

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// PruneResult holds the outcome of a prune operation.
type PruneResult struct {
	Removed []Metadata
	Kept    []Metadata
}

// Prune removes checkpoints older than the given duration, but never removes
// the checkpoint linked to currentCommitHash. If dryRun is true, no changes
// are made.
func (s *Store) Prune(olderThan time.Duration, currentCommitHash string, dryRun bool) (*PruneResult, error) {
	cutoff := time.Now().Add(-olderThan)
	result := &PruneResult{}

	// Check branch exists
	_, err := s.git("rev-parse", "--verify", checkpointBranch)
	if err != nil {
		return result, nil
	}

	// List all shards
	shards, err := s.git("ls-tree", "--name-only", checkpointBranch)
	if err != nil || shards == "" {
		return result, nil
	}

	// Collect all checkpoints and classify them
	var all []cpEntry

	for _, shard := range strings.Split(shards, "\n") {
		if shard == "" {
			continue
		}
		entries, err := s.git("ls-tree", "--name-only", checkpointBranch+":"+shard)
		if err != nil {
			continue
		}
		for _, entry := range strings.Split(entries, "\n") {
			if entry == "" {
				continue
			}
			metaJSON, err := s.git("show", checkpointBranch+":"+shard+"/"+entry+"/metadata.json")
			if err != nil {
				continue
			}
			var meta Metadata
			if err := json.Unmarshal([]byte(metaJSON), &meta); err != nil {
				continue
			}
			all = append(all, cpEntry{shard: shard, rest: entry, meta: meta})
		}
	}

	// Classify: keep vs remove
	keepIDs := make(map[string]bool)
	for _, cp := range all {
		createdAt, err := time.Parse(time.RFC3339, cp.meta.CreatedAt)
		if err != nil {
			// Can't parse time, keep it to be safe
			keepIDs[cp.meta.ID] = true
			result.Kept = append(result.Kept, cp.meta)
			continue
		}

		// Never delete checkpoint linked to current HEAD
		if cp.meta.CommitHash == currentCommitHash {
			keepIDs[cp.meta.ID] = true
			result.Kept = append(result.Kept, cp.meta)
			continue
		}

		if createdAt.Before(cutoff) {
			result.Removed = append(result.Removed, cp.meta)
		} else {
			keepIDs[cp.meta.ID] = true
			result.Kept = append(result.Kept, cp.meta)
		}
	}

	if len(result.Removed) == 0 || dryRun {
		return result, nil
	}

	// Rebuild the tree without removed checkpoints
	currentTree, err := s.getCurrentTree()
	if err != nil {
		return nil, fmt.Errorf("getting current tree: %w", err)
	}

	newRoot, err := s.rebuildTreeWithout(currentTree, keepIDs, all)
	if err != nil {
		return nil, fmt.Errorf("rebuilding tree: %w", err)
	}

	// Create commit
	parentCommit, err := s.git("rev-parse", checkpointBranch)
	if err != nil {
		return nil, fmt.Errorf("getting parent commit: %w", err)
	}

	commitMsg := fmt.Sprintf("prune: removed %d checkpoint(s)", len(result.Removed))
	commitHash, err := s.git("commit-tree", newRoot, "-p", parentCommit, "-m", commitMsg)
	if err != nil {
		return nil, fmt.Errorf("creating commit: %w", err)
	}

	// Update ref
	_, err = s.git("update-ref", "refs/heads/"+checkpointBranch, commitHash)
	if err != nil {
		return nil, fmt.Errorf("updating ref: %w", err)
	}

	return result, nil
}

// rebuildTreeWithout rebuilds the root tree keeping only checkpoints in keepIDs.
func (s *Store) rebuildTreeWithout(currentTree string, keepIDs map[string]bool, all []cpEntry) (string, error) {
	// Group kept entries by shard
	shardEntries := make(map[string][]cpEntry)
	for _, cp := range all {
		if keepIDs[cp.meta.ID] {
			shardEntries[cp.shard] = append(shardEntries[cp.shard], cp)
		}
	}

	// Read current root tree to get existing entries (in case there are non-shard entries)
	rootListing, _ := s.git("ls-tree", currentTree)

	var rootEntries []treeEntry

	for _, line := range strings.Split(rootListing, "\n") {
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		tabParts := strings.SplitN(line, "\t", 2)
		if len(parts) < 4 || len(tabParts) < 2 {
			continue
		}
		name := tabParts[1]

		entries, isShard := shardEntries[name]
		if !isShard {
			// Not a shard we know about (or empty shard), skip it if it's a shard with no kept entries
			// Check if any checkpoint was in this shard
			hasCheckpoints := false
			for _, cp := range all {
				if cp.shard == name {
					hasCheckpoints = true
					break
				}
			}
			if hasCheckpoints {
				// This shard had checkpoints but none are kept - omit from tree
				continue
			}
			// Not a checkpoint shard, preserve as-is
			rootEntries = append(rootEntries, treeEntry{
				mode: parts[0],
				typ:  parts[1],
				hash: parts[2],
				name: name,
			})
			continue
		}

		// Rebuild shard tree with only kept entries
		var shardTreeEntries []treeEntry
		for _, cp := range entries {
			// Read the original tree hash for this checkpoint
			cpTreeHash, err := s.getCheckpointTreeHash(currentTree, cp.shard, cp.rest)
			if err != nil {
				continue
			}
			shardTreeEntries = append(shardTreeEntries, treeEntry{
				mode: "040000",
				typ:  "tree",
				hash: cpTreeHash,
				name: cp.rest,
			})
		}

		if len(shardTreeEntries) == 0 {
			continue
		}

		newShardTree, err := s.mktree(shardTreeEntries)
		if err != nil {
			return "", fmt.Errorf("creating shard tree for %s: %w", name, err)
		}

		rootEntries = append(rootEntries, treeEntry{
			mode: "040000",
			typ:  "tree",
			hash: newShardTree,
			name: name,
		})
	}

	if len(rootEntries) == 0 {
		// All checkpoints pruned - create empty tree
		return s.mktree(nil)
	}

	return s.mktree(rootEntries)
}

// getCheckpointTreeHash gets the tree hash for a specific checkpoint entry.
func (s *Store) getCheckpointTreeHash(rootTree, shard, rest string) (string, error) {
	shardListing, err := s.git("ls-tree", rootTree)
	if err != nil {
		return "", err
	}

	var shardTreeHash string
	for _, line := range strings.Split(shardListing, "\n") {
		tabParts := strings.SplitN(line, "\t", 2)
		if len(tabParts) < 2 {
			continue
		}
		if tabParts[1] == shard {
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				shardTreeHash = parts[2]
			}
			break
		}
	}

	if shardTreeHash == "" {
		return "", fmt.Errorf("shard %s not found", shard)
	}

	entryListing, err := s.git("ls-tree", shardTreeHash)
	if err != nil {
		return "", err
	}

	for _, line := range strings.Split(entryListing, "\n") {
		tabParts := strings.SplitN(line, "\t", 2)
		if len(tabParts) < 2 {
			continue
		}
		if tabParts[1] == rest {
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				return parts[2], nil
			}
		}
	}

	return "", fmt.Errorf("checkpoint %s/%s not found", shard, rest)
}

type cpEntry struct {
	shard string
	rest  string
	meta  Metadata
}
