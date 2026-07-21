package checkpoint

import (
	"encoding/json"
	"strings"

	"github.com/partio-io/cli/internal/git"
)

// FindByCommit searches all stored checkpoints for those whose commit hash matches commitSHA.
// Returns a slice of checkpoint IDs. Returns nil (no error) when the checkpoint branch is
// absent or no checkpoints match.
func FindByCommit(commitSHA string) ([]string, error) {
	shards, err := git.ExecGit("ls-tree", "--name-only", git.CheckpointBranch)
	if err != nil || strings.TrimSpace(shards) == "" {
		return nil, nil
	}

	var ids []string
	for _, shard := range strings.Split(strings.TrimSpace(shards), "\n") {
		if shard == "" {
			continue
		}
		entries, err := git.ExecGit("ls-tree", "--name-only", git.CheckpointBranch+":"+shard)
		if err != nil {
			continue
		}
		for _, entry := range strings.Split(strings.TrimSpace(entries), "\n") {
			if entry == "" {
				continue
			}
			metaJSON, err := git.ExecGit("show", git.CheckpointBranch+":"+shard+"/"+entry+"/metadata.json")
			if err != nil {
				continue
			}
			var meta Metadata
			if err := json.Unmarshal([]byte(metaJSON), &meta); err != nil {
				continue
			}
			if meta.CommitHash == commitSHA {
				ids = append(ids, shard+entry)
			}
		}
	}

	return ids, nil
}
