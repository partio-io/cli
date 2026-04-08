package hooks

import (
	"errors"
	"testing"
	"time"

	"github.com/partio-io/cli/internal/agent"
)

func TestFindSessionWithRetry_SuccessFirstTry(t *testing.T) {
	calls := 0
	finder := func(repoRoot string) (string, *agent.SessionData, error) {
		calls++
		return "/path/session.jsonl", &agent.SessionData{SessionID: "abc123"}, nil
	}

	path, data, err := findSessionWithRetry(finder, "/repo", 3*time.Second)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if path != "/path/session.jsonl" {
		t.Errorf("unexpected path: %s", path)
	}
	if data.SessionID != "abc123" {
		t.Errorf("unexpected session ID: %s", data.SessionID)
	}
	if calls != 1 {
		t.Errorf("expected 1 call, got %d", calls)
	}
}

func TestFindSessionWithRetry_SuccessAfterRetry(t *testing.T) {
	calls := 0
	finder := func(repoRoot string) (string, *agent.SessionData, error) {
		calls++
		if calls < 3 {
			return "", nil, errors.New("not ready")
		}
		return "/path/session.jsonl", &agent.SessionData{SessionID: "abc123"}, nil
	}

	path, data, err := findSessionWithRetry(finder, "/repo", 5*time.Second)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if data.SessionID != "abc123" {
		t.Errorf("unexpected session ID: %s", data.SessionID)
	}
	if path != "/path/session.jsonl" {
		t.Errorf("unexpected path: %s", path)
	}
	if calls < 3 {
		t.Errorf("expected at least 3 calls, got %d", calls)
	}
}

func TestFindSessionWithRetry_TimeoutExhausted(t *testing.T) {
	calls := 0
	finder := func(repoRoot string) (string, *agent.SessionData, error) {
		calls++
		return "", nil, errors.New("not ready")
	}

	start := time.Now()
	_, _, err := findSessionWithRetry(finder, "/repo", 300*time.Millisecond)
	elapsed := time.Since(start)

	if err == nil {
		t.Error("expected error after timeout")
	}
	if calls < 2 {
		t.Errorf("expected retries, got %d calls", calls)
	}
	if elapsed < 290*time.Millisecond {
		t.Errorf("returned too early: %v", elapsed)
	}
	if elapsed > 2*time.Second {
		t.Errorf("took too long: %v", elapsed)
	}
}

func TestFindSessionWithRetry_ZeroTimeout_NoRetry(t *testing.T) {
	calls := 0
	finder := func(repoRoot string) (string, *agent.SessionData, error) {
		calls++
		return "", nil, errors.New("not ready")
	}

	_, _, err := findSessionWithRetry(finder, "/repo", 0)

	if err == nil {
		t.Error("expected error")
	}
	if calls != 1 {
		t.Errorf("expected exactly 1 call with zero timeout, got %d", calls)
	}
}

func TestFindSessionWithRetry_EmptySessionID_Retries(t *testing.T) {
	calls := 0
	finder := func(repoRoot string) (string, *agent.SessionData, error) {
		calls++
		if calls < 2 {
			// File found but no session ID yet (empty/not flushed)
			return "/path/session.jsonl", &agent.SessionData{SessionID: ""}, nil
		}
		return "/path/session.jsonl", &agent.SessionData{SessionID: "abc123"}, nil
	}

	_, data, err := findSessionWithRetry(finder, "/repo", 5*time.Second)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if data.SessionID != "abc123" {
		t.Errorf("unexpected session ID: %s", data.SessionID)
	}
	if calls < 2 {
		t.Errorf("expected at least 2 calls, got %d", calls)
	}
}
