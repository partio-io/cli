package session

import (
	"os"
	"testing"
	"time"
)

func TestCleanupStale_NoSession(t *testing.T) {
	dir := t.TempDir()
	mgr := NewManager(dir)

	result, err := mgr.CleanupStale(10 * time.Minute)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Cleaned {
		t.Error("expected no cleanup when there is no session")
	}
}

func TestCleanupStale_EndedSession(t *testing.T) {
	dir := t.TempDir()
	mgr := NewManager(dir)

	_, err := mgr.Start("claude-code", "main", "/tmp/test")
	if err != nil {
		t.Fatalf("start error: %v", err)
	}
	if err := mgr.End(); err != nil {
		t.Fatalf("end error: %v", err)
	}

	// Backdate the file so it looks old.
	past := time.Now().Add(-20 * time.Minute)
	if err := os.Chtimes(mgr.currentPath(), past, past); err != nil {
		t.Fatalf("chtimes error: %v", err)
	}

	result, err := mgr.CleanupStale(10 * time.Minute)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Cleaned {
		t.Error("should not clean up an already-ended session")
	}
}

func TestCleanupStale_RecentActiveSession(t *testing.T) {
	dir := t.TempDir()
	mgr := NewManager(dir)

	_, err := mgr.Start("claude-code", "main", "/tmp/test")
	if err != nil {
		t.Fatalf("start error: %v", err)
	}

	// File is fresh — should not be cleaned up.
	result, err := mgr.CleanupStale(10 * time.Minute)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Cleaned {
		t.Error("should not clean up a recently updated session")
	}

	cur, _ := mgr.Current()
	if cur.State != StateActive {
		t.Errorf("expected state=active, got %s", cur.State)
	}
}

func TestCleanupStale_OldActiveSession_NoPID(t *testing.T) {
	dir := t.TempDir()
	mgr := NewManager(dir)

	_, err := mgr.Start("claude-code", "main", "/tmp/test")
	if err != nil {
		t.Fatalf("start error: %v", err)
	}

	// Backdate the state file so it appears stale.
	past := time.Now().Add(-20 * time.Minute)
	if err := os.Chtimes(mgr.currentPath(), past, past); err != nil {
		t.Fatalf("chtimes error: %v", err)
	}

	result, err := mgr.CleanupStale(10 * time.Minute)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Cleaned {
		t.Error("expected stale session to be cleaned up")
	}
	if result.Session == nil {
		t.Fatal("expected session to be set in result")
	}

	cur, _ := mgr.Current()
	if cur.State != StateEnded {
		t.Errorf("expected state=ended after cleanup, got %s", cur.State)
	}
	if cur.EndedAt.IsZero() {
		t.Error("expected ended_at to be set")
	}
}

func TestCleanupStale_OldIdleSession_NoPID(t *testing.T) {
	dir := t.TempDir()
	mgr := NewManager(dir)

	s, err := mgr.Start("claude-code", "main", "/tmp/test")
	if err != nil {
		t.Fatalf("start error: %v", err)
	}
	s.State = StateIdle
	if err := mgr.save(s); err != nil {
		t.Fatalf("save error: %v", err)
	}

	past := time.Now().Add(-20 * time.Minute)
	if err := os.Chtimes(mgr.currentPath(), past, past); err != nil {
		t.Fatalf("chtimes error: %v", err)
	}

	result, err := mgr.CleanupStale(10 * time.Minute)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Cleaned {
		t.Error("expected stale idle session to be cleaned up")
	}

	cur, _ := mgr.Current()
	if cur.State != StateEnded {
		t.Errorf("expected state=ended after cleanup, got %s", cur.State)
	}
}

func TestCleanupStale_OldActiveSession_AlivePID(t *testing.T) {
	dir := t.TempDir()
	mgr := NewManager(dir)

	// Use current process PID — we know it's alive.
	pid := os.Getpid()
	s, err := mgr.Start("claude-code", "main", "/tmp/test")
	if err != nil {
		t.Fatalf("start error: %v", err)
	}
	s.AgentPID = pid
	if err := mgr.save(s); err != nil {
		t.Fatalf("save error: %v", err)
	}

	past := time.Now().Add(-20 * time.Minute)
	if err := os.Chtimes(mgr.currentPath(), past, past); err != nil {
		t.Fatalf("chtimes error: %v", err)
	}

	result, err := mgr.CleanupStale(10 * time.Minute)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Cleaned {
		t.Error("should not clean up a session whose process is still alive")
	}

	cur, _ := mgr.Current()
	if cur.State != StateActive {
		t.Errorf("expected state=active (process still alive), got %s", cur.State)
	}
}

func TestCleanupStale_OldActiveSession_DeadPID(t *testing.T) {
	dir := t.TempDir()
	mgr := NewManager(dir)

	// PID 1 is init — always alive; use a clearly dead PID instead.
	// On Linux, PIDs > /proc/sys/kernel/pid_max don't exist.
	// Use a very high number unlikely to be running.
	const deadPID = 999999999

	s, err := mgr.Start("claude-code", "main", "/tmp/test")
	if err != nil {
		t.Fatalf("start error: %v", err)
	}
	s.AgentPID = deadPID
	if err := mgr.save(s); err != nil {
		t.Fatalf("save error: %v", err)
	}

	past := time.Now().Add(-20 * time.Minute)
	if err := os.Chtimes(mgr.currentPath(), past, past); err != nil {
		t.Fatalf("chtimes error: %v", err)
	}

	result, err := mgr.CleanupStale(10 * time.Minute)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Cleaned {
		t.Error("expected session with dead PID to be cleaned up")
	}

	cur, _ := mgr.Current()
	if cur.State != StateEnded {
		t.Errorf("expected state=ended, got %s", cur.State)
	}
}
