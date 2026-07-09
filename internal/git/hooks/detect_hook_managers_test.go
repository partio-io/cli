package hooks

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectHookManagers_none(t *testing.T) {
	dir := t.TempDir()
	managers := DetectHookManagers(dir)
	if len(managers) != 0 {
		t.Errorf("expected no managers, got %v", managers)
	}
}

func TestDetectHookManagers_husky_dir(t *testing.T) {
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, ".husky"), 0o755); err != nil {
		t.Fatal(err)
	}
	managers := DetectHookManagers(dir)
	if len(managers) != 1 || managers[0].Name != "Husky" {
		t.Errorf("expected Husky, got %v", managers)
	}
}

func TestDetectHookManagers_husky_package_json_devdeps(t *testing.T) {
	dir := t.TempDir()
	pkg := `{"devDependencies": {"husky": "^8.0.0"}}`
	if err := os.WriteFile(filepath.Join(dir, "package.json"), []byte(pkg), 0o644); err != nil {
		t.Fatal(err)
	}
	managers := DetectHookManagers(dir)
	if len(managers) != 1 || managers[0].Name != "Husky" {
		t.Errorf("expected Husky, got %v", managers)
	}
}

func TestDetectHookManagers_husky_package_json_config_key(t *testing.T) {
	dir := t.TempDir()
	pkg := `{"husky": {"hooks": {"pre-commit": "lint-staged"}}}`
	if err := os.WriteFile(filepath.Join(dir, "package.json"), []byte(pkg), 0o644); err != nil {
		t.Fatal(err)
	}
	managers := DetectHookManagers(dir)
	if len(managers) != 1 || managers[0].Name != "Husky" {
		t.Errorf("expected Husky, got %v", managers)
	}
}

func TestDetectHookManagers_lefthook_yml(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "lefthook.yml"), []byte("pre-commit:\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	managers := DetectHookManagers(dir)
	if len(managers) != 1 || managers[0].Name != "Lefthook" {
		t.Errorf("expected Lefthook, got %v", managers)
	}
}

func TestDetectHookManagers_lefthook_yaml(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "lefthook.yaml"), []byte("pre-commit:\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	managers := DetectHookManagers(dir)
	if len(managers) != 1 || managers[0].Name != "Lefthook" {
		t.Errorf("expected Lefthook, got %v", managers)
	}
}

func TestDetectHookManagers_overcommit(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, ".overcommit.yml"), []byte("PreCommit:\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	managers := DetectHookManagers(dir)
	if len(managers) != 1 || managers[0].Name != "Overcommit" {
		t.Errorf("expected Overcommit, got %v", managers)
	}
}

func TestDetectHookManagers_multiple(t *testing.T) {
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, ".husky"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "lefthook.yml"), []byte(""), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".overcommit.yml"), []byte(""), 0o644); err != nil {
		t.Fatal(err)
	}
	managers := DetectHookManagers(dir)
	if len(managers) != 3 {
		t.Errorf("expected 3 managers, got %d: %v", len(managers), managers)
	}
}

func TestDetectHookManagers_instructions_not_empty(t *testing.T) {
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, ".husky"), 0o755); err != nil {
		t.Fatal(err)
	}
	managers := DetectHookManagers(dir)
	if len(managers) == 0 {
		t.Fatal("expected at least one manager")
	}
	if managers[0].Instructions == "" {
		t.Error("expected non-empty instructions for Husky")
	}
}
