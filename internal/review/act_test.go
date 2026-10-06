package review

import (
	"slices"
	"strings"
	"testing"
)

func labelNames(issue map[string]any) []string {
	var out []string
	for _, l := range issue["labels"].([]map[string]any) {
		out = append(out, l["name"].(string))
	}
	return out
}

// A keep posts one evidence comment, drops the build labels the issue
// carries, adds minion-reviewed, and writes its row. Labels the issue
// does not carry cost no request.
func TestGateKeepActsOnTheIssue(t *testing.T) {
	gh := newFakeGitHub()
	gh.labels = append(gh.labels, ReviewedLabel)
	gh.issues[12] = issueJSON(12, "Add a retry to the pre-push hook", "body",
		[]string{"minion-proposal", "minion-approved", "do-not-build"})
	tracking := gh.trackingIssue(77)

	res := runAct(t, gh, 12, writeVerdict(t, keepVerdict))

	if !res.Valid || res.Outcome != OutcomeKeep {
		t.Fatalf("Result = %+v, want a valid keep", res)
	}
	want := []string{
		"GET /repos/partio-io/cli/issues/12",
		"GET /repos/partio-io/cli/issues/12/timeline",
		"GET /repos/partio-io/cli/pulls",
		"GET /repos/partio-io/cli/labels/minion-reviewed",
		"GET /repos/partio-io/cli/issues/12/comments",
		"POST /repos/partio-io/cli/issues/12/comments",
		"DELETE /repos/partio-io/cli/issues/12/labels/minion-approved",
		"DELETE /repos/partio-io/cli/issues/12/labels/do-not-build",
		"POST /repos/partio-io/cli/issues/12/labels",
		"GET /repos/partio-io/cli/issues",
		"GET /repos/partio-io/cli/issues/77/comments",
		"POST /repos/partio-io/cli/issues/77/comments",
		"PATCH /repos/partio-io/cli/issues/77",
	}
	if !slices.Equal(gh.requests, want) {
		t.Fatalf("requests:\n%s\nwant:\n%s", strings.Join(gh.requests, "\n"), strings.Join(want, "\n"))
	}
	if got, want := labelNames(gh.issues[12]), []string{"minion-proposal", ReviewedLabel}; !slices.Equal(got, want) {
		t.Errorf("labels = %q, want %q", got, want)
	}
	if got := len(gh.comments[12]); got != 1 {
		t.Fatalf("reviewed issue has %d comments, want 1", got)
	}
	if gh.issues[12]["state"] != "open" {
		t.Errorf("a keep changed the state to %v", gh.issues[12]["state"])
	}
	row := gh.comments[tracking][0]["body"].(string)
	if !strings.Contains(row, "**keep** ·") || strings.Contains(row, "dry run") {
		t.Errorf("tracking row is not an acted keep:\n%s", row)
	}
}

// closeVerdictFor is a valid close verdict for issue 12 with reason.
func closeVerdictFor(reason string) string {
	claims, premise := `[]`, "no-claims"
	switch reason {
	case ReasonFalsePremise:
		claims, premise = `[{"claim": "push has no retry", "evidence": "internal/hooks/prepush.go", "verdict": "fails", "excerpt": "retry(push)", "correction": "internal/hooks/retry.go already retries the push once"}]`, "fails"
	case ReasonCouldNotVerify:
		claims, premise = `[{"claim": "pushes fail often", "evidence": "no push logs in the repo", "verdict": "unresolved", "excerpt": "(no log files)"}]`, "unresolved"
	}
	fit := `{"applies": true, "reason": "in scope"}`
	if reason == ReasonDoesNotApply {
		fit = `{"applies": false, "reason": "partio has no server"}`
	}
	built := `{"built": false, "evidence": ""}`
	if reason == ReasonBuilt {
		built = `{"built": true, "evidence": "internal/hooks/retry.go already retries"}`
	}
	dupOf, dupes := "", `[]`
	if reason == ReasonDuplicate {
		dupOf, dupes = `"duplicate_of": 13,`, `[{"issue": 13, "same": true, "why": "same retry idea"}]`
	}
	return `{"issue": 12, "outcome": "close", "close_reason": "` + reason + `", ` + dupOf +
		`"premise": {"verdict": "` + premise + `", "claims": ` + claims + `},` +
		`"fit": ` + fit + `, "built": ` + built + `, "duplicates": ` + dupes + `}`
}

