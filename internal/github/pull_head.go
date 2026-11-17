package github

import (
	"fmt"
	"net/http"
)

// maxPullCommitPages bounds the pull request commit listing. The endpoint
// serves at most 250 commits, so three pages of PerPage cover all of them.
const maxPullCommitPages = 3

// PullRepo is the part of a pull request's head or base repository that
// fork sync reads.
type PullRepo struct {
	FullName string `json:"full_name"`
	CloneURL string `json:"clone_url"`
}

// PullHead describes where a pull request comes from and where it goes.
// Head.Repo is a pointer because GitHub may report no head repository, which
// is what a deleted fork looks like.
type PullHead struct {
	Head struct {
		SHA  string    `json:"sha"`
		Repo *PullRepo `json:"repo"`
	} `json:"head"`
	Base struct {
		Repo PullRepo `json:"repo"`
	} `json:"base"`
}

// PullCommit is one commit listed on a pull request.
type PullCommit struct {
	SHA    string `json:"sha"`
	Commit struct {
		Message string `json:"message"`
	} `json:"commit"`
}

// GetPullHead reads pull request number in repo.
func (c Client) GetPullHead(repo string, number int) (PullHead, error) {
	var pr PullHead
	url := fmt.Sprintf("%s/repos/%s/pulls/%d", c.BaseURL, repo, number)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return pr, fmt.Errorf("read pull request: %w", err)
	}
	if err := c.Do(req, &pr); err != nil {
		return pr, fmt.Errorf("read pull request: %w", err)
	}
	return pr, nil
}

// ListPullCommits lists the commits of pull request number in repo, up to
// the 250 the endpoint serves.
func (c Client) ListPullCommits(repo string, number int) ([]PullCommit, error) {
	var all []PullCommit
	path := fmt.Sprintf("/repos/%s/pulls/%d/commits", repo, number)
	err := GetPages(c, path, maxPullCommitPages, func(page []PullCommit) bool {
		all = append(all, page...)
		return true
	})
	if err != nil {
		return nil, fmt.Errorf("list pull request commits: %w", err)
	}
	return all, nil
}
