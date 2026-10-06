package review

import (
	"fmt"
	"net/url"

	"github.com/partio-io/cli/internal/github"
)

// proposalLabel marks an issue the proposer filed.
const proposalLabel = "minion-proposal"

// proposal is one minion-proposal issue as the list endpoint sends it.
type proposal struct {
	Number      int            `json:"number"`
	Title       string         `json:"title"`
	Body        string         `json:"body"`
	State       string         `json:"state"`
	StateReason string         `json:"state_reason"`
	Labels      []github.Label `json:"labels"`
	PullRequest *struct{}      `json:"pull_request,omitempty"`
}

func (p proposal) labelNames() []string {
	names := make([]string, len(p.Labels))
	for i, l := range p.Labels {
		names[i] = l.Name
	}
	return names
}

// listProposals returns every minion-proposal issue in repo, open and
// closed, from the list endpoint and not from the search index, which
// lags behind new issues.
func listProposals(gh github.Client, repo string) ([]proposal, error) {
	path := fmt.Sprintf("/repos/%s/issues?state=all&labels=%s", repo, url.QueryEscape(proposalLabel))
	var out []proposal
	err := github.GetPages(gh, path, 0, func(page []proposal) bool {
		for _, p := range page {
			if p.PullRequest == nil {
				out = append(out, p)
			}
		}
		return true
	})
	if err != nil {
		return nil, fmt.Errorf("list %s issues: %w", proposalLabel, err)
	}
	return out, nil
}