// A close posts the evidence comment, adds minion-reviewed and closes
// the issue: completed for built, not_planned for every other reason.
// A duplicate close reads the issue that stays, to name and link it.
func TestGateCloseActsOnTheIssue(t *testing.T) {
	tests := []struct {
		reason      string
		stateReason string
	}{
		{ReasonBuilt, "completed"},
		{ReasonFalsePremise, "not_planned"},
		{ReasonDoesNotApply, "not_planned"},
		{ReasonDuplicate, "not_planned"},
		{ReasonCouldNotVerify, "not_planned"},
	}
	for _, tt := range tests {
		t.Run(tt.reason, func(t *testing.T) {
			gh := newFakeGitHub()
			gh.labels = append(gh.labels, ReviewedLabel)
			gh.issues[12] = issueJSON(12, "Add a retry to the pre-push hook", "body",
				[]string{"minion-proposal", "minion-approved"})
			gh.trackingIssue(77)

			res := runAct(t, gh, 12, writeVerdict(t, closeVerdictFor(tt.reason)))

			if !res.Valid || res.Outcome != OutcomeClose {
				t.Fatalf("Result = %+v, want a valid close", res)
			}
			want := []string{"GET /repos/partio-io/cli/issues/12", "GET /repos/partio-io/cli/issues/12/timeline"}
			if tt.reason == ReasonDuplicate {
				want = append(want, "GET /repos/partio-io/cli/issues/13")
			}
			want = append(want,
				"GET /repos/partio-io/cli/pulls",
				"GET /repos/partio-io/cli/labels/minion-reviewed",
				"GET /repos/partio-io/cli/issues/12/comments",
				"POST /repos/partio-io/cli/issues/12/comments",
				"POST /repos/partio-io/cli/issues/12/labels",
				"PATCH /repos/partio-io/cli/issues/12",
				"GET /repos/partio-io/cli/issues",
				"GET /repos/partio-io/cli/issues/77/comments",
				"POST /repos/partio-io/cli/issues/77/comments",
				"PATCH /repos/partio-io/cli/issues/77",
			)
			if !slices.Equal(gh.requests, want) {
				t.Fatalf("requests:\n%s\nwant:\n%s", strings.Join(gh.requests, "\n"), strings.Join(want, "\n"))
			}
			issue := gh.issues[12]
			if issue["state"] != "closed" || issue["state_reason"] != tt.stateReason {
				t.Errorf("state = %v/%v, want closed/%s", issue["state"], issue["state_reason"], tt.stateReason)
			}
			if !slices.Contains(labelNames(issue), ReviewedLabel) {
				t.Errorf("labels = %q, want %s", labelNames(issue), ReviewedLabel)
			}
			if got := len(gh.comments[12]); got != 1 {
				t.Fatalf("reviewed issue has %d comments, want 1", got)
			}
			body := gh.comments[12][0]["body"].(string)
			if tt.reason == ReasonDuplicate {
				link := "[Cache the session index](https://github.com/partio-io/cli/issues/13)"
				if !strings.Contains(body, "#13") || !strings.Contains(body, link) {
					t.Errorf("duplicate comment does not name and link #13:\n%s", body)
				}
			}
		})
	}
}

// The evidence comment opens with the marker and carries every claim
// with its evidence, verdict and excerpt, and the fit, built and
// duplicate decisions.
func TestGateEvidenceCommentCarriesTheVerdict(t *testing.T) {
	gh := newFakeGitHub()
	gh.trackingIssue(77)
	verdict := `{
	"issue": 12,
	"outcome": "keep",
	"premise": {"verdict": "holds", "claims": [
		{"claim": "the pre-push hook has no retry", "evidence": "internal/hooks/prepush.go", "verdict": "holds", "excerpt": "return push(remote)"},
		{"claim": "pushes run over https", "evidence": "internal/git/push.go", "verdict": "holds", "excerpt": "url := \"https://\" + host"}
	]},
	"fit": {"applies": true, "reason": "pre-push is in scope"},
	"built": {"built": false, "evidence": "no retry in internal/hooks"},
	"duplicates": [{"issue": 13, "same": false, "why": "caches sessions, not pushes"}]
}`

	runAct(t, gh, 12, writeVerdict(t, verdict))

	body := gh.comments[12][0]["body"].(string)
	if !strings.HasPrefix(body, "<!-- partio:review:v1 -->\n") {
		t.Errorf("comment does not start with the review marker:\n%s", body)
	}
	for _, want := range []string{
		"keep", "pre-push is in scope",
		"the pre-push hook has no retry", "internal/hooks/prepush.go", "return push(remote)",
		"pushes run over https", "internal/git/push.go", `url := "https://" + host`,
		"**holds**",
		"**Fit:** applies",
		"**Built:** no · no retry in internal/hooks",
		"#13 · different idea · caches sessions, not pushes",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("comment missing %q:\n%s", want, body)
		}
	}
}

