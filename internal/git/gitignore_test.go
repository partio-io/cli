package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestFilterGitIgnored(t *testing.T) {
	dir := t.TempDir()

	if out, err := exec.Command("git", "-C", dir, "init").CombinedOutput(); err != nil {
		t.Fatalf("git init: %v\n%s", err, out)
	}

	if err := os.WriteFile(filepath.Join(dir, ".gitignore"), []byte("*.log\n"), 0o644); err != nil {
		t.Fatalf("writing .gitignore: %v", err)
	}

	paths := []string{"main.go", "debug.log", "server.go"}
	got, err := FilterGitIgnored(dir, paths)
	if err != nil {
		t.Fatalf("FilterGitIgnored: %v", err)
	}

	// debug.log matches *.log and must be absent from checkpoint metadata.
	for _, p := range got {
		if p == "debug.log" {
			t.Errorf("gitignored path %q should be absent from checkpoint metadata", p)
		}
	}

	found := make(map[string]bool)
	for _, p := range got {
		found[p] = true
	}
	if !found["main.go"] {
		t.Error("main.go should be present in checkpoint metadata")
	}
	if !found["server.go"] {
		t.Error("server.go should be present in checkpoint metadata")
	}
}

func TestFilterGitIgnoredNoneIgnored(t *testing.T) {
	dir := t.TempDir()

	if out, err := exec.Command("git", "-C", dir, "init").CombinedOutput(); err != nil {
		t.Fatalf("git init: %v\n%s", err, out)
	}

	if err := os.WriteFile(filepath.Join(dir, ".gitignore"), []byte("*.log\n"), 0o644); err != nil {
		t.Fatalf("writing .gitignore: %v", err)
	}

	paths := []string{"main.go", "server.go"}
	got, err := FilterGitIgnored(dir, paths)
	if err != nil {
		t.Fatalf("FilterGitIgnored: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("expected all 2 paths returned, got %v", got)
	}
}

func TestFilterGitIgnoredEmpty(t *testing.T) {
	dir := t.TempDir()

	if out, err := exec.Command("git", "-C", dir, "init").CombinedOutput(); err != nil {
		t.Fatalf("git init: %v\n%s", err, out)
	}

	got, err := FilterGitIgnored(dir, nil)
	if err != nil {
		t.Fatalf("FilterGitIgnored with nil paths: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty result, got %v", got)
	}
}

func TestFilterGitIgnoredNestedGitignore(t *testing.T) {
	dir := t.TempDir()

	if out, err := exec.Command("git", "-C", dir, "init").CombinedOutput(); err != nil {
		t.Fatalf("git init: %v\n%s", err, out)
	}

	// Root .gitignore
	if err := os.WriteFile(filepath.Join(dir, ".gitignore"), []byte("*.log\n"), 0o644); err != nil {
		t.Fatalf("writing root .gitignore: %v", err)
	}

	// Nested .gitignore in subdir/
	if err := os.MkdirAll(filepath.Join(dir, "subdir"), 0o755); err != nil {
		t.Fatalf("mkdir subdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "subdir", ".gitignore"), []byte("*.tmp\n"), 0o644); err != nil {
		t.Fatalf("writing nested .gitignore: %v", err)
	}

	paths := []string{"main.go", "debug.log", "subdir/cache.tmp", "subdir/code.go"}
	got, err := FilterGitIgnored(dir, paths)
	if err != nil {
		t.Fatalf("FilterGitIgnored: %v", err)
	}

	for _, p := range got {
		if p == "debug.log" || p == "subdir/cache.tmp" {
			t.Errorf("gitignored path %q should be absent from checkpoint metadata", p)
		}
	}

	found := make(map[string]bool)
	for _, p := range got {
		found[p] = true
	}
	if !found["main.go"] {
		t.Error("main.go should be present in checkpoint metadata")
	}
	if !found["subdir/code.go"] {
		t.Error("subdir/code.go should be present in checkpoint metadata")
	}
}
