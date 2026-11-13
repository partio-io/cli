package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestSaveRepoSetting_NewFile(t *testing.T) {
	dir := t.TempDir()

	if err := SaveRepoSetting(dir, "commit_linking", CommitLinkingAlways); err != nil {
		t.Fatalf("SaveRepoSetting: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, PartioDir, "settings.json"))
	if err != nil {
		t.Fatalf("reading settings.json: %v", err)
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("unmarshaling: %v", err)
	}

	var val string
	if err := json.Unmarshal(raw["commit_linking"], &val); err != nil {
		t.Fatalf("unmarshaling commit_linking: %v", err)
	}
	if val != CommitLinkingAlways {
		t.Errorf("expected %s, got %s", CommitLinkingAlways, val)
	}
}

func TestSaveRepoSetting_MalformedExisting(t *testing.T) {
	dir := t.TempDir()
	partioDir := filepath.Join(dir, PartioDir)
	if err := os.MkdirAll(partioDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	malformed := `{"enabled": true,` // truncated JSON
	settingsPath := filepath.Join(partioDir, "settings.json")
	if err := os.WriteFile(settingsPath, []byte(malformed), 0o644); err != nil {
		t.Fatalf("writing malformed settings: %v", err)
	}

	if err := SaveRepoSetting(dir, "commit_linking", CommitLinkingAlways); err == nil {
		t.Fatal("expected error for malformed settings.json, got nil")
	}

	data, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("reading settings.json: %v", err)
	}
	if string(data) != malformed {
		t.Errorf("malformed settings.json was rewritten: %q", string(data))
	}
}

func TestSaveRepoSetting_PreservesExisting(t *testing.T) {
	dir := t.TempDir()
	partioDir := filepath.Join(dir, PartioDir)
	if err := os.MkdirAll(partioDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	// Write initial settings
	initial := `{"enabled": true, "strategy": "manual-commit"}`
	if err := os.WriteFile(filepath.Join(partioDir, "settings.json"), []byte(initial), 0o644); err != nil {
		t.Fatalf("writing initial settings: %v", err)
	}

	if err := SaveRepoSetting(dir, "commit_linking", CommitLinkingAlways); err != nil {
		t.Fatalf("SaveRepoSetting: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(partioDir, "settings.json"))
	if err != nil {
		t.Fatalf("reading settings.json: %v", err)
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("unmarshaling: %v", err)
	}

	// Check commit_linking was added
	var linking string
	if err := json.Unmarshal(raw["commit_linking"], &linking); err != nil {
		t.Fatalf("unmarshaling commit_linking: %v", err)
	}
	if linking != CommitLinkingAlways {
		t.Errorf("expected commit_linking=%s, got %s", CommitLinkingAlways, linking)
	}

	// Check existing settings preserved
	var enabled bool
	if err := json.Unmarshal(raw["enabled"], &enabled); err != nil {
		t.Fatalf("unmarshaling enabled: %v", err)
	}
	if !enabled {
		t.Error("expected enabled to be preserved as true")
	}
}

func TestSaveGlobalStrategyOptions(t *testing.T) {
	stealth := map[string]any{"push_sessions": false, "disable_trailers": true}

	tests := []struct {
		name     string
		existing string // empty means no file and no directory
		wantTop  map[string]string
		wantOpts map[string]string
	}{
		{
			name:     "creates directory and file",
			wantOpts: map[string]string{"push_sessions": "false", "disable_trailers": "true"},
		},
		{
			name:     "keeps unrelated keys",
			existing: `{"agent": "codex", "strategy_options": {"push_sessions": true, "other": "x"}}`,
			wantTop:  map[string]string{"agent": `"codex"`},
			wantOpts: map[string]string{"push_sessions": "false", "disable_trailers": "true", "other": `"x"`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			home := t.TempDir()
			t.Setenv("HOME", home)
			path := filepath.Join(home, ".config", "partio", "settings.json")

			if tt.existing != "" {
				if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(path, []byte(tt.existing), 0o644); err != nil {
					t.Fatal(err)
				}
			}

			if _, err := SaveGlobalStrategyOptions(stealth); err != nil {
				t.Fatalf("SaveGlobalStrategyOptions: %v", err)
			}
			first, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("reading global settings: %v", err)
			}

			var raw map[string]json.RawMessage
			if err := json.Unmarshal(first, &raw); err != nil {
				t.Fatalf("unmarshaling: %v", err)
			}
			var opts map[string]json.RawMessage
			if err := json.Unmarshal(raw["strategy_options"], &opts); err != nil {
				t.Fatalf("unmarshaling strategy_options: %v", err)
			}
			for key, want := range tt.wantTop {
				if got := string(raw[key]); got != want {
					t.Errorf("%s = %s, want %s", key, got, want)
				}
			}
			for key, want := range tt.wantOpts {
				if got := string(opts[key]); got != want {
					t.Errorf("strategy_options.%s = %s, want %s", key, got, want)
				}
			}

			// Idempotent: a second write leaves the file unchanged.
			if _, err := SaveGlobalStrategyOptions(stealth); err != nil {
				t.Fatalf("second SaveGlobalStrategyOptions: %v", err)
			}
			second, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if string(first) != string(second) {
				t.Errorf("second write changed the file:\nfirst:  %s\nsecond: %s", first, second)
			}

			// Load reads what the writer wrote.
			cfg, err := Load("")
			if err != nil {
				t.Fatalf("Load: %v", err)
			}
			if cfg.StrategyOptions.PushSessions || !cfg.StrategyOptions.DisableTrailers {
				t.Errorf("Load strategy_options = %+v, want push_sessions=false, disable_trailers=true", cfg.StrategyOptions)
			}
		})
	}
}

func TestSaveGlobalStrategyOptions_Malformed(t *testing.T) {
	tests := []struct {
		name     string
		existing string
	}{
		{name: "invalid JSON", existing: `{"strategy_options": {`},
		{name: "strategy_options not an object", existing: `{"strategy_options": true}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			home := t.TempDir()
			t.Setenv("HOME", home)
			path := filepath.Join(home, ".config", "partio", "settings.json")
			if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(path, []byte(tt.existing), 0o644); err != nil {
				t.Fatal(err)
			}

			if _, err := SaveGlobalStrategyOptions(map[string]any{"push_sessions": false}); err == nil {
				t.Fatal("expected error for malformed global settings, got nil")
			}

			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if string(data) != tt.existing {
				t.Errorf("malformed global settings were rewritten: %q", data)
			}
		})
	}
}
