package checksverdict

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// programPath is the repair session's program, read from the repository
// for the same reason the workflow is: these are guarantees about what
// CI actually runs, and a copy would drift away from that silently.
const programPath = "../../.minions/programs/checks-fix.md"

func program(t *testing.T) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Clean(programPath))
	if err != nil {
		t.Fatalf("read the checks fix program: %v", err)
	}
	return string(data)
}

// The rule that separates a repair from a bot that reaches green by
// deleting an assertion. The session may correct a test the pull
// request added — a minion's own wrong test is the common case — and
// may not touch one that predates it. A failure in a pre-existing test
// survives as a finding, and the check stays red for a person to read.
func TestTheProgramBoundsWhichTestsMayChange(t *testing.T) {
	p := program(t)

	if !strings.Contains(p, "added-tests.txt") {
		t.Error("the program never reads the added-test list, so its boundary is its own judgment")
	}
	for _, rule := range []string{
		"only test files you may modify",
		"predates this pull request is out of bounds",
		"survives",
		"Never weaken a test",
	} {
		if !strings.Contains(p, rule) {
			t.Errorf("the program does not state %q", rule)
		}
	}
}

// Implementation code is what a repair usually has to change, and a
// program that only talked about tests would leave the session unsure
// it may touch anything else.
func TestTheProgramAllowsImplementationRepairs(t *testing.T) {
	p := program(t)

	if !strings.Contains(p, "Implementation code is yours to change") {
		t.Error("the program does not say implementation code may be changed")
	}
	if !strings.Contains(p, "Repair the implementation first") {
		t.Error("the program does not prefer an implementation repair over a test change")
	}
}

// The session contract, unchanged from the dead-code fixer: a patch,
// proven, and nothing else. The workflow owns every write to the
// repository, so a session that commits or pushes would be pushing work
// no one proved and no one counted.
func TestTheProgramProducesAPatchAndNothingElse(t *testing.T) {
	p := program(t)

	for _, rule := range []string{
		"The session runs no git write commands: no commit, no push, no branches, no PRs, no comments",
		"The skip marker file is written as the session's final act on every path",
		"make lint and make test pass in the patched tree before the patch is exported",
	} {
		if !strings.Contains(p, rule) {
			t.Errorf("the program's acceptance criteria do not carry %q", rule)
		}
	}
	if !strings.Contains(p, "You do not commit or push") {
		t.Error("the program does not tell the session it must not commit or push")
	}
	if !strings.Contains(p, "skip_marker: .minion-fix-done") {
		t.Error("the program declares no skip marker, so leftover edits would become a stray pull request")
	}
}
