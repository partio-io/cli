package premise

import (
	"path/filepath"
	"strings"
	"testing"
)

// commentCommands maps each workflow that a comment command starts to the
// command it answers. Both run on the self-hosted runner, which holds GH_PAT.
var commentCommands = map[string]string{
	"../../.github/workflows/minion.yml":   "/minion build",
	"../../.github/workflows/research.yml": "/minion research",
}

// trustedCommenter is the author check of a comment trigger.
const trustedCommenter = `contains(fromJSON('["OWNER", "MEMBER", "COLLABORATOR"]'), github.event.comment.author_association)`

// TestCommentCommandsNeedALeadingCommandAndATrustedAuthor guards the comment
// triggers. The repository is public, so anyone can comment: without the
// author check, a stranger's issue and a "/minion build" comment ran a build
// on the runner. The proposal review posts hundreds of comments as the
// operator, and 86 open proposals end with "Comment `/minion build` ...", so a
// command quoted inside a comment must not start a run either.
func TestCommentCommandsNeedALeadingCommandAndATrustedAuthor(t *testing.T) {
	for path, command := range commentCommands {
		src := readRepoFile(t, path)
		for _, want := range []string{
			"startsWith(github.event.comment.body, '" + command + "')",
			trustedCommenter,
		} {
			if !strings.Contains(src, want) {
				t.Errorf("%s does not guard its comment trigger with %s", path, want)
			}
		}
	}

	paths, err := filepath.Glob("../../.github/workflows/*.yml")
	if err != nil || len(paths) == 0 {
		t.Fatalf("list the workflows: %v (found %d)", err, len(paths))
	}
	for _, path := range paths {
		if strings.Contains(readRepoFile(t, path), "contains(github.event.comment.body") {
			t.Errorf("%s starts a run on a command anywhere in a comment; match it with startsWith", path)
		}
	}
}
