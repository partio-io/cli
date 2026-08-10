package repairround

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
)

// perPage is GitHub's maximum page size for the pull-request commits
// endpoint.
const perPage = 100

// maxPages bounds the walk. The endpoint serves at most 250 commits,
// so three pages cover every branch it can describe. Undercounting
// would hand out rounds the cap already spent, so the bound is
// deliberately above what the API can return rather than at it.
const maxPages = 3

// prSubjects returns the subject line of every commit on the pull
// request's branch.
func prSubjects(cfg Config) ([]string, error) {
	var subjects []string
	for page := 1; page <= maxPages; page++ {
		url := fmt.Sprintf("%s/repos/%s/pulls/%d/commits?per_page=%d&page=%d",
			cfg.APIBaseURL, cfg.Repo, cfg.PRNumber, perPage, page)
		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			return nil, fmt.Errorf("list pull request commits: %w", err)
		}
		var commits []struct {
			Commit struct {
				Message string `json:"message"`
			} `json:"commit"`
		}
		if err := doGitHub(cfg, req, &commits); err != nil {
			return nil, fmt.Errorf("list pull request commits: %w", err)
		}
		for _, c := range commits {
			subjects = append(subjects, subjectOf(c.Commit.Message))
		}
		if len(commits) < perPage {
			break
		}
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
	if err := doGitHub(cfg, req, &pr); err != nil {
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

// doGitHub executes one GitHub API request and decodes the response
// into out when out is non-nil.
func doGitHub(cfg Config, req *http.Request, out any) error {
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Authorization", "Bearer "+cfg.Token)
	client := cfg.HTTPClient
	if client == nil {
		client = http.DefaultClient
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			slog.Debug("closing response body", "url", req.URL.Path, "error", closeErr)
		}
	}()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		detail, readErr := io.ReadAll(io.LimitReader(resp.Body, 512))
		if readErr != nil {
			detail = fmt.Appendf(nil, "(error body unreadable: %v)", readErr)
		}
		return fmt.Errorf("github: %s %s: %s: %s", req.Method, req.URL.Path, resp.Status, detail)
	}
	if out == nil {
		return nil
	}
	return json.NewDecoder(resp.Body).Decode(out)
}
