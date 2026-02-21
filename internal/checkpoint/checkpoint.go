package checkpoint

import (
	"crypto/rand"
	"encoding/hex"
	"time"
)

// Checkpoint represents a captured point-in-time snapshot.
type Checkpoint struct {
	ID          string    `json:"id"`
	SessionID   string    `json:"session_id"`
	CommitHash  string    `json:"commit_hash"`
	Branch      string    `json:"branch"`
	CreatedAt   time.Time `json:"created_at"`
	Agent       string    `json:"agent"`
	AgentPct    int       `json:"agent_percent"`
	ContentHash string    `json:"content_hash"`
	PlanSlug    string    `json:"plan_slug,omitempty"`
}

// Metadata is the JSON schema for checkpoint metadata stored on the orphan branch.
type Metadata struct {
	ID           string `json:"id"`
	SessionID    string `json:"session_id"`
	CommitHash   string `json:"commit_hash"`
	Branch       string `json:"branch"`
	CreatedAt    string `json:"created_at"`
	Agent        string `json:"agent"`
	AgentPercent int    `json:"agent_percent"`
	ContentHash  string `json:"content_hash"`
	PlanSlug     string `json:"plan_slug,omitempty"`
}

// NewID generates a 12-character hex checkpoint ID.
func NewID() string {
	b := make([]byte, 6)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// Shard returns the 2-char shard prefix for a checkpoint ID.
func Shard(id string) string {
	if len(id) < 2 {
		return id
	}
	return id[:2]
}

// Rest returns the remaining chars after the shard prefix.
func Rest(id string) string {
	if len(id) <= 2 {
		return ""
	}
	return id[2:]
}
