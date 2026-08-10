package patchapply

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/partio-io/cli/internal/repairround"
)

// headRef is the pull request head branch every fixture pushes to.
const headRef = "feature"

// gitRun runs git in dir and returns its trimmed output. It fails the
// test on a non-zero exit, because every call here is fixture setup or
// an assertion, and neither has a meaningful failure mode.
func gitRun(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = gitEnv()
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v in %s: %v\n%s", args, dir, err, out)
	}
	return strings.TrimSpace(string(out))
}

// gitOutput is gitRun for output that must stay byte-exact. A patch
// loses its trailing newline under TrimSpace, and git then rejects it.
func gitOutput(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = gitEnv()
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("git %v in %s: %v", args, dir, err)
	}
	return string(out)
}

// gitEnv pins the identity and drops the developer's own git config,
// so a global setting cannot change what these tests observe.
func gitEnv() []string {
	return append(os.Environ(),
		"GIT_AUTHOR_NAME=Test",
		"GIT_AUTHOR_EMAIL=test@example.com",
		"GIT_COMMITTER_NAME=Test",
		"GIT_COMMITTER_EMAIL=test@example.com",
		"GIT_CONFIG_GLOBAL=/dev/null",
		"GIT_CONFIG_SYSTEM=/dev/null",
	)
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

// makefile returns a Makefile with the two targets the default prove
// step runs. It is real, so the tests exercise the real default rather
// than a stubbed one.
func makefile(lintFails bool) string {
	lint := "\t@true\n"
	if lintFails {
		lint = "\t@echo 'lint found problems' >&2; exit 1\n"
	}
	return ".PHONY: lint test\n\nlint:\n" + lint + "\ntest:\n\t@true\n"
}

// fixture is a self-contained stand-in for the pull request the gate
// repairs. Nothing in it touches the network.
type fixture struct {
	// origin is a bare repository standing in for GitHub.
	origin string
	// seed is the clone that authors the branch and the patch.
	seed string
	// runner is the checkout the package operates from. It is left
	// deliberately behind origin, so a run that reads the checkout
	// instead of fetching the real head is visible in the result.
	runner string
}

func newFixture(t *testing.T, lintFails bool) *fixture {
	t.Helper()
	root := t.TempDir()
	f := &fixture{
		origin: filepath.Join(root, "origin.git"),
		seed:   filepath.Join(root, "seed"),
		runner: filepath.Join(root, "runner"),
	}

	gitRun(t, root, "init", "--bare", f.origin)
	gitRun(t, root, "init", f.seed)
	writeFile(t, filepath.Join(f.seed, "Makefile"), makefile(lintFails))
	writeFile(t, filepath.Join(f.seed, "notes.txt"), "one\n")
	gitRun(t, f.seed, "add", "-A")
	gitRun(t, f.seed, "commit", "-m", "initial")
	gitRun(t, f.seed, "branch", "-M", headRef)
	gitRun(t, f.seed, "push", f.origin, headRef)
	gitRun(t, root, "clone", "--branch", headRef, f.origin, f.runner)

	return f
}

// advanceHead pushes a further commit to the branch, so the pull
// request head is ahead of the runner checkout.
func (f *fixture) advanceHead(t *testing.T, content string) {
	t.Helper()
	writeFile(t, filepath.Join(f.seed, "notes.txt"), content)
	gitRun(t, f.seed, "add", "-A")
	gitRun(t, f.seed, "commit", "-m", "real head")
	gitRun(t, f.seed, "push", f.origin, headRef)
}

// makePatch exports the same patch a repair session exports, then
// restores the seed clone.
func (f *fixture) makePatch(t *testing.T, name, content string) string {
	t.Helper()
	writeFile(t, filepath.Join(f.seed, name), content)
	gitRun(t, f.seed, "add", "-A")
	patch := gitOutput(t, f.seed, "diff", "--cached", "--binary")
	gitRun(t, f.seed, "reset", "--hard", "HEAD")

	path := filepath.Join(t.TempDir(), "fix.patch")
	writeFile(t, path, patch)
	return path
}

func (f *fixture) headSHA(t *testing.T) string {
	t.Helper()
	return gitRun(t, f.origin, "rev-parse", headRef)
}

func (f *fixture) headSubject(t *testing.T) string {
	t.Helper()
	return gitRun(t, f.origin, "log", "-1", "--format=%s", headRef)
}

func (f *fixture) fileAtHead(t *testing.T, name string) string {
	t.Helper()
	return gitRun(t, f.origin, "show", headRef+":"+name)
}

// config returns the configuration every test starts from: a local
// remote, an explicit head, and the dead-code check.
func (f *fixture) config(patch string) Config {
	return Config{
		RepoDir:   f.runner,
		PatchPath: patch,
		Body:      "removed the unused helper",
		Repo:      "partio-io/cli",
		HeadRepo:  "partio-io/cli",
		HeadRef:   headRef,
		PRNumber:  622,
		Check:     "dead-code",
		Remote:    f.origin,
	}
}

// TestRunAppliesProvesAndPushes is the end-to-end path: the package
// fetches the real head, applies the patch, proves it with the
// repository's own targets, and pushes one marked commit.
func TestRunAppliesProvesAndPushes(t *testing.T) {
	f := newFixture(t, false)
	f.advanceHead(t, "one\ntwo\n")
	patch := f.makePatch(t, "notes.txt", "one\ntwo\nthree\n")
	realHead := f.headSHA(t)

	res, err := Run(f.config(patch))
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !res.Pushed {
		t.Fatalf("nothing pushed: %s", res.Reason)
	}

	if got, want := f.headSubject(t), "audit(dead-code): fix findings (#622)"; got != want {
		t.Errorf("commit subject = %q, want %q", got, want)
	}

	// The subject is only useful if the package that counts rounds
	// reads the check back out of it. Asserting the literal alone
	// would survive a marker change that stops the counting.
	if check, ok := repairround.CheckOf(f.headSubject(t)); !ok || check != "dead-code" {
		t.Errorf("repairround.CheckOf(%q) = %q, %t; want \"dead-code\", true",
			f.headSubject(t), check, ok)
	}
	if got, want := f.fileAtHead(t, "notes.txt"), "one\ntwo\nthree"; got != want {
		t.Errorf("notes.txt at head = %q, want %q", got, want)
	}
	if got := gitRun(t, f.origin, "rev-parse", headRef+"^"); got != realHead {
		t.Errorf("parent = %s, want the real head %s", got, realHead)
	}
}

// TestRunReportsAPatchThatDoesNotApply covers the round that races the
// author: the patch was built against a tree the head no longer has.
// Nothing is pushed, and the round says so instead of failing the job.
func TestRunReportsAPatchThatDoesNotApply(t *testing.T) {
	f := newFixture(t, false)
	patch := f.makePatch(t, "notes.txt", "one\ntwo\n")
	f.advanceHead(t, "something else entirely\n")
	before := f.headSHA(t)

	res, err := Run(f.config(patch))
	if err != nil {
		t.Fatalf("Run failed the job, want a reported round: %v", err)
	}
	if res.Pushed {
		t.Fatal("pushed a patch that does not apply")
	}
	if !strings.Contains(res.Reason, "did not apply") {
		t.Errorf("reason = %q, want it to say the patch did not apply", res.Reason)
	}
	if got := f.headSHA(t); got != before {
		t.Errorf("head moved to %s, want it left at %s", got, before)
	}
}

// TestRunPushesNothingWhenTheChecksFail covers the repair that applies
// cleanly and is still wrong. The default prove step is the
// repository's own `make lint`, and this fixture's lint target fails.
func TestRunPushesNothingWhenTheChecksFail(t *testing.T) {
	f := newFixture(t, true)
	f.advanceHead(t, "one\ntwo\n")
	patch := f.makePatch(t, "notes.txt", "one\ntwo\nthree\n")
	before := f.headSHA(t)

	res, err := Run(f.config(patch))
	if err != nil {
		t.Fatalf("Run failed the job, want a reported round: %v", err)
	}
	if res.Pushed {
		t.Fatal("pushed a patch that leaves the checks red")
	}
	if !strings.Contains(res.Reason, "checks red") || !strings.Contains(res.Reason, "make lint") {
		t.Errorf("reason = %q, want it to name the red check", res.Reason)
	}
	if got := f.headSHA(t); got != before {
		t.Errorf("head moved to %s, want it left at %s", got, before)
	}
}

// TestRunRefusesAForkHeadBeforeFetching proves the order, not only the
// refusal: the remote is a path that does not exist, so a run that
// fetched first would fail instead of reporting the fork.
func TestRunRefusesAForkHeadBeforeFetching(t *testing.T) {
	f := newFixture(t, false)
	patch := f.makePatch(t, "notes.txt", "one\ntwo\n")

	cfg := f.config(patch)
	cfg.HeadRepo = "someone-else/cli"
	cfg.Remote = filepath.Join(t.TempDir(), "no-such-remote.git")

	res, err := Run(cfg)
	if err != nil {
		t.Fatalf("Run failed the job, want a reported round: %v", err)
	}
	if res.Pushed {
		t.Fatal("pushed to a fork head")
	}
	if !strings.Contains(res.Reason, "fork") || !strings.Contains(res.Reason, "someone-else/cli") {
		t.Errorf("reason = %q, want it to name the fork", res.Reason)
	}
	if _, err := os.Stat(filepath.Join(f.runner, ".git", "FETCH_HEAD")); !errors.Is(err, os.ErrNotExist) {
		t.Error("the run fetched before it refused the fork head")
	}
}

// TestRunReadsTheHeadFromTheAPI covers the production path, where the
// workflow supplies a pull request number and nothing else. The head
// branch and the head repository come from GitHub, and the fork
// refusal is made on what GitHub reports.
func TestRunReadsTheHeadFromTheAPI(t *testing.T) {
	tests := []struct {
		name     string
		fullName string
		wantPush bool
		wantIn   string
	}{
		{name: "same repository", fullName: "partio-io/cli", wantPush: true},
		{name: "fork", fullName: "someone-else/cli", wantIn: "fork"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newFixture(t, false)
			f.advanceHead(t, "one\ntwo\n")
			patch := f.makePatch(t, "notes.txt", "one\ntwo\nthree\n")

			// Only the one path answers. A run that asks for
			// anything else gets a 404 and fails the test.
			mux := http.NewServeMux()
			mux.HandleFunc("/repos/partio-io/cli/pulls/622", func(w http.ResponseWriter, _ *http.Request) {
				body := fmt.Sprintf(`{"head":{"ref":%q,"repo":{"full_name":%q}}}`, headRef, tt.fullName)
				if _, err := w.Write([]byte(body)); err != nil {
					t.Errorf("write pull request response: %v", err)
				}
			})
			srv := httptest.NewServer(mux)
			defer srv.Close()

			cfg := f.config(patch)
			cfg.HeadRepo = ""
			cfg.HeadRef = ""
			cfg.APIBaseURL = srv.URL
			cfg.HTTPClient = srv.Client()

			res, err := Run(cfg)
			if err != nil {
				t.Fatalf("Run: %v", err)
			}
			if res.Pushed != tt.wantPush {
				t.Fatalf("pushed = %t, want %t (%s)", res.Pushed, tt.wantPush, res.Reason)
			}
			if tt.wantIn != "" && !strings.Contains(res.Reason, tt.wantIn) {
				t.Errorf("reason = %q, want it to mention %q", res.Reason, tt.wantIn)
			}
		})
	}
}
