package premise

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
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

	for _, want := range []string{`--issue "$ISSUE"`, `--dry-run="$DRY_RUN"`, `--night "$NIGHT"`, "--verdict"} {
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

// TestReviewExtractsTheAssumedFacts guards the miss of the second dry run on
// #31: the review listed six claims that hold and skipped the fact that the
// issue takes for granted, "append-only checkpoint data". The claims in the
// verdict are those of the issue as filed, not those of its rewrite.
func TestReviewExtractsTheAssumedFacts(t *testing.T) {
	agent := reviewer(t)

	for _, want := range []string{
		"Do not skip the assumed facts",
		"Start the claim with the quoted phrase",
		"the claims of the issue as it is filed",
		"The claims of a rewrite go into its new body, not into `premise`.",
	} {
		if !containsPhrase(agent, want) {
			t.Errorf("%s does not carry %q", reviewAgentH3, want)
		}
	}
}

// TestReviewCorrectsAFalseFactTheIdeaDoesNotRestOn pins the decisions of
// 2026-10-06: a failing claim closes the issue only when the idea rests on
// it. When the problem survives the correction, the review rewrites the issue,
// and the verdict carries the correction that the gate requires. The rewrite
// designs nothing new: it adds a criterion for each correction and removes the
// design that breaks one. The third dry run on #31 found the false fact but
// kept the union merge that brings pruned checkpoints back.
func TestReviewCorrectsAFalseFactTheIdeaDoesNotRestOn(t *testing.T) {
	agent := reviewer(t)

	for _, want := range []string{
		"`false-premise` when a claim fails and the idea rests on it",
		"`rewrite` with a corrected premise",
		"decide whether the idea rests on it",
		"the fact that holds instead",
		`"correction": "…"`,
		"A rewrite with a corrected premise designs nothing new.",
		"add an acceptance criterion that a build can test",
		"Remove each one that breaks a criterion or takes the false fact for granted.",
		"Add nothing in its place",
		"each criterion you added and each part you removed in `rewrite.changes`",
	} {
		if !containsPhrase(agent, want) {
			t.Errorf("%s does not carry %q", reviewAgentH3, want)
		}
	}
}

// TestReviewChecksItsOwnVerdict checks that the review runs the checks of the
// gate on its verdict before its session ends. Two of the first four dry runs
// on #31 ended with no verdict for a cause the gate checks in code: a body the
// premise parser refused, and a failing claim with no correction.
func TestReviewChecksItsOwnVerdict(t *testing.T) {
	agent := reviewer(t)

	for _, want := range []string{
		`go run ./cmd/minion-review check --issue <issue number> --verdict "$MINION_REVIEW_DIR/verdict.json"`,
		"fix the file and run the check again",
		"Run it at most three times",
	} {
		if !containsPhrase(agent, want) {
			t.Errorf("%s does not carry %q", reviewAgentH3, want)
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

// TestSweepRunsByHandAndAtNight pins the two triggers: a schedule at 23:00 UTC
// and a dispatch with an issue list, a sample size and the dry-run flag.
func TestSweepRunsByHandAndAtNight(t *testing.T) {
	src := readRepoFile(t, sweepWorkflow)
	on := sweepTriggers(t, src)
	for _, want := range []string{
		"workflow_dispatch:", "issues:", "sample:", "dry_run:", "default: true", "type: boolean",
		"schedule:", `- cron: "0 23 * * *"`,
	} {
		if !strings.Contains(on, want) {
			t.Errorf("%s triggers do not carry %q", sweepWorkflow, want)
		}
	}
	if n := strings.Count(on, "cron:"); n != 1 {
		t.Errorf("%s has %d schedules, want the one at 23:00 UTC", sweepWorkflow, n)
	}
	for _, trigger := range []string{"pull_request:", "issues:\n    types"} {
		if strings.Contains(on, trigger) {
			t.Errorf("%s carries the trigger %q", sweepWorkflow, trigger)
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

// sweepTriggers returns the text of the sweep workflow before its jobs.
func sweepTriggers(t *testing.T, src string) string {
	t.Helper()
	jobs := strings.Index(src, "\njobs:")
	if jobs < 0 {
		t.Fatalf("%s has no jobs: section", sweepWorkflow)
	}
	return src[:jobs]
}

// TestSweepActsAtNightOnlyWhenSwitchedOn pins the operator's switch: the job
// of a scheduled run is skipped unless PROPOSAL_REVIEW_SWEEP is on, so the run
// ends at once. A scheduled run has no inputs, so it is not a dry run.
func TestSweepActsAtNightOnlyWhenSwitchedOn(t *testing.T) {
	src := readRepoFile(t, sweepWorkflow)
	job := src[strings.Index(src, "\njobs:"):]
	steps := strings.Index(job, "\n    steps:")
	if steps < 0 {
		t.Fatalf("%s job has no steps", sweepWorkflow)
	}
	const gate = "if: github.event_name != 'schedule' || vars.PROPOSAL_REVIEW_SWEEP == 'on'"
	if !strings.Contains(job[:steps], "\n    "+gate+"\n") {
		t.Errorf("%s job does not carry %q", sweepWorkflow, gate)
	}
	const dryRun = "DRY_RUN: ${{ github.event_name == 'schedule' && 'false' || inputs.dry_run }}"
	if !strings.Contains(src, dryRun) {
		t.Errorf("%s does not carry %q", sweepWorkflow, dryRun)
	}
}

func TestSweepPassesTheDryRunInput(t *testing.T) {
	src := readRepoFile(t, sweepWorkflow)
	if !strings.Contains(src, "|| inputs.dry_run }}") {
		t.Errorf("%s does not hand the dry_run input to the step", sweepWorkflow)
	}
	if strings.Contains(src, "Only a dry run") {
		t.Errorf("%s still refuses a dispatch with dry_run false", sweepWorkflow)
	}
	loop := sweepLoop(t)
	gate := loop[strings.Index(loop, reviewGateRun):]
	if !strings.Contains(gate, `--dry-run="$DRY_RUN"`) {
		t.Errorf("the gate call does not pass the dry_run input:\n%s", gate)
	}
}

// TestSweepTakesNoIssueAfterTheWindow runs the window check of the loop with a
// fake clock: a scheduled run takes an issue from 23:00 to 04:59 UTC and none
// after, and a dispatch takes one at any hour.
func TestSweepTakesNoIssueAfterTheWindow(t *testing.T) {
	src := readRepoFile(t, sweepWorkflow)
	if !strings.Contains(src, "EVENT: ${{ github.event_name }}") {
		t.Fatalf("%s does not hand the event name to the step", sweepWorkflow)
	}
	loop := sweepLoop(t)
	start := strings.Index(loop, "HOUR=$(date -u +%H)")
	if start < 0 || start > strings.Index(loop, reviewRun) {
		t.Fatalf("the loop reads no UTC hour before the review:\n%s", loop)
	}
	end := strings.Index(loop[start:], "\n            fi")
	if end < 0 {
		t.Fatalf("the window check has no fi:\n%s", loop[start:])
	}
	check := loop[start : start+end+len("\n            fi")]

	bin := t.TempDir()
	clock := "#!/bin/sh\necho \"$FAKE_HOUR\"\n"
	if err := os.WriteFile(filepath.Join(bin, "date"), []byte(clock), 0o755); err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		event, hour string
		take        bool
	}{
		{"schedule", "23", true},
		{"schedule", "00", true},
		{"schedule", "04", true},
		{"schedule", "05", false},
		{"schedule", "08", false},
		{"schedule", "12", false},
		{"schedule", "22", false},
		{"workflow_dispatch", "12", true},
		{"workflow_dispatch", "05", true},
	}
	for _, tt := range tests {
		script := "for ISSUE in 30; do\n" + check + "\necho took\ndone\n"
		cmd := exec.Command("bash", "-e", "-c", script)
		cmd.Env = append(os.Environ(), "PATH="+bin+":"+os.Getenv("PATH"), "EVENT="+tt.event, "FAKE_HOUR="+tt.hour)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("%s at %s: the window check failed: %v\n%s", tt.event, tt.hour, err, out)
		}
		if took := strings.Contains(string(out), "took"); took != tt.take {
			t.Errorf("%s at %s:00 UTC: took an issue = %v, want %v\n%s", tt.event, tt.hour, took, tt.take, out)
		}
	}
}

// TestSweepPicksItsOwnBatch pins the batch: without an issue list the sweep
// asks minion-review next, with the sample size when the dispatch gives one.
func TestSweepPicksItsOwnBatch(t *testing.T) {
	src := readRepoFile(t, sweepWorkflow)
	for _, want := range []string{
		"SAMPLE: ${{ inputs.sample }}",
		"go run ./cmd/minion-review next)",
		`go run ./cmd/minion-review next --sample "$SAMPLE")`,
	} {
		if !strings.Contains(src, want) {
			t.Errorf("%s does not carry %q", sweepWorkflow, want)
		}
	}
}

// TestSweepRunsAlone pins the concurrency group, so two sweep runs never edit
// the tracking issue at once, and a job timeout that covers the window, 23:00
// to 05:00 UTC, plus the last issue.
func TestSweepRunsAlone(t *testing.T) {
	src := readRepoFile(t, sweepWorkflow)
	on := sweepTriggers(t, src)
	for _, want := range []string{"\nconcurrency:\n  group: proposal-review\n  cancel-in-progress: false\n"} {
		if !strings.Contains(on, want) {
			t.Errorf("%s does not carry the top-level %q", sweepWorkflow, want)
		}
	}
	m := regexp.MustCompile(`(?m)^    timeout-minutes: (\d+)$`).FindStringSubmatch(src)
	if m == nil {
		t.Fatalf("%s job has no timeout-minutes", sweepWorkflow)
	}
	if minutes, _ := strconv.Atoi(m[1]); minutes < 6*60+60 {
		t.Errorf("job timeout is %d minutes, want the 6-hour window plus an hour for the last issue", minutes)
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
