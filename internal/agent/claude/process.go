package claude

import (
	"log/slog"
	"os/exec"
	"strings"

	"github.com/partio-io/cli/internal/agent"
	"github.com/partio-io/cli/internal/wsl"
)

// IsRunning checks if a Claude Code process is currently running.
// On WSL, it also checks for a Windows-native Claude Code process via tasklist.exe.
func (d *Detector) IsRunning() (bool, error) {
	out, err := exec.Command("pgrep", "-f", "claude").Output()
	if err != nil {
		// pgrep returns exit code 1 if no processes found
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			// pgrep found nothing; fall through to Windows check if in WSL.
		} else {
			return false, err
		}
	} else if strings.TrimSpace(string(out)) != "" {
		return true, nil
	}

	// In WSL, supplement with a Windows-side process query.
	if wsl.IsWSL() {
		csvOut, taskErr := exec.Command("tasklist.exe", "/FO", "CSV", "/NH").Output()
		if taskErr != nil {
			slog.Warn("tasklist.exe failed during Claude process check", "err", taskErr)
		} else if parseTasklistOutput(string(csvOut)) {
			return true, nil
		}
	}

	return false, nil
}

// parseTasklistOutput returns true if the tasklist CSV output contains a row
// whose image-name field (field 0) contains "claude" (case-insensitive).
func parseTasklistOutput(csv string) bool {
	for _, line := range strings.Split(csv, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// Each CSV line looks like: "image.exe","PID","Session","#","Mem Usage"
		// Split on comma and examine the first field.
		fields := strings.Split(line, ",")
		imageName := strings.Trim(fields[0], `"`)
		if strings.Contains(strings.ToLower(imageName), "claude") {
			return true
		}
	}
	return false
}

// AgentPID returns the PID of a running Claude Code process, or (0, false) when
// none is found. When only a Windows-native Claude Code process is running
// (detected via tasklist.exe), returns (0, false) because PIDs do not cross the
// WSL process-namespace boundary.
func (d *Detector) AgentPID() (int, bool) {
	return agent.PgrepFirst("claude")
}
