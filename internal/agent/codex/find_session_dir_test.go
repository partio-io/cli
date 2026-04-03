package codex

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindSessionDir(t *testing.T) {
	tests := []struct {
		name      string
		createDir bool
		wantErr   bool
	}{
		{
			name:      "returns directory when .codex exists",
			createDir: true,
			wantErr:   false,
		},
		{
			name:      "returns error when .codex does not exist",
			createDir: false,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			home := filepath.Join(tmpDir, "home")

			if tt.createDir {
				codexDir := filepath.Join(home, ".codex")
				if err := os.MkdirAll(codexDir, 0o755); err != nil {
					t.Fatal(err)
				}
			}

			t.Setenv("HOME", home)

			d := &Detector{}
			got, err := d.FindSessionDir(filepath.Join(tmpDir, "repo"))
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			want := filepath.Join(home, ".codex")
			if got != want {
				t.Errorf("got %s, want %s", got, want)
			}
		})
	}
}
