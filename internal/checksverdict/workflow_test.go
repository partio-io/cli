package checksverdict

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/partio-io/cli/internal/repairround"
)

// workflowPath is the checks workflow, read from the repository rather
// than from a copy: the guarantees below are about what CI actually
// runs, and a copy would drift away from that silently.
const workflowPath = "../../.github/workflows/checks.yml"

func workflow(t *testing.T) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Clean(workflowPath))
	if err != nil {
		t.Fatalf("read checks workflow: %v", err)
	}
	return string(data)
}

// The check has to cover every pull request, including hand-written
// ones. The audits guard on the "minion" label; this one must not, or
// a broken build merges whenever a human opens the pull request.
func TestChecksRunOnEveryPullRequest(t *testing.T) {
	wf := workflow(t)

	if !strings.Contains(wf, "pull_request:") {
		t.Error("the workflow does not trigger on pull requests")
	}
	for _, event := range []string{"opened", "synchronize", "reopened"} {
		if !strings.Contains(wf, event) {
			t.Errorf("the workflow does not trigger on %q, so it would miss those pull requests", event)
		}
	}
	if strings.Contains(wf, "github.event.pull_request.labels") {
		t.Error("the workflow guards on a label, so unlabelled pull requests would go unchecked")
	}
}

// Nothing in this workflow moves a branch by hand. Repair pushes only
// through the shared applier, which fetches the true head, proves the
// tree, and pushes with the token that starts the follow-up run. Inline
// git writes would be a second implementation of that sequence, and the
// checkouts keep no credentials so a stray step cannot push at all.
func TestChecksPushOnlyThroughTheSharedApplier(t *testing.T) {
	wf := workflow(t)

	for _, forbidden := range []string{"git push", "git commit"} {
		if strings.Contains(wf, forbidden) {
			t.Errorf("the checks workflow runs %q inline; the shared applier owns apply-prove-push", forbidden)
		}
	}
	if !strings.Contains(wf, "persist-credentials: false") {
		t.Error("a checkout persists credentials, so a later step could push without meaning to")
	}
}

// The tracer: the whole repair chain, end to end, in the order it has
// to happen. The guard decides whether this pull request may be
// repaired at all and which round it is, the verdict decides whether
// there is anything to repair, the fix session produces a patch, the
// shared applier proves and pushes it, and the gate reports the round.
// A link missing here is a repair loop that stops halfway and says
// nothing.
func TestChecksRepairChain(t *testing.T) {
	wf := workflow(t)

	links := []struct {
		fragment string
		why      string
	}{
		{"minion-repair-round", "nothing counts the repair rounds, so the loop has no cap"},
		{"--check checks", "rounds are not counted under this check's own name"},
		{"--repairable", "nothing decides whether this pull request may be repaired at all"},
		{"needs.guard.outputs.repairable == 'true'", "the repair job is not gated on that decision"},
		{"needs.guard.outputs.skip != 'true'", "the repair job ignores a spent repair budget"},
		{"minions run .minions/programs/checks-fix.md", "no fix session runs, so nothing produces a patch"},
		{"minion-apply-fix", "no patch is applied, proven, or pushed"},
		{"--round", "the gate cannot tell a spent budget from a first-pass failure"},
	}
	for _, link := range links {
		if !strings.Contains(wf, link.fragment) {
			t.Errorf("the workflow is missing %q: %s", link.fragment, link.why)
		}
	}
}

// Each check gets its own three rounds. A dead-code repair must not
// spend the budget a failing test needs, and the counting is by name,
// so the two workflows have to count under different ones.
func TestChecksKeepsItsOwnRepairBudget(t *testing.T) {
	wf := workflow(t)

	audit, err := os.ReadFile(filepath.Clean("../../.github/workflows/deadcode-audit.yml"))
	if err != nil {
		t.Fatalf("read the dead-code workflow: %v", err)
	}

	if strings.Contains(wf, "--check dead-code") {
		t.Error("the checks workflow counts rounds under the dead-code name, so the two budgets interfere")
	}
	if !strings.Contains(string(audit), "--check dead-code") {
		t.Error("the dead-code workflow no longer counts under its own name")
	}
	// The cap comes from the shared default. A workflow that passes
	// --max sets a second cap the gate does not know about, and the two
	// sides of one round would disagree about when the budget is spent.
	if strings.Contains(wf, "--max") {
		t.Error("the checks workflow overrides the round cap; the gate still assumes the default")
	}
	if repairround.DefaultMaxRounds != 3 {
		t.Errorf("the shared cap is %d rounds, and this check is specified at three",
			repairround.DefaultMaxRounds)
	}
}

// The give-up comment needs the number of rounds that have already run.
// This gate reports before its own run's repair, so that number is the
// count on the branch — prior, never the round about to start. Reading
// the wrong one would announce a spent budget a round early.
func TestChecksGivesTheGateTheRoundsAlreadyRun(t *testing.T) {
	wf := workflow(t)

	if !strings.Contains(wf, "needs.guard.outputs.prior") {
		t.Error("the gate is not given the rounds already on the branch, so it cannot say the pipeline gave up")
	}
	if strings.Contains(wf, `--round "${ROUND`) {
		t.Error("the gate is given the round about to start; it would report a spent budget a round early")
	}
}

// The outcome belongs to the gate, not to the step that runs the
// commands: a converter that crashes must still leave the gate to fail
// the check closed and say so on the pull request.
func TestTheGateOwnsTheOutcome(t *testing.T) {
	wf := workflow(t)

	if !strings.Contains(wf, "minion-checks-verdict") {
		t.Error("the workflow never writes a verdict")
	}
	if !strings.Contains(wf, "minion-audit-gate") || !strings.Contains(wf, "--audit checks") {
		t.Error("the workflow does not report through the existing verdict gate")
	}
	if !strings.Contains(wf, "if: always()") {
		t.Error("the gate step is skippable, so a crashed converter would report nothing at all")
	}
}

// The commands are the repository's own targets. If these drift, the
// check stops measuring what the developers measure.
func TestDefaultCommandsAreTheRepositoryTargets(t *testing.T) {
	cmds := DefaultCommands()

	if len(cmds) != 2 {
		t.Fatalf("got %d default commands, want lint and test: %+v", len(cmds), cmds)
	}
	want := map[string]string{"lint": "make lint", "test": "make test"}
	for _, c := range cmds {
		if got := strings.Join(c.Args, " "); got != want[c.Name] {
			t.Errorf("command %q runs %q, want %q", c.Name, got, want[c.Name])
		}
	}
}
