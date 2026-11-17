package main

import (
	"fmt"
	"log/slog"
	"os/exec"
	"strings"

	"github.com/partio-io/cli/internal/checkpoint"
)

const checkpointTrailerKey = "Partio-Checkpoint"

// checkpointIDsFromMessages returns the Partio-Checkpoint trailer values of
// messages, deduplicated in first-seen order. Trailers are read with git's
// own rules (only the final trailer block counts, and a "---" line does not
// end the message); the key matches case-insensitively. Values that are not
// valid checkpoint IDs are warned about once each and skipped.
func checkpointIDsFromMessages(dir string, messages []string) ([]string, error) {
	var ids []string
	seen := map[string]bool{}
	rejected := map[string]bool{}
	for _, msg := range messages {
		values, err := parseTrailerValues(dir, msg, checkpointTrailerKey)
		if err != nil {
			return nil, err
		}
		for _, v := range values {
			if !checkpoint.ValidID(v) {
				if !rejected[v] {
					rejected[v] = true
					slog.Warn("ignoring invalid checkpoint trailer value", "value", v)
				}
				continue
			}
			if !seen[v] {
				seen[v] = true
				ids = append(ids, v)
			}
		}
	}
	return ids, nil
}

// parseTrailerValues returns the values of trailers named key in msg.
func parseTrailerValues(dir, msg, key string) ([]string, error) {
	cmd := exec.Command("git", "interpret-trailers", "--parse", "--no-divider")
	cmd.Dir = dir
	cmd.Stdin = strings.NewReader(msg)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("parsing commit trailers: %w", err)
	}
	var values []string
	for _, line := range strings.Split(string(out), "\n") {
		k, v, ok := strings.Cut(line, ":")
		if ok && strings.EqualFold(strings.TrimSpace(k), key) {
			values = append(values, strings.TrimSpace(v))
		}
	}
	return values, nil
}
