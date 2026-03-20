package session

import (
	"time"

	"github.com/google/uuid"
)

// Session represents an AI agent coding session.
type Session struct {
	ID        string    `json:"id"`
	Agent     string    `json:"agent"`
	State     State     `json:"state"`
	StartedAt time.Time `json:"started_at"`
	EndedAt   time.Time `json:"ended_at,omitempty"`
	Branch    string    `json:"branch"`
	SourceDir string    `json:"source_dir"`

	// Condensed is true when the session has been fully captured in a checkpoint
	// and has ended. Post-commit sets this after creating a checkpoint so future
	// commits with the same session can be skipped.
	Condensed         bool      `json:"condensed,omitempty"`
	CapturedSessionID string    `json:"captured_session_id,omitempty"`
	CapturedAt        time.Time `json:"captured_at,omitempty"`
}

// New creates a new session with a generated UUID.
func New(agent, branch, sourceDir string) *Session {
	return &Session{
		ID:        uuid.New().String(),
		Agent:     agent,
		State:     StateActive,
		StartedAt: time.Now(),
		Branch:    branch,
		SourceDir: sourceDir,
	}
}
