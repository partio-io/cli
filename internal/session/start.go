package session

import (
	"fmt"
	"os"
)

// Start begins a new session, ending any existing one.
func (m *Manager) Start(agent, branch, sourceDir string) (*Session, error) {
	if err := os.MkdirAll(m.stateDir, 0o755); err != nil {
		return nil, fmt.Errorf("creating session directory: %w", err)
	}

	s := New(agent, branch, sourceDir)
	return s, m.save(s)
}