// An open pull request from an older build of the issue is named in the
// comment, for a keep and for a close. The gate only lists pull
// requests: it never writes to one.
func TestGateNamesTheOlderBuildPullRequest(t *testing.T) {
	for name, verdict := range map[string]string{"keep": keepVerdict, "close": closeVerdictFor(ReasonBuilt)} {
		t.Run(name, func(t *testing.T) {
			gh := newFakeGitHub()
			gh.trackingIssue(77)
			gh.pulls = []map[string]any{
				{"number": 45, "state": "open", "head": "partio-io:minion/implement-implement-12",
					"title": "Add a retry to the pre-push hook", "html_url": "https://github.com/partio-io/cli/pull/45"},
				{"number": 46, "state": "open", "head": "partio-io:minion/implement-implement-13",
					"title": "Cache the session index", "html_url": "https://github.com/partio-io/cli/pull/46"},
			}

			runAct(t, gh, 12, writeVerdict(t, verdict))

			body := gh.comments[12][0]["body"].(string)
			if !strings.Contains(body, "#45") || !strings.Contains(body, "https://github.com/partio-io/cli/pull/45") {
				t.Errorf("comment does not name pull request #45:\n%s", body)
			}
			if strings.Contains(body, "#46") {
				t.Errorf("comment names #46, the build of another issue:\n%s", body)
			}
			for _, r := range gh.requests {
				method, path, _ := strings.Cut(r, " ")
				if method == "GET" || path == "/repos/partio-io/cli/labels" {
					continue
				}
				// A pull request is also an issue: only #12 and the
				// tracking issue may take a write.
				if !strings.HasPrefix(path, "/repos/partio-io/cli/issues/12/") && path != "/repos/partio-io/cli/issues/12" &&
					!strings.HasPrefix(path, "/repos/partio-io/cli/issues/77/") && path != "/repos/partio-io/cli/issues/77" {
					t.Errorf("gate wrote outside the reviewed and tracking issues: %s", r)
				}
			}
		})
	}
}

// The first acting run on a repository without minion-reviewed creates
// it; the next run finds it and creates nothing.
func TestGateCreatesTheReviewedLabelOnce(t *testing.T) {
	gh := newFakeGitHub()
	gh.trackingIssue(77)

	runAct(t, gh, 12, writeVerdict(t, keepVerdict))
	first := requestsTo(gh.requests, "/repos/partio-io/cli/labels")
	gh.requests = nil
	runAct(t, gh, 13, writeVerdict(t, strings.Replace(keepVerdict, `"issue": 12`, `"issue": 13`, 1)))
	second := requestsTo(gh.requests, "/repos/partio-io/cli/labels")

	wantFirst := []string{"GET /repos/partio-io/cli/labels/minion-reviewed", "POST /repos/partio-io/cli/labels"}
	if !slices.Equal(first, wantFirst) {
		t.Errorf("first run label requests = %q, want %q", first, wantFirst)
	}
	wantSecond := []string{"GET /repos/partio-io/cli/labels/minion-reviewed"}
	if !slices.Equal(second, wantSecond) {
		t.Errorf("second run label requests = %q, want %q", second, wantSecond)
	}
	if n := strings.Count(strings.Join(gh.labels, ","), ReviewedLabel); n != 1 {
		t.Errorf("repository has %d %s labels, want 1", n, ReviewedLabel)
	}
}

