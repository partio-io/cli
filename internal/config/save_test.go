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
