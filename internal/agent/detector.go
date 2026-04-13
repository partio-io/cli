package agent

// Detector detects whether an AI agent is currently active.
type Detector interface {
	// Name returns the agent name (e.g. "claude-code").
	Name() string

	// IsRunning returns true if the agent process is currently active.
	IsRunning() (bool, error)

	// FindSessionDir returns the path to the agent's session data for the given repo.
	FindSessionDir(repoRoot string) (string, error)
}

// SessionParser is implemented by detectors that can extract session data.
// The hooks use this to capture session transcripts from any agent.
type SessionParser interface {
	// FindLatestSession returns the path to the most recent session file
	// and the parsed session data for the given repo.
	FindLatestSession(repoRoot string) (path string, data *SessionData, err error)
}
