package hooks

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectExternalHookManagers(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(t *testing.T, dir string)
		expected []string // expected manager names
	}{
		{
			name:     "no hook managers",
			setup:    func(t *testing.T, dir string) {},
			expected: nil,
		},
		{
			name: "husky directory",
			setup: func(t *testing.T, dir string) {
				if err := os.Mkdir(filepath.Join(dir, ".husky"), 0o755); err != nil {
					t.Fatal(err)
				}
			},
			expected: []string{"Husky"},
		},
		{
			name: "husky in package.json",
			setup: func(t *testing.T, dir string) {
				pkg := `{"scripts":{"prepare":"husky install"}}`
				if err := os.WriteFile(filepath.Join(dir, "package.json"), []byte(pkg), 0o644); err != nil {
					t.Fatal(err)
				}
			},
			expected: []string{"Husky"},
		},
		{
			name: "husky directory takes precedence over package.json",
			setup: func(t *testing.T, dir string) {
				if err := os.Mkdir(filepath.Join(dir, ".husky"), 0o755); err != nil {
					t.Fatal(err)
				}
				pkg := `{"scripts":{"prepare":"husky install"}}`
				if err := os.WriteFile(filepath.Join(dir, "package.json"), []byte(pkg), 0o644); err != nil {
					t.Fatal(err)
				}
			},
			expected: []string{"Husky"},
		},
		{
			name: "lefthook.yml",
			setup: func(t *testing.T, dir string) {
				if err := os.WriteFile(filepath.Join(dir, "lefthook.yml"), []byte(""), 0o644); err != nil {
					t.Fatal(err)
				}
			},
			expected: []string{"Lefthook"},
		},
		{
			name: "dot lefthook.yml",
			setup: func(t *testing.T, dir string) {
				if err := os.WriteFile(filepath.Join(dir, ".lefthook.yml"), []byte(""), 0o644); err != nil {
					t.Fatal(err)
				}
			},
			expected: []string{"Lefthook"},
		},
		{
			name: "overcommit",
			setup: func(t *testing.T, dir string) {
				if err := os.WriteFile(filepath.Join(dir, ".overcommit.yml"), []byte(""), 0o644); err != nil {
					t.Fatal(err)
				}
			},
			expected: []string{"Overcommit"},
		},
		{
			name: "multiple managers",
			setup: func(t *testing.T, dir string) {
				if err := os.Mkdir(filepath.Join(dir, ".husky"), 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(dir, "lefthook.yml"), []byte(""), 0o644); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(dir, ".overcommit.yml"), []byte(""), 0o644); err != nil {
					t.Fatal(err)
				}
			},
			expected: []string{"Husky", "Lefthook", "Overcommit"},
		},
		{
			name: "invalid package.json ignored",
			setup: func(t *testing.T, dir string) {
				if err := os.WriteFile(filepath.Join(dir, "package.json"), []byte("not json"), 0o644); err != nil {
					t.Fatal(err)
				}
			},
			expected: nil,
		},
		{
			name: "package.json without prepare script",
			setup: func(t *testing.T, dir string) {
				pkg := `{"scripts":{"test":"jest"}}`
				if err := os.WriteFile(filepath.Join(dir, "package.json"), []byte(pkg), 0o644); err != nil {
					t.Fatal(err)
				}
			},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			tt.setup(t, dir)

			managers := DetectExternalHookManagers(dir)

			if len(managers) != len(tt.expected) {
				t.Fatalf("got %d managers, want %d", len(managers), len(tt.expected))
			}

			for i, m := range managers {
				if m.Name != tt.expected[i] {
					t.Errorf("manager[%d].Name = %q, want %q", i, m.Name, tt.expected[i])
				}
				if m.Reason == "" {
					t.Errorf("manager[%d].Reason is empty", i)
				}
			}
		})
	}
}
