package github

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type comment struct {
	ID   int64  `json:"id"`
	Body string `json:"body"`
}

// UpsertComment creates a comment on issue or pull request number in
// repo, or updates the existing one in place. The existing comment is
// the first whose body starts with prefix, never one found by author:
// minion comments are posted with a human token, so the author
// distinguishes nothing.
func (c Client) UpsertComment(repo string, number int, prefix, body string) error {
	existing, err := c.findComment(repo, number, prefix)
	if err != nil {
		return err
	}
	if existing == nil {
		url := fmt.Sprintf("%s/repos/%s/issues/%d/comments", c.BaseURL, repo, number)
		return c.sendComment(http.MethodPost, url, body)
	}
	url := fmt.Sprintf("%s/repos/%s/issues/comments/%d", c.BaseURL, repo, existing.ID)
	return c.sendComment(http.MethodPatch, url, body)
}

// findComment scans the comments for one whose body starts with
// prefix, following pagination so a busy thread cannot hide it.
func (c Client) findComment(repo string, number int, prefix string) (*comment, error) {
	var found *comment
	path := fmt.Sprintf("/repos/%s/issues/%d/comments", repo, number)
	err := GetPages(c, path, 0, func(comments []comment) bool {
		for _, cm := range comments {
			if strings.HasPrefix(cm.Body, prefix) {
				found = &cm
				return false
			}
		}
		return true
	})
	if err != nil {
		return nil, fmt.Errorf("list comments: %w", err)
	}
	return found, nil
}

func (c Client) sendComment(method, url, body string) error {
	payload, err := json.Marshal(map[string]string{"body": body})
	if err != nil {
		return fmt.Errorf("encode comment: %w", err)
	}
	req, err := http.NewRequest(method, url, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("send comment: %w", err)
	}
	if err := c.Do(req, nil); err != nil {
		return fmt.Errorf("send comment: %w", err)
	}
	return nil
}
