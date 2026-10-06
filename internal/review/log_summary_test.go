package review

import "testing"

// TestSummarizeReadsTheVerdictColumn checks that a reason which quotes a
// bold verdict does not take the place of the verdict of its row.
func TestSummarizeReadsTheVerdictColumn(t *testing.T) {
	night := NightMarker("2026-10-06") + "\n### Review night 2026-10-06\n\n" +
		"- #5 · A title · **keep** · a · **close** · b · [link](https://x/5)\n"
	s := summarize([]string{night})
	if got := s.verdicts[OutcomeKeep].rows; got != 1 {
		t.Errorf("keep rows = %d, want 1", got)
	}
	if got := s.verdicts[OutcomeClose].rows; got != 0 {
		t.Errorf("close rows = %d, want 0", got)
	}
}

// A build row is an acted verdict, not a dry run: it counts as acted
// and resets the "no verdict" count of its issue.
func TestSummarizeCountsABuildRowAsActed(t *testing.T) {
	night := NightMarker("2026-10-06") + "\n### Review night 2026-10-06\n\n" +
		"- #5 · A title · **no verdict** · bad · [link](https://x/5)\n" +
		"- #5 · A title · **no verdict** (build) · bad · [link](https://x/5)\n" +
		"- #5 · A title · **keep** (build) · in scope · [link](https://x/5)\n"
	s := summarize([]string{night})
	if got := s.verdicts[noVerdict]; got.rows != 2 || got.dryRun != 0 {
		t.Errorf("no verdict tally = %+v, want 2 rows, 0 dry run", got)
	}
	if got := s.verdicts[OutcomeKeep]; got.rows != 1 || got.dryRun != 0 {
		t.Errorf("keep tally = %+v, want 1 row, 0 dry run", got)
	}
	if len(s.needsYou) != 0 {
		t.Errorf("needsYou = %v, want none after an acted build keep", s.needsYou)
	}
}
