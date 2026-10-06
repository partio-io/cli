package repairround

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/partio-io/cli/internal/github"
)

// maxPages bounds the walk. The endpoint serves at most 250 commits,
// so three pages cover every branch it can describe. Undercounting
// would hand out rounds the cap already spent, so the bound is
// deliberately above what the API can return rather than at it.
const maxPages = 3

// prSubjects returns the subject line of every commit on the pull
// request's branch.
func prSubjects(cfg Config) ([]string, error) {
	type prCommit struct {
		Commit struct {
			Message string `json:"message"`
		} `json:"commit"`
	}
	var subjects []string
	path := fmt.Sprintf("/repos/%s/pulls/%d/commits", cfg.Repo, cfg.PRNumber)
	err := github.GetPages(cfg.client(), path, maxPages, func(commits []prCommit) bool {
		for _, c := range commits {
			subjects = append(subjects, subjectOf(c.Commit.Message))
		}
		return true
	})
	if err != nil {
		return nil, fmt.Errorf("list pull request commits: %w", err)
	}
	return subjects, nil
}

// prMeta returns the pull request's label names and the repository its
// head branch lives in.
//
// The head repository is empty when GitHub reports none, which is what
// a deleted fork looks like. That is an unknown head, not a local one,
// and Eligible refuses it.
func prMeta(cfg Config) (labels []string, headRepo string, err error) {
	url := fmt.Sprintf("%s/repos/%s/pulls/%d", cfg.APIBaseURL, cfg.Repo, cfg.PRNumber)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, "", fmt.Errorf("read pull request: %w", err)
	}
	var pr struct {
		Labels []struct {
			Name string `json:"name"`
		} `json:"labels"`
		Head struct {
			Repo struct {
				FullName string `json:"full_name"`
			} `json:"repo"`
		} `json:"head"`
	}
	if err := cfg.client().Do(req, &pr); err != nil {
		return nil, "", fmt.Errorf("read pull request: %w", err)
	}
	for _, l := range pr.Labels {
		labels = append(labels, l.Name)
	}
	return labels, pr.Head.Repo.FullName, nil
}

// subjectOf returns a commit message's first line. A body that quotes
// a marker must not read as one, so only the subject is kept.
func subjectOf(message string) string {
	subject, _, _ := strings.Cut(message, "\n")
	return subject
}

// client is the GitHub client cfg describes.
func (c Config) client() github.Client {
	return github.Client{BaseURL: c.APIBaseURL, Token: c.Token, HTTPClient: c.HTTPClient}
}
