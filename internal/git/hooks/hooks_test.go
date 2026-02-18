package hooks

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func initGitRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	cmd := exec.Command("git", "init", dir)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git init: %v\n%s", err, out)
	}
	return dir
}

func TestInstallAndUninstall(t *testing.T) {
	dir := initGitRepo(t)
	hooksDir := filepath.Join(dir, ".git", "hooks")

	// Install hooks
	if err := Install(dir); err != nil {
		t.Fatalf("Install error: %v", err)
	}

	// Verify all hooks exist
	for _, name := range hookNames {
		path := filepath.Join(hooksDir, name)
		data, err := os.ReadFile(path)
		if err != nil {
			t.Errorf("hook %s not found: %v", name, err)
			continue
		}

		content := string(data)
		if !isPartioHook(content) {
			t.Errorf("hook %s missing partio marker", name)
		}

		// Check executable permission
		info, _ := os.Stat(path)
		if info.Mode()&0o111 == 0 {
			t.Errorf("hook %s is not executable", name)
		}
	}

	// Uninstall hooks
	if err := Uninstall(dir); err != nil {
		t.Fatalf("Uninstall error: %v", err)
	}

	// Verify hooks are removed
	for _, name := range hookNames {
		path := filepath.Join(hooksDir, name)
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Errorf("hook %s should be removed after uninstall", name)
		}
	}
}

func TestInstallBackupChaining(t *testing.T) {
	dir := initGitRepo(t)
	hooksDir := filepath.Join(dir, ".git", "hooks")

	// Write an existing hook
	existingHook := "#!/bin/bash\necho 'existing hook'\n"
	hookPath := filepath.Join(hooksDir, "pre-commit")
	if err := os.WriteFile(hookPath, []byte(existingHook), 0o755); err != nil {
		t.Fatalf("writing existing hook: %v", err)
	}

	// Install partio hooks
	if err := Install(dir); err != nil {
		t.Fatalf("Install error: %v", err)
	}

	// Original should be backed up
	backupPath := hookPath + ".partio-backup"
	data, err := os.ReadFile(backupPath)
	if err != nil {
		t.Fatalf("backup not found: %v", err)
	}
	if string(data) != existingHook {
		t.Errorf("backup content mismatch: %q", string(data))
	}

	// New hook should be ours
	data, err = os.ReadFile(hookPath)
	if err != nil {
		t.Fatalf("reading hook: %v", err)
	}
	if !isPartioHook(string(data)) {
		t.Error("installed hook missing partio marker")
	}

	// Uninstall should restore original
	if err := Uninstall(dir); err != nil {
		t.Fatalf("Uninstall error: %v", err)
	}

	data, err = os.ReadFile(hookPath)
	if err != nil {
		t.Fatalf("hook should exist after uninstall (restored): %v", err)
	}
	if string(data) != existingHook {
		t.Errorf("original hook not restored: %q", string(data))
	}
}

func TestInstallWorktree(t *testing.T) {
	// Create a main repo
	mainDir := initGitRepo(t)

	// Need at least one commit to create a worktree
	cmd := exec.Command("git", "-C", mainDir, "commit", "--allow-empty", "-m", "init")
	cmd.Env = append(os.Environ(), "GIT_AUTHOR_NAME=test", "GIT_AUTHOR_EMAIL=test@test.com",
		"GIT_COMMITTER_NAME=test", "GIT_COMMITTER_EMAIL=test@test.com")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git commit: %v\n%s", err, out)
	}

	// Create a worktree
	worktreeDir := t.TempDir()
	cmd = exec.Command("git", "-C", mainDir, "worktree", "add", worktreeDir, "-b", "test-branch")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git worktree add: %v\n%s", err, out)
	}

	// Install hooks in worktree — this should not fail
	if err := Install(worktreeDir); err != nil {
		t.Fatalf("Install in worktree error: %v", err)
	}

	// Verify hooks exist in the common git dir (shared across worktrees)
	for _, name := range hookNames {
		cmd := exec.Command("git", "rev-parse", "--git-common-dir")
		cmd.Dir = worktreeDir
		out, err := cmd.Output()
		if err != nil {
			t.Fatalf("rev-parse: %v", err)
		}
		gitDir := string(out[:len(out)-1]) // trim newline
		if !filepath.IsAbs(gitDir) {
			gitDir = filepath.Join(worktreeDir, gitDir)
		}
		hookPath := filepath.Join(gitDir, "hooks", name)
		data, err := os.ReadFile(hookPath)
		if err != nil {
			t.Errorf("hook %s not found in worktree: %v", name, err)
			continue
		}
		if !isPartioHook(string(data)) {
			t.Errorf("hook %s missing partio marker in worktree", name)
		}
	}
}

func TestIsPartioHook(t *testing.T) {
	tests := []struct {
		content  string
		expected bool
	}{
		{"#!/bin/bash\n# Installed by partio\npartio _hook pre-commit", true},
		{"#!/bin/bash\necho hello", false},
		{"", false},
		{partioMarker, true},
	}

	for _, tt := range tests {
		if got := isPartioHook(tt.content); got != tt.expected {
			t.Errorf("isPartioHook(%q) = %v, want %v", tt.content, got, tt.expected)
		}
	}
}
