package session

import (
	"fmt"
	"os"
)

// RecordActive persists the current agent session in ACTIVE state, refreshing
// the state file (and therefore its modification time) and recording the agent
// PID for later liveness checks. A non-ended session is refreshed in place
// (keeping its ID and start time); otherwise a new session is created.
func (m *Manager) RecordActive(agentName, branch, sourceDir string, pid int) error {
	if err := os.MkdirAll(m.stateDir, 0o755); err != nil {
		return fmt.Errorf("creating session directory: %w", err)
	}

	s, err := m.Current()
	if err != nil || s == nil || s.State == StateEnded {
		s = New(agentName, branch, sourceDir)
	} else {
		s.Agent = agentName
		s.Branch = branch
		s.SourceDir = sourceDir
		s.State = StateActive
	}
	s.AgentPID = pid
	return m.save(s)
}
