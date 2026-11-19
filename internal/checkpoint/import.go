package checkpoint

import (
	"encoding/json"
	"fmt"
	"strings"
)

// ImportOutcome is what Import did with one checkpoint ID.
type ImportOutcome int

const (
	// ImportSynced means the checkpoint was copied into the local branch.
	ImportSynced ImportOutcome = iota
	// ImportPresent means the local branch already holds an identical copy.
	ImportPresent
	// ImportConflict means the local branch holds a different copy, which
	// was kept untouched.
	ImportConflict
	// ImportMissing means the source tree has no checkpoint with this ID.
	ImportMissing
	// ImportRejected means the source checkpoint failed validation.
	ImportRejected
)

// ImportResult is the outcome for one checkpoint ID. Reason explains a
// conflict or a rejection.
type ImportResult struct {
	ID      string
	Outcome ImportOutcome
	Reason  string
}

// Import copies the checkpoints named by ids from the checkpoint tree at
// sourceRef (for example a fetched fork branch) into the local checkpoint
// branch. ids must be distinct, valid checkpoint IDs. Each source checkpoint
// must be a tree of regular-file blobs and trees whose metadata.json carries
// the same ID; it is grafted unchanged, so its tree hash is identical to the
// source's. Checkpoints already present locally are never overwritten.
//
// All new checkpoints land in one commit parented only on the current local
// tip, with the message message(synced). No commit is written when nothing
// new was copied. The branch is created first if it is missing and at least
// one checkpoint will be written.
func (s *Store) Import(sourceRef string, ids []string, message func(synced []string) string) ([]ImportResult, error) {
	sourceTree, err := s.git("rev-parse", "--verify", sourceRef+"^{tree}")
	if err != nil {
		return nil, fmt.Errorf("reading source tree %s: %w", sourceRef, err)
	}

	exists, err := s.branchExists()
	if err != nil {
		return nil, err
	}
	localTree := ""
	if exists {
		if localTree, err = s.getCurrentTree(); err != nil {
			return nil, fmt.Errorf("reading local checkpoint tree: %w", err)
		}
	}

	type graft struct{ id, tree string }
	var grafts []graft
	var results []ImportResult
	for _, id := range ids {
		res, tree, err := s.classify(sourceTree, localTree, id)
		if err != nil {
			return nil, err
		}
		results = append(results, res)
		if res.Outcome == ImportSynced {
			grafts = append(grafts, graft{id: id, tree: tree})
		}
	}

	if len(grafts) == 0 {
		return results, nil
	}

	if !exists {
		if err := s.InitBranch(); err != nil {
			return nil, err
		}
		if localTree, err = s.getCurrentTree(); err != nil {
			return nil, fmt.Errorf("reading local checkpoint tree: %w", err)
		}
	}

	parent, err := s.git("rev-parse", "--verify", "refs/heads/"+checkpointBranch)
	if err != nil {
		return nil, fmt.Errorf("getting parent commit: %w", err)
	}

	root := localTree
	synced := make([]string, 0, len(grafts))
	for _, g := range grafts {
		if root, err = s.addToTree(root, Shard(g.id), Rest(g.id), g.tree); err != nil {
			return nil, fmt.Errorf("grafting checkpoint %s: %w", g.id, err)
		}
		synced = append(synced, g.id)
	}

	commit, err := s.git("commit-tree", root, "-p", parent, "-m", message(synced))
	if err != nil {
		return nil, fmt.Errorf("creating commit: %w", err)
	}
	// Passing the old value makes the update fail rather than clobber a
	// concurrent writer that moved the branch since it was read.
	if _, err := s.git("update-ref", "refs/heads/"+checkpointBranch, commit, parent); err != nil {
		return nil, fmt.Errorf("updating ref: %w", err)
	}
	return results, nil
}

