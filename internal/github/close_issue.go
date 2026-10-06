package github

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// GitHub's state reasons for a closed issue.
const (
	StateReasonCompleted  = "completed"
	StateReasonNotPlanned = "not_planned"
)

// CloseIssue closes issue number in repo with state reason reason.
func (c Client) CloseIssue(repo string, number int, reason string) error {
	payload, err := json.Marshal(map[string]string{"state": "closed", "state_reason": reason})
	if err != nil {
		return fmt.Errorf("encode close: %w", err)
	}
	req, err := http.NewRequest(http.MethodPatch,
		fmt.Sprintf("%s/repos/%s/issues/%d", c.BaseURL, repo, number), bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("close issue: %w", err)
	}
	if err := c.Do(req, nil); err != nil {
		return fmt.Errorf("close issue: %w", err)
	}
	return nil
}
