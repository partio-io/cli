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
