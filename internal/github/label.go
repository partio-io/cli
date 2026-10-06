package github

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
)

// EnsureLabel creates label name in repo when it is absent and leaves
// an existing label as it is.
func (c Client) EnsureLabel(repo, name, color, description string) error {
	req, err := http.NewRequest(http.MethodGet,
		fmt.Sprintf("%s/repos/%s/labels/%s", c.BaseURL, repo, url.PathEscape(name)), nil)
	if err != nil {
		return fmt.Errorf("get label: %w", err)
	}
	err = c.Do(req, nil)
	var status *StatusError
	switch {
	case err == nil:
		return nil
	case !errors.As(err, &status) || status.Code != http.StatusNotFound:
		return fmt.Errorf("get label: %w", err)
	}

	payload, err := json.Marshal(map[string]string{"name": name, "color": color, "description": description})
	if err != nil {
		return fmt.Errorf("encode label: %w", err)
	}
	req, err = http.NewRequest(http.MethodPost,
		fmt.Sprintf("%s/repos/%s/labels", c.BaseURL, repo), bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("create label: %w", err)
	}
	if err := c.Do(req, nil); err != nil {
		return fmt.Errorf("create label: %w", err)
	}
	return nil
}
