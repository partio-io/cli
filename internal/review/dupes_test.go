package review

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"

	"github.com/partio-io/cli/internal/github"
)

// fakeIssue is one issue as GitHub's list endpoint sends it.
type fakeIssue struct {
	Number      int         `json:"number"`
	Title       string      `json:"title"`
	Body        string      `json:"body"`
	State       string      `json:"state"`
	StateReason string      `json:"state_reason,omitempty"`
	Labels      []fakeLabel `json:"labels"`
	PullRequest *struct{}   `json:"pull_request,omitempty"`
}

type fakeLabel struct {
	Name string `json:"name"`
}

func fakeProposal(number int, title, body, state, reason string, labels ...string) fakeIssue {
	ls := []fakeLabel{{Name: "minion-proposal"}}
	for _, l := range labels {
		ls = append(ls, fakeLabel{Name: l})
	}
	return fakeIssue{Number: number, Title: title, Body: body, State: state, StateReason: reason, Labels: ls}
}

// fillers returns n proposals that share nothing with the test ideas.
func fillers(from, n int) []fakeIssue {
	out := make([]fakeIssue, n)
	for i := range out {
		num := from + i
		out[i] = fakeProposal(num, fmt.Sprintf("Filler idea %d", num),
			fmt.Sprintf("**Origin:** elsewhere/other#%d", num), "open", "")
	}
	return out
}

// fakeBacklog serves issues on the repo's issue list endpoint, a page
// at a time, and records each query it served.
type fakeBacklog struct {
	srv     *httptest.Server
	queries []string
}

func newFakeBacklog(t *testing.T, issues []fakeIssue) *fakeBacklog {
	t.Helper()
	f := &fakeBacklog{}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /repos/partio-io/cli/issues", func(w http.ResponseWriter, r *http.Request) {
		f.queries = append(f.queries, r.URL.RawQuery)
		q := r.URL.Query()
		if q.Get("state") != "all" || q.Get("labels") != "minion-proposal" {
			http.Error(w, "unexpected filter", http.StatusBadRequest)
			return
		}
		page, _ := strconv.Atoi(q.Get("page"))
		per, _ := strconv.Atoi(q.Get("per_page"))
		lo := min((page-1)*per, len(issues))
		hi := min(lo+per, len(issues))
		_ = json.NewEncoder(w).Encode(issues[lo:hi])
	})
	f.srv = httptest.NewServer(mux)
	t.Cleanup(f.srv.Close)
	return f
}

func (f *fakeBacklog) client() github.Client {
	return github.Client{BaseURL: f.srv.URL, Token: "tok", HTTPClient: f.srv.Client()}
}

func TestFindDupesWalksEveryPageForASourceMatch(t *testing.T) {
	issues := append(fillers(1, 100),
		fakeProposal(688, "Add Goose coding agent integration",
			"## Source\n\n**Origin:** entireio/cli#2075\n", "closed", "completed"))
	f := newFakeBacklog(t, issues)

	got, err := FindDupes(f.client(), DupesQuery{
		Repo:    "partio-io/cli",
		Sources: []string{"entireio/cli#2075"},
		Title:   "Support a new tool",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(f.queries) != 2 {
		t.Errorf("served %d pages, want 2: %v", len(f.queries), f.queries)
	}
	want := []Dupe{{
		Number: 688, Title: "Add Goose coding agent integration",
		State: "closed", StateReason: "completed", Labels: []string{"minion-proposal"},
		Match: MatchSource, Shared: []string{"entireio/cli#2075"},
	}}
	if fmt.Sprint(got) != fmt.Sprint(want) {
		t.Errorf("FindDupes = %+v\nwant %+v", got, want)
	}
}

func TestFindDupesMatchesASourceOpenOrClosed(t *testing.T) {
	body := "Inspired by entireio/cli PR #2641, adapted to Partio."
	f := newFakeBacklog(t, []fakeIssue{
		fakeProposal(736, "Hook installer overwrites a backup", body, "open", ""),
		fakeProposal(740, "Hook installer drops a backup", body, "closed", "not_planned", "minion-reviewed"),
		fakeProposal(741, "Hook installer loses a backup", "Origin: entireio/cli#9", "closed", "completed"),
	})
	got, err := FindDupes(f.client(), DupesQuery{
		Repo: "partio-io/cli", Sources: []string{"https://github.com/entireio/cli/pull/2641"}, Title: "Keep hooks",
	})
	if err != nil {
		t.Fatal(err)
	}
	var nums []int
	for _, d := range got {
		if d.Match != MatchSource {
			t.Errorf("#%d matched by %q, want %q", d.Number, d.Match, MatchSource)
		}
		nums = append(nums, d.Number)
	}
	if fmt.Sprint(nums) != "[736 740]" {
		t.Errorf("matched %v, want [736 740]", nums)
	}
}

func TestFindDupesMatchesAStrongTitleOverlap(t *testing.T) {
	issues := []fakeIssue{
		fakeProposal(710, "installHooks silently overwrites the original hook backup when a foreign hook appears on reinstall",
			"Inspired by entireio/cli#2237.", "open", ""),
		fakeProposal(681, "Add Goose coding agent integration", "Origin: entireio/cli#2075", "open", ""),
		fakeProposal(700, "Redact secrets in session transcripts before upload", "Origin: entireio/cli#3", "open", ""),
		fakeProposal(701, "Encrypt checkpoint session", "Origin: entireio/cli#4", "open", ""),
	}
	// Words that most titles carry weigh little: sharing them alone is
	// no strong overlap.
	for i := range 20 {
		issues = append(issues, fakeProposal(800+i, fmt.Sprintf("Checkpoint session tweak %d", i), "", "open", ""))
	}
	f := newFakeBacklog(t, issues)
	tests := []struct {
		title string
		want  string
	}{
		{"Hook installer silently overwrites existing backup on re-install", "[710]"},
		{"Add Goose: Coding Agent integration!", "[681]"},
		{"Compact checkpoint session", "[]"},
		{"Show the binary size in the status output", "[]"},
	}
	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			got, err := FindDupes(f.client(), DupesQuery{Repo: "partio-io/cli", Sources: []string{"other/repo#1"}, Title: tt.title})
			if err != nil {
				t.Fatal(err)
			}
			nums := []int{}
			for _, d := range got {
				if d.Match != MatchTitle {
					t.Errorf("#%d matched by %q, want %q", d.Number, d.Match, MatchTitle)
				}
				nums = append(nums, d.Number)
			}
			if fmt.Sprint(nums) != tt.want {
				t.Errorf("matched %v, want %s", nums, tt.want)
			}
		})
	}
}

