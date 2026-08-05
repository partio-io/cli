package checkpoint

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// FindByBranch returns all checkpoints whose Branch field exactly matches the given branch name.
// Returns an empty slice (not an error) when the checkpoint branch does not exist or no checkpoints match.
// Returns an error if a git invocation on an existing checkpoint branch fails.
func FindByBranch(repoDir, branch string) ([]Checkpoint, error) {
	s := NewStore(repoDir)

	// If the checkpoint branch doesn't exist, treat as no checkpoints.
	_, err := s.git("rev-parse", "--verify", checkpointBranch)
	if err != nil {
		return []Checkpoint{}, nil
	}

	// List all shards.
	shards, err := s.git("ls-tree", "--name-only", checkpointBranch)
	if err != nil || shards == "" {
		return []Checkpoint{}, nil
	}

	var result []Checkpoint

	for _, shard := range strings.Split(shards, "\n") {
		if shard == "" {
			continue
		}
		entries, err := s.git("ls-tree", "--name-only", checkpointBranch+":"+shard)
		if err != nil {
			return nil, fmt.Errorf("listing shard %s: %w", shard, err)
		}
		for _, entry := range strings.Split(entries, "\n") {
			if entry == "" {
				continue
			}
			metaJSON, err := s.git("show", checkpointBranch+":"+shard+"/"+entry+"/metadata.json")
			if err != nil {
				return nil, fmt.Errorf("reading metadata for %s/%s: %w", shard, entry, err)
			}
			var meta Metadata
			if err := json.Unmarshal([]byte(metaJSON), &meta); err != nil {
				return nil, fmt.Errorf("parsing metadata for %s/%s: %w", shard, entry, err)
			}
			if meta.Branch == branch {
				cp, err := fromMetadata(meta)
				if err != nil {
					return nil, fmt.Errorf("converting metadata for %s/%s: %w", shard, entry, err)
				}
				result = append(result, cp)
			}
		}
	}

	return result, nil
}

// fromMetadata converts a Metadata value to a Checkpoint.
func fromMetadata(m Metadata) (Checkpoint, error) {
	createdAt, err := time.Parse(time.RFC3339, m.CreatedAt)
	if err != nil {
		return Checkpoint{}, fmt.Errorf("parsing created_at %q: %w", m.CreatedAt, err)
	}
	return Checkpoint{
		ID:          m.ID,
		SessionID:   m.SessionID,
		CommitHash:  m.CommitHash,
		Branch:      m.Branch,
		CreatedAt:   createdAt,
		Agent:       m.Agent,
		AgentPct:    m.AgentPercent,
		ContentHash: m.ContentHash,
		PlanSlug:    m.PlanSlug,
	}, nil
}
