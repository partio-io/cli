package claude

import (
	"log/slog"
	"os/exec"
	"strings"

	"github.com/partio-io/cli/internal/agent"
	"github.com/partio-io/cli/internal/wsl"
)

// IsRunning checks if a Claude Code process is currently running.
// On WSL, it additionally checks for a Windows-native Claude Code process
// via tasklist.exe when pgrep finds nothing.
func (d *Detector) IsRunning() (bool, error) {
	out, err := exec.Command("pgrep", "-f", "claude").Output()
	if err != nil {
		// pgrep returns exit code 1 if no processes found
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			// Fall through to WSL check below.
		} else {
			return false, err
		}
	} else if strings.TrimSpace(string(out)) != "" {
		return true, nil
	}

	// pgrep found nothing; check Windows-native process if running under WSL.
	if wsl.IsWSL() {
		tlOut, tlErr := exec.Command("tasklist.exe", "/FO", "CSV", "/NH").Output()
		if tlErr != nil {
			slog.Warn("tasklist.exe failed; skipping Windows process check", "err", tlErr)
			return false, nil
		}
		return parseTasklistOutput(string(tlOut)), nil
	}

	return false, nil
}

// parseTasklistOutput parses the CSV output of `tasklist.exe /FO CSV /NH` and
// returns true if any row's image name contains "claude" (case-insensitive).
func parseTasklistOutput(csv string) bool {
	for _, line := range strings.Split(csv, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.SplitN(line, ",", 2)
		if len(fields) < 1 {
			continue
		}
		imageName := strings.Trim(fields[0], "\"")
		if strings.Contains(strings.ToLower(imageName), "claude") {
			return true
		}
	}
	return false
}

// AgentPID returns the PID of a running Claude Code process, or (0, false) when
// none is found.
func (d *Detector) AgentPID() (int, bool) {
	return agent.PgrepFirst("claude")
}
