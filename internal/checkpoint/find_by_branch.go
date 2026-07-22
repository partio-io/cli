package checkpoint

import (
	"encoding/json"
	"strings"
)

// FindByBranch returns all checkpoints whose Branch field exactly matches the
// given branch name. The returned slice is unsorted. An empty slice (not an
// error) is returned when no checkpoints match.
func FindByBranch(repoDir, branch string) ([]Metadata, error) {
	s := NewStore(repoDir)

	// Check branch exists
	_, err := s.git("rev-parse", "--verify", checkpointBranch)
	if err != nil {
		return nil, nil
	}

	// List all shard directories
	shards, err := s.git("ls-tree", "--name-only", checkpointBranch)
	if err != nil || shards == "" {
		return nil, nil
	}

	var matches []Metadata

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
			if meta.Branch == branch {
				matches = append(matches, meta)
			}
		}
	}

	return matches, nil
}
