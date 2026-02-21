package claude

import "encoding/json"

// jsonlEntry represents a single line in Claude's JSONL transcript.
type jsonlEntry struct {
	Type      string          `json:"type"`
	Role      string          `json:"role,omitempty"`
	Message   json.RawMessage `json:"message,omitempty"`
	Content   json.RawMessage `json:"content,omitempty"`
	Timestamp float64         `json:"timestamp,omitempty"`
	SessionID string          `json:"sessionId,omitempty"`
	Slug      string          `json:"slug,omitempty"`

	// For content blocks
	ContentBlocks []contentBlock `json:"contentBlocks,omitempty"`
}

type contentBlock struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

// messageContent extracts text content from a JSONL entry.
type messageContent struct {
	Role    string         `json:"role"`
	Content []contentBlock `json:"content"`
}
