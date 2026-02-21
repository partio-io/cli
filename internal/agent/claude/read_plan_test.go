package claude

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadPlanFileExisting(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	plansDir := filepath.Join(home, ".claude", "plans")
	if err := os.MkdirAll(plansDir, 0o755); err != nil {
		t.Fatalf("creating plans dir: %v", err)
	}

	content := "# My Plan\n\nBuild a todo app with React."
	if err := os.WriteFile(filepath.Join(plansDir, "noble-mixing-unicorn.md"), []byte(content), 0o644); err != nil {
		t.Fatalf("writing plan file: %v", err)
	}

	got, err := ReadPlanFile("noble-mixing-unicorn")
	if err != nil {
		t.Fatalf("ReadPlanFile error: %v", err)
	}

	if got != content {
		t.Errorf("expected %q, got %q", content, got)
	}
}

func TestReadPlanFileMissing(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	got, err := ReadPlanFile("nonexistent-slug")
	if err != nil {
		t.Fatalf("ReadPlanFile error: %v", err)
	}

	if got != "" {
		t.Errorf("expected empty string for missing plan, got %q", got)
	}
}

func TestReadPlanFileEmptySlug(t *testing.T) {
	got, err := ReadPlanFile("")
	if err != nil {
		t.Fatalf("ReadPlanFile error: %v", err)
	}

	if got != "" {
		t.Errorf("expected empty string for empty slug, got %q", got)
	}
}
