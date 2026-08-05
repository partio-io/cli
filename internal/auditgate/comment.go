package auditgate

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
)

// commentPrefix is the shared body prefix for all minion audit
// comments. The gate appends the audit name so different audits on
// the same PR each keep their own comment.
const commentPrefix = "Minion audit — "

// failBody renders the single red-run comment from a fail verdict.
func failBody(auditName string, v Verdict) string {
	var b bytes.Buffer
	fmt.Fprintf(&b, "%s%s audit failed.\n", commentPrefix, auditName)
	for i, f := range v.Findings {
		fmt.Fprintf(&b, "\n%d. `%s` — %s\n", i+1, f.Location, f.Reasoning)
	}
	return b.String()
}

// failClosedBody renders the red-run comment for a run whose verdict
// was missing or malformed: the audit produced nothing the gate can
// trust, so the gate reports that instead of findings.
func failClosedBody(auditName string, cause error) string {
	return fmt.Sprintf("%s%s audit produced no usable verdict; failing closed.\n\nDetail: %v\n",
		commentPrefix, auditName, cause)
}

type comment struct {
	ID   int64  `json:"id"`
	Body string `json:"body"`
}

// upsertComment creates the audit comment on the PR, or updates the
// existing one in place. The existing comment is found by body
// prefix, never by author: minion comments are posted with a human
// token, so the author distinguishes nothing.
func upsertComment(cfg Config, body string) error {
	match := commentPrefix + cfg.AuditName
	existing, err := findComment(cfg, match)
	if err != nil {
		return err
	}
	if existing == nil {
		url := fmt.Sprintf("%s/repos/%s/issues/%d/comments", cfg.APIBaseURL, cfg.Repo, cfg.PRNumber)
		return sendComment(cfg, http.MethodPost, url, body)
	}
	url := fmt.Sprintf("%s/repos/%s/issues/comments/%d", cfg.APIBaseURL, cfg.Repo, existing.ID)
	return sendComment(cfg, http.MethodPatch, url, body)
}

// findComment scans the PR's comments for one whose body starts with
// match, following pagination so a busy PR cannot hide it.
func findComment(cfg Config, match string) (*comment, error) {
	for page := 1; ; page++ {
		url := fmt.Sprintf("%s/repos/%s/issues/%d/comments?per_page=100&page=%d",
			cfg.APIBaseURL, cfg.Repo, cfg.PRNumber, page)
		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			return nil, fmt.Errorf("list comments: %w", err)
		}
		var comments []comment
		if err := doGitHub(cfg, req, &comments); err != nil {
			return nil, fmt.Errorf("list comments: %w", err)
		}
		for _, c := range comments {
			if len(c.Body) >= len(match) && c.Body[:len(match)] == match {
				return &c, nil
			}
		}
		if len(comments) < 100 {
			return nil, nil
		}
	}
}

func sendComment(cfg Config, method, url, body string) error {
	payload, err := json.Marshal(map[string]string{"body": body})
	if err != nil {
		return fmt.Errorf("encode comment: %w", err)
	}
	req, err := http.NewRequest(method, url, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("send comment: %w", err)
	}
	if err := doGitHub(cfg, req, nil); err != nil {
		return fmt.Errorf("send comment: %w", err)
	}
	return nil
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