// classify decides what to do with checkpoint id. For ImportSynced it also
// returns the source subtree to graft. An empty localTree means the local
// branch does not exist yet.
func (s *Store) classify(sourceTree, localTree, id string) (ImportResult, string, error) {
	res := ImportResult{ID: id}
	path := Shard(id) + "/" + Rest(id)

	srcMode, srcType, srcHash, found, err := s.lookupEntry(sourceTree, path)
	if err != nil {
		return res, "", fmt.Errorf("looking up %s in source tree: %w", id, err)
	}
	if !found {
		res.Outcome = ImportMissing
		return res, "", nil
	}

	if localTree != "" {
		_, _, localHash, localFound, err := s.lookupEntry(localTree, path)
		if err != nil {
			return res, "", fmt.Errorf("looking up %s in local tree: %w", id, err)
		}
		if localFound {
			if localHash == srcHash {
				res.Outcome = ImportPresent
			} else {
				res.Outcome, res.Reason = ImportConflict, fmt.Sprintf("local tree %s differs from source tree %s", localHash, srcHash)
			}
			return res, "", nil
		}
	}

	if srcMode != "040000" || srcType != "tree" {
		res.Outcome, res.Reason = ImportRejected, fmt.Sprintf("entry is %s %s, not a tree", srcMode, srcType)
		return res, "", nil
	}
	if reason, err := s.validateCheckpointTree(srcHash, id); err != nil {
		return res, "", err
	} else if reason != "" {
		res.Outcome, res.Reason = ImportRejected, reason
		return res, "", nil
	}

	res.Outcome = ImportSynced
	return res, srcHash, nil
}

// validateCheckpointTree checks that tree holds only regular-file blobs and
// trees, recursively, and a metadata.json whose ID is id. It returns a
// non-empty reason when the tree is invalid, and an error only when git
// itself fails.
func (s *Store) validateCheckpointTree(tree, id string) (string, error) {
	listing, err := s.git("ls-tree", "-r", "-t", "-z", tree)
	if err != nil {
		return "", fmt.Errorf("listing checkpoint tree %s: %w", id, err)
	}
	for _, rec := range strings.Split(listing, "\x00") {
		if rec == "" {
			continue
		}
		mode, typ, _, name, ok := parseTreeLine(rec)
		if !ok {
			return "", fmt.Errorf("unexpected ls-tree output for %s: %q", id, rec)
		}
		regularFile := mode == "100644" && typ == "blob"
		directory := mode == "040000" && typ == "tree"
		if !regularFile && !directory {
			return fmt.Sprintf("entry %q has mode %s (%s); only regular files and directories are allowed", name, mode, typ), nil
		}
	}

	mode, typ, hash, found, err := s.lookupEntry(tree, "metadata.json")
	if err != nil {
		return "", fmt.Errorf("looking up metadata.json in %s: %w", id, err)
	}
	if !found || mode != "100644" || typ != "blob" {
		return "metadata.json is missing", nil
	}
	content, err := s.git("cat-file", "blob", hash)
	if err != nil {
		return "", fmt.Errorf("reading metadata.json of %s: %w", id, err)
	}
	var meta Metadata
	if err := json.Unmarshal([]byte(content), &meta); err != nil {
		return fmt.Sprintf("metadata.json does not parse: %v", err), nil
	}
	if meta.ID != id {
		return fmt.Sprintf("metadata.json names checkpoint %q", meta.ID), nil
	}
	return "", nil
}

// lookupEntry finds path in tree. found is false, with no error, when tree
// has no such entry.
func (s *Store) lookupEntry(tree, path string) (mode, typ, hash string, found bool, err error) {
	out, err := s.git("ls-tree", "-z", tree, "--", path)
	if err != nil {
		return "", "", "", false, err
	}
	for _, rec := range strings.Split(out, "\x00") {
		m, t, h, name, ok := parseTreeLine(rec)
		if ok && name == path {
			return m, t, h, true, nil
		}
	}
	return "", "", "", false, nil
}

// parseTreeLine splits one "<mode> <type> <hash>\t<name>" ls-tree record.
func parseTreeLine(rec string) (mode, typ, hash, name string, ok bool) {
	meta, name, found := strings.Cut(rec, "\t")
	if !found {
		return "", "", "", "", false
	}
	parts := strings.Fields(meta)
	if len(parts) != 3 {
		return "", "", "", "", false
	}
	return parts[0], parts[1], parts[2], name, true
}
