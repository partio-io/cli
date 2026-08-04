package checkpoint

import (
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/partio-io/cli/internal/git"
)

// CheckpointData holds all readable data from a stored checkpoint.
type CheckpointData struct {
	Metadata Metadata
	Prompt   string
	Plan     string
	Diff     string
	Context  string
}

// Read retrieves all checkpoint data from the orphan branch by ID.
func Read(id string) (*CheckpointData, error) {
	if len(id) != 12 {
		return nil, fmt.Errorf("checkpoint ID must be 12 characters (got %d)", len(id))
	}

	shard := Shard(id)
	rest := Rest(id)
	prefix := git.CheckpointBranch + ":" + shard + "/" + rest

	// Read metadata (required)
	metaJSON, err := git.ExecGit("show", prefix+"/metadata.json")
	if err != nil {
		return nil, fmt.Errorf("checkpoint %s not found", id)
	}

	var meta Metadata
	if err := json.Unmarshal([]byte(metaJSON), &meta); err != nil {
		return nil, fmt.Errorf("invalid checkpoint metadata: %w", err)
	}

	// Read optional files — missing files are not errors
	prompt := showOptional(prefix + "/0/prompt.txt")
	plan := showOptional(prefix + "/0/plan.md")
	diff := showOptional(prefix + "/0/diff.patch")
	context := showOptional(prefix + "/0/context.md")

	return &CheckpointData{
		Metadata: meta,
		Prompt:   prompt,
		Plan:     plan,
		Diff:     diff,
		Context:  context,
	}, nil
}

// showOptional reads one optional checkpoint file. Absence is expected, so
// the failure is logged at debug and the file treated as empty.
func showOptional(path string) string {
	out, err := git.ExecGit("show", path)
	if err != nil {
		slog.Debug("optional checkpoint file not read", "path", path, "error", err)
	}
	return out
}
