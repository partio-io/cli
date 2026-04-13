package hooks

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/partio-io/cli/internal/session"
)

func TestShouldSkipSession(t *testing.T) {
	tests := []struct {
		name         string
		setup        func(t *testing.T, partioDir, jsonlPath string)
		sessionID    string
		modifyJSONL  bool // touch JSONL after marking condensed to simulate new content
		wantSkip     bool
	}{
		{
			name: "skip when ended and condensed with matching session ID",
			setup: func(t *testing.T, partioDir, jsonlPath string) {
				mgr := session.NewManager(partioDir)
				if err := mgr.MarkCondensed("sess-123"); err != nil {
					t.Fatalf("MarkCondensed: %v", err)
				}
			},
			sessionID: "sess-123",
			wantSkip:  true,
		},
		{
			name: "do not skip when session ID does not match",
			setup: func(t *testing.T, partioDir, jsonlPath string) {
				mgr := session.NewManager(partioDir)
				if err := mgr.MarkCondensed("sess-other"); err != nil {
					t.Fatalf("MarkCondensed: %v", err)
				}
			},
			sessionID: "sess-123",
			wantSkip:  false,
		},
		{
			name:      "do not skip when no session state exists",
			setup:     func(t *testing.T, partioDir, jsonlPath string) {},
			sessionID: "sess-123",
			wantSkip:  false,
		},
		{
			name: "do not skip when session is active (not ended)",
			setup: func(t *testing.T, partioDir, jsonlPath string) {
				mgr := session.NewManager(partioDir)
				if _, err := mgr.Start("claude-code", "main", "/tmp/test"); err != nil {
					t.Fatalf("Start: %v", err)
				}
				// Active session — do NOT mark condensed
			},
			sessionID: "sess-123",
			wantSkip:  false,
		},
		{
			name: "do not skip when JSONL was modified after capture",
			setup: func(t *testing.T, partioDir, jsonlPath string) {
				mgr := session.NewManager(partioDir)
				if err := mgr.MarkCondensed("sess-123"); err != nil {
					t.Fatalf("MarkCondensed: %v", err)
				}
			},
			sessionID:   "sess-123",
			modifyJSONL: true,
			wantSkip:    false,
		},
		{
			name:      "do not skip when session ID is empty",
			setup:     func(t *testing.T, partioDir, jsonlPath string) {},
			sessionID: "",
			wantSkip:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			partioDir := t.TempDir()
			jsonlDir := t.TempDir()
			jsonlPath := filepath.Join(jsonlDir, "session.jsonl")

			if err := os.WriteFile(jsonlPath, []byte(`{"sessionId":"sess-123"}`+"\n"), 0o644); err != nil {
				t.Fatalf("writing JSONL: %v", err)
			}

			tt.setup(t, partioDir, jsonlPath)

			if tt.modifyJSONL {
				// Ensure modification time is strictly after CapturedAt by waiting a moment.
				// Filesystem timestamp granularity can be coarse (e.g. 1s on some systems).
				time.Sleep(50 * time.Millisecond)
				if err := os.WriteFile(jsonlPath, []byte(`{"sessionId":"sess-123"}`+"\n"+"new line\n"), 0o644); err != nil {
					t.Fatalf("modifying JSONL: %v", err)
				}
			}

			got := shouldSkipSession(partioDir, tt.sessionID, jsonlPath)
			if got != tt.wantSkip {
				t.Errorf("shouldSkipSession() = %v, want %v", got, tt.wantSkip)
			}
		})
	}
}
