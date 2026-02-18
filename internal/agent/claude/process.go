package claude

import (
	"os/exec"
	"strings"
)

// IsRunning checks if a Claude Code process is currently running.
func (d *Detector) IsRunning() (bool, error) {
	out, err := exec.Command("pgrep", "-f", "claude").Output()
	if err != nil {
		// pgrep returns exit code 1 if no processes found
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return false, nil
		}
		return false, err
	}
	return strings.TrimSpace(string(out)) != "", nil
}
