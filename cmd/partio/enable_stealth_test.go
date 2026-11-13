package main

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/partio-io/cli/internal/checkpoint"
	"github.com/partio-io/cli/internal/config"
)

// newEnableTestRepo makes a git repository with one commit in a temporary
// directory, points HOME at a separate temporary directory so the global
// settings file is isolated, changes into the repository and returns its root
// and a command runner.
func newEnableTestRepo(t *testing.T) (dir, home string, run func(args ...string) string) {
	t.Helper()

	home = t.TempDir()
	t.Setenv("HOME", home)

	dir = t.TempDir()
	run = func(args ...string) string {
		t.Helper()
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = dir
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("command %v failed: %v\n%s", args, err, out)
		}
		return strings.TrimSpace(string(out))
	}

	run("git", "init")
	run("git", "config", "user.email", "test@example.com")
	run("git", "config", "user.name", "Test")
	run("git", "commit", "--allow-empty", "-m", "initial")

	t.Chdir(dir)
	return dir, home, run
}

// runEnableWith runs `partio enable` with the given flags and returns its
// stdout.
func runEnableWith(t *testing.T, flags map[string]string) (string, error) {
	t.Helper()

	cmd := newEnableCmd()
	for name, value := range flags {
		if err := cmd.Flags().Set(name, value); err != nil {
			t.Fatalf("setting --%s: %v", name, err)
		}
	}
	var runErr error
	out := captureOutput(t, func() {
		runErr = runEnable(cmd, nil)
	})
	return out, runErr
}

func globalSettingsFile(home string) string {
	return filepath.Join(home, ".config", "partio", "settings.json")
}

func readJSONObject(t *testing.T, path string) map[string]json.RawMessage {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading %s: %v", path, err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("parsing %s: %v", path, err)
	}
	return raw
}

