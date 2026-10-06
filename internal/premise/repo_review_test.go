package premise

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

const (
	reviewProgram  = "../../.minions/programs/review.md"
	sweepWorkflow  = "../../.github/workflows/proposal-review.yml"
	reviewRun      = "minions run .minions/programs/review.md"
	reviewGateRun  = "go run ./cmd/minion-review gate"
	reviewAgentH3  = "### reviewer"
	reviewDirEnv   = "MINION_REVIEW_DIR"
	verdictFileRef = "$MINION_REVIEW_DIR/verdict.json"
)

// sweepLoop returns the body of the per-issue loop in the sweep workflow: the
// text between the loop's "do" and its "done". The program and the gate run in
// one step, so their order is an order inside this body, not a step order.
func sweepLoop(t *testing.T) string {
	t.Helper()

	steps := workflowStep.Split(readRepoFile(t, sweepWorkflow), -1)
	at := stepContaining(steps, reviewRun)
	if at < 0 {
		t.Fatalf("%s has no step that runs %q", sweepWorkflow, reviewRun)
	}
	step := steps[at]
	run := strings.Index(step, reviewRun)

	start := strings.LastIndex(step[:run], "for ")
	if start < 0 {
		t.Fatalf("the review program does not run inside a per-issue loop:\n%s", step)
	}
	end := strings.Index(step[run:], "\n          done")
	if end < 0 {
		t.Fatalf("the per-issue loop has no done:\n%s", step)
	}
	return step[start : run+end]
}

// TestSweepReviewsEachIssueBeforeItsGate is the tracer bullet: a dispatch runs
// one fresh review session per issue, and the gate records that session's
// verdict right after it, in the same loop turn.
func TestSweepReviewsEachIssueBeforeItsGate(t *testing.T) {
	loop := sweepLoop(t)

	runAt := strings.Index(loop, reviewRun)
	gateAt := strings.Index(loop, reviewGateRun)
	switch {
	case runAt < 0:
		t.Fatalf("the per-issue loop does not run the review program:\n%s", loop)
	case gateAt < 0:
		t.Fatalf("the per-issue loop does not run the gate:\n%s", loop)
	case gateAt < runAt:
		t.Errorf("the gate runs before the review program, so it reads the last issue's verdict:\n%s", loop)
	}

	for _, want := range []string{`--issue "$ISSUE"`, "--dry-run", `--night "$NIGHT"`, "--verdict"} {
		if !strings.Contains(loop[gateAt:], want) {
			t.Errorf("the gate call does not carry %s:\n%s", want, loop[gateAt:])
		}
	}
}

// TestReviewProgramIsOneAgent pins where the review's instructions live: in
// one agent under ## Agents. The runtime drops ## Steps sections and ####
// headings, so instructions there never reach the model.
func TestReviewProgramIsOneAgent(t *testing.T) {
	src := readRepoFile(t, reviewProgram)

	agents, ok := section(src, "## Agents")
	if !ok {
		t.Fatalf("%s has no ## Agents section", reviewProgram)
	}
	var h3 []string
	for line := range strings.SplitSeq(agents, "\n") {
		if strings.HasPrefix(line, "### ") {
			h3 = append(h3, line)
		}
	}
	if len(h3) != 1 || h3[0] != reviewAgentH3 {
		t.Errorf("%s ## Agents holds %q, want exactly %q", reviewProgram, h3, reviewAgentH3)
	}

	for line := range strings.SplitSeq(src, "\n") {
		if strings.HasPrefix(line, "## Steps") || strings.HasPrefix(line, "#### ") {
			t.Errorf("%s carries %q, which the runtime drops", reviewProgram, line)
		}
	}
}

// reviewer returns the review agent's section, the only text of the program
// the runtime gives the model besides the H1 prose and ## Context.
func reviewer(t *testing.T) string {
	t.Helper()

	agent, ok := section(readRepoFile(t, reviewProgram), reviewAgentH3)
	if !ok {
		t.Fatalf("%s has no %s section", reviewProgram, reviewAgentH3)
	}
	return agent
}

