package config

import "time"

// Defaults returns a Config with default values.
func Defaults() Config {
	return Config{
		Enabled:       true,
		Strategy:      "manual-commit",
		Agent:         "",
		LogLevel:      "info",
		CommitLinking: CommitLinkingAsk,
		StrategyOptions: StrategyOptions{
			PushSessions: true,
		},
		Redact: RedactOptions{
			Enabled:          true,
			EntropyThreshold: 4.5,
			EntropyMinLength: 20,
		},
		StaleSessionThreshold: Duration(10 * time.Minute),
	}
}
