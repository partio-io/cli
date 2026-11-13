package config

import (
	"os"
	"path/filepath"
)

// GlobalSettingsPath returns the path of the global settings file,
// ~/.config/partio/settings.json. Load reads it and SaveGlobalStrategyOptions
// writes it, so the read path and the write path cannot diverge.
func GlobalSettingsPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "partio", "settings.json"), nil
}
