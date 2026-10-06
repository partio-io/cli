package review

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

// validRewriteBody has the issue shape the proposer files today.
const validRewriteBody = "## What\n\nRetry the push once.\n\n" +
	"## Premise\n\n<!-- partio:premise:v1 -->\n\n" +
	"- the pre-push hook has no retry [evidence: `internal/hooks/prepush.go`]\n\n" +
	"## Acceptance Criteria\n\n- [ ] a push that fails on a network error runs once more\n\n" +
	"## Source\n\nInspired by entireio/cli#2237.\n\n" +
	"Proposal id: retry-pre-push\n"

// rewriteVerdictWith is a valid rewrite verdict for issue 12 with the
// given title and body.
func rewriteVerdictWith(t *testing.T, title, body string) string {
	t.Helper()
	rw, err := json.Marshal(Rewrite{Title: title, Body: body, Changes: []string{"added the premise block"}})
	if err != nil {
		t.Fatalf("encode rewrite: %v", err)
	}
	return `{"issue": 12, "outcome": "rewrite",
	"premise": {"verdict": "no-claims", "claims": []},
	"fit": {"applies": true, "reason": "in scope"},
	"built": {"built": false, "evidence": ""},
	"duplicates": [], "rewrite": ` + string(rw) + `}`
}

// A rewrite whose new text fails a shape check is no verdict, in dry
// run and in a real run: the gate writes a no-verdict row that names
// the check, and the issue does not change. In dry run the row also
// carries the proposed text.
func TestGateInvalidRewriteIsNoVerdict(t *testing.T) {
	tests := []struct {
		name  string
		title string
		body  string
		cause string
	}{
		{"no premise section", "Retry the push",
			strings.Replace(validRewriteBody, "## Premise", "## Facts", 1), "premise"},
		{"no premise marker", "Retry the push",
			strings.Replace(validRewriteBody, "<!-- partio:premise:v1 -->", "", 1), "premise"},
		{"claim without evidence", "Retry the push",
			strings.Replace(validRewriteBody, " [evidence: `internal/hooks/prepush.go`]", "", 1), "premise"},
		{"no acceptance criteria section", "Retry the push",
			strings.Replace(validRewriteBody, "## Acceptance Criteria", "## Done when", 1), "Acceptance Criteria"},
		{"no checklist item", "Retry the push",
			strings.Replace(validRewriteBody, "- [ ] a push", "- a push", 1), "Acceptance Criteria"},
		{"checklist item outside the section", "Retry the push",
			strings.Replace(strings.Replace(validRewriteBody, "- [ ] a push", "- a push", 1),
				"Retry the push once.", "- [ ] retry the push once", 1), "Acceptance Criteria"},
		{"no proposal id", "Retry the push",
			strings.Replace(validRewriteBody, "Proposal id: retry-pre-push\n", "", 1), "Proposal id"},
		{"proposal id not last", "Retry the push",
			validRewriteBody + "\nThanks.\n", "Proposal id"},
		{"proposal id without an id", "Retry the push",
			strings.Replace(validRewriteBody, "Proposal id: retry-pre-push", "Proposal id:", 1), "Proposal id"},
		{"source dropped", "Retry the push",
			strings.Replace(validRewriteBody, "## Source\n\nInspired by entireio/cli#2237.\n\n", "", 1), "entireio/cli#2237"},
		{"program pointer", "Retry the push",
			strings.Replace(validRewriteBody, "Proposal id:", "<!-- program: .minions/programs/retry-pre-push.md -->\n\nProposal id:", 1), "program"},
		{"body past the size limit", "Retry the push",
			strings.Replace(validRewriteBody, "Retry the push once.", strings.Repeat("x", maxIssueBody), 1), "limit"},
		{"empty title", "", validRewriteBody, "title"},
		{"blank title", "  ", validRewriteBody, "title"},
	}
	for _, tt := range tests {
		for _, dryRun := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/dry-run=%t", tt.name, dryRun), func(t *testing.T) {
				gh := newFakeGitHub()
				gh.issues[12]["body"] = "Inspired by entireio/cli#2237."
				tracking := gh.trackingIssue(77)

				res := runGateMode(t, gh, 12, writeVerdict(t, rewriteVerdictWith(t, tt.title, tt.body)), "2026-10-06", dryRun)

				if res.Valid || !strings.Contains(res.Cause, tt.cause) {
					t.Fatalf("Result = %+v, want no verdict naming %q", res, tt.cause)
				}
				for _, r := range gh.requests {
					method, path, _ := strings.Cut(r, " ")
					if method != "GET" && !strings.HasPrefix(path, "/repos/partio-io/cli/issues/77/") {
						t.Errorf("gate changed the issue on an invalid rewrite: %s", r)
					}
				}
				row := gh.comments[tracking][0]["body"].(string)
				if !strings.Contains(row, "**no verdict**") || !strings.Contains(row, tt.cause) {
					t.Errorf("tracking row does not name the failed check %q:\n%s", tt.cause, row)
				}
				// A rewrite without a title fails when the verdict
				// loads, so there is no proposed text to show.
				want := dryRun && strings.TrimSpace(tt.title) != ""
				if got := strings.Contains(row, "  ## What"); got != want {
					t.Errorf("row carries the proposed body = %t, want %t:\n%s", got, want, row)
				}
			})
		}
	}
}

// A valid rewrite keeps every source item of the old body, in any form
// SourceRefs reads. An old body without a source needs none.
func TestGateValidRewriteIsActedOn(t *testing.T) {
	noSource := strings.Replace(validRewriteBody, "## Source\n\nInspired by entireio/cli#2237.\n\n", "", 1)
	tests := []struct {
		name    string
		oldBody string
		body    string
	}{
		{"same source", "Inspired by entireio/cli#2237.", validRewriteBody},
		{"source as a URL", "Inspired by entireio/cli#2237.",
			strings.Replace(validRewriteBody, "entireio/cli#2237", "https://github.com/entireio/cli/issues/2237", 1)},
		{"old body without a source", "Retry pushes. No source.", noSource},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gh := newFakeGitHub()
			gh.issues[12]["body"] = tt.oldBody
			gh.trackingIssue(77)

			res := runAct(t, gh, 12, writeVerdict(t, rewriteVerdictWith(t, "Retry the push", tt.body)))

			if !res.Valid || res.Outcome != OutcomeRewrite {
				t.Fatalf("Result = %+v, want a valid rewrite", res)
			}
			if gh.issues[12]["body"] != tt.body || gh.issues[12]["title"] != "Retry the push" {
				t.Errorf("issue was not rewritten: %q\n%s", gh.issues[12]["title"], gh.issues[12]["body"])
			}
		})
	}
}
