package agent

import "testing"

func TestPgrepFirst_NoMatch(t *testing.T) {
	// A pattern that no real process should match must yield (0, false).
	if pid, ok := PgrepFirst("zzzz-partio-no-such-process-zzzz"); ok || pid != 0 {
		t.Errorf("PgrepFirst(no match) = (%d, %v), want (0, false)", pid, ok)
	}
}
