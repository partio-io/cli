package checkpoint

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/partio-io/cli/internal/git"
)

// ErrNoBranch is returned when the checkpoint branch does not exist.
var ErrNoBranch = fmt.Errorf("checkpoint branch does not exist")

// List enumerates all checkpoints on the orphan branch and returns them
// sorted newest-first by CreatedAt.
func List() ([]Metadata, error) {
	branch := git.CheckpointBranch

	// Check branch exists
	_, err := git.ExecGit("rev-parse", "--verify", branch)
	if err != nil {
		return nil, ErrNoBranch
	}

	// List top-level shard directories
	shards, err := git.ExecGit("ls-tree", "--name-only", branch)
	if err != nil || shards == "" {
		return nil, nil
	}

	var results []Metadata

	for _, shard := range strings.Split(shards, "\n") {
		entries, err := git.ExecGit("ls-tree", "--name-only", branch+":"+shard)
		if err != nil {
			continue
		}
		for _, entry := range strings.Split(entries, "\n") {
			if entry == "" {
				continue
			}
			metaJSON, err := git.ExecGit("show", branch+":"+shard+"/"+entry+"/metadata.json")
			if err != nil {
				continue
			}

			var meta Metadata
			if err := json.Unmarshal([]byte(metaJSON), &meta); err != nil {
				continue
			}
			results = append(results, meta)
		}
	}

	// Sort newest-first by CreatedAt
	sort.Slice(results, func(i, j int) bool {
		ti, _ := time.Parse(time.RFC3339, results[i].CreatedAt)
		tj, _ := time.Parse(time.RFC3339, results[j].CreatedAt)
		return ti.After(tj)
	})

	return results, nil
}
