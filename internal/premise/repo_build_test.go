package premise

import (
	"regexp"
	"strings"
	"testing"
)

const (
	implementProgram = "../../.minions/programs/implement.md"
	gateProgram      = "../../.minions/programs/premise-gate.md"
	buildWorkflow    = "../../.github/workflows/minion.yml"
)

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

// TestBuildStageVerifiesBeforeItWritesCode is the tracer bullet: the premise is
// checked by its own program, that program runs before the one that writes
// code, and the verification it applies is the shared description rather than a
// second idea of what verification means.
func TestBuildStageVerifiesBeforeItWritesCode(t *testing.T) {
	gate := readRepoFile(t, gateProgram)

	context, ok := section(gate, "## Context")
	if !ok {
		t.Fatalf("%s has no ## Context section, so the shared descriptions are never read", gateProgram)
	}
	for _, want := range []string{VerifierPath, GatePath} {
		if !strings.Contains(context, want) {
			t.Errorf("%s ## Context does not carry %s", gateProgram, want)
		}
	}

	steps := buildSteps(t)

	gateAt := stepContaining(steps, "premise-gate.md")
	if gateAt < 0 {
		t.Fatalf("%s never runs the premise gate", buildWorkflow)
	}

	buildAt := stepContaining(steps, "minions $ARGS")
	if buildAt < 0 {
		t.Fatalf("%s never runs the implement program", buildWorkflow)
	}

	if gateAt > buildAt {
		t.Errorf("the premise gate runs after the build: gate is step %d, build is step %d", gateAt, buildAt)
	}
}

// gateGuard is the condition that keeps a step from running once the premise
// has been found not to hold.
const gateGuard = "steps.gate.outputs.blocked != 'true'"

// TestBuildRestatesNeitherVerificationNorGating pins that the build stage
// applies the two shared descriptions rather than carrying its own copies. A
// pasted copy is what a stray marker looks like.
func TestBuildRestatesNeitherVerificationNorGating(t *testing.T) {
	src := readRepoFile(t, gateProgram)

	for _, marker := range []string{VerifierMarker, GateMarker} {
		if strings.Contains(src, marker) {
			t.Errorf("%s carries %s, so it is a copy rather than a reference", gateProgram, marker)
		}
	}

	checker, ok := section(src, "### premise-checker")
	if !ok {
		t.Fatalf("%s has no ### premise-checker agent", gateProgram)
	}
	for _, want := range []string{VerifierPath, GatePath, "as written"} {
		if !containsPhrase(checker, want) {
			t.Errorf("### premise-checker does not name %q, so it does not reuse the shared behaviour", want)
		}
	}
}