// TestReviewJudgesWithTheSharedRules checks that the review judges a premise
// with the verifier, assumed facts included, and fit with the ingest prompt's
// relevance rule, rather than with a second idea of either. A keep or a rewrite
// is described in the issue shape the proposer files, so a build reads a
// reviewed issue the way it reads a new one.
func TestReviewJudgesWithTheSharedRules(t *testing.T) {
	src := readRepoFile(t, reviewProgram)
	context, ok := section(src, "## Context")
	if !ok {
		t.Fatalf("%s has no ## Context section", reviewProgram)
	}
	for _, want := range []string{VerifierPath, ".minions/ingest-prompt.md"} {
		if !strings.Contains(context, want) {
			t.Errorf("%s ## Context does not carry %s", reviewProgram, want)
		}
	}

	agent := reviewer(t)
	for _, want := range []string{
		VerifierPath,
		NoBlockSection,
		"assumed fact",
		".minions/ingest-prompt.md",
		"Git workflows, AI agent sessions, code attribution, checkpoints",
	} {
		if !containsPhrase(agent, want) {
			t.Errorf("%s does not carry %q", reviewAgentH3, want)
		}
	}

	for _, want := range []string{
		"## Acceptance Criteria",
		"`- [ ]`",
		"## Premise",
		"<!-- partio:premise:v1 -->",
		"[evidence:",
		"Proposal id: <id>",
	} {
		if !containsPhrase(agent, want) {
			t.Errorf("%s does not describe the proposer's issue shape: no %q", reviewAgentH3, want)
		}
	}
}

// TestReviewWritesOnlyItsVerdict checks that a review session changes nothing.
// It reads GitHub and writes one file outside its working directory: a file in
// the working directory is read as build output and turns the run into a pull
// request, and a GitHub change is the gate's to make, after the verdict passes.
func TestReviewWritesOnlyItsVerdict(t *testing.T) {
	agent := reviewer(t)

	for _, want := range []string{
		verdictFileRef,
		"Make no change on GitHub",
		"Write nothing in your working directory",
		"The verdict file is the only file you write",
	} {
		if !containsPhrase(agent, want) {
			t.Errorf("%s does not carry %q", reviewAgentH3, want)
		}
	}

	if strings.Contains(agent, "/tmp") {
		t.Errorf("%s offers /tmp for a scratch file, so the verdict is not the only file it writes", reviewAgentH3)
	}

	for _, write := range []string{
		"gh issue edit", "gh issue close", "gh issue comment", "gh issue create",
		"gh pr create", "gh pr comment", "gh pr edit", "gh label", "gh api",
	} {
		if strings.Contains(agent, write) {
			t.Errorf("%s names %q, a call that can change GitHub", reviewAgentH3, write)
		}
	}
}

// TestReviewRewriteKeepsTheIdea checks that a rewrite changes the issue's shape
// and not its idea. It keeps the idea and the link to its source item, and it
// drops the pointer to a proposal file: proposal files no longer exist, so the
// pointer leads nowhere.
func TestReviewRewriteKeepsTheIdea(t *testing.T) {
	agent := flat(reviewer(t))

	for _, want := range []string{
		"keep the original idea",
		"source link",
		"<!-- program:",
		"drop",
	} {
		if !strings.Contains(agent, want) {
			t.Errorf("%s does not carry %q", reviewAgentH3, want)
		}
	}
}

