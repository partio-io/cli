package checkpoint

import "time"

// ToMetadata converts a Checkpoint to its storage Metadata format.
func (c *Checkpoint) ToMetadata() Metadata {
	return Metadata{
		ID:           c.ID,
		SessionID:    c.SessionID,
		CommitHash:   c.CommitHash,
		Branch:       c.Branch,
		CreatedAt:    c.CreatedAt.Format(time.RFC3339),
		Agent:        c.Agent,
		AgentPercent: c.AgentPct,
		ContentHash:  c.ContentHash,
		PlanSlug:     c.PlanSlug,
	}
}
