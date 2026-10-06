package github

import (
	"fmt"
	"time"
)

// TimelineEvent is the part of an issue timeline event the minion tools
// read. Event names its kind, such as "commented", "closed" or
// "reopened". ID and Body are set on a "commented" event only, and
// UpdatedAt is the time of the comment's last edit.
type TimelineEvent struct {
	Event     string    `json:"event"`
	ID        int64     `json:"id"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Timeline reads every event on issue number in repo, oldest first,
// following pagination.
func (c Client) Timeline(repo string, number int) ([]TimelineEvent, error) {
	var all []TimelineEvent
	path := fmt.Sprintf("/repos/%s/issues/%d/timeline", repo, number)
	err := GetPages(c, path, 0, func(page []TimelineEvent) bool {
		all = append(all, page...)
		return true
	})
	if err != nil {
		return nil, fmt.Errorf("read timeline of #%d: %w", number, err)
	}
	return all, nil
}
