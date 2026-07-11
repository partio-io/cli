package session

import (
	"log/slog"
	"os"
	"syscall"
	"time"
)

// CleanupResult holds the outcome of a stale-session cleanup run.
type CleanupResult struct {
	// Cleaned is true when a stale session was found and transitioned to ENDED.
	Cleaned bool
	// Session is the session that was cleaned up, or nil if none.
	Session *Session
}

// CleanupStale checks the current session and, if it is stale, transitions it
// to ENDED state. A session is considered stale when:
//   - Its state is ACTIVE or IDLE, and
//   - The session state file has not been updated within threshold, and
//   - If AgentPID is non-zero, the associated process is no longer alive.
//
// Returns a CleanupResult describing what happened.
func (m *Manager) CleanupStale(threshold time.Duration) (CleanupResult, error) {
	s, err := m.Current()
	if err != nil {
		return CleanupResult{}, err
	}
	if s == nil {
		return CleanupResult{}, nil
	}

	// Only clean ACTIVE or IDLE sessions — ENDED sessions are already done.
	if s.State != StateActive && s.State != StateIdle {
		return CleanupResult{}, nil
	}

	if !isStale(m.currentPath(), s, threshold) {
		return CleanupResult{}, nil
	}

	slog.Info("cleaning up stale session",
		"id", s.ID,
		"agent", s.Agent,
		"state", s.State,
		"pid", s.AgentPID,
	)

	s.State = StateEnded
	s.EndedAt = time.Now()
	if err := m.save(s); err != nil {
		return CleanupResult{}, err
	}

	return CleanupResult{Cleaned: true, Session: s}, nil
}

// isStale returns true when the session file is old enough AND (if PID is known)
// the agent process is no longer alive.
func isStale(path string, s *Session, threshold time.Duration) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}

	age := time.Since(info.ModTime())
	if age < threshold {
		// File was updated recently — session is not stale yet.
		return false
	}

	// File is old enough. If we have a PID, require the process to be gone too.
	if s.AgentPID > 0 {
		return !isProcessAlive(s.AgentPID)
	}

	// No PID recorded — rely on timestamp alone.
	return true
}

// isProcessAlive returns true if the process with the given PID is still alive.
// Uses signal 0, which checks process existence without sending an actual signal.
func isProcessAlive(pid int) bool {
	err := syscall.Kill(pid, 0)
	return err == nil
}
