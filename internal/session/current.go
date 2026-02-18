package session

import (
	"encoding/json"
	"fmt"
	"os"
)

// Current returns the active session, or nil if none.
func (m *Manager) Current() (*Session, error) {
	path := m.currentPath()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading session state: %w", err)
	}

	var s Session
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("parsing session state: %w", err)
	}
	return &s, nil
}
