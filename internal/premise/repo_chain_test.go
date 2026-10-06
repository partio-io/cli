package premise

import (
	"regexp"
	"strings"
	"testing"
)

// The research chain sits between the review and the build: an approved issue
// that carries no slice plan, or whose review rewrote it, is researched in the
// same run, and the build starts only once a plan exists. These tests live
// beside the build tests because they share the workflow step parser, and
// because one of the chain's two blocking outcomes is a premise block raised
// inside research.

// researchProgramRef is researchProgram as a workflow file spells it. It is
// derived so the two cannot drift apart.
var researchProgramRef = strings.TrimPrefix(researchProgram, "../../")

// runtimePlanMarker is the pattern partio-minions uses to find the slice-plan
// comment (internal/slices/slices.go, planMarkerRe, pinned at v0.0.13). The
// build workflow must look for the same thing. If the workflow reads "no plan"
// where the runtime reads a plan, research runs again on a planned issue. If
// it reads a plan where the runtime reads none, the build falls back to the
// single unplanned session that this chain exists to remove.
const runtimePlanMarker = `(?m)^\s*<!--\s*minion:research-slices(\s[^>]*)?-->\s*$`

// workflowMarker pulls the single job-level pattern the workflow greps
// comments with.
var workflowMarker = regexp.MustCompile(`SLICE_PLAN_MARKER: '([^']*)'`)

// TestApprovedIssueIsResearchedBeforeItIsBuilt pins the order the chain
// depends on. The plan is looked for, research fills the gap, the result is
// confirmed, and only then does code get written.
func TestApprovedIssueIsResearchedBeforeItIsBuilt(t *testing.T) {
	steps := buildSteps(t)

	order := []struct {
		token string
		what  string
	}{
		{"minions run " + reviewProgramRef, "the review session"},
		{buildGateCall, "the build gate"},
		{"present=true", "the slice-plan check"},
		{researchProgramRef, "the research run"},
		{"refusing to build unplanned", "the confirmation that research produced a plan"},
		{"minions $ARGS", "the build"},
	}

	prev := -1
	for _, want := range order {
		at := stepContaining(steps, want.token)
		if at < 0 {
			t.Fatalf("%s never runs %s", buildWorkflow, want.what)
		}
		if at <= prev {
			t.Errorf("%s runs at step %d, which is not after the step before it (%d)", want.what, at, prev)
		}
		prev = at
	}
}

// TestResearchRunsOnlyWhenTheIssueCarriesNoPlan pins that an issue which was
// already researched is not researched again. A second research run would post
// a duplicate PRD and a duplicate plan: research.md does not skip artifacts it
// has already written.
func TestResearchRunsOnlyWhenTheIssueCarriesNoPlan(t *testing.T) {
	steps := buildSteps(t)

	at := stepContaining(steps, researchProgramRef)
	if at < 0 {
		t.Fatalf("%s never runs %s", buildWorkflow, researchProgramRef)
	}

	const want = "if: steps.slices.outputs.present == 'false'"
	if !strings.Contains(steps[at], want) {
		t.Errorf("the research step is not guarded by %q, so an issue that already carries a plan is researched again", want)
	}
}

// TestRewriteVoidsTheOldPlan pins that a rewrite sends the issue back to
// research. The old plan describes the old text, so when the gate reports a
// change the plan check reads "no plan" whatever the comments hold.
func TestRewriteVoidsTheOldPlan(t *testing.T) {
	steps := buildSteps(t)
	at := stepContaining(steps, "present=true")
	if at < 0 {
		t.Fatalf("%s never looks for a slice plan", buildWorkflow)
	}
	step := steps[at]

	changedAt := strings.Index(step, "steps.gate.outputs.changed")
	grepAt := strings.Index(step, `grep -qE "$SLICE_PLAN_MARKER"`)
	switch {
	case changedAt < 0:
		t.Fatal("the plan check never reads steps.gate.outputs.changed, so a rewritten issue builds from its old plan")
	case grepAt >= 0 && changedAt > grepAt:
		t.Error("the plan check reads the comments before it reads the rewrite, so an old plan still counts")
	}
	branch := step[changedAt:]
	if end := strings.Index(branch, "\n          elif"); end >= 0 {
		branch = branch[:end]
	}
	if !strings.Contains(branch, "present=false") {
		t.Error("a rewrite does not set present=false, so research does not run again")
	}
}