// TestSweepIsDispatchedByHand checks the sweep's only trigger in this version:
// a manual dispatch with a list of issues and a dry-run flag that defaults to
// true. Every issue runs in one job, so the issues run in sequence.
func TestSweepIsDispatchedByHand(t *testing.T) {
	src := readRepoFile(t, sweepWorkflow)
	jobs := strings.Index(src, "\njobs:")
	if jobs < 0 {
		t.Fatalf("%s has no jobs: section", sweepWorkflow)
	}
	on := src[:jobs]
	for _, want := range []string{"workflow_dispatch:", "issues:", "dry_run:", "default: true", "type: boolean"} {
		if !strings.Contains(on, want) {
			t.Errorf("%s dispatch does not carry %q", sweepWorkflow, want)
		}
	}
	for _, trigger := range []string{"schedule:", "pull_request:", "issues:\n    types"} {
		if strings.Contains(on, trigger) {
			t.Errorf("%s carries the trigger %q; this version is dispatched by hand only", sweepWorkflow, trigger)
		}
	}

	if n := strings.Count(src, "runs-on:"); n != 1 {
		t.Errorf("%s has %d jobs, want one job that reviews every issue in sequence", sweepWorkflow, n)
	}
	for _, want := range []string{"runs-on: github-runner-partio-minion-ai-01", "GH_TOKEN: ${{ secrets.GH_PAT }}"} {
		if !strings.Contains(src, want) {
			t.Errorf("%s does not carry %q", sweepWorkflow, want)
		}
	}

	loop := sweepLoop(t)
	if !strings.Contains(loop, "for ISSUE in ") {
		t.Errorf("the review does not run once per issue:\n%s", loop)
	}
	if !strings.Contains(loop, `rm -rf "$`+reviewDirEnv+`"`) {
		t.Errorf("the loop does not clear the review directory, so one issue can read another's verdict:\n%s", loop)
	}
	if !strings.Contains(src, reviewDirEnv+": ${{ github.workspace }}/") {
		t.Errorf("%s does not set %s to an absolute path in the workspace", sweepWorkflow, reviewDirEnv)
	}
}

// TestSweepSurvivesAFailedReview checks that one issue cannot end the sweep. The
// step shell runs with -e, so an unguarded failing command ends the loop: a
// crashed session and a "no verdict" exit from the gate are each guarded, and
// nothing in the loop body exits.
func TestSweepSurvivesAFailedReview(t *testing.T) {
	loop := sweepLoop(t)

	for _, call := range []string{reviewRun, reviewGateRun} {
		line, ok := lineContaining(loop, call)
		if !ok {
			t.Fatalf("the per-issue loop does not run %q", call)
		}
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "if ! ") && !strings.Contains(line, "||") {
			t.Errorf("a failure of %q ends the loop; guard it with `if !` or `||`:\n%s", call, line)
		}
	}

	for line := range strings.SplitSeq(loop, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "exit") {
			t.Errorf("the per-issue loop exits, so the issues after this one are never reviewed:\n%s", line)
		}
	}
}

// minionsPin matches the pinned install of the minions runtime in a workflow.
var minionsPin = regexp.MustCompile(`go install github\.com/partio-io/minions/cmd/minions@(\S+)`)

// TestSweepPinsTheRuntimeOfTheOtherWorkflows checks that the sweep runs the
// review on the minions version every other workflow runs, so a program change
// is never tested on one runtime and swept on another. The gate runs from the
// checked-out tree with go run, like the other minion gates.
func TestSweepPinsTheRuntimeOfTheOtherWorkflows(t *testing.T) {
	sweep := readRepoFile(t, sweepWorkflow)
	m := minionsPin.FindStringSubmatch(sweep)
	if m == nil {
		t.Fatalf("%s does not install minions at a pinned version", sweepWorkflow)
	}
	pin := m[1]

	others, err := filepath.Glob("../../.github/workflows/*.yml")
	if err != nil {
		t.Fatalf("list the workflows: %v", err)
	}
	seen := 0
	for _, path := range others {
		if path == sweepWorkflow {
			continue
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		for _, other := range minionsPin.FindAllStringSubmatch(string(raw), -1) {
			seen++
			if other[1] != pin {
				t.Errorf("%s installs minions@%s, but %s installs minions@%s", sweepWorkflow, pin, path, other[1])
			}
		}
	}
	if seen == 0 {
		t.Fatal("no other workflow installs minions, so the pin has nothing to match")
	}

	if !strings.Contains(sweepLoop(t), reviewGateRun) {
		t.Errorf("the per-issue loop does not run the gate with %q", reviewGateRun)
	}
}
