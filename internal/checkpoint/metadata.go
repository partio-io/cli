package checkpoint

// SessionMetadata is stored per-session within a checkpoint directory.
type SessionMetadata struct {
	Agent       string `json:"agent"`
	Model       string `json:"model,omitempty"`
	TotalTokens int    `json:"total_tokens"`
	Duration    string `json:"duration"`
}
