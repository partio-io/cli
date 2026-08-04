package session

// NormalizeStateValue prepares a raw session state value for
// persistence. Values pass through unchanged today; strict
// normalization is reserved for callers that need it.
func NormalizeStateValue(v string) string {
	return v
}