func TestFindDupesExcludesTheIssueUnderReview(t *testing.T) {
	f := newFakeBacklog(t, []fakeIssue{
		fakeProposal(681, "Add Goose coding agent integration", "Origin: entireio/cli#2075", "open", ""),
		fakeProposal(688, "Add Goose coding agent integration", "Origin: entireio/cli#2075", "closed", "completed"),
	})
	got, err := FindDupes(f.client(), DupesQuery{
		Repo: "partio-io/cli", Sources: []string{"entireio/cli#2075"},
		Title: "Add Goose coding agent integration", Exclude: 681,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Number != 688 {
		t.Errorf("FindDupes = %+v, want only #688", got)
	}
}

func TestFindDupesPutsSourceMatchesFirst(t *testing.T) {
	f := newFakeBacklog(t, []fakeIssue{
		fakeProposal(681, "Add Goose coding agent integration", "Origin: other/repo#5", "open", ""),
		fakeProposal(690, "Rename a flag", "**Origin:** entireio/cli#2075", "closed", "not_planned", "minion-reviewed"),
	})
	got, err := FindDupes(f.client(), DupesQuery{
		Repo: "partio-io/cli", Sources: []string{"entireio/cli PR #2075"}, Title: "Add Goose coding agent integration",
	})
	if err != nil {
		t.Fatal(err)
	}
	want := []Dupe{
		{Number: 690, Title: "Rename a flag", State: "closed", StateReason: "not_planned",
			Labels: []string{"minion-proposal", "minion-reviewed"}, Match: MatchSource, Shared: []string{"entireio/cli#2075"}},
		{Number: 681, Title: "Add Goose coding agent integration", State: "open", StateReason: "",
			Labels: []string{"minion-proposal"}, Match: MatchTitle, Shared: []string{"agent", "coding", "goose", "integration"}},
	}
	if fmt.Sprint(got) != fmt.Sprint(want) {
		t.Errorf("FindDupes = %+v\nwant %+v", got, want)
	}
}

// TestFindDupesFindsConfirmedBacklogPairs serves the real bodies of two confirmed duplicate pairs
// from the backlog, among fillers, and checks that each issue finds its
// twin.
func TestFindDupesFindsConfirmedBacklogPairs(t *testing.T) {
	body := func(n int) string {
		raw, err := os.ReadFile(fmt.Sprintf("testdata/issue-%d-body.md", n))
		if err != nil {
			t.Fatal(err)
		}
		return string(raw)
	}
	titles := map[int]string{
		681: "Add Goose coding agent integration",
		688: "Add Goose coding agent integration",
		710: "installHooks silently overwrites the original hook backup when a foreign hook appears on reinstall",
		736: "Hook installer silently overwrites existing backup on re-install",
	}
	issues := fillers(1, 30)
	for _, n := range []int{681, 688, 710, 736} {
		state, reason := "open", ""
		if n == 688 {
			state, reason = "closed", "completed"
		}
		issues = append(issues, fakeProposal(n, titles[n], body(n), state, reason))
	}
	f := newFakeBacklog(t, issues)

	tests := []struct {
		issue, twin int
		match       string
	}{
		{681, 688, MatchSource},
		{688, 681, MatchSource},
		{710, 736, MatchTitle},
		{736, 710, MatchTitle},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprint(tt.issue), func(t *testing.T) {
			got, err := FindDupes(f.client(), DupesQuery{
				Repo: "partio-io/cli", Sources: SourceRefs(body(tt.issue)), Title: titles[tt.issue], Exclude: tt.issue,
			})
			if err != nil {
				t.Fatal(err)
			}
			if len(got) == 0 || got[0].Number != tt.twin || got[0].Match != tt.match {
				t.Errorf("#%d: FindDupes = %+v, want #%d first by %s", tt.issue, got, tt.twin, tt.match)
			}
		})
	}
}
