package codex

import (
	"os/exec"
	"strings"
)

// execCommand is the function used to run pgrep. It can be overridden in tests.
var execCommand = func(name string, args ...string) ([]byte, error) {
	return exec.Command(name, args...).Output()
}

// IsRunning checks if a Codex CLI process is currently running.
func (d *Detector) IsRunning() (bool, error) {
	out, err := execCommand("pgrep", "-f", "codex")
	if err != nil {
		// pgrep returns exit code 1 if no processes found
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return false, nil
		}
		return false, err
	}
	return strings.TrimSpace(string(out)) != "", nil
}
