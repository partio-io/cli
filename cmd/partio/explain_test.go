package main

import (
	"strings"
	"testing"
	"time"

	"github.com/partio-io/cli/internal/checkpoint"
)

func TestRunExplainChangedEntities(t *testing.T) {
	longParams := strings.Repeat("a", 120)

	tests := []struct {
		name            string
		diff            string
		wantInOutput    []string
		wantNotInOutput []string
	}{
		{
			name: "go diff lists entity changes",
			diff: strings.Join([]string{
				"diff --git a/internal/a.go b/internal/a.go",
				"index 1111111..2222222 100644",
				"--- a/internal/a.go",
				"+++ b/internal/a.go",
				"@@ -1,3 +1,4 @@",
				" package a",
				"-func Read(id string) error {",
				"+func Read(id string) (*Data, error) {",
				"+func New(x int) error {",
				" \treturn nil",
			}, "\n"),
			wantInOutput: []string{
				"\nChanged entities:\n  internal/a.go\n",
				"    signature changed  func Read(id string) (*Data, error)\n",
				"    added              func New(x int) error\n",
			},
		},
		{
			name: "long signature is cut",
			diff: strings.Join([]string{
				"diff --git a/a.go b/a.go",
				"--- a/a.go",
				"+++ b/a.go",
				"@@ -1,0 +1,1 @@",
				"+func Long(" + longParams + " int) {}",
			}, "\n"),
			wantInOutput:    []string{"    added              func Long(" + longParams[:89] + "…\n"},
			wantNotInOutput: []string{longParams},
		},
		{
			name: "unsupported-only diff has no section",
			diff: strings.Join([]string{
				"diff --git a/README.md b/README.md",
				"--- a/README.md",
				"+++ b/README.md",
				"@@ -1,1 +1,1 @@",
				"-func Old() {}",
				"+func New() {}",
			}, "\n"),
			wantNotInOutput: []string{"Changed entities:"},
		},
		{
			name:            "empty diff has no section",
			diff:            "",
			wantNotInOutput: []string{"Changed entities:"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			rewindInitTestRepo(t, dir)

			const id = "abcdef123456"
			cp := &checkpoint.Checkpoint{
				ID:          id,
				Branch:      "main",
				CommitHash:  "deadbeef",
				CreatedAt:   time.Now().Truncate(time.Second),
				Agent:       "claude-code",
				ContentHash: id,
				PlanSlug:    "my-plan",
			}
			sf := &checkpoint.SessionFiles{
				ContentHash: "hash",
				Diff:        tt.diff,
				Metadata:    checkpoint.SessionMetadata{Agent: cp.Agent},
				Prompt:      "the prompt",
			}
			if err := checkpoint.NewStore(dir).Write(cp, sf); err != nil {
				t.Fatalf("Store.Write: %v", err)
			}

			t.Chdir(dir)

			var runErr error
			out := captureOutput(t, func() {
				runErr = runExplain(nil, []string{id})
			})
			if runErr != nil {
				t.Fatalf("runExplain() error = %v", runErr)
			}

			for _, want := range tt.wantInOutput {
				if !strings.Contains(out, want) {
					t.Errorf("output missing %q; got:\n%s", want, out)
				}
			}
			for _, notWant := range tt.wantNotInOutput {
				if strings.Contains(out, notWant) {
					t.Errorf("output unexpectedly contains %q; got:\n%s", notWant, out)
				}
			}

			// The section sits between the metadata block and the prompt.
			if section := strings.Index(out, "Changed entities:"); section >= 0 {
				plan := strings.Index(out, "Plan:")
				prompt := strings.Index(out, "Prompt:")
				if plan < 0 || prompt < 0 || section < plan || section > prompt {
					t.Errorf("Changed entities: not between Plan: and Prompt:; got:\n%s", out)
				}
			}
			if !strings.Contains(out, "Prompt:\nthe prompt") {
				t.Errorf("prompt missing; got:\n%s", out)
			}
		})
	}
}
