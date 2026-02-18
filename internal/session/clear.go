package session

import "os"

// Clear removes the current session state file.
func (m *Manager) Clear() error {
	return os.Remove(m.currentPath())
}
