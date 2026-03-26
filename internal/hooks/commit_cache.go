package hooks

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

const (
	commitCacheFile    = "processed-commits.json"
	commitCacheMaxDays = 30
	commitCacheMaxLen  = 100
)

type commitCacheEntry struct {
	SHA         string    `json:"sha"`
	ProcessedAt time.Time `json:"processed_at"`
}

type commitCache struct {
	Entries []commitCacheEntry `json:"entries"`
}

func commitCachePath(partioDir string) string {
	return filepath.Join(partioDir, "state", commitCacheFile)
}

func loadCommitCache(partioDir string) commitCache {
	data, err := os.ReadFile(commitCachePath(partioDir))
	if err != nil {
		return commitCache{}
	}
	var cache commitCache
	if err := json.Unmarshal(data, &cache); err != nil {
		return commitCache{}
	}
	return cache
}

func (c *commitCache) contains(sha string) bool {
	for _, e := range c.Entries {
		if e.SHA == sha {
			return true
		}
	}
	return false
}

func (c *commitCache) add(sha string) {
	c.Entries = append(c.Entries, commitCacheEntry{
		SHA:         sha,
		ProcessedAt: time.Now(),
	})
	c.prune()
}

func (c *commitCache) prune() {
	cutoff := time.Now().AddDate(0, 0, -commitCacheMaxDays)
	pruned := c.Entries[:0]
	for _, e := range c.Entries {
		if e.ProcessedAt.After(cutoff) {
			pruned = append(pruned, e)
		}
	}
	c.Entries = pruned

	if len(c.Entries) > commitCacheMaxLen {
		c.Entries = c.Entries[len(c.Entries)-commitCacheMaxLen:]
	}
}

func saveCommitCache(partioDir string, cache commitCache) error {
	stateDir := filepath.Join(partioDir, "state")
	if err := os.MkdirAll(stateDir, 0o755); err != nil {
		return err
	}
	data, err := json.Marshal(cache)
	if err != nil {
		return err
	}
	return os.WriteFile(commitCachePath(partioDir), data, 0o644)
}
