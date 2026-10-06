package premise

import (
	"errors"
	"io/fs"
	"os"
	"regexp"
	"strings"
	"testing"
)

const (
	implementProgram = "../../.minions/programs/implement.md"
	buildWorkflow    = "../../.github/workflows/minion.yml"

	// removedGateProgram is the premise-gate program the review replaced.
	removedGateProgram = "../../.minions/programs/premise-gate.md"

	// buildGateCall runs the review gate in the mode a build uses.
	buildGateCall = "minion-review gate --mode build"
)

// reviewProgramRef is reviewProgram as a workflow file spells it.
var reviewProgramRef = strings.TrimPrefix(reviewProgram, "../../")

// workflowStep starts a step in the build workflow's step list. The runtime
// order of the steps is what the gate depends on, so the tests below compare
// step positions rather than positions in the raw file: the program path
// appears in an earlier "Determine program" step than the step that runs it.
var workflowStep = regexp.MustCompile(`(?m)^      - `)

// buildSteps splits the build workflow into its steps, in order. Element 0 is
// everything above the first step and is kept so the returned indices line up
// with nothing in particular — only their order matters.
func buildSteps(t *testing.T) []string {
	t.Helper()

	return workflowStep.Split(readRepoFile(t, buildWorkflow), -1)
}

// stepContaining reports the index of the first build step carrying want.
func stepContaining(steps []string, want string) int {
	for i, step := range steps {
		if strings.Contains(step, want) {
			return i
		}
	}
	return -1
}

// TestBuildStageReviewsBeforeItWritesCode is the tracer bullet: the build
// reviews its issue with the program and gate the sweep uses, as its first
// steps after the install, and before any step that writes code. The premise
// gate it replaced is gone, from the workflow and from the repository.
func TestBuildStageReviewsBeforeItWritesCode(t *testing.T) {
	context, ok := section(readRepoFile(t, reviewProgram), "## Context")
	if !ok {
		t.Fatalf("%s has no ## Context section, so the shared verifier is never read", reviewProgram)
	}
	if !strings.Contains(context, VerifierPath) {
		t.Errorf("%s ## Context does not carry %s", reviewProgram, VerifierPath)
	}

	steps := buildSteps(t)
	installAt := stepContaining(steps, "name: Install minions")
	reviewAt := stepContaining(steps, "minions run "+reviewProgramRef)
	gateAt := stepContaining(steps, buildGateCall)
	buildAt := stepContaining(steps, "minions $ARGS")
	switch {
	case installAt < 0 || buildAt < 0:
		t.Fatalf("%s lost its install or build step", buildWorkflow)
	case reviewAt != installAt+1:
		t.Errorf("the review session is step %d, want the first step after the install (%d)", reviewAt, installAt+1)
	case gateAt != reviewAt+1:
		t.Errorf("the build gate is step %d, want the step after the review session (%d)", gateAt, reviewAt+1)
	case gateAt > buildAt:
		t.Errorf("the build gate runs after the build: gate is step %d, build is step %d", gateAt, buildAt)
	}
	if gateAt >= 0 && !strings.Contains(steps[gateAt], "id: gate") {
		t.Errorf("the build gate step is not id gate, so steps.gate.outputs.blocked reads nothing")
	}

	if at := stepContaining(steps, "premise-gate.md"); at >= 0 {
		t.Errorf("step %d still runs the premise-gate program", at)
	}
	if _, err := os.Stat(removedGateProgram); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("%s still exists (stat: %v); the review replaced it", removedGateProgram, err)
	}
}

