package hooks

import (
	"bufio"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/partio-io/cli/internal/config"
)

// shouldLinkCommit determines whether the current commit should be linked to
// the active agent session. It returns true if the commit should be linked.
//
// When commit_linking is "ask", it prompts the user interactively via /dev/tty.
// The prompt offers [Y/n/a]: Y (default) links, n skips, a links and persists
// "always" to settings so future commits are linked without prompting.
func shouldLinkCommit(repoRoot string, cfg config.Config) bool {
	switch cfg.CommitLinking {
	case config.CommitLinkingAlways:
		slog.Debug("commit linking: always (auto-linking)")
		return true
	case config.CommitLinkingNever:
		slog.Debug("commit linking: never (skipping)")
		return false
	default: // "ask" or unset
		return promptCommitLinking(repoRoot)
	}
}

// promptCommitLinking opens /dev/tty to ask the user whether to link this
// commit. If /dev/tty is not available (non-interactive), it defaults to
// linking the commit.
func promptCommitLinking(repoRoot string) bool {
	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		slog.Debug("commit linking: no TTY available, auto-linking")
		return true
	}
	defer func() { _ = tty.Close() }()

	_, _ = fmt.Fprint(tty, "partio: Link this commit to the active AI session? [Y/n/a] ")

	reader := bufio.NewReader(tty)
	line, _ := reader.ReadString('\n')
	answer := strings.TrimSpace(strings.ToLower(line))

	switch answer {
	case "n":
		slog.Debug("commit linking: user chose not to link")
		return false
	case "a":
		slog.Debug("commit linking: user chose always")
		if err := config.SaveRepoSetting(repoRoot, "commit_linking", config.CommitLinkingAlways); err != nil {
			slog.Warn("could not persist commit_linking=always", "error", err)
		}
		return true
	default: // "", "y", "Y"
		return true
	}
}
