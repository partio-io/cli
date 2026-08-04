package claude

import (
	"bufio"
	"encoding/json"
	"log/slog"
	"os"
)

// PeekSessionID reads just enough of a JSONL file to extract the session ID
// without parsing the full transcript. Returns "" if the ID cannot be determined.
func PeekSessionID(path string) string {
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer func() {
		if closeErr := f.Close(); closeErr != nil {
			slog.Debug("closing session file", "path", path, "error", closeErr)
		}
	}()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		var entry struct {
			SessionID string `json:"sessionId"`
		}
		if json.Unmarshal(scanner.Bytes(), &entry) == nil && entry.SessionID != "" {
			return entry.SessionID
		}
	}
	return ""
}
