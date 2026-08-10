// Package patchapply turns a repair session's patch into one pushed
// commit, or into a stated reason why nothing was pushed.
//
// A repair session never commits and never pushes: it writes a patch
// and stops. Everything after that is deterministic, and it lives here
// so that every gate shares one implementation. Run fetches the pull
// request's real head, detaches a worktree onto it, applies the patch,
// proves the result with the repository's own lint and test targets,
// commits it with the marker internal/repairround counts, and pushes
// it with the personal access token.
//
// A patch that does not apply, or that leaves the checks red, is a
// spent round and not a job error. Run reports it and pushes nothing.
package patchapply

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/partio-io/cli/internal/repairround"
)

// Config carries everything one apply-prove-push run needs.
type Config struct {
	RepoDir     string // repository the worktree is added from; "" means the current directory
	WorktreeDir string // where the head is detached; "" means a new temporary directory
	PatchPath   string // patch the repair session exported
	Body        string // commit body, usually the session's fix summary

	Repo     string // owner/name the workflow runs in
	PRNumber int    // pull request the patch repairs
	Check    string // check that produced the patch, e.g. "dead-code"

	HeadRepo string // owner/name of the pull request head; empty means ask the API
	HeadRef  string // head branch name; empty means ask the API

	Token  string // personal access token; the push must use it
	Remote string // fetch and push target; "" means the GitHub URL for Repo

	Prove [][]string // commands that prove the applied patch; nil means make lint and make test

	APIBaseURL string       // GitHub API root, e.g. https://api.github.com
	HTTPClient *http.Client // nil means http.DefaultClient
}

// Result is what the round did. A false Pushed is an outcome and not a
// failure: Reason says why nothing was pushed.
type Result struct {
	Pushed bool
	Reason string
}

// Run applies the patch to the pull request's real head, proves it,
// and pushes one marked commit.
func Run(cfg Config) (Result, error) {
	res, err := run(cfg)

	// One choke point for the token: git echoes the remote URL when a
	// fetch or a push fails, and both the reason and the error reach
	// a job log.
	res.Reason = redact(res.Reason, cfg.Token)
	if err != nil {
		err = errors.New(redact(err.Error(), cfg.Token))
	}
	return res, err
}

func run(cfg Config) (Result, error) {
	if cfg.Repo == "" || cfg.Check == "" || cfg.PRNumber == 0 {
		return Result{}, errors.New("repository, check name and pull request number are required")
	}

	head, err := resolveHead(cfg)
	if err != nil {
		return Result{}, err
	}

	// The fork is refused before the fetch. The workflow holds no
	// write access to a fork, so a round started there can only fail
	// after doing all of the work.
	if head.repo != cfg.Repo {
		return Result{Reason: fmt.Sprintf(
			"the pull request head lives in %s and not in %s; a repair round cannot push to a fork",
			head.repo, cfg.Repo)}, nil
	}

	patch, err := patchFile(cfg.PatchPath)
	if err != nil {
		return Result{}, err
	}
	if patch == "" {
		return Result{Reason: "the repair session exported no patch"}, nil
	}

	remote := pushRemote(cfg)

	wt, err := detach(cfg.RepoDir, cfg.WorktreeDir, remote, head.ref)
	if err != nil {
		return Result{}, err
	}
	defer wt.remove()

	if out, err := git(wt.dir, "apply", "--index", patch); err != nil {
		return Result{Reason: "the patch did not apply to the pull request head: " + tail(out)}, nil
	}

	if out, err := prove(wt.dir, cfg.Prove); err != nil {
		return Result{Reason: fmt.Sprintf("the applied patch left the checks red (%v): %s", err, tail(out))}, nil
	}

	subject := repairround.Subject(cfg.Check, cfg.PRNumber)
	if err := commit(wt.dir, subject, cfg.Body); err != nil {
		return Result{}, err
	}

	if out, err := git(wt.dir, "push", remote, "HEAD:"+head.ref); err != nil {
		return Result{}, fmt.Errorf("push to %s: %w: %s", head.ref, err, tail(out))
	}

	return Result{
		Pushed: true,
		Reason: fmt.Sprintf("pushed one %s repair commit to %s", cfg.Check, head.ref),
	}, nil
}

// patchFile returns the absolute path of a patch that has content, or
// "" when the repair session exported nothing.
//
// The path must be absolute. Git runs with the worktree as its working
// directory, and would resolve a relative path against that instead of
// against the job's workspace.
func patchFile(path string) (string, error) {
	if path == "" {
		return "", nil
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("resolve patch path: %w", err)
	}
	info, err := os.Stat(abs)
	if errors.Is(err, os.ErrNotExist) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("read patch: %w", err)
	}
	if info.Size() == 0 {
		return "", nil
	}
	return abs, nil
}

// maxReportLines bounds what a reason carries out of a failed round. A
// full `make test` log buries the reason it is attached to.
const maxReportLines = 20

// tail returns the last lines of out, which is where a failing tool
// says what went wrong.
func tail(out string) string {
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) <= maxReportLines {
		return strings.Join(lines, "\n")
	}
	return "...\n" + strings.Join(lines[len(lines)-maxReportLines:], "\n")
}
