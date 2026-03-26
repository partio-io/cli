package hooks

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCommitCacheContains(t *testing.T) {
	c := commitCache{
		Entries: []commitCacheEntry{
			{SHA: "abc123", ProcessedAt: time.Now()},
		},
	}

	if !c.contains("abc123") {
		t.Error("expected cache to contain abc123")
	}
	if c.contains("def456") {
		t.Error("expected cache not to contain def456")
	}
}

func TestCommitCacheAdd(t *testing.T) {
	c := commitCache{}
	c.add("abc123")

	if !c.contains("abc123") {
		t.Error("expected cache to contain abc123 after add")
	}
	if len(c.Entries) != 1 {
		t.Errorf("expected 1 entry, got %d", len(c.Entries))
	}
}

func TestCommitCachePruneByAge(t *testing.T) {
	old := time.Now().AddDate(0, 0, -(commitCacheMaxDays + 1))
	c := commitCache{
		Entries: []commitCacheEntry{
			{SHA: "old1", ProcessedAt: old},
			{SHA: "old2", ProcessedAt: old},
			{SHA: "new1", ProcessedAt: time.Now()},
		},
	}
	c.prune()

	if len(c.Entries) != 1 {
		t.Errorf("expected 1 entry after pruning old entries, got %d", len(c.Entries))
	}
	if c.Entries[0].SHA != "new1" {
		t.Errorf("expected new1 to remain, got %s", c.Entries[0].SHA)
	}
}

func TestCommitCachePruneByLength(t *testing.T) {
	c := commitCache{}
	for i := 0; i < commitCacheMaxLen+10; i++ {
		c.Entries = append(c.Entries, commitCacheEntry{
			SHA:         string(rune('a'+i%26)) + string(rune('0'+i%10)),
			ProcessedAt: time.Now(),
		})
	}
	c.prune()

	if len(c.Entries) > commitCacheMaxLen {
		t.Errorf("expected at most %d entries, got %d", commitCacheMaxLen, len(c.Entries))
	}
}

func TestLoadSaveCommitCache(t *testing.T) {
	dir := t.TempDir()

	// Load from non-existent file returns empty cache
	c := loadCommitCache(dir)
	if len(c.Entries) != 0 {
		t.Errorf("expected empty cache, got %d entries", len(c.Entries))
	}

	// Add and save
	c.add("sha1")
	c.add("sha2")
	if err := saveCommitCache(dir, c); err != nil {
		t.Fatalf("saveCommitCache: %v", err)
	}

	// Reload and verify
	c2 := loadCommitCache(dir)
	if !c2.contains("sha1") {
		t.Error("expected reloaded cache to contain sha1")
	}
	if !c2.contains("sha2") {
		t.Error("expected reloaded cache to contain sha2")
	}
}

func TestLoadCommitCacheCorruptFile(t *testing.T) {
	dir := t.TempDir()

	// Write corrupt JSON
	stateDir := filepath.Join(dir, "state")
	if err := os.MkdirAll(stateDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(commitCachePath(dir), []byte("not json"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	c := loadCommitCache(dir)
	if len(c.Entries) != 0 {
		t.Errorf("expected empty cache on corrupt file, got %d entries", len(c.Entries))
	}
}
