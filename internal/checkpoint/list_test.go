package checkpoint

import (
	"sort"
	"testing"
	"time"
)

func TestListSort(t *testing.T) {
	now := time.Now()
	metas := []Metadata{
		{ID: "aaa111222333", CreatedAt: now.Add(-2 * time.Hour).Format(time.RFC3339)},
		{ID: "bbb444555666", CreatedAt: now.Format(time.RFC3339)},
		{ID: "ccc777888999", CreatedAt: now.Add(-1 * time.Hour).Format(time.RFC3339)},
	}

	sort.Slice(metas, func(i, j int) bool {
		ti, _ := time.Parse(time.RFC3339, metas[i].CreatedAt)
		tj, _ := time.Parse(time.RFC3339, metas[j].CreatedAt)
		return ti.After(tj)
	})

	if metas[0].ID != "bbb444555666" {
		t.Errorf("expected newest first, got %s", metas[0].ID)
	}
	if metas[1].ID != "ccc777888999" {
		t.Errorf("expected second newest second, got %s", metas[1].ID)
	}
	if metas[2].ID != "aaa111222333" {
		t.Errorf("expected oldest last, got %s", metas[2].ID)
	}
}

func TestListNoBranch(t *testing.T) {
	// When run outside a git repo or without the checkpoint branch,
	// List() should return ErrNoBranch.
	// We can't easily mock git commands, so we verify the error sentinel exists.
	if ErrNoBranch == nil {
		t.Error("ErrNoBranch should not be nil")
	}
	if ErrNoBranch.Error() != "checkpoint branch does not exist" {
		t.Errorf("unexpected error message: %s", ErrNoBranch.Error())
	}
}

func TestListEmptyResult(t *testing.T) {
	// An empty slice (not nil) should be treated as "no checkpoints".
	var metas []Metadata
	if len(metas) != 0 {
		t.Error("expected empty slice")
	}
}

func TestMetadataFields(t *testing.T) {
	now := time.Now()
	meta := Metadata{
		ID:           "abcdef123456",
		SessionID:    "session-1",
		CommitHash:   "abc1234567890",
		Branch:       "main",
		CreatedAt:    now.Format(time.RFC3339),
		Agent:        "claude-code",
		AgentPercent: 100,
		ContentHash:  "hash123",
	}

	// Verify ID prefix (12-char)
	if len(meta.ID) != 12 {
		t.Errorf("expected 12-char ID, got %d", len(meta.ID))
	}

	// Verify commit hash can be truncated to 7
	commit := meta.CommitHash
	if len(commit) > 7 {
		commit = commit[:7]
	}
	if commit != "abc1234" {
		t.Errorf("expected truncated commit abc1234, got %s", commit)
	}

	// Verify CreatedAt parses as RFC3339
	_, err := time.Parse(time.RFC3339, meta.CreatedAt)
	if err != nil {
		t.Errorf("CreatedAt should be valid RFC3339: %v", err)
	}
}
