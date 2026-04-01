package codex

import (
	"os/exec"
	"strings"
)

// IsRunning checks if a Codex CLI process is currently running.
func (d *Detector) IsRunning() (bool, error) {
	out, err := exec.Command("pgrep", "-f", "codex").Output()
	return parseIsRunning(out, err)
}

// parseIsRunning interprets the output and error from pgrep.
func parseIsRunning(out []byte, err error) (bool, error) {
	if err != nil {
		// pgrep returns exit code 1 if no processes found — not an error.
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return false, nil
		}
		return false, err
	}
	return strings.TrimSpace(string(out)) != "", nil
}
