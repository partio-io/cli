package review

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"unicode/utf8"
)

func TestFindDupesRanksTheBestCandidatesAfterTheStrongMatches(t *testing.T) {
	issues := []fakeIssue{
		fakeProposal(690, "Goose agent hook", "**Origin:** entireio/cli#2075", "open", ""),
		fakeProposal(681, "Add Goose coding agent integration", "Origin: other/repo#5", "open", ""),
	}
	// Twelve titles: six share "coding" and "goose" with the query, six
	// share "goose" only. Only ten of them come back.
	for i := range 12 {
		title := fmt.Sprintf("Goose plugin variant%d", i)
		if i%2 == 0 {
			title = fmt.Sprintf("Goose coding variant%d", i)
		}
		issues = append(issues, fakeProposal(701+i, title, "Origin: other/repo#9", "open", ""))
	}
	issues = append(issues, fillers(1, 20)...)
	f := newFakeBacklog(t, issues)

	got, err := FindDupes(f.client(), DupesQuery{
		Repo: "partio-io/cli", Sources: []string{"entireio/cli#2075"}, Title: "Add Goose coding agent integration",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 12 {
		t.Fatalf("got %d candidates, want 2 strong and 10 ranked: %+v", len(got), got)
	}
	if got[0].Number != 690 || got[0].Match != MatchSource || got[1].Number != 681 || got[1].Match != MatchTitle {
		t.Errorf("strong matches = %+v, %+v; want #690 by source, #681 by title", got[0], got[1])
	}
	var ranked []int
	for i, d := range got[2:] {
		if d.Match != MatchRanked {
			t.Errorf("#%d matched by %q, want %q", d.Number, d.Match, MatchRanked)
		}
		if d.Number == 690 || d.Number == 681 {
			t.Errorf("strong match #%d comes again in the ranked part", d.Number)
		}
		if i > 0 && d.Score > got[i+1].Score {
			t.Errorf("#%d (%.3f) ranks after #%d (%.3f)", d.Number, d.Score, got[i+1].Number, got[i+1].Score)
		}
		ranked = append(ranked, d.Number)
	}
	if fmt.Sprint(ranked) != "[701 703 705 707 709 711 702 704 706 708]" {
		t.Errorf("ranked %v, want the coding titles first, then the older plugin titles", ranked)
	}
	if fmt.Sprint(got[2].Shared) != "[coding goose]" || fmt.Sprint(got[8].Shared) != "[goose]" {
		t.Errorf("shared = %v and %v, want [coding goose] and [goose]", got[2].Shared, got[8].Shared)
	}
}

func TestFindDupesGivesTheSameCandidatesOnEveryRun(t *testing.T) {
	issues := fillers(1, 20)
	for i := range 30 {
		issues = append(issues, fakeProposal(700+i,
			fmt.Sprintf("Respect config dir in session discovery step%d hook%d agent%d", i, i%3, i%5), "", "open", ""))
	}
	f := newFakeBacklog(t, issues)
	q := DupesQuery{Repo: "partio-io/cli", Title: "Respect CLAUDE_CONFIG_DIR in Claude session discovery hook1 agent2"}

	first, err := FindDupes(f.client(), q)
	if err != nil {
		t.Fatal(err)
	}
	for range 50 {
		got, err := FindDupes(f.client(), q)
		if err != nil {
			t.Fatal(err)
		}
		if fmt.Sprintf("%v", got) != fmt.Sprintf("%v", first) {
			t.Fatalf("a second run gave\n%+v\nwant\n%+v", got, first)
		}
	}
}

func TestFindDupesRanksNoCandidateWithAZeroScore(t *testing.T) {
	f := newFakeBacklog(t, append(fillers(1, 5),
		fakeProposal(700, "Compact the session transcript", "", "open", "")))
	got, err := FindDupes(f.client(), DupesQuery{Repo: "partio-io/cli", Title: "Compact checkpoints"})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Number != 700 || got[0].Score <= 0 {
		t.Errorf("FindDupes = %+v, want only #700 with a score above zero", got)
	}
}

// A changelog entry has no source item: its title alone finds candidates.
func TestFindDupesRanksByTitleForAQueryWithNoSource(t *testing.T) {
	f := newFakeBacklog(t, append(fillers(1, 5),
		fakeProposal(750, "Honor the user config folder for session lookup", "Origin: entireio/cli#2627", "open", "")))
	got, err := FindDupes(f.client(), DupesQuery{
		Repo: "partio-io/cli", Title: "Respect CLAUDE_CONFIG_DIR in Claude session discovery",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Number != 750 || got[0].Match != MatchRanked {
		t.Errorf("FindDupes = %+v, want #750 ranked", got)
	}
}

func TestFindDupesGivesEachCandidateAScoreAndAnExcerpt(t *testing.T) {
	f := newFakeBacklog(t, append(fillers(1, 10),
		fakeProposal(690, "Rename a flag", "Origin: entireio/cli#2075. Rename the flag.", "open", ""),
		fakeProposal(681, "Add Goose coding agent integration", "## Summary\n\nGoose fires hooks.", "open", ""),
		fakeProposal(700, "Goose hook tweaks", "## What\n\nTweak the Goose hooks.", "open", ""),
	))
	got, err := FindDupes(f.client(), DupesQuery{
		Repo: "partio-io/cli", Sources: []string{"entireio/cli#2075"}, Title: "Add Goose coding agent integration",
	})
	if err != nil {
		t.Fatal(err)
	}
	want := map[int]struct {
		match, excerpt string
		scored         bool
	}{
		690: {MatchSource, "Origin: entireio/cli#2075. Rename the flag.", false},
		681: {MatchTitle, "Goose fires hooks.", true},
		700: {MatchRanked, "Tweak the Goose hooks.", true},
	}
	if len(got) != len(want) {
		t.Fatalf("FindDupes = %+v, want %d candidates", got, len(want))
	}
	for _, d := range got {
		w := want[d.Number]
		if d.Match != w.match || d.Excerpt != w.excerpt || (d.Score > 0) != w.scored {
			t.Errorf("#%d = %q, score %.3f, excerpt %q; want %q, scored %v, excerpt %q",
				d.Number, d.Match, d.Score, d.Excerpt, w.match, w.scored, w.excerpt)
		}
	}
}

func TestFindDupesExcerptIsTheStartOfTheDescriptionAsPlainText(t *testing.T) {
	raw, err := os.ReadFile("testdata/issue-710-body.md")
	if err != nil {
		t.Fatal(err)
	}
	escaped := "## Summary\n\nThe hook reads \\`core.hooksPath\\` and skips \\*.sample files \\#1.\n\n## Why\n\nIt breaks."
	f := newFakeBacklog(t, []fakeIssue{
		fakeProposal(710, "Hook backup overwrite", string(raw), "open", ""),
		fakeProposal(711, "Hook path sample", escaped, "open", ""),
	})
	got, err := FindDupes(f.client(), DupesQuery{Repo: "partio-io/cli", Title: "Hook"})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("FindDupes = %+v, want #710 and #711", got)
	}
	long, short := got[0].Excerpt, got[1].Excerpt
	if !strings.HasPrefix(long, "`installHooks` in `internal/git/hooks/install.go` backs up") {
		t.Errorf("#710 excerpt = %q, want the first words under ## Description", long)
	}
	if n := utf8.RuneCountInString(long); n < 250 || n > 300 {
		t.Errorf("#710 excerpt has %d characters, want about 300", n)
	}
	if !strings.Contains(string(raw), long+" ") {
		t.Errorf("#710 excerpt = %q, want it to stop at the end of a word", long)
	}
	if want := "The hook reads `core.hooksPath` and skips *.sample files #1. It breaks."; short != want {
		t.Errorf("#711 excerpt = %q, want %q", short, want)
	}
	for _, d := range got {
		if strings.Contains(d.Excerpt, "##") || strings.Contains(d.Excerpt, `\`) {
			t.Errorf("#%d excerpt = %q, want no heading marks and no escapes", d.Number, d.Excerpt)
		}
	}
}
