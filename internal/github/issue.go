package github

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// Issue is the part of a GitHub issue the minion tools read.
type Issue struct {
	Number  int    `json:"number"`
	Title   string `json:"title"`
	Body    string `json:"body"`
	HTMLURL string `json:"html_url"`
	// PullRequest is set when the item is a pull request: the issues
	// API lists pull requests as issues.
	PullRequest *struct{} `json:"pull_request,omitempty"`
}

// GetIssue reads issue number in repo.
func (c Client) GetIssue(repo string, number int) (Issue, error) {
	req, err := http.NewRequest(http.MethodGet,
		fmt.Sprintf("%s/repos/%s/issues/%d", c.BaseURL, repo, number), nil)
	if err != nil {
		return Issue{}, fmt.Errorf("get issue: %w", err)
	}
	var issue Issue
	if err := c.Do(req, &issue); err != nil {
		return Issue{}, fmt.Errorf("get issue: %w", err)
	}
	return issue, nil
}

// OpenIssuesWithLabel lists every open issue in repo that carries
// label, following pagination. It leaves out pull requests, which the
// issues API also returns.
func (c Client) OpenIssuesWithLabel(repo, label string) ([]Issue, error) {
	var all []Issue
	path := fmt.Sprintf("/repos/%s/issues?state=open&labels=%s", repo, url.QueryEscape(label))
	err := GetPages(c, path, 0, func(page []Issue) bool {
		for _, is := range page {
			if is.PullRequest == nil {
				all = append(all, is)
			}
		}
		return true
	})
	if err != nil {
		return nil, fmt.Errorf("list issues: %w", err)
	}
	return all, nil
}

// CreateIssue opens a new issue in repo with the given labels.
func (c Client) CreateIssue(repo, title, body string, labels []string) (Issue, error) {
	payload, err := json.Marshal(map[string]any{"title": title, "body": body, "labels": labels})
	if err != nil {
		return Issue{}, fmt.Errorf("encode issue: %w", err)
	}
	req, err := http.NewRequest(http.MethodPost,
		fmt.Sprintf("%s/repos/%s/issues", c.BaseURL, repo), bytes.NewReader(payload))
	if err != nil {
		return Issue{}, fmt.Errorf("create issue: %w", err)
	}
	var issue Issue
	if err := c.Do(req, &issue); err != nil {
		return Issue{}, fmt.Errorf("create issue: %w", err)
	}
	return issue, nil
}
