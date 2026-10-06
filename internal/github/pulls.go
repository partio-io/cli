package github

import (
	"fmt"
	"net/url"
	"strings"
)

// PullRequest is the part of a GitHub pull request the minion tools
// read.
type PullRequest struct {
	Number  int    `json:"number"`
	Title   string `json:"title"`
	HTMLURL string `json:"html_url"`
}

// OpenPullsByHead lists the open pull requests in repo whose head is
// branch in repo itself.
func (c Client) OpenPullsByHead(repo, branch string) ([]PullRequest, error) {
	owner, _, _ := strings.Cut(repo, "/")
	var all []PullRequest
	path := fmt.Sprintf("/repos/%s/pulls?state=open&head=%s", repo, url.QueryEscape(owner+":"+branch))
	err := GetPages(c, path, 0, func(page []PullRequest) bool {
		all = append(all, page...)
		return true
	})
	if err != nil {
		return nil, fmt.Errorf("list pulls: %w", err)
	}
	return all, nil
}
