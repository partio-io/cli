package patchapply

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
)

// prHead is the branch a repair round pushes to, and the repository
// that branch lives in.
type prHead struct {
	repo string // owner/name
	ref  string // branch name
}

// resolveHead reports the pull request's head.
//
// A caller that already knows the head says so and no request is made.
// That is how the tests stay off the network, and how a workflow that
// has already read the head avoids asking for it twice.
func resolveHead(cfg Config) (prHead, error) {
	if cfg.HeadRepo != "" && cfg.HeadRef != "" {
		return prHead{repo: cfg.HeadRepo, ref: cfg.HeadRef}, nil
	}
	if cfg.APIBaseURL == "" {
		return prHead{}, errors.New("the pull request head is unknown and no GitHub API root is set")
	}

	url := fmt.Sprintf("%s/repos/%s/pulls/%d", cfg.APIBaseURL, cfg.Repo, cfg.PRNumber)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return prHead{}, fmt.Errorf("read pull request head: %w", err)
	}

	var pr struct {
		Head struct {
			Ref  string `json:"ref"`
			Repo struct {
				FullName string `json:"full_name"`
			} `json:"repo"`
		} `json:"head"`
	}
	if err := doGitHub(cfg, req, &pr); err != nil {
		return prHead{}, fmt.Errorf("read pull request head: %w", err)
	}
	if pr.Head.Ref == "" || pr.Head.Repo.FullName == "" {
		return prHead{}, fmt.Errorf(
			"read pull request head: GitHub named no branch or no repository for pull request %d", cfg.PRNumber)
	}
	return prHead{repo: pr.Head.Repo.FullName, ref: pr.Head.Ref}, nil
}

// doGitHub executes one GitHub API request and decodes the response
// into out.
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
	return json.NewDecoder(resp.Body).Decode(out)
}