// A valid rewrite edits the title and the body in place with one
// request, posts the evidence comment and applies the keep label rules.
func TestGateRewriteActsOnTheIssue(t *testing.T) {
	gh := newFakeGitHub()
	gh.labels = append(gh.labels, ReviewedLabel)
	gh.issues[12] = issueJSON(12, "Add a retry to the pre-push hook", "body",
		[]string{"minion-proposal", "minion-failed"})
	tracking := gh.trackingIssue(77)

	res := runAct(t, gh, 12, writeVerdict(t, rewriteVerdict))

	if !res.Valid || res.Outcome != OutcomeRewrite {
		t.Fatalf("Result = %+v, want a valid rewrite", res)
	}
	want := []string{
		"GET /repos/partio-io/cli/issues/12",
		"GET /repos/partio-io/cli/issues/12/timeline",
		"GET /repos/partio-io/cli/pulls",
		"GET /repos/partio-io/cli/labels/minion-reviewed",
		"GET /repos/partio-io/cli/issues/12/comments",
		"POST /repos/partio-io/cli/issues/12/comments",
		"PATCH /repos/partio-io/cli/issues/12",
		"DELETE /repos/partio-io/cli/issues/12/labels/minion-failed",
		"POST /repos/partio-io/cli/issues/12/labels",
		"GET /repos/partio-io/cli/issues",
		"GET /repos/partio-io/cli/issues/77/comments",
		"POST /repos/partio-io/cli/issues/77/comments",
		"PATCH /repos/partio-io/cli/issues/77",
	}
	if !slices.Equal(gh.requests, want) {
		t.Fatalf("requests:\n%s\nwant:\n%s", strings.Join(gh.requests, "\n"), strings.Join(want, "\n"))
	}
	issue := gh.issues[12]
	if issue["title"] != "Retry the pre-push once on a network error" {
		t.Errorf("title = %q, want the rewrite title", issue["title"])
	}
	if body := issue["body"].(string); !strings.HasPrefix(body, "## What\n") || !strings.HasSuffix(body, "Proposal id: retry-pre-push") {
		t.Errorf("body is not the rewrite body:\n%s", body)
	}
	if issue["state"] != "open" {
		t.Errorf("a rewrite changed the state to %v", issue["state"])
	}
	if got, want := labelNames(issue), []string{"minion-proposal", ReviewedLabel}; !slices.Equal(got, want) {
		t.Errorf("labels = %q, want %q", got, want)
	}
	if row := gh.comments[tracking][0]["body"].(string); !strings.Contains(row, "**rewrite** ·") || strings.Contains(row, "dry run") {
		t.Errorf("tracking row is not an acted rewrite:\n%s", row)
	}
}

// The evidence comment of a rewrite lists each change the review made,
// one item per change.
func TestGateRewriteCommentListsTheChanges(t *testing.T) {
	gh := newFakeGitHub()
	gh.trackingIssue(77)

	runAct(t, gh, 12, writeVerdict(t, rewriteVerdict))

	body := gh.comments[12][0]["body"].(string)
	want := "### What changed\n\n- narrowed to network errors\n- dropped the config flag\n"
	if !strings.Contains(body, want) {
		t.Errorf("comment does not list the changes %q:\n%s", want, body)
	}
}

// A false-premise close shows the correction under each failing claim,
// and tells the operator how to keep the idea: reopen the issue and
// correct its text.
func TestGateFalsePremiseCloseShowsTheCorrection(t *testing.T) {
	gh := newFakeGitHub()
	gh.trackingIssue(77)

	res := runAct(t, gh, 12, writeVerdict(t, closeVerdictFor(ReasonFalsePremise)))

	if !res.Valid || res.Outcome != OutcomeClose {
		t.Fatalf("Result = %+v, want a valid close", res)
	}
	body := gh.comments[12][0]["body"].(string)
	for _, want := range []string{
		"### Premise: fails\n\nTo keep the idea, reopen the issue and correct its text.",
		"- **fails** · push has no retry\n  - Correction: internal/hooks/retry.go already retries the push once\n",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("comment does not carry %q:\n%s", want, body)
		}
	}
}

