package claude

import (
	"log/slog"
	"os/exec"
	"strings"

	"github.com/partio-io/cli/internal/agent"
	"github.com/partio-io/cli/internal/wsl"
)

// IsRunning checks if a Claude Code process is currently running.
// On WSL with Windows interop enabled, it additionally queries the Windows
// process list via tasklist.exe when pgrep finds no Linux-side process.
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

	// No Linux-side process found. Try Windows process list when in WSL.
	if !wsl.IsWSL() {
		return false, nil
	}

	csvOut, err := exec.Command("tasklist.exe", "/FO", "CSV", "/NH").Output()
	if err != nil {
		slog.Warn("tasklist.exe failed; skipping Windows process check", "err", err)
		return false, nil
	}
	return parseTasklistOutput(string(csvOut)), nil
}

// parseTasklistOutput reports whether any row in tasklist.exe CSV output has
// an image name containing "claude" (case-insensitive). The CSV format is:
//
//	"Image Name","PID","Session Name","Session#","Mem Usage"
func parseTasklistOutput(csv string) bool {
	for _, line := range strings.Split(csv, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Split(line, ",")
		imageName := strings.Trim(fields[0], `"`)
		if strings.Contains(strings.ToLower(imageName), "claude") {
			return true
		}
	}
	return false
}

// AgentPID returns the PID of a running Claude Code process, or (0, false) when
// none is found. In WSL with a Windows-native Claude Code, pgrep finds no Linux
// process and (0, false) is returned; IsRunning handles the Windows-side check.
func (d *Detector) AgentPID() (int, bool) {
	return agent.PgrepFirst("claude")
}
