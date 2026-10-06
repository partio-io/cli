package review

import (
	"fmt"
	"strings"

	"github.com/partio-io/cli/internal/github"
)

const (
	// TrackingLabel marks the tracking issue. The gate lists open
	// issues with it, then picks the one whose body starts with
	// TrackingMarker: never by title or author, which anyone can match.
	TrackingLabel = "minion-review-log"
	// TrackingMarker is the first line of the tracking issue body.
	TrackingMarker = "<!-- minion-review-log -->"

	trackingTitle = "Minion review log"
	trackingBody  = TrackingMarker + "\n" +
		"The proposal review records every verdict here: one comment per night, one row per reviewed issue.\n"
)

// NightMarker is the first line of the night comment for night, a UTC
// date in YYYY-MM-DD form.
func NightMarker(night string) string {
	return "<!-- minion-review-night " + night + " -->"
}

// findOrCreateTracking returns the number of the tracking issue. When
// there is none it creates the label, if absent, and the issue.
func findOrCreateTracking(gh github.Client, repo string) (int, error) {
	issues, err := gh.OpenIssuesWithLabel(repo, TrackingLabel)
	if err != nil {
		return 0, fmt.Errorf("find tracking issue: %w", err)
	}
	for _, is := range issues {
		first, _, _ := strings.Cut(is.Body, "\n")
		if strings.TrimSpace(first) == TrackingMarker {
			return is.Number, nil
		}
	}
	if err := gh.EnsureLabel(repo, TrackingLabel, "5319e7", "Tracking issue of the minion proposal review"); err != nil {
		return 0, fmt.Errorf("create tracking label: %w", err)
	}
	created, err := gh.CreateIssue(repo, trackingTitle, trackingBody, []string{TrackingLabel})
	if err != nil {
		return 0, fmt.Errorf("create tracking issue: %w", err)
	}
	return created.Number, nil
}

// maxCommentBody is GitHub's limit on a comment body. GitHub counts
// characters; the gate counts bytes, which is never fewer.
const maxCommentBody = 65536

// appendRow adds r to the last night comment on the tracking issue. It
// creates that comment for the first row of the night, and a
// continuation comment when r would push the last one past GitHub's
// size limit: rewrite rows carry full bodies, so a busy night outgrows
// one comment.
func appendRow(gh github.Client, repo string, tracking int, night, r string) error {
	marker := NightMarker(night)
	last, err := lastComment(gh, repo, tracking, marker)
	if err != nil {
		return fmt.Errorf("find night comment: %w", err)
	}
	if last != nil {
		body := strings.TrimRight(last.Body, "\n") + "\n" + r
		if len(body) <= maxCommentBody {
			if err := gh.UpdateComment(repo, last.ID, body); err != nil {
				return fmt.Errorf("update night comment: %w", err)
			}
			return nil
		}
	}
	heading := "### Review night " + night
	if last != nil {
		heading += " (continued)"
	}
	if err := gh.CreateComment(repo, tracking, marker+"\n"+heading+"\n\n"+r); err != nil {
		return fmt.Errorf("create night comment: %w", err)
	}
	return nil
}

// lastComment returns the last comment on issue number whose body
// starts with prefix, or nil when none does.
func lastComment(gh github.Client, repo string, number int, prefix string) (*github.Comment, error) {
	var last *github.Comment
	path := fmt.Sprintf("/repos/%s/issues/%d/comments", repo, number)
	err := github.GetPages(gh, path, 0, func(comments []github.Comment) bool {
		for _, cm := range comments {
			if strings.HasPrefix(cm.Body, prefix) {
				last = &cm
			}
		}
		return true
	})
	if err != nil {
		return nil, fmt.Errorf("list comments: %w", err)
	}
	return last, nil
}
