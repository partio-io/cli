package claude

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/partio-io/cli/internal/agent"
)

// ParseJSONL reads a Claude Code JSONL transcript and extracts session data.
func ParseJSONL(path string) (*agent.SessionData, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening JSONL file: %w", err)
	}
	defer func() { _ = f.Close() }()

	var (
		messages    []agent.Message
		prompt      string
		sessionID   string
		totalTokens int
		firstTS     time.Time
		lastTS      time.Time
	)

	scanner := bufio.NewScanner(f)
	// Allow for large lines (Claude transcripts can be big)
	buf := make([]byte, 0, 1024*1024)
	scanner.Buffer(buf, 10*1024*1024)

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var entry jsonlEntry
		if err := json.Unmarshal(line, &entry); err != nil {
			continue // Skip malformed lines
		}

		if entry.SessionID != "" && sessionID == "" {
			sessionID = entry.SessionID
		}

		ts := time.Unix(int64(entry.Timestamp), 0)
		if firstTS.IsZero() {
			firstTS = ts
		}
		lastTS = ts

		// Extract message content
		text := extractText(entry)
		if text == "" {
			continue
		}

		role := entry.Role
		if role == "" {
			role = entry.Type
		}

		msg := agent.Message{
			Role:      role,
			Content:   text,
			Timestamp: ts,
		}
		messages = append(messages, msg)

		// First human message is the prompt
		if prompt == "" && role == "human" {
			prompt = text
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scanning JSONL: %w", err)
	}

	duration := lastTS.Sub(firstTS)

	return &agent.SessionData{
		SessionID:   sessionID,
		Agent:       "claude-code",
		Prompt:      prompt,
		Transcript:  messages,
		Context:     generateContext(messages),
		TotalTokens: totalTokens,
		Duration:    duration,
	}, nil
}

// extractText pulls text content from various JSONL entry formats.
func extractText(entry jsonlEntry) string {
	// Try content blocks first
	if len(entry.ContentBlocks) > 0 {
		var text string
		for _, b := range entry.ContentBlocks {
			if b.Type == "text" {
				text += b.Text
			}
		}
		return text
	}

	// Try message.content
	if entry.Message != nil {
		var mc messageContent
		if json.Unmarshal(entry.Message, &mc) == nil && len(mc.Content) > 0 {
			var text string
			for _, b := range mc.Content {
				if b.Type == "text" {
					text += b.Text
				}
			}
			return text
		}

		// Try as plain string
		var s string
		if json.Unmarshal(entry.Message, &s) == nil {
			return s
		}
	}

	// Try direct content
	if entry.Content != nil {
		// Try as array of blocks
		var blocks []contentBlock
		if json.Unmarshal(entry.Content, &blocks) == nil && len(blocks) > 0 {
			var text string
			for _, b := range blocks {
				if b.Type == "text" {
					text += b.Text
				}
			}
			return text
		}

		// Try as plain string
		var s string
		if json.Unmarshal(entry.Content, &s) == nil {
			return s
		}
	}

	return ""
}

// generateContext creates a human-readable summary of the session.
func generateContext(messages []agent.Message) string {
	if len(messages) == 0 {
		return ""
	}

	summary := "AI coding session"
	for _, m := range messages {
		if m.Role == "human" {
			if len(m.Content) > 200 {
				summary = m.Content[:200] + "..."
			} else {
				summary = m.Content
			}
			break
		}
	}

	return summary
}