// TestBuildStopBehaviourIsTheSameAsResearch pins criterion 3: the build stops
// the way research stops. Both stages point at the one description of stopping,
// and neither holds its own idea of the label or the comment shape.
func TestBuildStopBehaviourIsTheSameAsResearch(t *testing.T) {
	restated := []struct {
		what  string
		token string
	}{
		{"the gate description", GateMarker},
		{"the blocking label", BlockingLabel},
		{"the gate comment shape", GateCommentMarker},
	}

	for _, path := range []string{gateProgram, researchProgram} {
		src := readRepoFile(t, path)

		if !strings.Contains(src, GatePath) {
			t.Errorf("%s does not apply %s", path, GatePath)
		}
		for _, r := range restated {
			if strings.Contains(src, r.token) {
				t.Errorf("%s carries %s (%s); stopping is described once, in %s",
					path, r.token, r.what, GatePath)
			}
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
		t.Errorf("the step that runs the build is not guarded by %q, so a blocked premise still opens a pull request", gateGuard)
	}

	gate := readRepoFile(t, gateProgram)

	// The gate must not be a slice-aware program: the slice path commits an
	// empty marker and pushes whatever the agent did or did not do, so a
	// checking program declared that way opens a pull request of its own.
	if strings.Contains(gate, "slices: true") {
		t.Errorf("%s declares slices: true, so the runtime pushes and opens a PR even when it writes nothing", gateProgram)
	}

	// On the path the gate does take, the runtime opens a pull request for any
	// worktree that is not clean, and it reads that with `git status
	// --porcelain`, which counts a file nobody tracked yet. So the instruction
	// to keep the working directory alone is what makes the check free of
	// artefacts, not a detail of style.
	for _, want := range []string{
		"Do not create or modify any file in the working directory",
		"/tmp",
	} {
		if !containsPhrase(gate, want) {
			t.Errorf("%s does not say %q, so a file left behind turns the check into a pull request", gateProgram, want)
		}
	}
}

// TestBlockedBuildNeverClosesTheIssue pins criterion 6 across both halves of
// the stage: the program does not close the issue, and neither does the
// workflow once the gate has blocked the run.
func TestBlockedBuildNeverClosesTheIssue(t *testing.T) {
	src := readRepoFile(t, gateProgram)
	for _, token := range []string{"gh issue close", "minion-done"} {
		if strings.Contains(src, token) {
			t.Errorf("%s carries %q; the stage reports, the operator decides", gateProgram, token)
		}
	}

	// The workflow does close the issue on a normal run. Every step that ends
	// the issue's life must stand down when the premise did not hold.
	steps := buildSteps(t)
	for _, token := range []string{"gh issue close", "minion-done", "minion-failed"} {
		at := stepContaining(steps, token)
		if at < 0 {
			continue
		}
		if !strings.Contains(steps[at], gateGuard) {
			t.Errorf("the step carrying %q is not guarded by %q, so a blocked run still changes the issue's state",
				token, gateGuard)
		}
	}
}

// TestOperatorOverrulesTheBuildByRemovingTheLabel pins that the verdict acted
// on is this run's, not a label left over from an earlier one. The gate runs
// first and the label is read afterwards, so an operator who removes the label
// gets a fresh verdict rather than a remembered one.
func TestOperatorOverrulesTheBuildByRemovingTheLabel(t *testing.T) {
	if src := readRepoFile(t, gateProgram); strings.Contains(src, BlockingLabel) {
		t.Errorf("%s names %s; a program that reads the label can skip on it instead of verifying again",
			gateProgram, BlockingLabel)
	}

	steps := buildSteps(t)
	at := stepContaining(steps, "premise-gate.md")
	if at < 0 {
		t.Fatalf("%s never runs the premise gate", buildWorkflow)
	}

	runAt := strings.Index(steps[at], "minions run .minions/programs/premise-gate.md")
	readAt := strings.Index(steps[at], "--json labels")
	switch {
	case readAt < 0:
		t.Fatal("the gate step never reads the issue's labels, so it has no verdict to act on")
	case runAt > readAt:
		t.Error("the gate step reads the labels before it verifies, so a stale label decides the build")
	}
}

// TestHoldingPremiseLetsTheBuildProceedUnchanged pins criterion 8: a premise
// that holds records its evidence, and changes nothing about the build.
func TestHoldingPremiseLetsTheBuildProceedUnchanged(t *testing.T) {
	checker, ok := section(readRepoFile(t, gateProgram), "### premise-checker")
	if !ok {
		t.Fatalf("%s has no ### premise-checker agent", gateProgram)
	}
	for _, want := range []string{
		"record every claim",
		"the excerpt that produced it",
		// "premise", not "block": the gate also verifies claims extracted
		// from the prose of a proposal filed before the block format, and
		// those record their evidence on the holding path too.
		"for a premise that holds",
	} {
		if !containsPhrase(checker, want) {
			t.Errorf("### premise-checker does not record %q, so a build that proceeds carries no evidence", want)
		}
	}

	// Only a blocked verdict stops the build. Any other condition on that step
	// would make a holding premise change how the build runs.
	steps := buildSteps(t)
	at := stepContaining(steps, "minions $ARGS")
	if at < 0 {
		t.Fatalf("%s never runs the implement program", buildWorkflow)
	}
	for _, line := range strings.Split(steps[at], "\n") {
		if !strings.HasPrefix(strings.TrimSpace(line), "if:") {
			continue
		}
		if strings.TrimSpace(line) != "if: "+gateGuard {
			t.Errorf("the build step is conditioned on %q, not only on a blocked premise", strings.TrimSpace(line))
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
// whole backlog. No open proposal carried a premise block when the gate
// shipped, so a gate that treats a blockless issue as out of scope labels
// nothing. The workflow reads a missing label as "not blocked" and builds, and
// the gate passes every issue it was added to stop. The verifier already
// describes where those claims come from, so the gate routes to it and does
// not stop.
func TestBuildVerifiesAProposalThatCarriesNoBlock(t *testing.T) {
	if !containsPhrase(readRepoFile(t, verifierDoc), NoBlockSection) {
		t.Fatalf("%s no longer carries %q, so no stage has a described route for a blockless proposal",
			VerifierPath, NoBlockSection)
	}

	checker, ok := section(readRepoFile(t, gateProgram), "### premise-checker")
	if !ok {
		t.Fatalf("%s has no ### premise-checker agent", gateProgram)
	}

	if !containsPhrase(checker, NoBlockSection) {
		t.Errorf("### premise-checker never routes to %q in %s, so a proposal with no block is never verified",
			NoBlockSection, VerifierPath)
	}

	// The exact wording that shipped the bug. A gate that tells the checker to
	// stop on a blockless issue reaches no verdict, and a build with no verdict
	// runs.
	for _, escape := range []string{"out of scope", "Leave the issue alone"} {
		if containsPhrase(checker, escape) {
			t.Errorf("### premise-checker says %q of a blockless issue; that is every open proposal, so every build goes through ungated",
				escape)
		}
	}
}
