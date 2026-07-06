package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestEnsureSettingsEnabled(t *testing.T) {
	tests := []struct {
		name            string
		existingContent string // empty means no file
		wantEnabled     bool
		wantStrategy    string // if non-empty, check this key is preserved
	}{
		{
			name:        "no existing file creates defaults",
			wantEnabled: true,
		},
		{
			name:            "existing file with enabled false sets to true",
			existingContent: `{"enabled": false, "strategy": "manual-commit", "agent": "claude-code"}`,
			wantEnabled:     true,
			wantStrategy:    "manual-commit",
		},
		{
			name:            "existing file with enabled true is preserved",
			existingContent: `{"enabled": true, "strategy": "custom-strategy", "agent": "claude-code"}`,
			wantEnabled:     true,
			wantStrategy:    "custom-strategy",
		},
		{
			name:            "corrupted JSON is overwritten with defaults",
			existingContent: `{not valid json`,
			wantEnabled:     true,
		},
		{
			name:            "missing enabled key sets it to true",
			existingContent: `{"strategy": "manual-commit"}`,
			wantEnabled:     true,
			wantStrategy:    "manual-commit",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			settingsPath := filepath.Join(dir, "settings.json")

			if tt.existingContent != "" {
				if err := os.WriteFile(settingsPath, []byte(tt.existingContent), 0o644); err != nil {
					t.Fatal(err)
				}
			}

			if err := ensureSettingsEnabled(settingsPath); err != nil {
				t.Fatalf("ensureSettingsEnabled() error = %v", err)
			}

			data, err := os.ReadFile(settingsPath)
			if err != nil {
				t.Fatalf("reading settings.json: %v", err)
			}

			var raw map[string]json.RawMessage
			if err := json.Unmarshal(data, &raw); err != nil {
				t.Fatalf("parsing settings.json: %v", err)
			}

			var enabled bool
			if v, ok := raw["enabled"]; !ok {
				t.Fatal("settings.json missing 'enabled' key")
			} else if err := json.Unmarshal(v, &enabled); err != nil {
				t.Fatalf("parsing 'enabled' value: %v", err)
			}

			if enabled != tt.wantEnabled {
				t.Errorf("enabled = %v, want %v", enabled, tt.wantEnabled)
			}

			if tt.wantStrategy != "" {
				var strategy string
				if v, ok := raw["strategy"]; !ok {
					t.Error("settings.json missing 'strategy' key")
				} else if err := json.Unmarshal(v, &strategy); err != nil {
					t.Errorf("parsing 'strategy' value: %v", err)
				} else if strategy != tt.wantStrategy {
					t.Errorf("strategy = %q, want %q", strategy, tt.wantStrategy)
				}
			}
		})
	}
}
