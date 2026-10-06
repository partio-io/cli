package github

// GitHub's state reasons for a closed issue.
const (
	StateReasonCompleted  = "completed"
	StateReasonNotPlanned = "not_planned"
)

// CloseIssue closes issue number in repo with state reason reason.
func (c Client) CloseIssue(repo string, number int, reason string) error {
	return c.patchIssue(repo, number, map[string]string{"state": "closed", "state_reason": reason}, "close")
}
