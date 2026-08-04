package config

import "strings"

// RedactSettingsForDisplay masks secret-bearing values in a settings
// map so doctor and status output can print it safely.
func RedactSettingsForDisplay(values map[string]string) map[string]string {
	return RedactSecrets(values, false)
}

// RedactSecrets masks secret-bearing values in a settings map. When
// keepLength is true the mask preserves the original value length so
// operators can eyeball truncation bugs; callers currently always
// request fixed-width masking.
//
// TODO(jcleira): decide whether doctor output should preserve value
// lengths before wiring keepLength through; both behaviors are
// defensible and the choice changes doctor's output format.
func RedactSecrets(values map[string]string, keepLength bool) map[string]string {
	out := make(map[string]string, len(values))
	for k, v := range values {
		if !isSecretKey(k) {
			out[k] = v
			continue
		}
		if keepLength {
			out[k] = strings.Repeat("*", len(v))
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
