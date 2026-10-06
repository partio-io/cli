package github

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
)

// AddLabels adds labels to issue number in repo. Labels the issue
// already carries stay as they are.
func (c Client) AddLabels(repo string, number int, labels ...string) error {
	payload, err := json.Marshal(map[string][]string{"labels": labels})
	if err != nil {
		return fmt.Errorf("encode labels: %w", err)
	}
	req, err := http.NewRequest(http.MethodPost,
		fmt.Sprintf("%s/repos/%s/issues/%d/labels", c.BaseURL, repo, number), bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("add labels: %w", err)
	}
	if err := c.Do(req, nil); err != nil {
		return fmt.Errorf("add labels: %w", err)
	}
	return nil
}

// RemoveLabel removes label name from issue number in repo. A label the
// issue does not carry is not an error: GitHub answers 404 for it.
func (c Client) RemoveLabel(repo string, number int, name string) error {
	req, err := http.NewRequest(http.MethodDelete,
		fmt.Sprintf("%s/repos/%s/issues/%d/labels/%s", c.BaseURL, repo, number, url.PathEscape(name)), nil)
	if err != nil {
		return fmt.Errorf("remove label: %w", err)
	}
	err = c.Do(req, nil)
	var status *StatusError
	if err != nil && (!errors.As(err, &status) || status.Code != http.StatusNotFound) {
		return fmt.Errorf("remove label: %w", err)
	}
	return nil
}