// Every acted outcome writes its row, without the dry-run mark, to the
// night comment of the tracking issue.
func TestGateEveryActionWritesItsRow(t *testing.T) {
	for _, reason := range []string{"", ReasonBuilt, ReasonFalsePremise, ReasonDoesNotApply, ReasonDuplicate, ReasonCouldNotVerify} {
		verdict, outcome, name := keepVerdict, OutcomeKeep, OutcomeKeep
		if reason != "" {
			verdict, outcome, name = closeVerdictFor(reason), OutcomeClose, "close-"+reason
		}
		t.Run(name, func(t *testing.T) {
			gh := newFakeGitHub()
			tracking := gh.trackingIssue(77)

			runAct(t, gh, 12, writeVerdict(t, verdict))

			if got := len(gh.comments[tracking]); got != 1 {
				t.Fatalf("tracking issue has %d comments, want 1", got)
			}
			body := gh.comments[tracking][0]["body"].(string)
			if !strings.HasPrefix(body, NightMarker("2026-10-06")) {
				t.Errorf("night comment does not start with its marker:\n%s", body)
			}
			if !strings.Contains(body, "- #12 · ") || !strings.Contains(body, "**"+outcome+"** · ") || strings.Contains(body, "dry run") {
				t.Errorf("row is not an acted %s for #12:\n%s", outcome, body)
			}
		})
	}
}

// Outside dry-run the gate refuses a verdict it cannot act on: a closed
// issue, a pull request, or a duplicate close whose kept issue is
// closed or absent. It writes a no-verdict row and changes nothing.
func TestGateRefusesWhatItCannotActOn(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(gh *fakeGitHub)
		verdict string
		cause   string
	}{
		{"closed issue", func(gh *fakeGitHub) { gh.issues[12]["state"] = "closed" }, keepVerdict, "open issues only"},
		{"pull request", func(gh *fakeGitHub) { gh.issues[12]["pull_request"] = map[string]any{} }, keepVerdict, "pull request"},
		{"closed kept issue", func(gh *fakeGitHub) { gh.issues[13]["state"] = "closed" }, closeVerdictFor(ReasonDuplicate), "#13 is not an open issue"},
		{"absent kept issue", func(gh *fakeGitHub) { delete(gh.issues, 13) }, closeVerdictFor(ReasonDuplicate), "#13 does not exist"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gh := newFakeGitHub()
			tracking := gh.trackingIssue(77)
			tt.setup(gh)

			res := runAct(t, gh, 12, writeVerdict(t, tt.verdict))

			if res.Valid || !strings.Contains(res.Cause, tt.cause) {
				t.Fatalf("Result = %+v, want no verdict with %q", res, tt.cause)
			}
			for _, r := range gh.requests {
				method, path, _ := strings.Cut(r, " ")
				if method != "GET" && !strings.HasPrefix(path, "/repos/partio-io/cli/issues/77/") && path != "/repos/partio-io/cli/issues/77" {
					t.Errorf("gate changed something it refused: %s", r)
				}
			}
			if row := gh.comments[tracking][0]["body"].(string); !strings.Contains(row, "**no verdict**") {
				t.Errorf("tracking row is not a no-verdict row:\n%s", row)
			}
		})
	}
}

// A second acting run on the same issue, after a run that failed past
// the comment, updates the evidence comment and does not post another.
func TestGateKeepsOneEvidenceComment(t *testing.T) {
	gh := newFakeGitHub()
	gh.trackingIssue(77)
	gh.comments[12] = []map[string]any{{"id": int64(9), "body": ReviewMarker + "\nolder run"}}

	runAct(t, gh, 12, writeVerdict(t, keepVerdict))

	if got := len(gh.comments[12]); got != 1 {
		t.Fatalf("reviewed issue has %d comments, want 1", got)
	}
	if body := gh.comments[12][0]["body"].(string); strings.Contains(body, "older run") {
		t.Errorf("evidence comment was not updated:\n%s", body)
	}
}

// Excerpts that push the evidence comment past GitHub's size limit are
// left out, and the comment says so.
func TestEvidenceCommentStaysUnderTheSizeLimit(t *testing.T) {
	v := Verdict{Outcome: OutcomeKeep, Premise: Premise{Verdict: Holds, Claims: []Claim{
		{Claim: "big", Evidence: "a.go", Verdict: Holds, Excerpt: strings.Repeat("x", maxCommentBody)},
	}}}

	body := evidenceComment(v, nil, nil)

	if len(body) > maxCommentBody {
		t.Fatalf("comment is %d bytes, over %d", len(body), maxCommentBody)
	}
	if !strings.Contains(body, "Excerpts left out") || !strings.Contains(body, "a.go") {
		t.Errorf("comment does not keep the evidence and say the excerpts are out:\n%s", body[:200])
	}
}