// TestReviewSessionCannotEndTheJobBeforeTheGate pins the fail-closed half of
// the review: a crashed session leaves no verdict, and the gate must still run
// to turn that into a failed job. So the session step continues on error, and
// the gate does not.
func TestReviewSessionCannotEndTheJobBeforeTheGate(t *testing.T) {
	steps := buildSteps(t)
	reviewAt := stepContaining(steps, "minions run "+reviewProgramRef)
	gateAt := stepContaining(steps, buildGateCall)
	if reviewAt < 0 || gateAt < 0 {
		t.Fatalf("%s does not run both the review session and the build gate", buildWorkflow)
	}

	if !strings.Contains(steps[reviewAt], "continue-on-error: true") {
		t.Error("the review session step does not continue on error, so a crash ends the job before the gate")
	}
	if strings.Contains(steps[gateAt], "continue-on-error") {
		t.Error("the build gate continues on error, so no verdict lets the build run")
	}
	if !strings.Contains(readRepoFile(t, buildWorkflow), "MINION_REVIEW_DIR: ${{ github.workspace }}/") {
		t.Error("MINION_REVIEW_DIR is not an absolute path in the workspace, so the gate cannot find the verdict")
	}
	if !strings.Contains(steps[gateAt], `--verdict "$MINION_REVIEW_DIR/verdict.json"`) {
		t.Error("the build gate does not read the verdict the review program writes")
	}
}

// TestGateReadsThisRunsVerdict pins that the verdict acted on is this run's,
// not one an earlier run left behind: the session step clears the verdict
// directory before the review runs.
func TestGateReadsThisRunsVerdict(t *testing.T) {
	steps := buildSteps(t)
	at := stepContaining(steps, "minions run "+reviewProgramRef)
	if at < 0 {
		t.Fatalf("%s never runs the review program", buildWorkflow)
	}
	clearAt := strings.Index(steps[at], `rm -rf "$MINION_REVIEW_DIR"`)
	runAt := strings.Index(steps[at], "minions run "+reviewProgramRef)
	if clearAt < 0 || clearAt > runAt {
		t.Error("the review step does not clear MINION_REVIEW_DIR before the session, so a stale verdict decides the build")
	}
}

// gateGuard is the condition that keeps a step from running once the review
// has closed the issue.
const gateGuard = "steps.gate.outputs.blocked != 'true'"

// researchGuard is the condition that keeps a step from running once research
// has ended without a slice plan. The runtime falls back to a single
// whole-issue session when no plan comment exists, so a build left unguarded
// here would silently become the unplanned build the chain exists to remove.
const researchGuard = "steps.researched.outputs.blocked != 'true'"

// TestReviewRestatesNoVerification pins that the review applies the shared
// verifier rather than carrying its own copy. A pasted copy is what a stray
// marker looks like.
func TestReviewRestatesNoVerification(t *testing.T) {
	src := readRepoFile(t, reviewProgram)

	for _, marker := range []string{VerifierMarker, GateMarker} {
		if strings.Contains(src, marker) {
			t.Errorf("%s carries %s, so it is a copy rather than a reference", reviewProgram, marker)
		}
	}

	reviewer, ok := section(src, "### reviewer")
	if !ok {
		t.Fatalf("%s has no ### reviewer agent", reviewProgram)
	}
	for _, want := range []string{VerifierPath, "as written"} {
		if !containsPhrase(reviewer, want) {
			t.Errorf("### reviewer does not name %q, so it does not reuse the shared behaviour", want)
		}
	}
}

// TestResearchStopsAsTheGateDescribes pins that research stops the one way the
// stage gate describes, and holds no idea of its own of the label or the
// comment shape. The build no longer stops that way: its review gate closes
// the issue and explains the close.
func TestResearchStopsAsTheGateDescribes(t *testing.T) {
	src := readRepoFile(t, researchProgram)

	if !strings.Contains(src, GatePath) {
		t.Errorf("%s does not apply %s", researchProgram, GatePath)
	}
	for _, r := range []struct {
		what  string
		token string
	}{
		{"the gate description", GateMarker},
		{"the blocking label", BlockingLabel},
		{"the gate comment shape", GateCommentMarker},
	} {
		if strings.Contains(src, r.token) {
			t.Errorf("%s carries %s (%s); stopping is described once, in %s",
				researchProgram, r.token, r.what, GatePath)
		}
	}
}

