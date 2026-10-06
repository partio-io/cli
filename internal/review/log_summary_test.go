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
