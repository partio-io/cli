package review

import (
	"slices"
	"strings"
	"testing"
)

const (
	closeVerdict = `{
	"issue": 12,
	"outcome": "close",
	"close_reason": "built",
	"premise": {"verdict": "no-claims", "claims": []},
	"fit": {"applies": true, "reason": "in scope"},
	"built": {"built": true, "evidence": "internal/hooks/retry.go already retries"},
	"duplicates": []
}`
	rewriteVerdict = `{
	"issue": 12,
	"outcome": "rewrite",
	"premise": {"verdict": "no-claims", "claims": []},
	"fit": {"applies": true, "reason": "in scope"},
	"built": {"built": false, "evidence": ""},
	"duplicates": [],
	"rewrite": {
		"title": "Retry the pre-push once on a network error",
		"body": "## What\n\n- retry once\n- log the retry\n\n## Why\n\nFlaky pushes fail the hook.\n\n## Premise\n\n<!-- partio:premise:v1 -->\n\n- the pre-push hook has no retry [evidence: ` + "`internal/hooks/prepush.go`" + `]\n\n## Acceptance Criteria\n\n- [ ] a push that fails on a network error runs once more\n\nProposal id: retry-pre-push",
		"changes": ["narrowed to network errors", "dropped the config flag"]
	}
}`
)

func requestsTo(requests []string, path string) []string {
	var out []string
	for _, r := range requests {
		if _, p, _ := strings.Cut(r, " "); p == path || strings.HasPrefix(p, path+"/") {
			out = append(out, r)
		}
	}
	return out
}

// With the tracking issue in place, every outcome costs one read of the
// reviewed issue and one comment write on the tracking issue: no
// comment, label, edit or close on the reviewed issue.
func TestGateDryRunChangesNothingOnTheReviewedIssue(t *testing.T) {
	for name, verdict := range map[string]string{"keep": keepVerdict, "close": closeVerdict, "rewrite": rewriteVerdict} {
		t.Run(name, func(t *testing.T) {
			gh := newFakeGitHub()
			tracking := gh.trackingIssue(77)

			res := runGate(t, gh, 12, writeVerdict(t, verdict), "2026-10-06")

			if !res.Valid || res.Outcome != name {
				t.Fatalf("Result = %+v, want a valid %s", res, name)
			}
			want := []string{
				"GET /repos/partio-io/cli/issues/12",
				"GET /repos/partio-io/cli/issues",
				"GET /repos/partio-io/cli/issues/77/comments",
				"POST /repos/partio-io/cli/issues/77/comments",
				"PATCH /repos/partio-io/cli/issues/77",
			}
			if !slices.Equal(gh.requests, want) {
				t.Fatalf("requests:\n%s\nwant:\n%s", strings.Join(gh.requests, "\n"), strings.Join(want, "\n"))
			}
			if got := len(gh.comments[tracking]); got != 1 {
				t.Errorf("tracking issue has %d comments, want 1", got)
			}
			if got := len(gh.comments[12]); got != 0 {
				t.Errorf("reviewed issue has %d comments, want 0", got)
			}
		})
	}
}