// TestAfterARewriteOnlyANewerPlanCounts pins that the confirmation after
// research accepts only a plan published after the rewrite. The plan check
// records the time once the gate has rewritten the issue, and the
// confirmation reads only the comments created at or after it. Without a
// rewrite the time is empty, and every comment counts.
func TestAfterARewriteOnlyANewerPlanCounts(t *testing.T) {
	steps := buildSteps(t)
	slicesAt := stepContaining(steps, "present=true")
	confirmAt := stepContaining(steps, "refusing to build unplanned")
	if slicesAt < 0 || confirmAt < 0 {
		t.Fatalf("%s lost its plan check or its confirmation", buildWorkflow)
	}

	if !strings.Contains(steps[slicesAt], `echo "since=`) {
		t.Error("the plan check records no rewrite time, so the confirmation cannot tell an old plan from a new one")
	}
	confirm := steps[confirmAt]
	for _, want := range []string{
		"PLAN_SINCE: ${{ steps.slices.outputs.since }}",
		"select(.createdAt >= env.PLAN_SINCE)",
	} {
		if !strings.Contains(confirm, want) {
			t.Errorf("the confirmation does not carry %q, so a plan from before the rewrite still counts", want)
		}
	}
	if strings.Contains(confirm, "'.comments[].body'") {
		t.Error("the confirmation still reads every comment, so a plan from before the rewrite still counts")
	}
}

// TestNoPlanAndNoBlockFailsRatherThanBuildsUnplanned pins the case that makes
// the chain worth having. The runtime falls back to one whole-issue session
// when no plan comment exists, and says nothing on the issue. A research run
// that ends with no plan and no premise block must stop the job red instead.
func TestNoPlanAndNoBlockFailsRatherThanBuildsUnplanned(t *testing.T) {
	steps := buildSteps(t)

	at := stepContaining(steps, "refusing to build unplanned")
	if at < 0 {
		t.Fatalf("%s never confirms that research produced a slice plan", buildWorkflow)
	}
	step := steps[at]

	for _, want := range []struct {
		token string
		why   string
	}{
		{"exit 1", "a run that produced no plan must fail rather than build unplanned"},
		{BlockingLabel, "a premise blocked inside research is a block, not a failure, and must be told apart"},
		{"blocked=true", "a premise blocked inside research must stop the build the way the gate does"},
	} {
		if !strings.Contains(step, want.token) {
			t.Errorf("the confirmation step does not carry %q: %s", want.token, want.why)
		}
	}
}

// TestWorkflowFindsThePlanTheRuntimeFinds pins the two readers together. The
// workflow decides whether to research; the runtime decides whether to slice.
// A disagreement between them is invisible in both repositories.
func TestWorkflowFindsThePlanTheRuntimeFinds(t *testing.T) {
	source := readRepoFile(t, buildWorkflow)

	found := workflowMarker.FindStringSubmatch(source)
	if found == nil {
		t.Fatalf("%s carries no MARKER='...' pattern, so nothing pins it to the runtime's", buildWorkflow)
	}

	// The workflow greps line by line; Go needs (?m) to read the same way.
	workflow, err := regexp.Compile("(?m)" + found[1])
	if err != nil {
		t.Fatalf("the workflow's MARKER does not compile: %v", err)
	}
	runtime := regexp.MustCompile(runtimePlanMarker)

	for _, tc := range []struct {
		name string
		body string
		want bool
	}{
		{"publisher form", "<!-- minion:research-slices parent=#42 -->", true},
		{"bare marker", "<!-- minion:research-slices -->", true},
		{"indented and padded", "   <!-- minion:research-slices -->   ", true},
		{"no inner spaces", "<!--minion:research-slices-->", true},
		{"marker below other lines", "one\ntwo\n<!-- minion:research-slices parent=#7 -->\nfour", true},
		{"prose mention", "A comment about minion:research-slices markers.", false},
		{"marker inside a line", "text <!-- minion:research-slices --> trailing", false},
		// Distinguishes a pattern anchored at the line start from one that
		// matches anywhere: only the anchor rejects this.
		{"marker at the end of a line", "text <!-- minion:research-slices -->", false},
		{"plan heading alone", "## Proposed slices\nno marker here", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := workflow.MatchString(tc.body)
			if got != tc.want {
				t.Errorf("the workflow reads match=%v, want %v", got, tc.want)
			}
			if runtime.MatchString(tc.body) != got {
				t.Errorf("the workflow reads match=%v and the runtime reads match=%v; "+
					"the two readers disagree about what a slice plan is", got, runtime.MatchString(tc.body))
			}
		})
	}
}
