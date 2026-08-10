package patchapply

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// git runs git in dir and returns its combined output either way, so a
// caller can report what git said without running the command again.
//
// The error names no argument. A fetch and a push carry the token in
// their argument list, and an error travels further than the output
// does.
func git(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	text := strings.TrimSpace(string(out))
	if err != nil {
		return text, fmt.Errorf("git: %w", err)
	}
	return text, nil
}

// worktree is a detached checkout of the pull request head.
type worktree struct {
	dir     string
	repoDir string
}

// detach fetches ref from remote and checks it out, detached, in its
// own worktree.
//
// The fetch is what makes the round operate on the pull request's real
// head. The job's own checkout is a merge commit or a stale tip, and a
// patch applied there pushes work the author never wrote.
func detach(repoDir, dir, remote, ref string) (*worktree, error) {
	if out, err := git(repoDir, "fetch", "--no-tags", remote, ref); err != nil {
		return nil, fmt.Errorf("fetch %s: %w: %s", ref, err, tail(out))
	}

	if dir == "" {
		parent, err := os.MkdirTemp("", "patchapply-")
		if err != nil {
			return nil, fmt.Errorf("create worktree directory: %w", err)
		}
		dir = filepath.Join(parent, "head")
	}

	// A worktree left behind by an earlier round blocks the add.
	w := &worktree{dir: dir, repoDir: repoDir}
	w.remove()

	if out, err := git(repoDir, "worktree", "add", "--detach", dir, "FETCH_HEAD"); err != nil {
		return nil, fmt.Errorf("detach %s: %w: %s", ref, err, tail(out))
	}
	return w, nil
}

// remove unregisters the worktree and deletes its directory.
//
// It is best effort by design. It runs before the add and again on the
// way out of every path, and the wanted state is "this directory is
// not a worktree" — which a failure to remove one that was never there
// already satisfies.
func (w *worktree) remove() {
	if _, err := git(w.repoDir, "worktree", "remove", "--force", w.dir); err != nil {
		slog.Debug("removing repair worktree", "dir", w.dir, "error", err)
	}
	if err := os.RemoveAll(w.dir); err != nil {
		slog.Debug("deleting repair worktree directory", "dir", w.dir, "error", err)
	}
}

// commit records the applied patch under the marker that names the
// check.
//
// The identity is explicit because a CI worktree inherits none, and
// signing is off because the runner holds no key.
func commit(dir, subject, body string) error {
	args := []string{
		"-c", "user.name=minion audit",
		"-c", "user.email=minions@partio.io",
		"-c", "commit.gpgsign=false",
		"commit", "-m", subject,
	}
	if body != "" {
		args = append(args, "-m", body)
	}
	if out, err := git(dir, args...); err != nil {
		return fmt.Errorf("commit the repair: %w: %s", err, tail(out))
	}
	return nil
}