// TestBlockedBuildOpensNoPullRequestAndCreatesNoBranch pins the criterion the
// runtime cannot enforce from inside a program. The runtime creates the branch
// before the agent session starts and pushes unconditionally, so the only place
// a build can be stopped without leaving artefacts behind is before it starts.
func TestBlockedBuildOpensNoPullRequestAndCreatesNoBranch(t *testing.T) {
	steps := buildSteps(t)

	buildAt := stepContaining(steps, "minions $ARGS")
	if buildAt < 0 {
		t.Fatalf("%s never runs the implement program", buildWorkflow)
	}
	if !strings.Contains(steps[buildAt], gateGuard) {
		t.Errorf("the step that runs the build is not guarded by %q, so a closed issue still opens a pull request", gateGuard)
	}

	review := readRepoFile(t, reviewProgram)

	// The review must not be a slice-aware program: the slice path commits
	// an empty marker and pushes whatever the agent did or did not do, so a
	// checking program declared that way opens a pull request of its own.
	if strings.Contains(review, "slices: true") {
		t.Errorf("%s declares slices: true, so the runtime pushes and opens a PR even when it writes nothing", reviewProgram)
	}

	// The runtime opens a pull request for any worktree that is not clean,
	// and it reads that with `git status --porcelain`, which counts a file
	// nobody tracked yet. So the instruction to keep the working directory
	// alone is what makes the review free of artefacts.
	for _, want := range []string{
		"Write nothing in your working directory",
		"do not open a pull request",
	} {
		if !containsPhrase(review, want) {
			t.Errorf("%s does not say %q, so a file left behind turns the review into a pull request", reviewProgram, want)
		}
	}
}

// TestBlockedBuildIsNeitherDoneNorFailed pins that a close is a block, not a
// completion and not a failure: the gate's comment explains the close, and no
// workflow step marks the issue done or failed after it.
func TestBlockedBuildIsNeitherDoneNorFailed(t *testing.T) {
	src := readRepoFile(t, reviewProgram)
	for _, token := range []string{"gh issue close", "minion-done", BlockingLabel} {
		if strings.Contains(src, token) {
			t.Errorf("%s carries %q; the session judges, the gate acts", reviewProgram, token)
		}
	}

	steps := buildSteps(t)
	for _, token := range []string{"gh issue close", "minion-done", "minion-failed"} {
		at := stepContaining(steps, token)
		if at < 0 {
			continue
		}
		if !strings.Contains(steps[at], gateGuard) {
			t.Errorf("the step carrying %q is not guarded by %q, so a closed issue is still marked", token, gateGuard)
		}
	}
}

// TestNoVerdictFailsTheBuildAndMarksTheIssue pins the other half of failing
// closed. The gate exits non-zero on no verdict and writes no blocked output,
// so the job fails, and the failure step, which only a block skips, marks the
// issue.
func TestNoVerdictFailsTheBuildAndMarksTheIssue(t *testing.T) {
	steps := buildSteps(t)
	gateAt := stepContaining(steps, buildGateCall)
	failAt := stepContaining(steps, "minion-failed")
	if gateAt < 0 || failAt < 0 {
		t.Fatalf("%s lost its build gate or its failure step", buildWorkflow)
	}
	if strings.Contains(steps[gateAt], "|| true") || strings.Contains(steps[gateAt], "if !") {
		t.Error("the build gate step swallows the gate's exit code, so no verdict lets the build run")
	}
	if failAt < gateAt || !strings.Contains(steps[failAt], "if: failure()") {
		t.Error("the failure step does not run on a failed gate, so no verdict leaves the issue unmarked")
	}
}

