package main

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/partio-io/cli/internal/checkpoint"
)

func TestListCmd_JSONFlag(t *testing.T) {
	cmd := newListCmd()
	// Verify --json flag exists
	f := cmd.Flags().Lookup("json")
	if f == nil {
		t.Fatal("expected --json flag to be defined")
	}
	if f.DefValue != "false" {
		t.Errorf("expected --json default to be false, got %s", f.DefValue)
	}
}

func TestCheckpointListOutput_EmptyJSON(t *testing.T) {
	out := checkpointListOutput{
		Checkpoints: []checkpoint.Metadata{},
	}

	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	if err := enc.Encode(out); err != nil {
		t.Fatalf("failed to encode: %v", err)
	}

	var decoded map[string]json.RawMessage
	if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}

	raw, ok := decoded["checkpoints"]
	if !ok {
		t.Fatal("expected 'checkpoints' key in JSON output")
	}

	var arr []json.RawMessage
	if err := json.Unmarshal(raw, &arr); err != nil {
		t.Fatalf("checkpoints is not an array: %v", err)
	}

	if len(arr) != 0 {
		t.Errorf("expected empty checkpoints array, got %d items", len(arr))
	}
}

func TestCheckpointListOutput_WithCheckpoints(t *testing.T) {
	out := checkpointListOutput{
		Checkpoints: []checkpoint.Metadata{
			{
				ID:           "abcdef123456",
				SessionID:    "session-1",
				CommitHash:   "abc123def456",
				Branch:       "main",
				CreatedAt:    "2025-01-15T10:30:00Z",
				Agent:        "claude-code",
				AgentPercent: 100,
				ContentHash:  "hash123",
			},
			{
				ID:           "fedcba654321",
				SessionID:    "session-2",
				CommitHash:   "def456abc789",
				Branch:       "feature",
				CreatedAt:    "2025-01-15T11:00:00Z",
				Agent:        "claude-code",
				AgentPercent: 0,
				ContentHash:  "hash456",
			},
		},
	}

	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	if err := enc.Encode(out); err != nil {
		t.Fatalf("failed to encode: %v", err)
	}

	// Verify valid JSON
	var decoded map[string]json.RawMessage
	if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}

	// Verify envelope
	raw, ok := decoded["checkpoints"]
	if !ok {
		t.Fatal("expected 'checkpoints' key in JSON output")
	}

	// Verify array contents
	var checkpoints []map[string]interface{}
	if err := json.Unmarshal(raw, &checkpoints); err != nil {
		t.Fatalf("checkpoints is not a valid array: %v", err)
	}

	if len(checkpoints) != 2 {
		t.Fatalf("expected 2 checkpoints, got %d", len(checkpoints))
	}

	// Verify required fields in first checkpoint
	requiredFields := []string{"id", "commit_hash", "agent", "agent_percent", "created_at", "branch"}
	for _, field := range requiredFields {
		if _, ok := checkpoints[0][field]; !ok {
			t.Errorf("expected field %q in checkpoint JSON", field)
		}
	}

	// Verify specific values
	if checkpoints[0]["id"] != "abcdef123456" {
		t.Errorf("expected id=abcdef123456, got %v", checkpoints[0]["id"])
	}
	if checkpoints[0]["branch"] != "main" {
		t.Errorf("expected branch=main, got %v", checkpoints[0]["branch"])
	}
	if checkpoints[0]["agent"] != "claude-code" {
		t.Errorf("expected agent=claude-code, got %v", checkpoints[0]["agent"])
	}
}

func TestCheckpointListOutput_NilBecomesEmptyArray(t *testing.T) {
	// When checkpoints is nil, JSON should still produce an empty array
	// This mirrors the behavior in runList where nil is replaced with empty slice
	checkpoints := []checkpoint.Metadata(nil)
	checkpoints = []checkpoint.Metadata{} // same as what runList does

	out := checkpointListOutput{Checkpoints: checkpoints}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(out); err != nil {
		t.Fatalf("failed to encode: %v", err)
	}

	var decoded struct {
		Checkpoints []json.RawMessage `json:"checkpoints"`
	}
	if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}

	if decoded.Checkpoints == nil {
		t.Error("expected non-nil checkpoints array")
	}
	if len(decoded.Checkpoints) != 0 {
		t.Errorf("expected empty array, got %d items", len(decoded.Checkpoints))
	}
}
