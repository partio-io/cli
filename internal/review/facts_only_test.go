package review

import (
	"bytes"
	"slices"
	"strings"
	"testing"

	"github.com/partio-io/cli/internal/premise"
)

// runBuild runs the gate as a build does and returns its result and the
// step outputs it writes.
func runBuild(t *testing.T, gh *fakeGitHub, issue int, verdict string) (Result, string) {
	t.Helper()
	srv := gh.server(t)
	res, err := Run(Config{
		VerdictPath: writeVerdict(t, verdict),
		Repo:        repo,
		Issue:       issue,
		Night:       "2026-10-06",
		Build:       true,
		APIBaseURL:  srv.URL,
		Token:       "test-token",
		HTTPClient:  srv.Client(),
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	var out bytes.Buffer
	if err := WriteBuildOutputs(&out, res); err != nil {
		t.Fatal(err)
	}
	return res, out.String()
}

// operatorIssue seeds issue 12 as the operator wrote it: no
// minion-proposal label.
func operatorIssue(gh *fakeGitHub, labels ...string) {
	gh.issues[12] = issueJSON(12, "Add a retry to the pre-push hook", "the operator's body", labels)
}

// A failing premise on an operator's issue blocks the build. The gate
// removes and then adds do-not-build, so the label event fires again,
// and posts the marked comment with every claim. It does not close the
// issue, edit it, or change any other label, though the verdict says
// close.
func TestFactsOnlyFailingPremiseBlocksTheBuild(t *testing.T) {
	gh := newFakeGitHub()
	operatorIssue(gh, "minion-approved", "do-not-build")
	gh.trackingIssue(77)

	res, outputs := runBuild(t, gh, 12, closeVerdictFor(ReasonFalsePremise))

	if !res.Valid || !res.FactsOnly || !res.Blocked {
		t.Fatalf("Result = %+v, want a valid facts-only block", res)
	}
	if outputs != "blocked=true\n" {
		t.Errorf("outputs = %q, want blocked=true", outputs)
	}
	want := []string{
		"GET /repos/partio-io/cli/issues/12",
		"DELETE /repos/partio-io/cli/issues/12/labels/do-not-build",
		"POST /repos/partio-io/cli/issues/12/labels",
		"POST /repos/partio-io/cli/issues/12/comments",
		"GET /repos/partio-io/cli/issues",
		"GET /repos/partio-io/cli/issues/77/comments",
		"POST /repos/partio-io/cli/issues/77/comments",
		"PATCH /repos/partio-io/cli/issues/77",
	}
	if !slices.Equal(gh.requests, want) {
		t.Fatalf("requests =\n%s\nwant\n%s", strings.Join(gh.requests, "\n"), strings.Join(want, "\n"))
	}
	if got := labelNames(gh.issues[12]); !slices.Equal(got, []string{"minion-approved", "do-not-build"}) {
		t.Errorf("labels = %v, want minion-approved and do-not-build only", got)
	}
	if gh.issues[12]["state"] != "open" || gh.issues[12]["body"] != "the operator's body" {
		t.Errorf("issue changed: %v", gh.issues[12])
	}
	body := gh.comments[12][0]["body"].(string)
	if !strings.HasPrefix(body, premise.GateCommentMarker+"\n") {
		t.Errorf("comment does not start with %s:\n%s", premise.GateCommentMarker, body)
	}
	for _, want := range []string{"push has no retry", "internal/hooks/prepush.go", "**fails**", "retry(push)"} {
		if !strings.Contains(body, want) {
			t.Errorf("comment lacks %q:\n%s", want, body)
		}
	}
}

// On an operator's issue the gate ignores keep, rewrite and close and
// applies only the premise: a premise that holds, or has no checkable
// claim, proceeds; one that fails or is unresolved blocks. No verdict
// closes the issue, edits its title or body, or changes a label other
// than do-not-build, and every run leaves the marked comment with its
// evidence. A rewrite's shape does not matter: its text is never used.
func TestFactsOnlyAppliesOnlyThePremise(t *testing.T) {
	tests := []struct {
		name    string
		verdict func(t *testing.T) string
		blocked bool
	}{
		{"keep, holds", func(*testing.T) string { return keepVerdict }, false},
		{"rewrite, no claims", func(*testing.T) string { return rewriteVerdict }, false},
		{"rewrite of a bad shape", func(t *testing.T) string { return rewriteVerdictWith(t, "New title", "no premise block") }, false},
		{"close, built", func(*testing.T) string { return closeVerdictFor(ReasonBuilt) }, false},
		{"close, duplicate of a closed issue", func(*testing.T) string { return closeVerdictFor(ReasonDuplicate) }, false},
		{"close, false premise", func(*testing.T) string { return closeVerdictFor(ReasonFalsePremise) }, true},
		{"close, could not verify", func(*testing.T) string { return closeVerdictFor(ReasonCouldNotVerify) }, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gh := newFakeGitHub()
			operatorIssue(gh, "minion-approved")
			gh.issues[13]["state"] = "closed"
			gh.trackingIssue(77)

			res, outputs := runBuild(t, gh, 12, tt.verdict(t))

			if !res.Valid || !res.FactsOnly || res.Blocked != tt.blocked {
				t.Fatalf("Result = %+v, want a valid facts-only check with Blocked %v", res, tt.blocked)
			}
			wantOut, wantLabels := "blocked=false\nchanged=false\n", []string{"minion-approved"}
			if tt.blocked {
				wantOut, wantLabels = "blocked=true\n", []string{"minion-approved", "do-not-build"}
			}
			if outputs != wantOut {
				t.Errorf("outputs = %q, want %q", outputs, wantOut)
			}
			var onIssue []string
			for _, r := range requestsTo(gh.requests, "/repos/partio-io/cli/issues/12") {
				if !strings.HasPrefix(r, "GET ") {
					onIssue = append(onIssue, r)
				}
			}
			want := []string{"POST /repos/partio-io/cli/issues/12/comments"}
			if tt.blocked {
				want = append([]string{
					"DELETE /repos/partio-io/cli/issues/12/labels/do-not-build",
					"POST /repos/partio-io/cli/issues/12/labels",
				}, want...)
			}
			if !slices.Equal(onIssue, want) {
				t.Errorf("writes to #12 =\n%s\nwant\n%s", strings.Join(onIssue, "\n"), strings.Join(want, "\n"))
			}
			if got := labelNames(gh.issues[12]); !slices.Equal(got, wantLabels) {
				t.Errorf("labels = %v, want %v", got, wantLabels)
			}
			issue := gh.issues[12]
			if issue["state"] != "open" || issue["title"] != "Add a retry to the pre-push hook" || issue["body"] != "the operator's body" {
				t.Errorf("issue changed: %v", issue)
			}
			body := gh.comments[12][0]["body"].(string)
			if !strings.HasPrefix(body, premise.GateCommentMarker+"\n") {
				t.Errorf("comment does not start with %s:\n%s", premise.GateCommentMarker, body)
			}
			result := "The build proceeds."
			if tt.blocked {
				result = "The build stops"
			}
			if !strings.Contains(body, result) {
				t.Errorf("comment does not say %q:\n%s", result, body)
			}
		})
	}
}

// A reopen after the review's close is the operator's overrule: the
// next run checks facts only, though the issue is still a proposal. The
// close comment is the one the gate itself posts. A reopen that comes
// before the review's close, or on an issue the review kept, does not
// count.
func TestFactsOnlyAfterAReopen(t *testing.T) {
	// closeComment is the body the gate posts when it closes #12.
	gh := newFakeGitHub()
	gh.labels = append(gh.labels, ReviewedLabel)
	gh.trackingIssue(77)
	runAct(t, gh, 12, writeVerdict(t, closeVerdictFor(ReasonDoesNotApply)))
	closeComment := gh.comments[12][0]["body"].(string)
	keepComment := strings.Replace(closeComment, "## Proposal review: close", "## Proposal review: keep", 1)

	comment := func(body, created, updated string) map[string]any {
		return map[string]any{"event": "commented", "id": 600, "body": body, "created_at": created, "updated_at": updated}
	}
	event := func(kind, at string) map[string]any {
		return map[string]any{"event": kind, "created_at": at}
	}
	tests := []struct {
		name      string
		timeline  []map[string]any
		factsOnly bool
	}{
		{"no timeline", nil, false},
		{"reopened after the close", []map[string]any{
			comment(closeComment, "2026-10-01T10:00:00Z", "2026-10-01T10:00:00Z"),
			event("closed", "2026-10-01T10:00:01Z"),
			event("reopened", "2026-10-02T09:00:00Z"),
		}, true},
		{"a keep comment edited into a close, then reopened", []map[string]any{
			comment(closeComment, "2026-09-20T10:00:00Z", "2026-10-01T10:00:00Z"),
			event("closed", "2026-10-01T10:00:01Z"),
			event("reopened", "2026-10-02T09:00:00Z"),
		}, true},
		{"reopened before the review closed it", []map[string]any{
			event("closed", "2026-09-01T10:00:00Z"),
			event("reopened", "2026-09-02T10:00:00Z"),
			comment(closeComment, "2026-10-01T10:00:00Z", "2026-10-01T10:00:00Z"),
		}, false},
		{"reopened, but the review kept it", []map[string]any{
			comment(keepComment, "2026-10-01T10:00:00Z", "2026-10-01T10:00:00Z"),
			event("closed", "2026-10-01T11:00:00Z"),
			event("reopened", "2026-10-02T09:00:00Z"),
		}, false},
		{"a close comment from another tool", []map[string]any{
			comment("## Proposal review: close", "2026-10-01T10:00:00Z", "2026-10-01T10:00:00Z"),
			event("reopened", "2026-10-02T09:00:00Z"),
		}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gh := newFakeGitHub()
			gh.labels = append(gh.labels, ReviewedLabel)
			gh.issues[12] = issueJSON(12, "Add a retry to the pre-push hook", "body",
				[]string{proposalLabel, ReviewedLabel})
			gh.timeline[12] = tt.timeline
			gh.trackingIssue(77)

			res, outputs := runBuild(t, gh, 12, closeVerdictFor(ReasonFalsePremise))

			if !res.Valid || res.FactsOnly != tt.factsOnly || !res.Blocked {
				t.Fatalf("Result = %+v, want a valid block with FactsOnly %v", res, tt.factsOnly)
			}
			if outputs != "blocked=true\n" {
				t.Errorf("outputs = %q, want blocked=true", outputs)
			}
			if !slices.Contains(gh.requests, "GET /repos/partio-io/cli/issues/12/timeline") {
				t.Errorf("the gate never read the timeline of #12: %v", gh.requests)
			}
			wantState, wantMarker := "closed", ReviewMarker
			if tt.factsOnly {
				wantState, wantMarker = "open", premise.GateCommentMarker
			}
			if gh.issues[12]["state"] != wantState {
				t.Errorf("state = %v, want %s", gh.issues[12]["state"], wantState)
			}
			if body := gh.comments[12][0]["body"].(string); !strings.HasPrefix(body, wantMarker) {
				t.Errorf("comment does not start with %s:\n%s", wantMarker, body)
			}
		})
	}
}

