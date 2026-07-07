package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// SaveRepoSetting updates a single key in the repo-level settings.json,
// preserving all other settings. The file is created if it does not exist.
func SaveRepoSetting(repoRoot, key string, value any) error {
	settingsPath := filepath.Join(repoRoot, PartioDir, "settings.json")

	raw := make(map[string]json.RawMessage)

	existing, err := os.ReadFile(settingsPath)
	if err == nil {
		_ = json.Unmarshal(existing, &raw)
	}

	encoded, err := json.Marshal(value)
	if err != nil {
		return err
	}
	raw[key] = json.RawMessage(encoded)

	data, err := json.MarshalIndent(raw, "", "  ")
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(settingsPath), 0o755); err != nil {
		return err
	}

	return os.WriteFile(settingsPath, data, 0o644)
}
