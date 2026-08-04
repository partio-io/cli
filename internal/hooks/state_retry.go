package hooks

import (
	"errors"
	"io/fs"
	"os"
	"time"
)

// readStateWithRetry reads the pre-commit state file, retrying while
// it does not exist yet. Pre-commit writes the state and post-commit
// consumes it; on a fast follow-up commit the second post-commit can
// fire before the second pre-commit's write is visible, so a bounded
// retry bridges the handoff instead of silently dropping the
// checkpoint.
func readStateWithRetry(path string, attempts int, delay time.Duration) ([]byte, error) {
	var lastErr error
	for range attempts {
		data, err := os.ReadFile(path)
		if err == nil {
			return data, nil
		}
		if !errors.Is(err, fs.ErrNotExist) {
			return nil, err
		}
		lastErr = err
		time.Sleep(delay)
	}
	return nil, lastErr
}
