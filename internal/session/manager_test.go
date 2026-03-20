package session

import (
	"testing"
)

func TestManagerStartAndCurrent(t *testing.T) {
	dir := t.TempDir()
	mgr := NewManager(dir)

	// No current session initially
	s, err := mgr.Current()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s != nil {
		t.Error("expected nil session initially")
	}

	// Start a session
	s, err = mgr.Start("claude-code", "main", "/tmp/test")
	if err != nil {
		t.Fatalf("start error: %v", err)
	}

	if s.Agent != "claude-code" {
		t.Errorf("expected agent=claude-code, got %s", s.Agent)
	}
	if s.State != StateActive {
		t.Errorf("expected state=active, got %s", s.State)
	}
	if s.Branch != "main" {
		t.Errorf("expected branch=main, got %s", s.Branch)
	}

	// Current should return the session
	cur, err := mgr.Current()
	if err != nil {
		t.Fatalf("current error: %v", err)
	}
	if cur == nil {
		t.Fatal("expected current session")
	}
	if cur.ID != s.ID {
		t.Errorf("expected same session ID, got %s vs %s", cur.ID, s.ID)
	}
}

func TestManagerEnd(t *testing.T) {
	dir := t.TempDir()
	mgr := NewManager(dir)

	_, err := mgr.Start("claude-code", "main", "/tmp/test")
	if err != nil {
		t.Fatalf("start error: %v", err)
	}

	if err := mgr.End(); err != nil {
		t.Fatalf("end error: %v", err)
	}

	cur, err := mgr.Current()
	if err != nil {
		t.Fatalf("current error: %v", err)
	}
	if cur.State != StateEnded {
		t.Errorf("expected state=ended, got %s", cur.State)
	}
}

func TestManagerMarkCondensed(t *testing.T) {
	t.Run("marks existing session as condensed and ended", func(t *testing.T) {
		dir := t.TempDir()
		mgr := NewManager(dir)

		_, err := mgr.Start("claude-code", "main", "/tmp/test")
		if err != nil {
			t.Fatalf("start error: %v", err)
		}

		if err := mgr.MarkCondensed("sess-abc"); err != nil {
			t.Fatalf("mark condensed error: %v", err)
		}

		cur, err := mgr.Current()
		if err != nil {
			t.Fatalf("current error: %v", err)
		}
		if cur.State != StateEnded {
			t.Errorf("expected state=ended, got %s", cur.State)
		}
		if !cur.Condensed {
			t.Error("expected condensed=true")
		}
		if cur.CapturedSessionID != "sess-abc" {
			t.Errorf("expected captured_session_id=sess-abc, got %s", cur.CapturedSessionID)
		}
		if cur.CapturedAt.IsZero() {
			t.Error("expected captured_at to be set")
		}
	})

	t.Run("creates session entry when none exists", func(t *testing.T) {
		dir := t.TempDir()
		mgr := NewManager(dir)

		if err := mgr.MarkCondensed("sess-xyz"); err != nil {
			t.Fatalf("mark condensed error: %v", err)
		}

		cur, err := mgr.Current()
		if err != nil {
			t.Fatalf("current error: %v", err)
		}
		if cur == nil {
			t.Fatal("expected session to be created")
		}
		if cur.State != StateEnded || !cur.Condensed || cur.CapturedSessionID != "sess-xyz" {
			t.Errorf("unexpected session state: %+v", cur)
		}
	})
}

func TestManagerClear(t *testing.T) {
	dir := t.TempDir()
	mgr := NewManager(dir)

	_, err := mgr.Start("claude-code", "main", "/tmp/test")
	if err != nil {
		t.Fatalf("start error: %v", err)
	}

	if err := mgr.Clear(); err != nil {
		t.Fatalf("clear error: %v", err)
	}

	cur, err := mgr.Current()
	if err != nil {
		t.Fatalf("current error: %v", err)
	}
	if cur != nil {
		t.Error("expected nil session after clear")
	}
}
