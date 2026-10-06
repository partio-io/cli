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

// findTracking returns the tracking issue, or nil when there is none.
func findTracking(gh github.Client, repo string) (*github.Issue, error) {
	issues, err := gh.OpenIssuesWithLabel(repo, TrackingLabel)
	if err != nil {
		return nil, fmt.Errorf("find tracking issue: %w", err)
	}
	for _, is := range issues {
		first, _, _ := strings.Cut(is.Body, "\n")
		if strings.TrimSpace(first) == TrackingMarker {
			return &is, nil
		}
	}
	return nil, nil
}

// findOrCreateTracking returns the tracking issue. When there is none
// it creates the label, if absent, and the issue.
func findOrCreateTracking(gh github.Client, repo string) (github.Issue, error) {
	found, err := findTracking(gh, repo)
	if err != nil {
		return github.Issue{}, err
	}
	if found != nil {
		return *found, nil
	}
	if err := gh.EnsureLabel(repo, TrackingLabel, "5319e7", "Tracking issue of the minion proposal review"); err != nil {
		return github.Issue{}, fmt.Errorf("create tracking label: %w", err)
	}
	created, err := gh.CreateIssue(repo, trackingTitle, trackingBody, []string{TrackingLabel})
	if err != nil {
		return github.Issue{}, fmt.Errorf("create tracking issue: %w", err)
	}
	return created, nil
}

// maxCommentBody is GitHub's limit on a comment body. GitHub counts
// characters; the gate counts bytes, which is never fewer.
const maxCommentBody = 65536

// nightPrefix starts the marker of every night comment.
var nightPrefix, _, _ = strings.Cut(NightMarker(""), " -->")

// appendRow adds r to the last night comment on the tracking issue. It
// creates that comment for the first row of the night, and a
// continuation comment when r would push the last one past GitHub's
// size limit: rewrite rows carry full bodies, so a busy night outgrows
// one comment. It returns the bodies of every night comment, oldest
// first, with r in place.
func appendRow(gh github.Client, repo string, tracking int, night, r string) ([]string, error) {
	marker := NightMarker(night)
	nights, err := nightComments(gh, repo, tracking)
	if err != nil {
		return nil, fmt.Errorf("find night comment: %w", err)
	}
	last := -1
	for i, cm := range nights {
		if strings.HasPrefix(cm.Body, marker) {
			last = i
		}
	}
	bodies := make([]string, len(nights))
	for i, cm := range nights {
		bodies[i] = cm.Body
	}
	if last >= 0 {
		body := strings.TrimRight(nights[last].Body, "\n") + "\n" + r
		if len(body) <= maxCommentBody {
			if err := gh.UpdateComment(repo, nights[last].ID, body); err != nil {
				return nil, fmt.Errorf("update night comment: %w", err)
			}
			bodies[last] = body
			return bodies, nil
		}
	}
	heading := "### Review night " + night
	if last >= 0 {
		heading += " (continued)"
	}
	body := marker + "\n" + heading + "\n\n" + r
	if err := gh.CreateComment(repo, tracking, body); err != nil {
		return nil, fmt.Errorf("create night comment: %w", err)
	}
	return append(bodies, body), nil
}

// nightComments returns the night comments on the tracking issue,
// oldest first.
func nightComments(gh github.Client, repo string, tracking int) ([]github.Comment, error) {
	var out []github.Comment
	path := fmt.Sprintf("/repos/%s/issues/%d/comments", repo, tracking)
	err := github.GetPages(gh, path, 0, func(comments []github.Comment) bool {
		for _, cm := range comments {
			if strings.HasPrefix(cm.Body, nightPrefix) {
				out = append(out, cm)
			}
		}
		return true
	})
	if err != nil {
		return nil, fmt.Errorf("list comments: %w", err)
	}
	return out, nil
}

// writeSummary replaces the tracking issue body with the summary of
// the night comments, so the body cannot drift from the rows.
func writeSummary(gh github.Client, repo string, tracking github.Issue, nights []string) error {
	body := trackingBody + "\n" + summarize(nights).markdown()
	if err := gh.EditIssue(repo, tracking.Number, tracking.Title, body); err != nil {
		return fmt.Errorf("write tracking summary: %w", err)
	}
	return nil
}
