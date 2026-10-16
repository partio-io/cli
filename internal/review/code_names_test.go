package review

import (
	"fmt"
	"math"
	"testing"
)

// rankedDupe returns the ranked candidate with the number n, or fails.
func rankedDupe(t *testing.T, got []Dupe, n int) Dupe {
	t.Helper()
	for _, d := range got {
		if d.Number == n {
			return d
		}
	}
	t.Fatalf("#%d not in %+v", n, got)
	return Dupe{}
}

func TestFindDupesAddsTheCodeNameOverlapToTheScore(t *testing.T) {
	issues := append(fillers(1, 8),
		fakeProposal(750, "Respect the config dir of Claude", "Read `CLAUDE_CONFIG_DIR` first.", "open", ""),
		fakeProposal(751, "Respect the config dir of Claude", "No code here.", "open", ""))
	f := newFakeBacklog(t, issues)
	got, err := FindDupes(f.client(), DupesQuery{
		Repo: "partio-io/cli", Sources: []string{"other/repo#1"},
		Title: "Claude sessions path", Body: "Honour `claude_config_dir`.",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) < 2 || got[0].Number != 750 || got[1].Number != 751 {
		t.Fatalf("FindDupes = %+v, want #750 then #751", got)
	}
	a, b := got[0], got[1]
	// A score above 1 from the code names is still no strong match: the
	// title alone decides that.
	if a.Score < titleMatch || a.Match != MatchRanked {
		t.Errorf("#750 score %v matched by %q, want a score of 0.6 or more and %q", a.Score, a.Match, MatchRanked)
	}
	if got := a.Score - b.Score; math.Abs(got-1) > 1e-9 {
		t.Errorf("code-name part of the score = %v, want 1", got)
	}
	if fmt.Sprint(a.Shared) != "[claude claude_config_dir]" {
		t.Errorf("shared = %v, want [claude claude_config_dir]", a.Shared)
	}
	if fmt.Sprint(b.Shared) != "[claude]" {
		t.Errorf("shared = %v, want [claude]", b.Shared)
	}
}

func TestFindDupesRanksByTitleAloneWithNoCodeNames(t *testing.T) {
	issues := append(fillers(1, 8),
		fakeProposal(760, "Faster checkpoint upload", "Use `upload_batch` here.", "open", ""),
		fakeProposal(761, "Faster checkpoint upload", "No code here.", "open", ""))
	f := newFakeBacklog(t, issues)
	for _, body := range []string{"", "Plain words, no `ab` code names."} {
		got, err := FindDupes(f.client(), DupesQuery{
			Repo: "partio-io/cli", Sources: []string{"other/repo#1"}, Title: "Faster upload", Body: body,
		})
		if err != nil {
			t.Fatal(err)
		}
		a, b := rankedDupe(t, got, 760), rankedDupe(t, got, 761)
		if a.Score != b.Score || fmt.Sprint(a.Shared) != fmt.Sprint(b.Shared) {
			t.Errorf("body %q: #760 %v %v, #761 %v %v, want the same title-only score", body, a.Score, a.Shared, b.Score, b.Shared)
		}
	}
	// A proposal with no code names gets its title overlap alone.
	got, err := FindDupes(f.client(), DupesQuery{
		Repo: "partio-io/cli", Sources: []string{"other/repo#1"}, Title: "Faster upload", Body: "`upload_batch`",
	})
	if err != nil {
		t.Fatal(err)
	}
	if a, b := rankedDupe(t, got, 760), rankedDupe(t, got, 761); math.Abs(a.Score-b.Score-1) > 1e-9 || fmt.Sprint(b.Shared) != "[faster upload]" {
		t.Errorf("#760 %v, #761 %v %v, want #761 one lower with [faster upload]", a.Score, b.Score, b.Shared)
	}
}

func TestFindDupesRanksARareCodeNameAboveACommonOne(t *testing.T) {
	issues := append(fillers(1, 4),
		fakeProposal(770, "Tune one thing", "Touch `rare_name` and `extra_a`.", "open", ""),
		fakeProposal(771, "Tune one thing", "Touch `common_name` and `extra_b`.", "open", ""))
	for i := range 10 {
		issues = append(issues, fakeProposal(780+i, fmt.Sprintf("Unrelated %d", i), "Uses `common_name`.", "open", ""))
	}
	f := newFakeBacklog(t, issues)
	got, err := FindDupes(f.client(), DupesQuery{
		Repo: "partio-io/cli", Sources: []string{"other/repo#1"},
		Title: "Tune it", Body: "Change `rare_name` and `common_name`.",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) < 2 || got[0].Number != 770 || got[1].Number != 771 {
		t.Fatalf("FindDupes = %+v, want #770 then #771", got)
	}
	if got[0].Score <= got[1].Score {
		t.Errorf("rare score %v, common score %v, want rare higher", got[0].Score, got[1].Score)
	}
}

func TestFindDupesSharesEachCodeNameCase(t *testing.T) {
	tests := []struct {
		name   string
		body   string // the body of the proposal
		query  string // the body of the query
		shared string // the shared code names, after the shared title word "tweak"
	}{
		{"escaped backticks count", `Set \` + "`" + `escaped_name\` + "`" + `.`, "`escaped_name`", "[tweak escaped_name]"},
		{"fenced block does not count", "```\nfenced_name\n```\nSee `kept_name`.", "`fenced_name` `kept_name`", "[tweak kept_name]"},
		{"fenced inline span does not count", "```go\nx := `fenced_name`\n```", "`fenced_name`", "[tweak]"},
		{"path counts by its base name", "Edit `internal/hooks/post_commit.go`.", "`cmd/post_commit.go`", "[tweak post_commit.go]"},
		{"trailing () is cut", "Call `FindDupes()`.", "`finddupes`", "[tweak finddupes]"},
		{"short spans drop", "Flag `-v` and `ab`.", "`-v` `ab`", "[tweak]"},
		{"digit spans drop", "Exit `2` or `404`.", "`404`", "[tweak]"},
		{"common names drop", "`partio` `true` `false` `nil` `git` `go`", "`partio` `true` `false` `nil` `git` `go`", "[tweak]"},
		{"a span with spaces stays whole", "Run `git commit --amend`.", "`git commit --amend`", "[tweak git commit --amend]"},
		{"a path with spaces stays whole", "Run `cd a/b c`.", "`cd a/b c`", "[tweak cd a/b c]"},
		{"trimmed and in lower case", "Read ` Claude_Config_Dir `.", "`CLAUDE_CONFIG_DIR`", "[tweak claude_config_dir]"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newFakeBacklog(t, append(fillers(1, 8), fakeProposal(790, "Small tweak", tt.body, "open", "")))
			got, err := FindDupes(f.client(), DupesQuery{
				Repo: "partio-io/cli", Sources: []string{"other/repo#1"}, Title: "One tweak more", Body: tt.query,
			})
			if err != nil {
				t.Fatal(err)
			}
			if d := rankedDupe(t, got, 790); fmt.Sprint(d.Shared) != tt.shared {
				t.Errorf("shared = %v, want %s", d.Shared, tt.shared)
			}
		})
	}
}

func TestFindDupesByIssueTakesTheCodeNamesOfItsBody(t *testing.T) {
	issues := append(fillers(1, 8),
		fakeProposal(750, "Respect the config dir", "Read `CLAUDE_CONFIG_DIR` first.", "open", ""),
		fakeProposal(771, "Claude sessions path", "Honour `claude_config_dir`.", "open", ""))
	f := newFakeBacklog(t, issues)
	got, err := FindDupes(f.client(), DupesQuery{Repo: "partio-io/cli", Issue: 771})
	if err != nil {
		t.Fatal(err)
	}
	if d := rankedDupe(t, got, 750); fmt.Sprint(d.Shared) != "[claude_config_dir]" {
		t.Errorf("shared = %v, want [claude_config_dir]", d.Shared)
	}
}
