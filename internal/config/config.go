package config

import "time"

// Config holds all partio configuration.
type Config struct {
	Enabled                bool            `json:"enabled"`
	Strategy               string          `json:"strategy"`
	Agent                  string          `json:"agent"`
	LogLevel               string          `json:"log_level"`
	CommitLinking          string          `json:"commit_linking"`
	StrategyOptions        StrategyOptions `json:"strategy_options"`
	Redact                 RedactOptions   `json:"redact"`
	StaleSessionThreshold  time.Duration   `json:"stale_session_threshold"`
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

// RedactOptions controls secret redaction in checkpoint data.
type RedactOptions struct {
	// Enabled toggles redaction. Defaults to true.
	Enabled bool `json:"enabled"`
	// EntropyThreshold is the minimum Shannon entropy (bits/char) at which a
	// whitespace-delimited token is considered a potential secret and redacted.
	// Defaults to 4.5.
	EntropyThreshold float64 `json:"entropy_threshold"`
	// EntropyMinLength is the minimum token length considered for entropy
	// scanning. Defaults to 20.
	EntropyMinLength int `json:"entropy_min_length"`
}

// PartioDir is the directory name for partio config within a repo.
const PartioDir = ".partio"
