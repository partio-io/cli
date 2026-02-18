package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaults(t *testing.T) {
	d := Defaults()

	if d.Enabled != true {
		t.Errorf("expected enabled=true, got %v", d.Enabled)
	}
	if d.Strategy != "manual-commit" {
		t.Errorf("expected strategy=manual-commit, got %s", d.Strategy)
	}
	if d.Agent != "claude-code" {
		t.Errorf("expected agent=claude-code, got %s", d.Agent)
	}
	if d.LogLevel != "info" {
		t.Errorf("expected log_level=info, got %s", d.LogLevel)
	}
	if d.StrategyOptions.PushSessions != true {
		t.Errorf("expected push_sessions=true, got %v", d.StrategyOptions.PushSessions)
	}
}

func TestLoadDefaults(t *testing.T) {
	cfg, err := Load("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Strategy != "manual-commit" {
		t.Errorf("expected default strategy, got %s", cfg.Strategy)
	}
}

func TestMergeFromFile(t *testing.T) {
	dir := t.TempDir()
	settingsPath := filepath.Join(dir, "settings.json")

	err := os.WriteFile(settingsPath, []byte(`{"strategy": "auto-commit", "log_level": "debug"}`), 0o644)
	if err != nil {
		t.Fatalf("writing test settings: %v", err)
	}

	cfg := Defaults()
	mergeFromFile(&cfg, settingsPath)

	if cfg.Strategy != "auto-commit" {
		t.Errorf("expected strategy=auto-commit, got %s", cfg.Strategy)
	}
	if cfg.LogLevel != "debug" {
		t.Errorf("expected log_level=debug, got %s", cfg.LogLevel)
	}
	// Unset fields should retain defaults
	if cfg.Agent != "claude-code" {
		t.Errorf("expected agent=claude-code (default), got %s", cfg.Agent)
	}
}

func TestEnvOverrides(t *testing.T) {
	t.Setenv("PARTIO_STRATEGY", "env-strategy")
	t.Setenv("PARTIO_LOG_LEVEL", "error")
	t.Setenv("PARTIO_ENABLED", "false")

	cfg := Defaults()
	applyEnv(&cfg)

	if cfg.Strategy != "env-strategy" {
		t.Errorf("expected strategy=env-strategy, got %s", cfg.Strategy)
	}
	if cfg.LogLevel != "error" {
		t.Errorf("expected log_level=error, got %s", cfg.LogLevel)
	}
	if cfg.Enabled != false {
		t.Errorf("expected enabled=false, got %v", cfg.Enabled)
	}
}

func TestLoadLayered(t *testing.T) {
	dir := t.TempDir()
	partioDir := filepath.Join(dir, ".partio")
	if err := os.MkdirAll(partioDir, 0o755); err != nil {
		t.Fatalf("creating partio dir: %v", err)
	}

	// Write repo-level config
	if err := os.WriteFile(filepath.Join(partioDir, "settings.json"),
		[]byte(`{"strategy": "repo-strategy"}`), 0o644); err != nil {
		t.Fatalf("writing settings.json: %v", err)
	}

	// Write local override
	if err := os.WriteFile(filepath.Join(partioDir, "settings.local.json"),
		[]byte(`{"log_level": "debug"}`), 0o644); err != nil {
		t.Fatalf("writing settings.local.json: %v", err)
	}

	cfg, err := Load(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Strategy != "repo-strategy" {
		t.Errorf("expected strategy=repo-strategy, got %s", cfg.Strategy)
	}
	if cfg.LogLevel != "debug" {
		t.Errorf("expected log_level=debug, got %s", cfg.LogLevel)
	}
}
