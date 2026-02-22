package claude

import (
	"encoding/json"
	"time"
)

// flexTimestamp handles both Unix epoch (float64) and ISO 8601 (string) timestamps
// found in Claude Code JSONL entries.
type flexTimestamp struct {
	Time time.Time
}

func (ft *flexTimestamp) UnmarshalJSON(data []byte) error {
	// Try as float64 (Unix epoch) first
	var f float64
	if err := json.Unmarshal(data, &f); err == nil {
		ft.Time = time.Unix(int64(f), 0)
		return nil
	}

	// Try as string (ISO 8601)
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		t, err := time.Parse(time.RFC3339Nano, s)
		if err == nil {
			ft.Time = t
			return nil
		}
		// Also try without nanoseconds
		t, err = time.Parse(time.RFC3339, s)
		if err == nil {
			ft.Time = t
			return nil
		}
	}

	return nil
}

// jsonlEntry represents a single line in Claude's JSONL transcript.
type jsonlEntry struct {
	Type      string          `json:"type"`
	Role      string          `json:"role,omitempty"`
	Message   json.RawMessage `json:"message,omitempty"`
	Content   json.RawMessage `json:"content,omitempty"`
	Timestamp flexTimestamp   `json:"timestamp"`
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
