package patchapply

import (
	"fmt"
	"os/exec"
	"strings"
)

// defaultProve is how a repair is proved when a caller names no
// commands: the repository's own targets, in the order CI runs them.
// Lint comes first because it is the cheaper of the two.
func defaultProve() [][]string {
	return [][]string{
		{"make", "lint"},
		{"make", "test"},
	}
}

// prove runs each command in dir and stops at the first failure. It
// returns that command's output, so the caller can say why the patch
// was dropped rather than only that it was.
func prove(dir string, commands [][]string) (string, error) {
	if len(commands) == 0 {
		commands = defaultProve()
	}
	for _, args := range commands {
		if len(args) == 0 {
			continue
		}
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = dir
		out, err := cmd.CombinedOutput()
		if err != nil {
			return strings.TrimSpace(string(out)), fmt.Errorf("%s: %w", strings.Join(args, " "), err)
		}
	}
	return "", nil
}
