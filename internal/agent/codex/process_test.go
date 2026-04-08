package codex

import "testing"

func TestDetector_Name(t *testing.T) {
	d := New()
	if d.Name() != "codex" {
		t.Errorf("expected name=codex, got %s", d.Name())
	}
}

func TestDetector_IsRunning(t *testing.T) {
	d := New()
	// We can't control whether codex is actually running,
	// but we can verify the function returns without unexpected errors.
	_, err := d.IsRunning()
	if err != nil {
		t.Errorf("IsRunning() returned unexpected error: %v", err)
	}
}

func TestDetector_FindSessionDir(t *testing.T) {
	t.Run("returns error when no session directory exists", func(t *testing.T) {
		// Use a fake HOME so ~/.codex/sessions/ doesn't exist
		t.Setenv("HOME", t.TempDir())
		d := New()
		_, err := d.FindSessionDir(t.TempDir())
		if err == nil {
			t.Error("FindSessionDir() expected error, got nil")
		}
	})
}
