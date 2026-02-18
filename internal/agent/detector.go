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
