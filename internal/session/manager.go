package session

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Manager handles session lifecycle transitions.
type Manager struct {
	stateDir string
}

// NewManager creates a session manager that persists state to the given directory.
func NewManager(partioDir string) *Manager {
	return &Manager{
		stateDir: filepath.Join(partioDir, "sessions"),
	}
}

func (m *Manager) save(s *Session) error {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling session: %w", err)
	}
	return os.WriteFile(m.currentPath(), data, 0o644)
}

func (m *Manager) currentPath() string {
	return filepath.Join(m.stateDir, "current.json")
}

// MarkCondensed marks the current session as fully condensed with the given
// agent session ID. When agentRunning is false, the session state is also
// transitioned to StateEnded so future hook runs can skip it. When
// agentRunning is true, StateEnded is not set — the agent may still make
// additional commits in this session that should receive checkpoints.
func (m *Manager) MarkCondensed(capturedSessionID string, agentRunning bool) error {
	s, err := m.Current()
	if err != nil {
		return err
	}
	if s == nil {
		// No tracked session yet — create a minimal one to persist condensed state.
		if err := os.MkdirAll(m.stateDir, 0o755); err != nil {
			return fmt.Errorf("creating session directory: %w", err)
		}
		s = &Session{}
	}
	if !agentRunning {
		s.State = StateEnded
	}
	s.Condensed = true
	s.CapturedSessionID = capturedSessionID
	s.CapturedAt = time.Now()
	return m.save(s)
}
