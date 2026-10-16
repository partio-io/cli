package premise

import (
	"strings"
	"testing"
)

const reviewDupesRun = "go run ./cmd/minion-review dupes"

// TestReviewSearchesForDuplicates checks that the review finds the same idea
// in the backlog with the deterministic search, not with its own idea of a
// search, and before it decides. It passes only the issue number: the tool
// reads the title, body and source items of the issue itself.
func TestReviewSearchesForDuplicates(t *testing.T) {
	agent := reviewer(t)

	run := strings.Index(agent, reviewDupesRun)
	if run < 0 {
		t.Fatalf("%s does not run %q", reviewAgentH3, reviewDupesRun)
	}
	line := agent[run:]
	if end := strings.Index(line, "\n"); end >= 0 {
		line = line[:end]
	}
	if want := reviewDupesRun + " --issue <issue number>"; strings.TrimSpace(line) != want {
		t.Errorf("%s runs %q, want %q", reviewAgentH3, strings.TrimSpace(line), want)
	}
	for _, flag := range []string{"--source", "--title", "--exclude"} {
		if strings.Contains(agent, flag) {
			t.Errorf("%s still names %s: --issue replaces it", reviewAgentH3, flag)
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

// TestReviewReadsTheRankedCandidates checks that the review knows the three
// kinds of match and the fields of a candidate, reads each title and excerpt
// before it opens an issue, and keeps the issue when it is not sure.
func TestReviewReadsTheRankedCandidates(t *testing.T) {
	agent := flat(reviewer(t))

	for _, want := range []string{
		`"match": "source"`,
		`"match": "title"`,
		`"match": "ranked"`,
		"is one of the 10 other proposals with the best score",
		"`title`, `state`, `state_reason`, `labels`, `score`, `shared` and `excerpt`",
		"read the title and the excerpt of each candidate",
		"open a candidate with `gh issue view <n>` only when it can be the same idea",
		"judge a ranked candidate that is the same idea as a strong match",
		"when you are not sure that a candidate is the same idea, keep the issue",
		`"same": false`,
		"the reason in `why`",
		"no issue closes on a weak guess",
		"the list is the backlog you judge against",
	} {
		if !strings.Contains(strings.ToLower(agent), strings.ToLower(want)) {
			t.Errorf("%s does not carry %q", reviewAgentH3, want)
		}
	}
}
