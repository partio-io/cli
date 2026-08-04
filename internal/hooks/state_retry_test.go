package hooks

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestReadStateWithRetryReturnsDataOnceFileAppears(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "pre-commit.json")

	go func() {
		time.Sleep(120 * time.Millisecond)
		if err := os.WriteFile(path, []byte(`{"agent_active":true}`), 0o644); err != nil {
			t.Error(err)
		}
	}()

	data, err := readStateWithRetry(path, 10, 50*time.Millisecond)
	if err != nil {
		t.Fatalf("expected retry to pick up the late state file, got %v", err)
	}
	if string(data) != `{"agent_active":true}` {
		t.Fatalf("unexpected state contents: %s", data)
	}
}

func TestReadStateWithRetryGivesUpAfterAttempts(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "pre-commit.json")

	start := time.Now()
	_, err := readStateWithRetry(path, 3, 20*time.Millisecond)
	if err == nil {
		t.Fatal("expected an error when the state file never appears")
	}
	if elapsed := time.Since(start); elapsed > 2*time.Second {
		t.Fatalf("retry took too long: %v", elapsed)
	}
}

func TestReadStateWithRetryPropagatesNonNotExistErrors(t *testing.T) {
	dir := t.TempDir()
	// A directory at the path makes ReadFile fail with a non-NotExist
	// error, which must not be retried.
	path := filepath.Join(dir, "state-as-dir")
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatal(err)
	}

	start := time.Now()
	_, err := readStateWithRetry(path, 10, 200*time.Millisecond)
	if err == nil {
		t.Fatal("expected an error reading a directory")
	}
	if elapsed := time.Since(start); elapsed > time.Second {
		t.Fatalf("non-NotExist error should fail fast, took %v", elapsed)
	}
}
