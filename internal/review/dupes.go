package review

import (
	"slices"

	"github.com/partio-io/cli/internal/github"
)

// Match reasons, in the order the search returns them.
const (
	MatchSource = "source" // cites the same source item
	MatchTitle  = "title"  // has a strongly overlapping title
)

// DupesQuery is one idea to look for in the backlog.
type DupesQuery struct {
	Repo    string   // owner/name of the backlog
	Sources []string // source references of the idea, in any form SourceRefs reads
	Title   string
	Exclude int // the issue under review; zero excludes nothing
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
}

// FindDupes lists every minion-proposal issue in q.Repo, open and
// closed, and returns those that match q by source item or by title.
func FindDupes(gh github.Client, q DupesQuery) ([]Dupe, error) {
	proposals, err := listProposals(gh, q.Repo)
	if err != nil {
		return nil, err
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

	var bySource, byTitle []Dupe
	for i, p := range proposals {
		if p.Number == q.Exclude {
			continue
		}
		var shared []string
		for _, ref := range SourceRefs(p.Body) {
			if slices.Contains(want, ref) {
				shared = append(shared, ref)
			}
		}
		if len(shared) > 0 {
			bySource = append(bySource, dupe(p, MatchSource, shared))
			continue
		}
		if score, words := weights.overlap(query, titles[i]); score >= titleMatch {
			byTitle = append(byTitle, dupe(p, MatchTitle, words))
		}
	}
	return append(bySource, byTitle...), nil
}

func dupe(p proposal, match string, shared []string) Dupe {
	return Dupe{
		Number: p.Number, Title: p.Title, State: p.State, StateReason: p.StateReason,
		Labels: p.labelNames(), Match: match, Shared: shared,
	}
}
