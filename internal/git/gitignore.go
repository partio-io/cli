package git

import (
	"errors"
	"os/exec"
	"strings"
)

// FilterGitIgnored returns the subset of paths that are NOT gitignored.
// It uses git check-ignore which respects repo .gitignore files, nested
// .gitignore files, and the global gitignore (core.excludesFile).
// repoRoot must be the top-level directory of the git repository.
func FilterGitIgnored(repoRoot string, paths []string) ([]string, error) {
	if len(paths) == 0 {
		return paths, nil
	}

	cmd := exec.Command("git", "check-ignore", "--stdin")
	cmd.Dir = repoRoot
	cmd.Stdin = strings.NewReader(strings.Join(paths, "\n") + "\n")
	out, err := cmd.Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
			// Exit code 1 means no paths are ignored — return all paths unchanged.
			return paths, nil
		}
		return nil, err
	}

	ignored := make(map[string]bool)
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if line != "" {
			ignored[line] = true
		}
	}

	result := make([]string, 0, len(paths))
	for _, p := range paths {
		if !ignored[p] {
			result = append(result, p)
		}
	}
	return result, nil
}
