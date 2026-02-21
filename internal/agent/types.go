package agent

import "time"

// SessionData holds extracted data from an agent session.
type SessionData struct {
	SessionID   string        `json:"session_id"`
	Agent       string        `json:"agent"`
	Prompt      string        `json:"prompt"`
	Transcript  []Message     `json:"transcript"`
	Context     string        `json:"context"`
	TotalTokens int           `json:"total_tokens"`
	Duration    time.Duration `json:"duration"`
	PlanSlug    string        `json:"plan_slug,omitempty"`
}

// Message represents a single message in an agent transcript.
type Message struct {
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
	Tokens    int       `json:"tokens,omitempty"`
}
