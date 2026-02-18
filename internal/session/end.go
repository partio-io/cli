package session

import "time"

// End marks the current session as ended.
func (m *Manager) End() error {
	s, err := m.Current()
	if err != nil {
		return err
	}
	if s == nil {
		return nil
	}

	s.State = StateEnded
	s.EndedAt = time.Now()
	return m.save(s)
}
