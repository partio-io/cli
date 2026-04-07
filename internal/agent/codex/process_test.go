package codex

import "testing"

func TestDetector_Name(t *testing.T) {
	d := New()
	if d.Name() != "codex" {
		t.Errorf("expected name=codex, got %s", d.Name())
	}
}

func TestDetector_IsRunning(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "returns without error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := New()
			// We can't control whether codex is actually running in CI,
			// but we can verify the function returns without unexpected errors.
			_, err := d.IsRunning()
			if err != nil {
				t.Errorf("IsRunning() returned unexpected error: %v", err)
			}
		})
	}
}

func TestDetector_FindSessionDir(t *testing.T) {
	tests := []struct {
		name     string
		repoRoot string
		wantErr  bool
	}{
		{
			name:     "returns error when no session directory exists",
			repoRoot: t.TempDir(),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := New()
			_, err := d.FindSessionDir(tt.repoRoot)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindSessionDir() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