func TestEnableStealth(t *testing.T) {
	tests := []struct {
		name           string
		existingGlobal string // empty means no global file
		wantKeptTop    map[string]string
		wantKeptOpts   map[string]string
	}{
		{
			name: "creates global settings on a fresh machine",
		},
		{
			name:           "merges into existing global settings",
			existingGlobal: `{"log_level": "debug", "strategy_options": {"push_sessions": true, "custom": 7}}`,
			wantKeptTop:    map[string]string{"log_level": `"debug"`},
			wantKeptOpts:   map[string]string{"custom": `7`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir, home, run := newEnableTestRepo(t)
			globalPath := globalSettingsFile(home)

			if tt.existingGlobal != "" {
				if err := os.MkdirAll(filepath.Dir(globalPath), 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(globalPath, []byte(tt.existingGlobal), 0o644); err != nil {
					t.Fatal(err)
				}
			}

			out, err := runEnableWith(t, map[string]string{"stealth": "true"})
			if err != nil {
				t.Fatalf("runEnable(--stealth) error = %v", err)
			}
			for _, want := range []string{"Stealth mode active", globalPath, "partio rewind --list", "partio explain <id>"} {
				if !strings.Contains(out, want) {
					t.Errorf("output missing %q; got:\n%s", want, out)
				}
			}

			raw := readJSONObject(t, globalPath)
			var opts map[string]json.RawMessage
			if err := json.Unmarshal(raw["strategy_options"], &opts); err != nil {
				t.Fatalf("parsing strategy_options: %v", err)
			}
			if got := string(opts["push_sessions"]); got != "false" {
				t.Errorf("push_sessions = %s, want false", got)
			}
			if got := string(opts["disable_trailers"]); got != "true" {
				t.Errorf("disable_trailers = %s, want true", got)
			}
			for key, want := range tt.wantKeptTop {
				if got := string(raw[key]); got != want {
					t.Errorf("top-level %s = %s, want %s", key, got, want)
				}
			}
			for key, want := range tt.wantKeptOpts {
				if got := string(opts[key]); got != want {
					t.Errorf("strategy_options.%s = %s, want %s", key, got, want)
				}
			}

			firstGlobal, err := os.ReadFile(globalPath)
			if err != nil {
				t.Fatal(err)
			}

			// A second run changes nothing.
			if _, err := runEnableWith(t, map[string]string{"stealth": "true"}); err != nil {
				t.Fatalf("second runEnable(--stealth) error = %v", err)
			}
			secondGlobal, err := os.ReadFile(globalPath)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(firstGlobal, secondGlobal) {
				t.Errorf("second run changed the global file:\nfirst:  %s\nsecond: %s", firstGlobal, secondGlobal)
			}

			if _, err := os.Stat(filepath.Join(dir, config.PartioDir, "settings.json")); !os.IsNotExist(err) {
				t.Errorf(".partio/settings.json exists or is unreadable (err = %v), want it absent", err)
			}
			if _, err := os.Stat(filepath.Join(dir, ".gitignore")); !os.IsNotExist(err) {
				t.Errorf(".gitignore exists or is unreadable (err = %v), want it absent", err)
			}

			excludePath := run("git", "rev-parse", "--git-path", "info/exclude")
			if !filepath.IsAbs(excludePath) {
				excludePath = filepath.Join(dir, excludePath)
			}
			exclude, err := os.ReadFile(excludePath)
			if err != nil {
				t.Fatalf("reading %s: %v", excludePath, err)
			}
			if n := strings.Count("\n"+string(exclude), "\n.partio/\n"); n != 1 {
				t.Errorf("info/exclude holds %d .partio/ lines, want 1:\n%s", n, exclude)
			}

			// Make a runtime file Partio writes so the exclude is exercised.
			if err := os.MkdirAll(filepath.Join(dir, config.PartioDir, "state"), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(dir, config.PartioDir, "state", "x.json"), []byte("{}"), 0o644); err != nil {
				t.Fatal(err)
			}
			if status := run("git", "status", "--porcelain", "--untracked-files=all"); status != "" {
				t.Errorf("git status --porcelain = %q, want nothing", status)
			}

			for _, hook := range []string{"pre-commit", "post-commit", "pre-push"} {
				if _, err := os.Stat(filepath.Join(dir, ".git", "hooks", hook)); err != nil {
					t.Errorf("hook %s not installed: %v", hook, err)
				}
			}
			run("git", "rev-parse", "--verify", "refs/heads/partio/checkpoints/v1")
		})
	}
}

func TestEnableStealthMalformedGlobal(t *testing.T) {
	dir, home, _ := newEnableTestRepo(t)
	globalPath := globalSettingsFile(home)

	malformed := `{"strategy_options": {` // truncated JSON
	if err := os.MkdirAll(filepath.Dir(globalPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(globalPath, []byte(malformed), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := runEnableWith(t, map[string]string{"stealth": "true"}); err == nil {
		t.Fatal("runEnable(--stealth) error = nil, want an error for a malformed global file")
	}

	data, err := os.ReadFile(globalPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != malformed {
		t.Errorf("malformed global file was rewritten: %q", data)
	}
	if _, err := os.Stat(filepath.Join(dir, config.PartioDir)); !os.IsNotExist(err) {
		t.Errorf(".partio/ exists or is unreadable (err = %v), want enable to stop before creating it", err)
	}
}

func TestEnableStealthExcludeWriteFailure(t *testing.T) {
	dir, _, _ := newEnableTestRepo(t)

	// A directory where the exclude file should be makes the write fail.
	excludePath := filepath.Join(dir, ".git", "info", "exclude")
	if err := os.RemoveAll(excludePath); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(excludePath, 0o755); err != nil {
		t.Fatal(err)
	}

	logs := &bytes.Buffer{}
	previous := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(logs, &slog.HandlerOptions{Level: slog.LevelWarn})))
	t.Cleanup(func() { slog.SetDefault(previous) })

	if _, err := runEnableWith(t, map[string]string{"stealth": "true"}); err != nil {
		t.Fatalf("runEnable(--stealth) error = %v, want success despite the exclude failure", err)
	}
	if !strings.Contains(logs.String(), "could not update git info/exclude") {
		t.Errorf("no exclude warning logged; logs:\n%s", logs.String())
	}
}

func TestEnablePlainUnchanged(t *testing.T) {
	dir, home, _ := newEnableTestRepo(t)

	out, err := runEnableWith(t, nil)
	if err != nil {
		t.Fatalf("runEnable() error = %v", err)
	}
	if strings.Contains(out, "Stealth") {
		t.Errorf("plain enable output mentions stealth:\n%s", out)
	}

	raw := readJSONObject(t, filepath.Join(dir, config.PartioDir, "settings.json"))
	if string(raw["enabled"]) != "true" {
		t.Errorf(".partio/settings.json enabled = %s, want true", raw["enabled"])
	}

	gitignore, err := os.ReadFile(filepath.Join(dir, ".gitignore"))
	if err != nil {
		t.Fatalf("reading .gitignore: %v", err)
	}
	for _, entry := range []string{".partio/settings.local.json", ".partio/sessions/", ".partio/state/"} {
		if !strings.Contains(string(gitignore), entry+"\n") {
			t.Errorf(".gitignore missing %q:\n%s", entry, gitignore)
		}
	}

	if _, err := os.Stat(globalSettingsFile(home)); !os.IsNotExist(err) {
		t.Errorf("global settings file exists or is unreadable (err = %v), want plain enable not to touch it", err)
	}
}

// TestStealthEndToEnd enables stealth, makes a commit, runs the post-commit
// hook with the effective config and checks that the commit is untouched while
// the checkpoint can still be listed and explained.
func TestStealthEndToEnd(t *testing.T) {
	dir, _, run := newEnableTestRepo(t)

	if _, err := runEnableWith(t, map[string]string{"stealth": "true"}); err != nil {
		t.Fatalf("runEnable(--stealth) error = %v", err)
	}

	effective, err := config.Load(dir)
	if err != nil {
		t.Fatalf("config.Load: %v", err)
	}
	if !stealthMode(effective) {
		t.Fatalf("effective config after enable --stealth is not stealth: %+v", effective.StrategyOptions)
	}
	previousCfg := cfg
	cfg = effective
	t.Cleanup(func() { cfg = previousCfg })

	if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("one\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	run("git", "add", "a.txt")
	// Keep git from running the installed shims: they would call whatever
	// partio binary is on PATH. The test runs the post-commit hook itself.
	run("git", "-c", "core.hooksPath=/dev/null", "commit", "-m", "work")

	before := run("git", "rev-parse", "HEAD")
	beforeMsg := run("git", "log", "-1", "--format=%B")
	branch := run("git", "rev-parse", "--abbrev-ref", "HEAD")

	state, err := json.Marshal(map[string]any{"agent_active": true, "agent_name": "claude-code", "branch": branch})
	if err != nil {
		t.Fatal(err)
	}
	stateDir := filepath.Join(dir, config.PartioDir, "state")
	if err := os.MkdirAll(stateDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(stateDir, "pre-commit.json"), state, 0o644); err != nil {
		t.Fatal(err)
	}

	if err := runHook(newHookCmd(), []string{"post-commit"}); err != nil {
		t.Fatalf("runHook(post-commit) error = %v", err)
	}

	if after := run("git", "rev-parse", "HEAD"); after != before {
		t.Errorf("HEAD = %s, want unchanged %s", after, before)
	}
	afterMsg := run("git", "log", "-1", "--format=%B")
	if afterMsg != beforeMsg || strings.Contains(afterMsg, "Partio-") {
		t.Errorf("commit message = %q, want unchanged %q with no trailers", afterMsg, beforeMsg)
	}

	cps, err := checkpoint.FindByBranch(dir, branch)
	if err != nil {
		t.Fatalf("checkpoint.FindByBranch: %v", err)
	}
	if len(cps) != 1 {
		t.Fatalf("got %d checkpoints, want 1", len(cps))
	}
	id := cps[0].ID

	var listErr error
	listOut := captureOutput(t, func() { listErr = runRewindList("") })
	if listErr != nil {
		t.Fatalf("runRewindList() error = %v", listErr)
	}
	if !strings.Contains(listOut, id) {
		t.Errorf("rewind --list output missing checkpoint %s:\n%s", id, listOut)
	}

	var explainErr error
	explainOut := captureOutput(t, func() { explainErr = runExplain(newExplainCmd(), []string{id}) })
	if explainErr != nil {
		t.Fatalf("runExplain(%s) error = %v", id, explainErr)
	}
	if !strings.Contains(explainOut, before) {
		t.Errorf("explain output missing commit %s:\n%s", before, explainOut)
	}
}