// TestKeptIssueBuildsUnchanged pins that only a blocking verdict stops the
// build. Exactly two can block: the review gate's close, and research when it
// ends with no slice plan. Any further condition on the build step would make
// a keep or a rewrite change how the build runs.
func TestKeptIssueBuildsUnchanged(t *testing.T) {
	wantGuard := "if: " + gateGuard + " && " + researchGuard

	steps := buildSteps(t)
	at := stepContaining(steps, "minions $ARGS")
	if at < 0 {
		t.Fatalf("%s never runs the implement program", buildWorkflow)
	}
	for _, line := range strings.Split(steps[at], "\n") {
		if !strings.HasPrefix(strings.TrimSpace(line), "if:") {
			continue
		}
		if strings.TrimSpace(line) != wantGuard {
			t.Errorf("the build step is conditioned on %q, want %q: only a blocking verdict may stop the build",
				strings.TrimSpace(line), wantGuard)
		}
	}
}

// topHeading matches a second-level heading in a program file.
var topHeading = regexp.MustCompile(`(?m)^## (.+)$`)

// deepHeading matches a heading below the level the program parser reads. The
// parser splits on every heading, whatever its depth, and keeps only the
// second-level sections it knows and the third-level agents inside ## Agents.
// A fourth-level heading therefore ends the agent's prose and its body is
// discarded without a word.
var deepHeading = regexp.MustCompile(`(?m)^#{4,} `)

// keptHeadings are the second-level sections the program parser carries into
// the prompt. Everything else is dropped, silently.
var keptHeadings = map[string]bool{"Context": true, "Planner": true, "Agents": true}

// TestImplementProgramInstructionsReachTheModel pins criterion 9. The parser
// keeps the H1 prose, ## Context, ## Planner and ## Agents, and drops every
// other second-level section without saying so. programshape.Check cannot catch
// this for the implement program: it returns nil for any program that defines
// an ## Agents section, so the instructions have to be pinned here.
func TestImplementProgramInstructionsReachTheModel(t *testing.T) {
	src := readRepoFile(t, implementProgram)

	for _, m := range topHeading.FindAllStringSubmatch(src, -1) {
		if heading := strings.TrimSpace(m[1]); !keptHeadings[heading] {
			t.Errorf("## %s is dropped by the parser, so nothing under it reaches the model", heading)
		}
	}

	if m := deepHeading.FindString(src); m != "" {
		t.Errorf("a %q heading ends the agent's prose, so everything below it is dropped", strings.TrimSpace(m))
	}

	agents, ok := section(src, "## Agents")
	if !ok {
		t.Fatalf("%s has no ## Agents section", implementProgram)
	}
	for _, want := range []string{
		"build only that slice",
		"Leave the tree green at your boundary",
		"never open a PR yourself",
		"conventional commit format",
		"Resolves #",
	} {
		if !containsPhrase(agents, want) {
			t.Errorf("the instruction %q does not live under ## Agents, so the parser drops it", want)
		}
	}
}

// TestBuildVerifiesAProposalThatCarriesNoBlock pins the case that covers the
// whole backlog. Most open proposals carry no premise block, so a review that
// treats a blockless issue as out of scope checks nothing, and the build runs
// on an unchecked premise. The verifier already describes where those claims
// come from, so the review routes to it and does not stop.
func TestBuildVerifiesAProposalThatCarriesNoBlock(t *testing.T) {
	if !containsPhrase(readRepoFile(t, verifierDoc), NoBlockSection) {
		t.Fatalf("%s no longer carries %q, so no stage has a described route for a blockless proposal",
			VerifierPath, NoBlockSection)
	}

	reviewer, ok := section(readRepoFile(t, reviewProgram), "### reviewer")
	if !ok {
		t.Fatalf("%s has no ### reviewer agent", reviewProgram)
	}

	if !containsPhrase(reviewer, NoBlockSection) {
		t.Errorf("### reviewer never routes to %q in %s, so a proposal with no block is never verified",
			NoBlockSection, VerifierPath)
	}

	// The exact wording that once shipped the bug in the premise gate.
	for _, escape := range []string{"out of scope", "Leave the issue alone"} {
		if containsPhrase(reviewer, escape) {
			t.Errorf("### reviewer says %q of a blockless issue; that is every open proposal, so every build goes through unchecked",
				escape)
		}
	}
}
