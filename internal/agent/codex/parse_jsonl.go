package codex

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/partio-io/cli/internal/agent"
)

// jsonlLine is the top-level structure of a Codex JSONL line.
type jsonlLine struct {
	Timestamp string          `json:"timestamp"`
	Type      string          `json:"type"`
	Payload   json.RawMessage `json:"payload"`
}

type sessionMeta struct {
	ID         string `json:"id"`
	CWD        string `json:"cwd"`
	CLIVersion string `json:"cli_version"`
}

type eventMsg struct {
	Type    string `json:"type"`
	Message string `json:"message,omitempty"`
}

type responseItem struct {
	Type      string `json:"type"`
	Name      string `json:"name,omitempty"`
	Arguments string `json:"arguments,omitempty"`
	Content   []struct {
		Text string `json:"text"`
	} `json:"content,omitempty"`
}

// ParseJSONL parses a Codex JSONL session file into SessionData.
func ParseJSONL(path string) (*agent.SessionData, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening codex session: %w", err)
	}
	defer f.Close()

	data := &agent.SessionData{
		Agent: "codex",
	}

	var firstTS, lastTS time.Time
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 1024*1024), 10*1024*1024)

	for scanner.Scan() {
		var line jsonlLine
		if err := json.Unmarshal(scanner.Bytes(), &line); err != nil {
			continue
		}

		// Track timestamps for duration
		if ts, err := time.Parse(time.RFC3339Nano, line.Timestamp); err == nil {
			if firstTS.IsZero() {
				firstTS = ts
			}
			lastTS = ts
		}

		switch line.Type {
		case "session_meta":
			var meta sessionMeta
			if err := json.Unmarshal(line.Payload, &meta); err == nil {
				data.SessionID = meta.ID
			}

		case "event_msg":
			var evt eventMsg
			if err := json.Unmarshal(line.Payload, &evt); err != nil {
				continue
			}

			switch evt.Type {
			case "user_message":
				if data.Prompt == "" {
					data.Prompt = evt.Message
				}
				data.Transcript = append(data.Transcript, agent.Message{
					Role:      "user",
					Content:   evt.Message,
					Timestamp: lastTS,
				})

			case "agent_message":
				data.Transcript = append(data.Transcript, agent.Message{
					Role:      "assistant",
					Content:   evt.Message,
					Timestamp: lastTS,
				})

			case "token_count":
				// Token counts come as a separate event; extract if needed
				var tc struct {
					InputTokens  int `json:"input_tokens"`
					OutputTokens int `json:"output_tokens"`
				}
				if err := json.Unmarshal(line.Payload, &tc); err == nil {
					data.TotalTokens += tc.InputTokens + tc.OutputTokens
				}
			}

		case "response_item":
			var item responseItem
			if err := json.Unmarshal(line.Payload, &item); err != nil {
				continue
			}

			if item.Type == "message" {
				var text string
				for _, c := range item.Content {
					text += c.Text
				}
				if text != "" {
					data.Transcript = append(data.Transcript, agent.Message{
						Role:      "assistant",
						Content:   text,
						Timestamp: lastTS,
					})
				}
			}
		}
	}

	if !firstTS.IsZero() && !lastTS.IsZero() {
		data.Duration = lastTS.Sub(firstTS)
	}

	return data, scanner.Err()
}

// PeekSessionID reads just the session ID from a Codex JSONL file.
func PeekSessionID(path string) string {
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	if scanner.Scan() {
		var line jsonlLine
		if err := json.Unmarshal(scanner.Bytes(), &line); err == nil && line.Type == "session_meta" {
			var meta sessionMeta
			if err := json.Unmarshal(line.Payload, &meta); err == nil {
				return meta.ID
			}
		}
	}
	return ""
}

// PeekCWD reads just the cwd from a Codex JSONL file's session_meta.
func PeekCWD(path string) string {
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	if scanner.Scan() {
		var line jsonlLine
		if err := json.Unmarshal(scanner.Bytes(), &line); err == nil && line.Type == "session_meta" {
			var meta sessionMeta
			if err := json.Unmarshal(line.Payload, &meta); err == nil {
				return meta.CWD
			}
		}
	}
	return ""
}
