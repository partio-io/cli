package github

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// Comment is one issue or pull request comment.
type Comment struct {
	ID   int64  `json:"id"`
	Body string `json:"body"`
}

// UpsertComment creates a comment on issue or pull request number in
// repo, or updates the existing one in place. The existing comment is
// the first whose body starts with prefix, never one found by author:
// minion comments are posted with a human token, so the author
// distinguishes nothing.
func (c Client) UpsertComment(repo string, number int, prefix, body string) error {
	existing, err := c.FindComment(repo, number, prefix)
	if err != nil {
		return err
	}
	if existing == nil {
		return c.CreateComment(repo, number, body)
	}
	return c.UpdateComment(repo, existing.ID, body)
}

// FindComment scans the comments on issue or pull request number for
// the first whose body starts with prefix, following pagination so a
// busy thread cannot hide it. It returns nil when none matches.
func (c Client) FindComment(repo string, number int, prefix string) (*Comment, error) {
	var found *Comment
	path := fmt.Sprintf("/repos/%s/issues/%d/comments", repo, number)
	err := GetPages(c, path, 0, func(comments []Comment) bool {
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

// CreateComment posts a new comment on issue or pull request number.
func (c Client) CreateComment(repo string, number int, body string) error {
	url := fmt.Sprintf("%s/repos/%s/issues/%d/comments", c.BaseURL, repo, number)
	return c.sendComment(http.MethodPost, url, body)
}

// UpdateComment replaces the body of comment id.
func (c Client) UpdateComment(repo string, id int64, body string) error {
	url := fmt.Sprintf("%s/repos/%s/issues/comments/%d", c.BaseURL, repo, id)
	return c.sendComment(http.MethodPatch, url, body)
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