// The sweep never selects an operator's issue, but a dispatch list can
// name one. The sweep then checks facts only too: a close or a rewrite
// does not close it, edit it, or mark it reviewed.
func TestSweepChecksFactsOnlyOnAnOperatorIssue(t *testing.T) {
	tests := []struct {
		name    string
		verdict string
		blocked bool
	}{
		{"close, does not apply", closeVerdictFor(ReasonDoesNotApply), false},
		{"close, false premise", closeVerdictFor(ReasonFalsePremise), true},
		{"rewrite", rewriteVerdict, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gh := newFakeGitHub()
			operatorIssue(gh)
			tracking := gh.trackingIssue(77)

			res := runAct(t, gh, 12, writeVerdict(t, tt.verdict))

			if !res.Valid || !res.FactsOnly || res.Blocked != tt.blocked {
				t.Fatalf("Result = %+v, want a valid facts-only check with Blocked %v", res, tt.blocked)
			}
			issue := gh.issues[12]
			if issue["state"] != "open" || issue["title"] != "Add a retry to the pre-push hook" || issue["body"] != "the operator's body" {
				t.Errorf("the sweep changed the operator's issue: %v", issue)
			}
			if slices.Contains(labelNames(issue), ReviewedLabel) {
				t.Errorf("the sweep marked the operator's issue reviewed: %v", labelNames(issue))
			}
			if body := gh.comments[12][0]["body"].(string); !strings.HasPrefix(body, premise.GateCommentMarker) {
				t.Errorf("comment does not start with %s:\n%s", premise.GateCommentMarker, body)
			}
			if row := gh.comments[tracking][0]["body"].(string); strings.Contains(row, "(build)") || strings.Contains(row, "(dry run)") {
				t.Errorf("sweep row carries a mark:\n%s", row)
			}
		})
	}
}

