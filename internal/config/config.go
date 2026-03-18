package config

// Config holds all partio configuration.
type Config struct {
	Enabled         bool            `json:"enabled"`
	Strategy        string          `json:"strategy"`
	Agent           string          `json:"agent"`
	LogLevel        string          `json:"log_level"`
	StrategyOptions StrategyOptions `json:"strategy_options"`
	HookOptions     HookOptions     `json:"hook_options"`
}

// StrategyOptions holds strategy-specific options.
type StrategyOptions struct {
	PushSessions bool `json:"push_sessions"`
}

// HookOptions holds hook-specific options.
type HookOptions struct {
	// SessionRetryTimeoutMs is the maximum time in milliseconds to retry
	// reading session data in the post-commit hook. 0 disables retries.
	SessionRetryTimeoutMs int `json:"session_retry_timeout_ms"`
}

// PartioDir is the directory name for partio config within a repo.
const PartioDir = ".partio"
