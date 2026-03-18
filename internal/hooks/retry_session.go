package hooks

import (
	"log/slog"
	"time"

	"github.com/partio-io/cli/internal/agent"
	"github.com/partio-io/cli/internal/agent/claude"
)

const initialRetryBackoff = 100 * time.Millisecond

// sessionFinder is a function that attempts to find a session.
type sessionFinder func(repoRoot string) (string, *agent.SessionData, error)

// sessionDataReady returns true when the session data is considered available.
func sessionDataReady(data *agent.SessionData, err error) bool {
	return err == nil && data != nil && data.SessionID != ""
}

// findSessionWithRetry calls finder repeatedly with exponential backoff until
// session data is available or the timeout expires. If the timeout is <= 0 no
// retries are performed.
func findSessionWithRetry(finder sessionFinder, repoRoot string, timeout time.Duration) (string, *agent.SessionData, error) {
	path, data, err := finder(repoRoot)
	if sessionDataReady(data, err) || timeout <= 0 {
		return path, data, err
	}

	deadline := time.Now().Add(timeout)
	backoff := initialRetryBackoff

	for time.Now().Before(deadline) {
		sleep := backoff
		if remaining := time.Until(deadline); sleep > remaining {
			sleep = remaining
		}
		time.Sleep(sleep)
		backoff *= 2

		path, data, err = finder(repoRoot)
		if sessionDataReady(data, err) {
			return path, data, err
		}
	}

	slog.Warn("session data not available after retry window", "timeout_ms", timeout.Milliseconds())
	return path, data, err
}

// newSessionFinder wraps the claude detector's FindLatestSession.
func newSessionFinder() sessionFinder {
	d := claude.New()
	return d.FindLatestSession
}
