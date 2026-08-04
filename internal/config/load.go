package config

import (
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
)

// Load reads config from all layers and merges them (lowest to highest priority):
// 1. Defaults
// 2. ~/.config/partio/settings.json
// 3. .partio/settings.json
// 4. .partio/settings.local.json
// 5. Environment variables
func Load(repoRoot string) (Config, error) {
	cfg := Defaults()

	// Global config
	home, err := os.UserHomeDir()
	if err == nil {
		mergeFromFile(&cfg, filepath.Join(home, ".config", "partio", "settings.json"))
	}

	// Repo config
	if repoRoot != "" {
		mergeFromFile(&cfg, filepath.Join(repoRoot, PartioDir, "settings.json"))
		mergeFromFile(&cfg, filepath.Join(repoRoot, PartioDir, "settings.local.json"))
	}

	// Env var overrides
	applyEnv(&cfg)

	return cfg, nil
}

func mergeFromFile(dst *Config, path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}

	// Use a map to detect which keys are present.
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return
	}

	mergeKey(raw, "enabled", path, &dst.Enabled)
	mergeKey(raw, "strategy", path, &dst.Strategy)
	mergeKey(raw, "agent", path, &dst.Agent)
	mergeKey(raw, "log_level", path, &dst.LogLevel)
	mergeKey(raw, "commit_linking", path, &dst.CommitLinking)
	mergeKey(raw, "strategy_options", path, &dst.StrategyOptions)
	mergeKey(raw, "stale_session_threshold", path, &dst.StaleSessionThreshold)
}

// mergeKey unmarshals raw[key] into dst if present. A malformed value keeps
// the previous config value; it is logged rather than silently dropped so a
// settings typo is visible instead of reverting to defaults without a trace.
func mergeKey(raw map[string]json.RawMessage, key, path string, dst any) {
	v, ok := raw[key]
	if !ok {
		return
	}
	if err := json.Unmarshal(v, dst); err != nil {
		slog.Warn("ignoring malformed config key", "path", path, "key", key, "error", err)
	}
}
