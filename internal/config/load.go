package config

import (
	"encoding/json"
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

	if v, ok := raw["enabled"]; ok {
		_ = json.Unmarshal(v, &dst.Enabled)
	}
	if v, ok := raw["strategy"]; ok {
		_ = json.Unmarshal(v, &dst.Strategy)
	}
	if v, ok := raw["agent"]; ok {
		_ = json.Unmarshal(v, &dst.Agent)
	}
	if v, ok := raw["log_level"]; ok {
		_ = json.Unmarshal(v, &dst.LogLevel)
	}
	if v, ok := raw["strategy_options"]; ok {
		_ = json.Unmarshal(v, &dst.StrategyOptions)
	}
}
