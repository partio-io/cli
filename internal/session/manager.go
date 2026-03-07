package session

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
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
	if err := os.WriteFile(m.idPath(s.ID), data, 0o644); err != nil {
		return fmt.Errorf("saving session by id: %w", err)
	}
	return os.WriteFile(m.currentPath(), data, 0o644)
}

func (m *Manager) currentPath() string {
	return filepath.Join(m.stateDir, "current.json")
}

func (m *Manager) idPath(id string) string {
	return filepath.Join(m.stateDir, id+".json")
}

// List returns all sessions sorted by most recent activity, newest first.
func (m *Manager) List() ([]*Session, error) {
	entries, err := os.ReadDir(m.stateDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading sessions dir: %w", err)
	}

	var sessions []*Session
	for _, e := range entries {
		if e.IsDir() || e.Name() == "current.json" || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(m.stateDir, e.Name()))
		if err != nil {
			continue
		}
		var s Session
		if err := json.Unmarshal(data, &s); err != nil {
			continue
		}
		sessions = append(sessions, &s)
	}

	sort.Slice(sessions, func(i, j int) bool {
		ti := latestTime(sessions[i])
		tj := latestTime(sessions[j])
		return ti.After(tj)
	})

	return sessions, nil
}

// RecordActivity updates a session's last activity time and cumulative files modified count.
func (m *Manager) RecordActivity(filesChanged int) error {
	s, err := m.Current()
	if err != nil || s == nil {
		return err
	}
	s.UpdatedAt = time.Now()
	s.FilesModified += filesChanged
	return m.save(s)
}

func latestTime(s *Session) time.Time {
	if !s.EndedAt.IsZero() {
		return s.EndedAt
	}
	if !s.UpdatedAt.IsZero() {
		return s.UpdatedAt
	}
	return s.StartedAt
}
