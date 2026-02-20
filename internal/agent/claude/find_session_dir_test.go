package claude

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestFindSessionDir(t *testing.T) {
	tests := []struct {
		name string
		// dirs maps sanitized directory names to JSONL filenames and their age
		// relative to "now". Positive means older (subtracted from now).
		dirs    map[string][]fileAge
		wantIdx int // index into dirs iteration isn't stable, so we use wantDir name
		wantDir string
		wantErr bool
	}{
		{
			name: "single candidate returns it",
			dirs: map[string][]fileAge{
				"child": {{name: "session.jsonl", age: time.Hour}},
			},
			wantDir: "child",
		},
		{
			name: "picks directory with freshest JSONL",
			dirs: map[string][]fileAge{
				"child":  {{name: "old.jsonl", age: 24 * time.Hour}},
				"parent": {{name: "fresh.jsonl", age: time.Minute}},
			},
			wantDir: "parent",
		},
		{
			name: "multiple files per directory — picks freshest overall",
			dirs: map[string][]fileAge{
				"child": {
					{name: "a.jsonl", age: 48 * time.Hour},
					{name: "b.jsonl", age: 2 * time.Hour},
				},
				"parent": {
					{name: "c.jsonl", age: 24 * time.Hour},
				},
			},
			wantDir: "child", // b.jsonl at 2h is fresher than c.jsonl at 24h
		},
		{
			name: "no JSONL files falls back to first candidate",
			dirs: map[string][]fileAge{
				"child":  {},
				"parent": {},
			},
			wantDir: "child",
		},
		{
			name:    "no candidates returns error",
			dirs:    map[string][]fileAge{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Build a fake filesystem:
			//   tmpDir/home/.claude/projects/<dir>/  — session directories
			//   tmpDir/repo/child/                   — fake repo root
			//   tmpDir/repo/                         — parent directory
			tmpDir := t.TempDir()
			home := filepath.Join(tmpDir, "home")
			projectsDir := filepath.Join(home, ".claude", "projects")

			// We simulate two path levels: tmpDir/repo/child (repoRoot)
			// and tmpDir/repo (parent). Their sanitized forms are computed
			// by sanitizePath, so we create directories matching those.
			repoParent := filepath.Join(tmpDir, "repo")
			repoRoot := filepath.Join(repoParent, "child")
			if err := os.MkdirAll(repoRoot, 0o755); err != nil {
				t.Fatal(err)
			}

			// Map logical names to actual paths
			pathMap := map[string]string{
				"child":  sanitizePath(repoRoot),
				"parent": sanitizePath(repoParent),
			}

			now := time.Now()
			for dirName, files := range tt.dirs {
				sanitized, ok := pathMap[dirName]
				if !ok {
					t.Fatalf("unknown dir name %q", dirName)
				}
				sessionDir := filepath.Join(projectsDir, sanitized)
				if err := os.MkdirAll(sessionDir, 0o755); err != nil {
					t.Fatal(err)
				}
				for _, f := range files {
					p := filepath.Join(sessionDir, f.name)
					if err := os.WriteFile(p, []byte("{}"), 0o644); err != nil {
						t.Fatal(err)
					}
					modTime := now.Add(-f.age)
					if err := os.Chtimes(p, modTime, modTime); err != nil {
						t.Fatal(err)
					}
				}
			}

			// Override HOME so FindSessionDir looks in our temp directory.
			t.Setenv("HOME", home)

			d := &Detector{}
			got, err := d.FindSessionDir(repoRoot)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			wantSanitized := pathMap[tt.wantDir]
			want := filepath.Join(projectsDir, wantSanitized)
			if got != want {
				t.Errorf("got %s, want %s", got, want)
			}
		})
	}
}

type fileAge struct {
	name string
	age  time.Duration
}
