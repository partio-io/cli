package config

import "strings"

// RedactSettingsForDisplay masks secret-bearing values in a settings
// map so doctor and status output can print it safely.
func RedactSettingsForDisplay(values map[string]string) map[string]string {
	return RedactSecrets(values)
}

// RedactSecrets masks secret-bearing values in a settings map using a
// fixed-width mask.
func RedactSecrets(values map[string]string) map[string]string {
	out := make(map[string]string, len(values))
	for k, v := range values {
		if !isSecretKey(k) {
			out[k] = v
			continue
		}
		out[k] = "********"
	}
	return out
}

// isSecretKey reports whether a settings key is expected to carry a
// secret value.
func isSecretKey(k string) bool {
	lower := strings.ToLower(k)
	return strings.Contains(lower, "token") ||
		strings.Contains(lower, "secret") ||
		strings.Contains(lower, "password")
}