// TestBlockedBuildNeverClosesTheIssue is the August "never closes" test,
// now for the issues the review checks facts only on: the operator's,
// and the ones the operator reopened. A premise that blocks stops the
// build and leaves the issue open with no minion-done, whatever the
// verdict's outcome. The workflow half, that no step marks a blocked
// build done or failed, is TestBlockedBuildIsNeitherDoneNorFailed in
// the premise package.
func TestBlockedBuildNeverClosesTheIssue(t *testing.T) {
	for _, reason := range []string{ReasonFalsePremise, ReasonCouldNotVerify} {
		t.Run(reason, func(t *testing.T) {
			gh := newFakeGitHub()
			operatorIssue(gh)
			gh.trackingIssue(77)

			res, outputs := runBuild(t, gh, 12, closeVerdictFor(reason))

			if !res.FactsOnly || outputs != "blocked=true\n" {
				t.Fatalf("Result = %+v, outputs %q, want a facts-only block", res, outputs)
			}
			if gh.issues[12]["state"] != "open" {
				t.Error("a facts-only block closed the issue; the gate reports, the operator decides")
			}
			if slices.Contains(labelNames(gh.issues[12]), "minion-done") {
				t.Error("a facts-only block marked the issue done")
			}
			for _, r := range gh.requests {
				if r == "PATCH /repos/partio-io/cli/issues/12" {
					t.Errorf("a facts-only block changed the issue: %s", r)
				}
			}
		})
	}
}

