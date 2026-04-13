package config

// Defaults returns a Config with default values.
func Defaults() Config {
	return Config{
		Enabled:  true,
		Strategy: "manual-commit",
		Agent:    "",
		LogLevel: "info",
		StrategyOptions: StrategyOptions{
			PushSessions: true,
		},
	}
}
