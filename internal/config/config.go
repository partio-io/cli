package config

// Config holds all partio configuration.
type Config struct {
	Enabled         bool            `json:"enabled"`
	Strategy        string          `json:"strategy"`
	Agent           string          `json:"agent"`
	LogLevel        string          `json:"log_level"`
	CommitLinking   string          `json:"commit_linking"`
	StrategyOptions StrategyOptions `json:"strategy_options"`
}

// CommitLinking values.
const (
	CommitLinkingAsk    = "ask"
	CommitLinkingAlways = "always"
	CommitLinkingNever  = "never"
)

// StrategyOptions holds strategy-specific options.
type StrategyOptions struct {
	PushSessions bool `json:"push_sessions"`
}

// PartioDir is the directory name for partio config within a repo.
const PartioDir = ".partio"