// TestOperatorOverrulesTheBuildByRemovingTheLabel is the August overrule
// test, now for a facts-only block. The gate never reads do-not-build
// as a decision: each run verifies again and reaches its own verdict. So
// an operator who removes the label, or corrects the issue text, and
// runs the build again gets a fresh verdict, not the remembered block.
func TestOperatorOverrulesTheBuildByRemovingTheLabel(t *testing.T) {
	gh := newFakeGitHub()
	operatorIssue(gh)
	gh.trackingIssue(77)

	if _, outputs := runBuild(t, gh, 12, closeVerdictFor(ReasonFalsePremise)); outputs != "blocked=true\n" {
		t.Fatalf("first run outputs = %q, want blocked=true", outputs)
	}
	if !slices.Contains(labelNames(gh.issues[12]), premise.BlockingLabel) {
		t.Fatalf("the block did not label the issue %s", premise.BlockingLabel)
	}

	// The operator overrules: remove the label and run the build again.
	gh.issues[12]["labels"] = []map[string]any{}
	gh.requests = nil
	res, outputs := runBuild(t, gh, 12, keepVerdict)

	if !res.FactsOnly || outputs != "blocked=false\nchanged=false\n" {
		t.Fatalf("Result = %+v, outputs %q, want a facts-only pass", res, outputs)
	}
	if got := labelNames(gh.issues[12]); len(got) != 0 {
		t.Errorf("labels after the overrule = %v, want none", got)
	}
	if n := len(gh.comments[12]); n != 2 {
		t.Errorf("%d comments on #12, want one per run: the pass leaves its evidence too", n)
	}
}

// A claim that fails or is unresolved blocks a facts-only issue, though
// the premise verdict says holds: the verdict check does not tie the
// premise verdict to its claims.
func TestPremiseBlocksOnAnyClaim(t *testing.T) {
	tests := []struct {
		name    string
		premise Premise
		want    bool
	}{
		{"holds, every claim holds", Premise{Verdict: Holds, Claims: []Claim{{Verdict: Holds}}}, false},
		{"no claims", Premise{Verdict: NoClaims}, false},
		{"fails", Premise{Verdict: Fails, Claims: []Claim{{Verdict: Fails}}}, true},
		{"unresolved", Premise{Verdict: Unresolved, Claims: []Claim{{Verdict: Unresolved}}}, true},
		{"holds with a failing claim", Premise{Verdict: Holds, Claims: []Claim{{Verdict: Holds}, {Verdict: Fails}}}, true},
		{"holds with an unresolved claim", Premise{Verdict: Holds, Claims: []Claim{{Verdict: Unresolved}}}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := premiseBlocks(Verdict{Premise: tt.premise}); got != tt.want {
				t.Errorf("premiseBlocks = %t, want %t", got, tt.want)
			}
		})
	}
}
