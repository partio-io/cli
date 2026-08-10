package checksverdict

import (
	"strings"
	"testing"
)

// Both commands run even when the first one fails. Stopping early
// would hide a failing test behind a lint finding, and the author
// would come back for a second round over a failure this run already
// knew about.
func TestRunKeepsGoingAfterAFailure(t *testing.T) {
	outcomes := Run(t.TempDir(), []Command{
		{Name: "lint", Args: []string{"sh", "-c", "echo lint spoke; exit 1"}},
		{Name: "test", Args: []string{"sh", "-c", "echo test spoke; exit 1"}},
	})

	if len(outcomes) != 2 {
		t.Fatalf("ran %d commands, want 2: %+v", len(outcomes), outcomes)
	}
	for _, o := range outcomes {
		if !o.Failed {
			t.Errorf("%s reported success for a command that exited 1", o.Name)
		}
		if !strings.Contains(o.Output, o.Name+" spoke") {
			t.Errorf("%s did not capture its output: %q", o.Name, o.Output)
		}
	}
}

// Output arrives whole: stderr is where both `go test` and
// golangci-lint say the things a repair session needs most.
func TestRunCapturesBothStreamsAndTheExitStatus(t *testing.T) {
	outcomes := Run(t.TempDir(), []Command{
		{Name: "test", Args: []string{"sh", "-c", "echo to stdout; echo to stderr 1>&2; exit 2"}},
		{Name: "lint", Args: []string{"sh", "-c", "exit 0"}},
	})

	if len(outcomes) != 2 {
		t.Fatalf("ran %d commands, want 2: %+v", len(outcomes), outcomes)
	}
	failed := outcomes[0]
	for _, want := range []string{"to stdout", "to stderr"} {
		if !strings.Contains(failed.Output, want) {
			t.Errorf("output is missing %q: %q", want, failed.Output)
		}
	}
	if failed.Command != "sh -c echo to stdout; echo to stderr 1>&2; exit 2" {
		t.Errorf("Command = %q, which is not what a reader would re-run", failed.Command)
	}
	if outcomes[1].Failed {
		t.Error("a command that exited 0 was recorded as failed")
	}
}

// A command that cannot start at all — no such binary — is a failure,
// not a silent pass. This is the shape of a runner missing make or
// golangci-lint, and it must turn the check red.
func TestRunTreatsAnUnstartableCommandAsAFailure(t *testing.T) {
	outcomes := Run(t.TempDir(), []Command{
		{Name: "test", Args: []string{"definitely-not-a-real-binary-9f3a"}},
	})

	if len(outcomes) != 1 {
		t.Fatalf("ran %d commands, want 1: %+v", len(outcomes), outcomes)
	}
	if !outcomes[0].Failed {
		t.Fatal("a command that could not start was recorded as a success")
	}
	v := Convert(outcomes)
	if v.Status != "fail" {
		t.Errorf("status = %q, want fail", v.Status)
	}
	if len(v.Findings) == 0 {
		t.Error("a command that could not start produced no findings to explain it")
	}
}

// With no commands named, the check measures what the repository
// measures.
func TestRunFallsBackToTheDefaultCommands(t *testing.T) {
	dir := t.TempDir()

	outcomes := Run(dir, nil)

	if len(outcomes) != len(DefaultCommands()) {
		t.Fatalf("ran %d commands, want %d: %+v", len(outcomes), len(DefaultCommands()), outcomes)
	}
	for i, c := range DefaultCommands() {
		if outcomes[i].Command != strings.Join(c.Args, " ") {
			t.Errorf("command %d ran %q, want %q", i, outcomes[i].Command, strings.Join(c.Args, " "))
		}
	}
}
