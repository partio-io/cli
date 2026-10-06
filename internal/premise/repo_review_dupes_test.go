package premise

import (
	"strings"
	"testing"
)

const reviewDupesRun = "go run ./cmd/minion-review dupes"

// TestReviewSearchesForDuplicates checks that the review finds the same idea
// in the backlog with the deterministic search, not with its own idea of a
// search, and before it decides. It passes the issue's sources and title, and
// leaves the issue itself out.
func TestReviewSearchesForDuplicates(t *testing.T) {
	agent := reviewer(t)

	run := strings.Index(agent, reviewDupesRun)
	if run < 0 {
		t.Fatalf("%s does not run %q", reviewAgentH3, reviewDupesRun)
	}
	for _, flag := range []string{"--source", "--title", "--exclude"} {
		if !strings.Contains(agent[run:], flag) {
			t.Errorf("%s runs the search without %s", reviewAgentH3, flag)
		}
	}
	if decide := strings.Index(agent, "**Decide.**"); decide < 0 || decide < run {
		t.Errorf("%s decides before it searches for duplicates", reviewAgentH3)
	}
	if strings.Contains(agent, "gh search") {
		t.Errorf("%s names gh search: the search index lags behind new issues", reviewAgentH3)
	}
}

// TestReviewRecordsEveryCandidate checks that the verdict lists every candidate
// the search returned, each with a judgment and its reason, so that a reader
// sees what the review considered and not only what it closed against.
func TestReviewRecordsEveryCandidate(t *testing.T) {
	agent := flat(reviewer(t))

	for _, want := range []string{
		"every candidate",
		`"duplicates": [`,
		`"same":`,
		`"why":`,
		`"duplicate_of":`,
		"| duplicate |",
	} {
		if !strings.Contains(agent, want) {
			t.Errorf("%s does not carry %q", reviewAgentH3, want)
		}
	}
}

// TestReviewAppliesTheDuplicateRules checks the rules that keep the first issue
// reviewed. A reviewed open issue stays and this one closes; an unreviewed one
// waits for its own review; an issue the review closed already has a verdict;
// and an issue closed as completed proves a build only when the tree shows it,
// because a build can close an issue whose pull request never merges.
func TestReviewAppliesTheDuplicateRules(t *testing.T) {
	agent := flat(reviewer(t))

	for _, want := range []string{
		"minion-reviewed",
		"the first issue reviewed stays",
		"closes when its own review runs",
		"not_planned",
		"already has a verdict",
		"completed",
		"only when the tree shows",
		"the closed state alone proves nothing",
		"names the issue that stays",
	} {
		if !strings.Contains(agent, want) {
			t.Errorf("%s does not carry the duplicate rule %q", reviewAgentH3, want)
		}
	}
}
