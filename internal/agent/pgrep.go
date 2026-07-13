package agent

import (
	"os/exec"
	"strconv"
	"strings"
)

// PgrepFirst runs `pgrep -f pattern` and returns the first matching PID and
// whether any process matched. Any error (including "no process found") yields
// (0, false).
func PgrepFirst(pattern string) (int, bool) {
	out, err := exec.Command("pgrep", "-f", pattern).Output()
	if err != nil {
		return 0, false
	}
	fields := strings.Fields(string(out))
	if len(fields) == 0 {
		return 0, false
	}
	pid, err := strconv.Atoi(fields[0])
	if err != nil {
		return 0, false
	}
	return pid, true
}
