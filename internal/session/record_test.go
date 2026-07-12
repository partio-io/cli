package session

import "testing"

func TestRecordActive_CreatesActiveSessionWithPID(t *testing.T) {
	mgr := NewManager(t.TempDir())

	if err := mgr.RecordActive("claude-code", "main", "/tmp/repo", 4242); err != nil {
		t.Fatalf("RecordActive: %v", err)
	}

	s, err := mgr.Current()
	if err != nil || s == nil {
		t.Fatalf("Current: %v (s=%v)", err, s)
	}
	if s.State != StateActive {
		t.Errorf("state = %s, want active", s.State)
	}
	if s.AgentPID != 4242 {
		t.Errorf("AgentPID = %d, want 4242", s.AgentPID)
	}
	if s.Agent != "claude-code" {
		t.Errorf("Agent = %q, want claude-code", s.Agent)
	}
}

func TestRecordActive_RefreshesExistingSession(t *testing.T) {
	mgr := NewManager(t.TempDir())

	if err := mgr.RecordActive("claude-code", "main", "/tmp/repo", 1); err != nil {
		t.Fatalf("first RecordActive: %v", err)
	}
	first, _ := mgr.Current()

	if err := mgr.RecordActive("claude-code", "main", "/tmp/repo", 2); err != nil {
		t.Fatalf("second RecordActive: %v", err)
	}
	second, _ := mgr.Current()

	if second.ID != first.ID {
		t.Errorf("session ID changed on refresh: %s -> %s", first.ID, second.ID)
	}
	if second.AgentPID != 2 {
		t.Errorf("AgentPID = %d, want 2 (refreshed)", second.AgentPID)
	}
	if second.State != StateActive {
		t.Errorf("state = %s, want active", second.State)
	}
}

func TestRecordActive_ReplacesEndedSession(t *testing.T) {
	mgr := NewManager(t.TempDir())

	// A previously ended/condensed session must not be refreshed — new agent
	// activity should start a fresh ACTIVE session.
	if err := mgr.MarkCondensed("sess-1"); err != nil {
		t.Fatalf("MarkCondensed: %v", err)
	}

	if err := mgr.RecordActive("claude-code", "main", "/tmp/repo", 7); err != nil {
		t.Fatalf("RecordActive: %v", err)
	}

	s, _ := mgr.Current()
	if s.State != StateActive {
		t.Errorf("state = %s, want active", s.State)
	}
	if s.AgentPID != 7 {
		t.Errorf("AgentPID = %d, want 7", s.AgentPID)
	}
	if s.Condensed {
		t.Error("expected Condensed to be reset on a fresh active session")
	}
}
