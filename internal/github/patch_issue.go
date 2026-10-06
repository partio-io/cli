package github

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// patchIssue sends fields as one PATCH to issue number in repo. what
// names the change in an error.
func (c Client) patchIssue(repo string, number int, fields map[string]string, what string) error {
	payload, err := json.Marshal(fields)
	if err != nil {
		return fmt.Errorf("encode %s: %w", what, err)
	}
	req, err := http.NewRequest(http.MethodPatch,
		fmt.Sprintf("%s/repos/%s/issues/%d", c.BaseURL, repo, number), bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("%s issue: %w", what, err)
	}
	if err := c.Do(req, nil); err != nil {
		return fmt.Errorf("%s issue: %w", what, err)
	}
	return nil
}
