package git

import (
	"fmt"
	"strings"
)

// AmendTrailers amends the current HEAD commit to add trailers.
func AmendTrailers(trailers map[string]string) error {
	// Get current commit message
	msg, err := execGit("log", "-1", "--format=%B")
	if err != nil {
		return fmt.Errorf("reading commit message: %w", err)
	}

	// Build the trailer lines
	var trailerLines []string
	for k, v := range trailers {
		trailerLines = append(trailerLines, fmt.Sprintf("%s: %s", k, v))
	}

	// Append trailers to message
	newMsg := strings.TrimRight(msg, "\n") + "\n\n" + strings.Join(trailerLines, "\n")

	// Amend commit with new message (--no-verify prevents re-triggering hooks)
	_, err = execGit("commit", "--amend", "--no-verify", "-m", newMsg)
	if err != nil {
		return fmt.Errorf("amending commit with trailers: %w", err)
	}

	return nil
}
