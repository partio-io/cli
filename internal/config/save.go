package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// SaveRepoSetting updates a single key in the repo-level settings.json,
// preserving all other settings. The file is created if it does not exist.
func SaveRepoSetting(repoRoot, key string, value any) error {
	settingsPath := filepath.Join(repoRoot, PartioDir, "settings.json")

	return updateSettingsFile(settingsPath, func(raw map[string]json.RawMessage) error {
		encoded, err := json.Marshal(value)
		if err != nil {
			return err
		}
		raw[key] = json.RawMessage(encoded)
		return nil
	})
}

// SaveGlobalStrategyOptions merges the given strategy_options keys into the
// global settings file (see GlobalSettingsPath). Every other top-level key and
// every other strategy_options key already in the file is kept. The file and
// its directory are created if they do not exist.
func SaveGlobalStrategyOptions(opts map[string]any) error {
	settingsPath, err := GlobalSettingsPath()
	if err != nil {
		return fmt.Errorf("resolving global settings path: %w", err)
	}

	return updateSettingsFile(settingsPath, func(raw map[string]json.RawMessage) error {
		strategyOptions := make(map[string]json.RawMessage)
		if existing, ok := raw["strategy_options"]; ok {
			if err := json.Unmarshal(existing, &strategyOptions); err != nil {
				return fmt.Errorf("strategy_options in %s is not a JSON object, not overwriting: %w", settingsPath, err)
			}
		}
		for key, value := range opts {
			encoded, err := json.Marshal(value)
			if err != nil {
				return err
			}
			strategyOptions[key] = json.RawMessage(encoded)
		}
		encoded, err := json.Marshal(strategyOptions)
		if err != nil {
			return err
		}
		raw["strategy_options"] = json.RawMessage(encoded)
		return nil
	})
}

// updateSettingsFile reads the settings file at path, applies update to its
// top-level keys and writes the result back with mode 0o644, creating the
// parent directory if needed. A missing file starts from an empty object. A
// malformed file is an error and is left unchanged: continuing with an empty
// map would rewrite it with only the new keys, dropping every setting the user
// already had.
func updateSettingsFile(path string, update func(raw map[string]json.RawMessage) error) error {
	raw := make(map[string]json.RawMessage)

	existing, err := os.ReadFile(path)
	if err == nil {
		if err := json.Unmarshal(existing, &raw); err != nil {
			return fmt.Errorf("existing %s is not valid JSON, not overwriting: %w", path, err)
		}
		if raw == nil {
			// The file holds JSON null; treat it as an empty object.
			raw = make(map[string]json.RawMessage)
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("reading %s: %w", path, err)
	}

	if err := update(raw); err != nil {
		return err
	}

	data, err := json.MarshalIndent(raw, "", "  ")
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	return os.WriteFile(path, data, 0o644)
}
