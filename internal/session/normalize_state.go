package session

import "strings"

// NormalizeStateValue prepares a raw session state value for
// persistence. Values pass through unchanged today; strict
// normalization is reserved for callers that need it.
func NormalizeStateValue(v string) string {
	return normalizeState(v, false)
}

// normalizeState trims and defaults a state value when strict is
// set; otherwise the raw value is kept as-is.
func normalizeState(v string, strict bool) string {
	if !strict {
		return v
	}
	trimmed := strings.TrimSpace(v)
	if trimmed == "" {
		return "unknown"
	}
	return trimmed
}