// The tracking issue is the open labeled issue whose body starts with
// the marker. A same-titled issue, or one with the marker further down,
// is not it.
func TestGateFindsTrackingIssueByLabelAndMarker(t *testing.T) {
	gh := newFakeGitHub()
	gh.labels = append(gh.labels, TrackingLabel)
	gh.issues[60] = issueJSON(60, "Minion review log", "a human copy\n"+TrackingMarker, []string{TrackingLabel})
	gh.issues[61] = issueJSON(61, "Minion review log", TrackingMarker+"\nno label", nil)
	gh.issues[62] = issueJSON(62, "Something else", TrackingMarker+"\nthe real one", []string{TrackingLabel})
	// The issues API lists pull requests too; a labeled one with the
	// marker is still not the tracking issue.
	gh.issues[59] = issueJSON(59, "Minion review log", TrackingMarker+"\na pull request", []string{TrackingLabel})
	gh.issues[59]["pull_request"] = map[string]any{"url": "https://api.github.com/repos/partio-io/cli/pulls/59"}

	runGate(t, gh, 12, writeVerdict(t, keepVerdict), "2026-10-06")

	if got := len(gh.comments[62]); got != 1 {
		t.Fatalf("tracking issue #62 has %d comments, want 1; requests:\n%s", got, strings.Join(gh.requests, "\n"))
	}
	for _, n := range []int{59, 60, 61} {
		if got := len(gh.comments[n]); got != 0 {
			t.Errorf("decoy #%d got %d comments", n, got)
		}
	}
	if creates := requestsTo(gh.requests, "/repos/partio-io/cli/labels"); len(creates) != 0 {
		t.Errorf("touched labels for a found tracking issue: %q", creates)
	}
}

// When the label exists and no issue carries the marker, the gate
// creates the issue but not the label.
func TestGateCreatesTrackingIssueWhenOnlyTheLabelExists(t *testing.T) {
	gh := newFakeGitHub()
	gh.labels = append(gh.labels, TrackingLabel)
	gh.issues[60] = issueJSON(60, "Minion review log", "no marker here", []string{TrackingLabel})

	runGate(t, gh, 12, writeVerdict(t, keepVerdict), "2026-10-06")

	want := []string{
		"GET /repos/partio-io/cli/issues/12",
		"GET /repos/partio-io/cli/issues",
		"GET /repos/partio-io/cli/labels/minion-review-log",
		"POST /repos/partio-io/cli/issues",
	}
	if got := gh.requests[:len(want)]; !slices.Equal(got, want) {
		t.Fatalf("requests:\n%s\nwant prefix:\n%s", strings.Join(gh.requests, "\n"), strings.Join(want, "\n"))
	}
}

// The label and the tracking issue are created once: a second run on a
// later night finds what the first created.
func TestGateCreatesTrackingIssueOnce(t *testing.T) {
	gh := newFakeGitHub()

	runGate(t, gh, 12, writeVerdict(t, keepVerdict), "2026-10-06")
	first := len(gh.requests)
	runGate(t, gh, 12, writeVerdict(t, keepVerdict), "2026-10-07")

	second := gh.requests[first:]
	for _, r := range second {
		if r == "POST /repos/partio-io/cli/issues" || r == "POST /repos/partio-io/cli/labels" {
			t.Errorf("second run created again: %s", r)
		}
	}
	if got := len(gh.issues); got != 3 {
		t.Errorf("repo has %d issues, want 3 (two reviewed and one tracking)", got)
	}
	if got := len(gh.comments[902]); got != 2 {
		t.Errorf("tracking issue has %d night comments, want 2 (one per night)", got)
	}
}

// One night is one comment: a second issue the same night appends its
// row to that comment in place; another night starts a new comment.
func TestGateNightCommentIsUpdatedInPlace(t *testing.T) {
	gh := newFakeGitHub()
	tracking := gh.trackingIssue(77)
	close13 := strings.Replace(closeVerdict, `"issue": 12`, `"issue": 13`, 1)

	runGate(t, gh, 12, writeVerdict(t, keepVerdict), "2026-10-06")
	runGate(t, gh, 13, writeVerdict(t, close13), "2026-10-06")

	comments := gh.comments[tracking]
	if len(comments) != 1 {
		t.Fatalf("tracking issue has %d comments after two issues on one night, want 1", len(comments))
	}
	if !slices.Contains(gh.requests, "PATCH /repos/partio-io/cli/issues/comments/501") {
		t.Errorf("second issue did not update the night comment in place:\n%s", strings.Join(gh.requests, "\n"))
	}
	body := comments[0]["body"].(string)
	first, second := strings.Index(body, "\n- #12 "), strings.Index(body, "\n- #13 ")
	if first < 0 || second < first {
		t.Errorf("night comment does not hold row #12 then row #13:\n%s", body)
	}
	if !strings.Contains(body, "Cache the session index") || !strings.Contains(body, "**close**") {
		t.Errorf("row #13 lacks its title or verdict:\n%s", body)
	}

	runGate(t, gh, 12, writeVerdict(t, keepVerdict), "2026-10-07")
	comments = gh.comments[tracking]
	if len(comments) != 2 || !strings.HasPrefix(comments[1]["body"].(string), NightMarker("2026-10-07")) {
		t.Errorf("a new night did not start its own comment: %v", comments)
	}
}

