package claude

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindAllSessions_SingleSession(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	repoRoot := "/tmp/myrepo"
	sessionDir := filepath.Join(home, ".claude", "projects", sanitizePath(repoRoot))
	if err := os.MkdirAll(sessionDir, 0o755); err != nil {
		t.Fatalf("creating session dir: %v", err)
	}

	content := `{"type":"human","role":"human","message":"hello","timestamp":1700000000,"sessionId":"primary-001"}
{"type":"assistant","role":"assistant","message":"hi","timestamp":1700000010}
`
	path := filepath.Join(sessionDir, "session1.jsonl")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("writing session: %v", err)
	}

	d := New()
	primaryPath, primaryData, subAgents, err := d.FindAllSessions(repoRoot)
	if err != nil {
		t.Fatalf("FindAllSessions error: %v", err)
	}

	if primaryPath != path {
		t.Errorf("expected primary path %s, got %s", path, primaryPath)
	}
	if primaryData == nil {
		t.Fatal("expected primary data, got nil")
	}
	if primaryData.SessionID != "primary-001" {
		t.Errorf("expected session ID primary-001, got %s", primaryData.SessionID)
	}
	if len(subAgents) != 0 {
		t.Errorf("expected 0 subagent sessions, got %d", len(subAgents))
	}
}

func TestFindAllSessions_WithSubAgent(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	repoRoot := "/tmp/myrepo"
	sessionDir := filepath.Join(home, ".claude", "projects", sanitizePath(repoRoot))
	if err := os.MkdirAll(sessionDir, 0o755); err != nil {
		t.Fatalf("creating session dir: %v", err)
	}

	// Primary session: earlier timestamps
	primaryContent := `{"type":"human","role":"human","message":"run task","timestamp":1700000000,"sessionId":"primary-001"}
{"type":"assistant","role":"assistant","message":"spawning subagent","timestamp":1700000010}
`
	// Subagent session: later timestamps
	subAgentContent := `{"type":"human","role":"human","message":"subtask","timestamp":1700001000,"sessionId":"subagent-002"}
{"type":"assistant","role":"assistant","message":"done","timestamp":1700001010}
`

	primaryPath := filepath.Join(sessionDir, "primary.jsonl")
	subPath := filepath.Join(sessionDir, "subagent.jsonl")
	if err := os.WriteFile(primaryPath, []byte(primaryContent), 0o644); err != nil {
		t.Fatalf("writing primary session: %v", err)
	}
	if err := os.WriteFile(subPath, []byte(subAgentContent), 0o644); err != nil {
		t.Fatalf("writing subagent session: %v", err)
	}

	d := New()
	gotPrimaryPath, primaryData, subAgents, err := d.FindAllSessions(repoRoot)
	if err != nil {
		t.Fatalf("FindAllSessions error: %v", err)
	}

	if gotPrimaryPath != primaryPath {
		t.Errorf("expected primary path %s, got %s", primaryPath, gotPrimaryPath)
	}
	if primaryData == nil || primaryData.SessionID != "primary-001" {
		t.Errorf("expected primary session ID primary-001, got %v", primaryData)
	}

	if len(subAgents) != 1 {
		t.Fatalf("expected 1 subagent session, got %d", len(subAgents))
	}
	if subAgents[0].Data == nil || subAgents[0].Data.SessionID != "subagent-002" {
		t.Errorf("expected subagent session ID subagent-002, got %v", subAgents[0].Data)
	}
	if subAgents[0].Path != subPath {
		t.Errorf("expected subagent path %s, got %s", subPath, subAgents[0].Path)
	}
}

func TestFindAllSessions_NoDuplicateAttribution(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	repoRoot := "/tmp/myrepo"
	sessionDir := filepath.Join(home, ".claude", "projects", sanitizePath(repoRoot))
	if err := os.MkdirAll(sessionDir, 0o755); err != nil {
		t.Fatalf("creating session dir: %v", err)
	}

	// Same session ID in both files — dedup logic in postcommit prevents double attribution.
	// Here we just verify FindAllSessions returns distinct paths correctly.
	primaryContent := `{"type":"human","role":"human","message":"task","timestamp":1700000000,"sessionId":"sess-abc"}
`
	subContent := `{"type":"human","role":"human","message":"subtask","timestamp":1700002000,"sessionId":"sess-xyz"}
`

	if err := os.WriteFile(filepath.Join(sessionDir, "a.jsonl"), []byte(primaryContent), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sessionDir, "b.jsonl"), []byte(subContent), 0o644); err != nil {
		t.Fatal(err)
	}

	d := New()
	_, primaryData, subAgents, err := d.FindAllSessions(repoRoot)
	if err != nil {
		t.Fatalf("FindAllSessions error: %v", err)
	}

	// Primary and subagent should have different session IDs.
	if primaryData == nil {
		t.Fatal("expected primary data")
	}
	if len(subAgents) != 1 {
		t.Fatalf("expected 1 subagent, got %d", len(subAgents))
	}
	if primaryData.SessionID == subAgents[0].Data.SessionID {
		t.Error("primary and subagent should have distinct session IDs")
	}
}
