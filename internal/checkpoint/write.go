package checkpoint

import (
	"encoding/json"
	"fmt"
)

// Write stores a checkpoint and its associated session data on the orphan branch.
func (s *Store) Write(cp *Checkpoint, sessionData *SessionFiles) error {
	shard := Shard(cp.ID)
	rest := Rest(cp.ID)

	// Hash all the blob objects
	metaJSON, err := json.MarshalIndent(cp.ToMetadata(), "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling metadata: %w", err)
	}

	metaHash, err := s.hashObject(string(metaJSON))
	if err != nil {
		return fmt.Errorf("hashing metadata: %w", err)
	}

	contentHashHash, err := s.hashObject(sessionData.ContentHash)
	if err != nil {
		return fmt.Errorf("hashing content_hash: %w", err)
	}

	contextHash, err := s.hashObject(sessionData.Context)
	if err != nil {
		return fmt.Errorf("hashing context: %w", err)
	}

	fullHash, err := s.hashObject(sessionData.FullJSONL)
	if err != nil {
		return fmt.Errorf("hashing full.jsonl: %w", err)
	}

	sessionMetaJSON, err := json.MarshalIndent(sessionData.Metadata, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling session metadata: %w", err)
	}
	sessionMetaHash, err := s.hashObject(string(sessionMetaJSON))
	if err != nil {
		return fmt.Errorf("hashing session metadata: %w", err)
	}

	promptHash, err := s.hashObject(sessionData.Prompt)
	if err != nil {
		return fmt.Errorf("hashing prompt: %w", err)
	}

	diffHash, err := s.hashObject(sessionData.Diff)
	if err != nil {
		return fmt.Errorf("hashing diff: %w", err)
	}

	// Build session subtree (0/)
	sessionTree, err := s.mktree([]treeEntry{
		{mode: "100644", typ: "blob", hash: contentHashHash, name: "content_hash.txt"},
		{mode: "100644", typ: "blob", hash: contextHash, name: "context.md"},
		{mode: "100644", typ: "blob", hash: diffHash, name: "diff.patch"},
		{mode: "100644", typ: "blob", hash: fullHash, name: "full.jsonl"},
		{mode: "100644", typ: "blob", hash: sessionMetaHash, name: "metadata.json"},
		{mode: "100644", typ: "blob", hash: promptHash, name: "prompt.txt"},
	})
	if err != nil {
		return fmt.Errorf("creating session tree: %w", err)
	}

	// Build checkpoint subtree (<rest>/)
	cpTree, err := s.mktree([]treeEntry{
		{mode: "100644", typ: "blob", hash: metaHash, name: "metadata.json"},
		{mode: "040000", typ: "tree", hash: sessionTree, name: "0"},
	})
	if err != nil {
		return fmt.Errorf("creating checkpoint tree: %w", err)
	}

	// Get current tree of the branch
	currentTree, err := s.getCurrentTree()
	if err != nil {
		return fmt.Errorf("getting current tree: %w", err)
	}

	// Build new root tree incorporating the existing tree + new shard entry
	newRoot, err := s.addToTree(currentTree, shard, rest, cpTree)
	if err != nil {
		return fmt.Errorf("building root tree: %w", err)
	}

	// Create commit
	parentCommit, err := s.git("rev-parse", checkpointBranch)
	if err != nil {
		return fmt.Errorf("getting parent commit: %w", err)
	}

	commitMsg := fmt.Sprintf("checkpoint: %s", cp.ID)
	commitHash, err := s.git("commit-tree", newRoot, "-p", parentCommit, "-m", commitMsg)
	if err != nil {
		return fmt.Errorf("creating commit: %w", err)
	}

	// Update ref
	_, err = s.git("update-ref", "refs/heads/"+checkpointBranch, commitHash)
	if err != nil {
		return fmt.Errorf("updating ref: %w", err)
	}

	return nil
}
