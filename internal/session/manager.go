package session

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
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