// A rewrite row carries the full proposed title and body in a collapsed
// section, nested under the row so the body's own list items do not
// read as rows.
func TestGateRewriteRowCarriesTheFullBody(t *testing.T) {
	gh := newFakeGitHub()
	tracking := gh.trackingIssue(77)

	runGate(t, gh, 12, writeVerdict(t, rewriteVerdict), "2026-10-06")

	body := gh.comments[tracking][0]["body"].(string)
	for _, want := range []string{
		"**rewrite** (dry run)",
		"narrowed to network errors; dropped the config flag",
		"  <details><summary>Proposed rewrite</summary>",
		"Retry the pre-push once on a network error",
		"  ## What",
		"  - retry once\n  - log the retry",
		"  Flaky pushes fail the hook.",
		"  </details>",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("night comment missing %q:\n%s", want, body)
		}
	}
	if got := strings.Count(body, "\n- "); got != 1 {
		t.Errorf("night comment has %d rows, want 1:\n%s", got, body)
	}
}

// A missing or malformed verdict still writes its row, as "no verdict"
// with the cause, and the result is not valid.
func TestGateNoVerdictWritesRowWithCause(t *testing.T) {
	tests := []struct {
		name      string
		path      func(t *testing.T) string
		wantCause string
	}{
		{name: "missing", path: func(t *testing.T) string { return t.TempDir() + "/absent.json" }, wantCause: "read verdict"},
		{name: "malformed", path: func(t *testing.T) string { return writeVerdict(t, "not json") }, wantCause: "malformed verdict"},
		{name: "other issue", path: func(t *testing.T) string {
			return writeVerdict(t, strings.Replace(keepVerdict, `"issue": 12`, `"issue": 13`, 1))
		}, wantCause: "verdict is for issue 13, not 12"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gh := newFakeGitHub()
			tracking := gh.trackingIssue(77)

			res := runGate(t, gh, 12, tt.path(t), "2026-10-06")

			if res.Valid {
				t.Fatalf("Run returned a valid result for a %s verdict", tt.name)
			}
			if !strings.Contains(res.Cause, tt.wantCause) {
				t.Errorf("Cause = %q, want it to contain %q", res.Cause, tt.wantCause)
			}
			if got := len(gh.comments[tracking]); got != 1 {
				t.Fatalf("tracking issue has %d comments, want 1", got)
			}
			body := gh.comments[tracking][0]["body"].(string)
			for _, want := range []string{"#12", "Add a retry to the pre-push hook", "**no verdict**", tt.wantCause} {
				if !strings.Contains(body, want) {
					t.Errorf("row missing %q:\n%s", want, body)
				}
			}
			if got := requestsTo(gh.requests, "/repos/partio-io/cli/issues/12"); !slices.Equal(got, []string{"GET /repos/partio-io/cli/issues/12"}) {
				t.Errorf("requests on the reviewed issue = %q, want only the read", got)
			}
		})
	}
}

