package config

import (
	"os"
	"strings"
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
}
