package config

import (
	"os"
	"strings"
	"time"
)

func applyEnv(cfg *Config) {
	if v := os.Getenv("PARTIO_ENABLED"); v != "" {
		cfg.Enabled = strings.EqualFold(v, "true") || v == "1"
	}
	if v := os.Getenv("PARTIO_STRATEGY"); v != "" {
		cfg.Strategy = v
	}
	if v := os.Getenv("PARTIO_LOG_LEVEL"); v != "" {
		cfg.LogLevel = v
	}
	if v := os.Getenv("PARTIO_AGENT"); v != "" {
		cfg.Agent = v
	}
	if v := os.Getenv("PARTIO_COMMIT_LINKING"); v != "" {
		cfg.CommitLinking = v
	}
	if v := os.Getenv("PARTIO_STALE_SESSION_THRESHOLD"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.StaleSessionThreshold = Duration(d)
		}
	}
}