// A night comment near GitHub's size limit is not grown past it: the
// row goes to a continuation comment with the same marker, and the next
// row of the night appends to that continuation.
func TestGateNightCommentContinuesPastTheSizeLimit(t *testing.T) {
	gh := newFakeGitHub()
	tracking := gh.trackingIssue(77)
	full := NightMarker("2026-10-06") + "\n" + strings.Repeat("x", maxCommentBody-100)
	gh.comments[tracking] = []map[string]any{{"id": int64(400), "body": full}}

	runGate(t, gh, 12, writeVerdict(t, rewriteVerdict), "2026-10-06")

	comments := gh.comments[tracking]
	if len(comments) != 2 {
		t.Fatalf("tracking issue has %d comments, want 2 (the full one and a continuation)", len(comments))
	}
	if comments[0]["body"] != full {
		t.Errorf("the full night comment changed")
	}
	cont := comments[1]["body"].(string)
	if !strings.HasPrefix(cont, NightMarker("2026-10-06")+"\n### Review night 2026-10-06 (continued)") || !strings.Contains(cont, "- #12 ") {
		t.Errorf("continuation comment is wrong:\n%s", cont)
	}

	close13 := strings.Replace(closeVerdict, `"issue": 12`, `"issue": 13`, 1)
	runGate(t, gh, 13, writeVerdict(t, close13), "2026-10-06")
	comments = gh.comments[tracking]
	if len(comments) != 2 || !strings.Contains(comments[1]["body"].(string), "- #13 ") {
		t.Errorf("the next row did not append to the continuation: %d comments", len(comments))
	}
}

// A rewrite body cannot break out of its collapsed section: its own
// closing tag and code fences stay inside a longer fence.
func TestGateRewriteBodyStaysInsideItsSection(t *testing.T) {
	gh := newFakeGitHub()
	tracking := gh.trackingIssue(77)
	hostile := strings.Replace(rewriteVerdict, "Flaky pushes fail the hook.", "</details>\\n\\n```go\\nunclosed", 1)

	runGate(t, gh, 12, writeVerdict(t, hostile), "2026-10-06")

	body := gh.comments[tracking][0]["body"].(string)
	if !strings.Contains(body, "  ````markdown\n") || !strings.Contains(body, "  unclosed\n") || !strings.Contains(body, "  Proposal id: retry-pre-push\n  ````\n") {
		t.Errorf("rewrite body is not fenced past its own fences:\n%s", body)
	}
}

// The tracking issue body carries totals per verdict and per close
// reason, recomputed from every night comment after each row.
func TestGateRecomputesTheTotalsAfterEachRow(t *testing.T) {
	gh := newFakeGitHub()
	tracking := gh.trackingIssue(77)
	body := func() string { return gh.issues[tracking]["body"].(string) }

	runGate(t, gh, 12, writeVerdict(t, keepVerdict), "2026-10-05")
	if !strings.Contains(body(), "| keep | 1 | 1 |") {
		t.Fatalf("totals after the first row:\n%s", body())
	}

	runGate(t, gh, 12, writeVerdict(t, closeVerdictFor(ReasonBuilt)), "2026-10-05")
	runGate(t, gh, 12, writeVerdict(t, closeVerdictFor(ReasonDuplicate)), "2026-10-06")
	runGate(t, gh, 12, writeVerdict(t, closeVerdictFor(ReasonDuplicate)), "2026-10-06")
	runGate(t, gh, 12, t.TempDir()+"/absent.json", "2026-10-06")
	runAct(t, gh, 12, writeVerdict(t, keepVerdict))

	for _, want := range []string{
		"## Totals",
		"| keep | 2 | 1 |",
		"| rewrite | 0 | 0 |",
		"| close | 3 | 3 |",
		"| no verdict | 1 | - |",
		"| built | 1 | 1 |",
		"| false-premise | 0 | 0 |",
		"| does-not-apply | 0 | 0 |",
		"| duplicate | 2 | 2 |",
		"| could-not-verify | 0 | 0 |",
		"## Needs you",
	} {
		if !strings.Contains(body(), want) {
			t.Errorf("tracking body lacks %q:\n%s", want, body())
		}
	}
	if !strings.HasPrefix(body(), TrackingMarker+"\n") {
		t.Errorf("tracking body lost its marker:\n%s", body())
	}
}
