package main

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/partio-io/cli/internal/config"
)

func writeRepoSettings(t *testing.T, dir string, files map[string]string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(dir, config.PartioDir), 0o755); err != nil {
		t.Fatal(err)
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, config.PartioDir, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func TestStealthOverrides(t *testing.T) {
	repoFile := filepath.Join(config.PartioDir, "settings.json")
	localFile := filepath.Join(config.PartioDir, "settings.local.json")
	push := "strategy_options.push_sessions"
	trailers := "strategy_options.disable_trailers"

	tests := []struct {
		name  string
		files map[string]string
		want  []stealthOverride
	}{
		{
			name: "no repo settings",
		},
		{
			name:  "repo settings agree with stealth",
			files: map[string]string{"settings.json": `{"strategy_options": {"push_sessions": false, "disable_trailers": true}}`},
		},
		{
			name:  "repo settings without strategy_options",
			files: map[string]string{"settings.json": `{"enabled": true}`},
		},
		{
			name:  "settings.json sets push_sessions=true",
			files: map[string]string{"settings.json": `{"strategy_options": {"push_sessions": true}}`},
			want:  []stealthOverride{{file: repoFile, key: push, value: true}},
		},
		{
			name:  "settings.json sets disable_trailers=false",
			files: map[string]string{"settings.json": `{"strategy_options": {"disable_trailers": false}}`},
			want:  []stealthOverride{{file: repoFile, key: trailers, value: false}},
		},
		{
			name:  "settings.local.json sets push_sessions=true",
			files: map[string]string{"settings.local.json": `{"strategy_options": {"push_sessions": true}}`},
			want:  []stealthOverride{{file: localFile, key: push, value: true}},
		},
		{
			name:  "settings.local.json sets disable_trailers=false",
			files: map[string]string{"settings.local.json": `{"strategy_options": {"disable_trailers": false}}`},
			want:  []stealthOverride{{file: localFile, key: trailers, value: false}},
		},
		{
			name: "both layers override, local first",
			files: map[string]string{
				"settings.json":       `{"strategy_options": {"push_sessions": true, "disable_trailers": false}}`,
				"settings.local.json": `{"strategy_options": {"disable_trailers": false}}`,
			},
			want: []stealthOverride{
				{file: localFile, key: trailers, value: false},
				{file: repoFile, key: push, value: true},
				{file: repoFile, key: trailers, value: false},
			},
		},
		{
			name:  "malformed file is skipped",
			files: map[string]string{"settings.json": `{"strategy_options": {"push_sessions": true`},
		},
		{
			name:  "malformed value is skipped",
			files: map[string]string{"settings.json": `{"strategy_options": {"push_sessions": "yes"}}`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			if tt.files != nil {
				writeRepoSettings(t, dir, tt.files)
			}
			if got := stealthOverrides(dir); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("stealthOverrides() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEnableStealthOverrideWarning(t *testing.T) {
	tests := []struct {
		name        string
		files       map[string]string
		wantWarning []string // empty means no override warning
	}{
		{
			name:        "settings.json sets push_sessions=true",
			files:       map[string]string{"settings.json": `{"enabled": true, "strategy_options": {"push_sessions": true}}`},
			wantWarning: []string{filepath.Join(config.PartioDir, "settings.json") + " sets strategy_options.push_sessions=true"},
		},
		{
			name:        "settings.local.json sets disable_trailers=false",
			files:       map[string]string{"settings.local.json": `{"strategy_options": {"disable_trailers": false}}`},
			wantWarning: []string{filepath.Join(config.PartioDir, "settings.local.json") + " sets strategy_options.disable_trailers=false"},
		},
		{
			name: "no overriding layer",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir, _, _ := newEnableTestRepo(t)
			if tt.files != nil {
				writeRepoSettings(t, dir, tt.files)
			}

			out, err := runEnableWith(t, map[string]string{"stealth": "true"})
			if err != nil {
				t.Fatalf("runEnable(--stealth) error = %v, want success", err)
			}

			if len(tt.wantWarning) == 0 {
				if strings.Contains(out, "stealth is not in effect") {
					t.Errorf("unexpected override warning:\n%s", out)
				}
			}
			for _, want := range tt.wantWarning {
				if !strings.Contains(out, want+"; stealth is not in effect in this repo") {
					t.Errorf("output missing warning %q; got:\n%s", want, out)
				}
			}

			// Repository settings files are left byte-for-byte unchanged.
			for name, content := range tt.files {
				got, err := os.ReadFile(filepath.Join(dir, config.PartioDir, name))
				if err != nil {
					t.Fatal(err)
				}
				if string(got) != content {
					t.Errorf("%s was changed:\ngot:  %s\nwant: %s", name, got, content)
				}
			}

			// partio status follows the effective config.
			effective, err := config.Load(dir)
			if err != nil {
				t.Fatalf("config.Load: %v", err)
			}
			previousCfg := cfg
			cfg = effective
			t.Cleanup(func() { cfg = previousCfg })

			var statusErr error
			statusOut := captureOutput(t, func() { statusErr = runStatus(newStatusCmd(), nil) })
			if statusErr != nil {
				t.Fatalf("runStatus() error = %v", statusErr)
			}
			hasMode := strings.Contains(statusOut, stealthModeLine)
			if overridden := len(tt.wantWarning) > 0; hasMode == overridden {
				t.Errorf("status shows Mode line = %v, want %v:\n%s", hasMode, !overridden, statusOut)
			}
		})
	}
}
