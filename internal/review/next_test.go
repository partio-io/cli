package review

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/partio-io/cli/internal/github"
)

// proposal seeds an open minion-proposal issue filed at created, an
// RFC 3339 time, with labels on top of minion-proposal.
func (f *fakeGitHub) proposal(number int, created string, labels ...string) {
	issue := issueJSON(number, fmt.Sprintf("Proposal %d", number), "body", append([]string{proposalLabel}, labels...))
	issue["created_at"] = created
	f.issues[number] = issue
}

func runNext(t *testing.T, gh *fakeGitHub, q NextQuery) []int {
	t.Helper()
	srv := gh.server(t)
	q.Repo = repo
	got, err := Next(github.Client{BaseURL: srv.URL, Token: "test-token", HTTPClient: srv.Client()}, q)
	if err != nil {
		t.Fatalf("Next: %v", err)
	}
	return got
}

func TestNextOrdersApprovedFirstThenOldestFirst(t *testing.T) {
	gh := newFakeGitHub()
	gh.issues = map[int]map[string]any{}
	gh.proposal(10, "2026-09-03T08:00:00Z")
	gh.proposal(11, "2026-08-20T08:00:00Z")
	gh.proposal(12, "2026-09-01T08:00:00Z", approvedLabel)
	gh.proposal(13, "2026-07-01T08:00:00Z", ReviewedLabel)
	gh.proposal(14, "2026-08-01T08:00:00Z", approvedLabel)
	gh.proposal(15, "2026-07-15T08:00:00Z", ReviewedLabel, approvedLabel)
	gh.issues[16] = issueJSON(16, "Not a proposal", "body", nil)

	got := runNext(t, gh, NextQuery{})

	if want := []int{14, 12, 11, 10}; !slices.Equal(got, want) {
		t.Errorf("Next = %v, want %v", got, want)
	}
}

// Two "no verdict" rows for one issue take it out of the batch and put
// it under "needs you" in the tracking issue body. One such row does
// not, and a later acted verdict clears the issue again.
func TestNextSkipsAnIssueWithTwoNoVerdictRows(t *testing.T) {
	gh := newFakeGitHub()
	gh.proposal(12, "2026-08-01T08:00:00Z")
	gh.proposal(13, "2026-08-02T08:00:00Z")
	tracking := gh.trackingIssue(77)
	absent := t.TempDir() + "/absent.json"

	runGate(t, gh, 12, absent, "2026-10-05")
	runGate(t, gh, 13, absent, "2026-10-05")
	if got := runNext(t, gh, NextQuery{}); !slices.Equal(got, []int{12, 13}) {
		t.Fatalf("after one no-verdict row each: Next = %v, want [12 13]", got)
	}
	runGate(t, gh, 12, absent, "2026-10-06")

	if got := runNext(t, gh, NextQuery{}); !slices.Equal(got, []int{13}) {
		t.Errorf("Next = %v, want [13]", got)
	}
	body := gh.issues[tracking]["body"].(string)
	_, needs, ok := strings.Cut(body, "## Needs you")
	if !ok {
		t.Fatalf("tracking body has no needs-you heading:\n%s", body)
	}
	if !strings.Contains(needs, "- #12") || strings.Contains(needs, "#13") {
		t.Errorf("needs-you list wrong:\n%s", needs)
	}
	if !strings.HasPrefix(body, TrackingMarker+"\n") {
		t.Errorf("tracking body lost its marker:\n%s", body)
	}

	runAct(t, gh, 12, writeVerdict(t, keepVerdict))
	body = gh.issues[tracking]["body"].(string)
	if _, needs, _ := strings.Cut(body, "## Needs you"); strings.Contains(needs, "#12") {
		t.Errorf("an acted keep left #12 under needs you:\n%s", body)
	}
}

// A sample takes one issue per filing month in turn, oldest month
// first and oldest issue first in each month, and the same backlog
// gives the same sample.
func TestNextSampleSpreadsAcrossFilingMonths(t *testing.T) {
	gh := newFakeGitHub()
	gh.issues = map[int]map[string]any{}
	gh.proposal(20, "2026-09-20T08:00:00Z", approvedLabel)
	gh.proposal(21, "2026-07-30T08:00:00Z")
	gh.proposal(22, "2026-07-02T08:00:00Z")
	gh.proposal(23, "2026-09-01T08:00:00Z")
	gh.proposal(24, "2026-07-15T08:00:00Z")
	gh.proposal(25, "2026-08-31T23:59:00Z")
	gh.proposal(26, "2026-09-05T08:00:00Z", ReviewedLabel)

	tests := []struct {
		sample int
		want   []int
	}{
		{sample: 1, want: []int{22}},
		{sample: 3, want: []int{22, 25, 23}},
		{sample: 5, want: []int{22, 25, 23, 24, 20}},
		{sample: 50, want: []int{22, 25, 23, 24, 20, 21}},
	}
	for _, tt := range tests {
		first := runNext(t, gh, NextQuery{Sample: tt.sample})
		again := runNext(t, gh, NextQuery{Sample: tt.sample})
		if !slices.Equal(first, tt.want) {
			t.Errorf("Sample %d = %v, want %v", tt.sample, first, tt.want)
		}
		if !slices.Equal(first, again) {
			t.Errorf("Sample %d differs between runs: %v then %v", tt.sample, first, again)
		}
	}
}
