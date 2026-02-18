package checkpoint

import (
	"testing"
	"time"
)

func TestNewID(t *testing.T) {
	id := NewID()

	if len(id) != 12 {
		t.Errorf("expected 12-char ID, got %d chars: %s", len(id), id)
	}

	// Should be hex characters only
	for _, c := range id {
		if (c < '0' || c > '9') && (c < 'a' || c > 'f') {
			t.Errorf("unexpected character in ID: %c", c)
		}
	}

	// Should be unique
	id2 := NewID()
	if id == id2 {
		t.Error("two consecutive IDs should not be equal")
	}
}

func TestShard(t *testing.T) {
	tests := []struct {
		id    string
		shard string
		rest  string
	}{
		{"abcdef123456", "ab", "cdef123456"},
		{"00112233aabb", "00", "112233aabb"},
		{"a", "a", ""},
		{"ab", "ab", ""},
	}

	for _, tt := range tests {
		if got := Shard(tt.id); got != tt.shard {
			t.Errorf("Shard(%q) = %q, want %q", tt.id, got, tt.shard)
		}
		if got := Rest(tt.id); got != tt.rest {
			t.Errorf("Rest(%q) = %q, want %q", tt.id, got, tt.rest)
		}
	}
}

func TestToMetadata(t *testing.T) {
	now := time.Now()
	cp := &Checkpoint{
		ID:          "abcdef123456",
		SessionID:   "session-1",
		CommitHash:  "abc123",
		Branch:      "main",
		CreatedAt:   now,
		Agent:       "claude-code",
		AgentPct:    85,
		ContentHash: "def456",
	}

	meta := cp.ToMetadata()

	if meta.ID != cp.ID {
		t.Errorf("ID mismatch: %s vs %s", meta.ID, cp.ID)
	}
	if meta.AgentPercent != 85 {
		t.Errorf("expected agent_percent=85, got %d", meta.AgentPercent)
	}
	if meta.CreatedAt != now.Format(time.RFC3339) {
		t.Errorf("created_at format mismatch")
	}
}
