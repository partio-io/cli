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
	if d.Agent != "" {
		t.Errorf("expected agent='' (auto-detect), got %s", d.Agent)
	}
	if d.LogLevel != "info" {
		t.Errorf("expected log_level=info, got %s", d.LogLevel)
	}
	if d.CommitLinking != CommitLinkingAsk {
		t.Errorf("expected commit_linking=ask, got %s", d.CommitLinking)
	}
	if d.StrategyOptions.PushSessions != true {
		t.Errorf("expected push_sessions=true, got %v", d.StrategyOptions.PushSessions)
	}
	if d.StrategyOptions.DisableTrailers != false {
		t.Errorf("expected disable_trailers=false, got %v", d.StrategyOptions.DisableTrailers)
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

	err := os.WriteFile(settingsPath, []byte(`{"strategy": "auto-commit", "log_level": "debug", "commit_linking": "always"}`), 0o644)
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
	if cfg.CommitLinking != CommitLinkingAlways {
		t.Errorf("expected commit_linking=always, got %s", cfg.CommitLinking)
	}
	// Unset fields should retain defaults
	if cfg.Agent != "" {
		t.Errorf("expected agent='' (auto-detect default), got %s", cfg.Agent)
	}
}

func TestEnvOverrides(t *testing.T) {
	t.Setenv("PARTIO_STRATEGY", "env-strategy")
	t.Setenv("PARTIO_LOG_LEVEL", "error")
	t.Setenv("PARTIO_ENABLED", "false")
	t.Setenv("PARTIO_COMMIT_LINKING", "never")

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
	if cfg.CommitLinking != CommitLinkingNever {
		t.Errorf("expected commit_linking=never, got %s", cfg.CommitLinking)
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

// TestLoadMergesStrategyOptionsPerField proves that a repo layer setting only
// push_sessions keeps disable_trailers from the global layer.
func TestLoadMergesStrategyOptionsPerField(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	globalDir := filepath.Join(home, ".config", "partio")
	if err := os.MkdirAll(globalDir, 0o755); err != nil {
		t.Fatalf("creating global dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(globalDir, "settings.json"),
		[]byte(`{"strategy_options": {"disable_trailers": true}}`), 0o644); err != nil {
		t.Fatalf("writing global settings: %v", err)
	}

	repo := t.TempDir()
	partioDir := filepath.Join(repo, PartioDir)
	if err := os.MkdirAll(partioDir, 0o755); err != nil {
		t.Fatalf("creating partio dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(partioDir, "settings.json"),
		[]byte(`{"strategy_options": {"push_sessions": false}}`), 0o644); err != nil {
		t.Fatalf("writing repo settings: %v", err)
	}

	cfg, err := Load(repo)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.StrategyOptions.PushSessions {
		t.Errorf("expected push_sessions=false from repo layer, got true")
	}
	if !cfg.StrategyOptions.DisableTrailers {
		t.Errorf("expected disable_trailers=true from global layer, got false")
	}
}
