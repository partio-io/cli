package review

import (
	"cmp"
	"fmt"
	"net/url"
	"slices"
	"time"

	"github.com/partio-io/cli/internal/github"
)

// approvedLabel marks a proposal the operator approved for a build. The
// sweep reviews those first, since a build waits on them.
const approvedLabel = "minion-approved"

// NextQuery picks the issues of one sweep run.
type NextQuery struct {
	Repo   string // owner/name
	Sample int    // when above 0, pick this many issues spread across the filing months
}

// openProposal is an open minion-proposal issue with its filing date,
// which github.Issue does not carry.
type openProposal struct {
	github.Issue
	CreatedAt time.Time `json:"created_at"`
}

// Next returns the issues the sweep reviews next, in order: the open
// minion-proposal issues that do not carry minion-reviewed, approved
// ones first, then by filing date, oldest first. GitHub's list filter
// cannot exclude a label, so Next filters the list itself. Next also
// skips an issue the tracking issue lists under "needs you". It changes
// nothing: the gate keeps that list current after each row.
//
// With q.Sample above 0, Next returns that many of the same issues,
// spread across the filing months for a dry run: one per month in
// turn, oldest month first and oldest issue first in each month.
func Next(gh github.Client, q NextQuery) ([]int, error) {
	open, err := listOpenProposals(gh, q.Repo)
	if err != nil {
		return nil, err
	}
	needsYou, err := readNeedsYou(gh, q.Repo)
	if err != nil {
		return nil, err
	}
	var batch []openProposal
	for _, p := range open {
		if !p.HasLabel(ReviewedLabel) && !slices.Contains(needsYou, p.Number) {
			batch = append(batch, p)
		}
	}
	slices.SortStableFunc(batch, func(a, b openProposal) int {
		if a, b := a.HasLabel(approvedLabel), b.HasLabel(approvedLabel); a != b {
			if a {
				return -1
			}
			return 1
		}
		return byFiling(a, b)
	})
	if q.Sample > 0 {
		batch = sample(batch, q.Sample)
	}
	numbers := make([]int, len(batch))
	for i, p := range batch {
		numbers[i] = p.Number
	}
	return numbers, nil
}

func byFiling(a, b openProposal) int {
	return cmp.Or(a.CreatedAt.Compare(b.CreatedAt), cmp.Compare(a.Number, b.Number))
}

// sample takes n proposals, one per filing month in turn. It depends
// only on the filing dates and numbers, so the same backlog gives the
// same sample.
func sample(batch []openProposal, n int) []openProposal {
	batch = slices.SortedStableFunc(slices.Values(batch), byFiling)
	var months [][]openProposal
	for i, p := range batch {
		if i == 0 || monthOf(p) != monthOf(batch[i-1]) {
			months = append(months, nil)
		}
		months[len(months)-1] = append(months[len(months)-1], p)
	}
	var out []openProposal
	for turn := 0; len(out) < n && len(out) < len(batch); turn++ {
		for _, month := range months {
			if turn < len(month) && len(out) < n {
				out = append(out, month[turn])
			}
		}
	}
	return out
}

func monthOf(p openProposal) string {
	return p.CreatedAt.UTC().Format("2006-01")
}

// readNeedsYou returns the issues with too many "no verdict" rows in
// the night comments of the tracking issue. With no tracking issue
// there are none.
func readNeedsYou(gh github.Client, repo string) ([]int, error) {
	tracking, err := findTracking(gh, repo)
	if err != nil || tracking == nil {
		return nil, err
	}
	nights, err := nightComments(gh, repo, tracking.Number)
	if err != nil {
		return nil, err
	}
	bodies := make([]string, len(nights))
	for i, cm := range nights {
		bodies[i] = cm.Body
	}
	return summarize(bodies).needsYou, nil
}

// listOpenProposals returns every open minion-proposal issue. The
// issues API lists pull requests too; it drops them.
func listOpenProposals(gh github.Client, repo string) ([]openProposal, error) {
	var out []openProposal
	path := fmt.Sprintf("/repos/%s/issues?state=open&labels=%s", repo, url.QueryEscape(proposalLabel))
	err := github.GetPages(gh, path, 0, func(page []openProposal) bool {
		for _, p := range page {
			if p.PullRequest == nil {
				out = append(out, p)
			}
		}
		return true
	})
	if err != nil {
		return nil, fmt.Errorf("list open proposals: %w", err)
	}
	return out, nil
}
