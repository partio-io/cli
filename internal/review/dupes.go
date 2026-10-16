package review

import (
	"cmp"
	"slices"

	"github.com/partio-io/cli/internal/github"
)

// Match reasons, in the order the search returns them.
const (
	MatchSource = "source" // cites the same source item
	MatchTitle  = "title"  // has a strongly overlapping title
	MatchRanked = "ranked" // has one of the best title scores below a strong match
)

// maxRanked is the most ranked candidates the search returns after the
// strong matches.
const maxRanked = 10

// DupesQuery is one idea to look for in the backlog.
type DupesQuery struct {
	Repo    string   // owner/name of the backlog
	Sources []string // source references of the idea, in any form SourceRefs reads
	Title   string
	Exclude int // the issue under review; zero excludes nothing
	// Issue, when set, is the issue under review: the search takes its
	// title and the source items of its body, and excludes it. It
	// replaces Title, Body and Exclude, and adds to Sources.
	Issue int
	Body  string // the body of the idea; with Issue, the body of that issue
}

// Dupe is one backlog issue that may hold the same idea.
type Dupe struct {
	Number      int      `json:"number"`
	Title       string   `json:"title"`
	State       string   `json:"state"`
	StateReason string   `json:"state_reason"`
	Labels      []string `json:"labels"`
	Match       string   `json:"match"`
	Shared      []string `json:"shared"` // the source items or title words both share
	Score       float64  `json:"score"`  // the title score, from 0 to 1
	Excerpt     string   `json:"excerpt"`
}

// FindDupes lists every minion-proposal issue in q.Repo, open and
// closed, and returns those that match q by source item or by title,
// then the best of the rest by title score.
func FindDupes(gh github.Client, q DupesQuery) ([]Dupe, error) {
	proposals, err := listProposals(gh, q.Repo)
	if err != nil {
		return nil, err
	}
	if q.Issue != 0 {
		if q, err = withIssue(gh, q, proposals); err != nil {
			return nil, err
		}
	}
	var want []string
	for _, s := range q.Sources {
		want = append(want, SourceRefs(s)...)
	}
	titles := make([]map[string]bool, len(proposals))
	for i, p := range proposals {
		titles[i] = titleWords(p.Title)
	}
	weights := newWordWeights(titles)
	query := titleWords(q.Title)

	// A ranked hit becomes a Dupe only when it makes the cut, so that
	// the excerpt of each body is built only for the candidates returned.
	type hit struct {
		p     proposal
		words []string
		score float64
	}
	var bySource, byTitle []Dupe
	var ranked []hit
	for i, p := range proposals {
		if p.Number == q.Exclude {
			continue
		}
		score, words := weights.overlap(query, titles[i])
		var shared []string
		for _, ref := range SourceRefs(p.Body) {
			if slices.Contains(want, ref) {
				shared = append(shared, ref)
			}
		}
		switch {
		case len(shared) > 0:
			bySource = append(bySource, dupe(p, MatchSource, shared, score))
		case score >= titleMatch:
			byTitle = append(byTitle, dupe(p, MatchTitle, words, score))
		case score > 0:
			ranked = append(ranked, hit{p, words, score})
		}
	}
	// The older issue of two with equal scores is the one that most
	// likely stays.
	slices.SortFunc(ranked, func(a, b hit) int {
		if a.score != b.score {
			return cmp.Compare(b.score, a.score)
		}
		return cmp.Compare(a.p.Number, b.p.Number)
	})
	out := append(bySource, byTitle...)
	for _, h := range ranked[:min(len(ranked), maxRanked)] {
		out = append(out, dupe(h.p, MatchRanked, h.words, h.score))
	}
	return out, nil
}

// withIssue fills q from the issue under review: from the listing when
// it holds the issue, else from one read of the issue.
func withIssue(gh github.Client, q DupesQuery, proposals []proposal) (DupesQuery, error) {
	i := slices.IndexFunc(proposals, func(p proposal) bool { return p.Number == q.Issue })
	if i >= 0 {
		q.Title, q.Body = proposals[i].Title, proposals[i].Body
	} else {
		is, err := gh.GetIssue(q.Repo, q.Issue)
		if err != nil {
			return q, err
		}
		q.Title, q.Body = is.Title, is.Body
	}
	q.Sources = append(slices.Clone(q.Sources), q.Body)
	q.Exclude = q.Issue
	return q, nil
}

func dupe(p proposal, match string, shared []string, score float64) Dupe {
	return Dupe{
		Number: p.Number, Title: p.Title, State: p.State, StateReason: p.StateReason,
		Labels: p.labelNames(), Match: match, Shared: shared, Score: score, Excerpt: excerpt(p.Body),
	}
}
