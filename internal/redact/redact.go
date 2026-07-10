// Package redact provides two-layer secret detection and redaction:
//  1. Pattern-based detection using gitleaks-compatible regexes for known secret formats.
//  2. Entropy-based scanning to catch high-entropy strings that don't match known patterns.
//
// Call [Text] to redact a single string, or [SessionFiles] to redact all fields
// of a checkpoint.SessionFiles value before it is written to the metadata branch.
package redact

import (
	"regexp"
	"strings"
	"unicode"

	"github.com/partio-io/cli/internal/checkpoint"
)

const (
	// Redacted is the placeholder substituted for detected secrets.
	Redacted = "[REDACTED]"

	// DefaultEntropyThreshold is the minimum Shannon entropy (bits/char) at which
	// a token is considered a potential secret. Strings at or above this threshold
	// that also satisfy DefaultEntropyMinLength are redacted.
	DefaultEntropyThreshold = 4.5

	// DefaultEntropyMinLength is the minimum token length considered for entropy
	// scanning. Short tokens produce unreliable entropy scores.
	DefaultEntropyMinLength = 20
)

// Options controls the redaction behaviour.
type Options struct {
	// Enabled toggles redaction entirely. When false, Text returns its input unchanged.
	Enabled bool

	// EntropyThreshold is the Shannon entropy value (bits per character) above
	// which a token is redacted. Defaults to DefaultEntropyThreshold.
	EntropyThreshold float64

	// EntropyMinLength is the minimum token length considered for entropy scanning.
	// Defaults to DefaultEntropyMinLength.
	EntropyMinLength int
}

// DefaultOptions returns Options with sensible defaults.
func DefaultOptions() Options {
	return Options{
		Enabled:          true,
		EntropyThreshold: DefaultEntropyThreshold,
		EntropyMinLength: DefaultEntropyMinLength,
	}
}

// Text applies pattern-based and entropy-based redaction to input and returns
// the sanitised result. When opts.Enabled is false the input is returned as-is.
func Text(input string, opts Options) string {
	if !opts.Enabled || input == "" {
		return input
	}

	// Layer 1 – pattern-based redaction.
	result := applyPatterns(input)

	// Layer 2 – entropy-based redaction.
	result = applyEntropy(result, opts.EntropyThreshold, opts.EntropyMinLength)

	return result
}

// SessionFiles redacts all user-visible text fields in sf in place and returns sf.
// The ContentHash field is a git commit hash and is intentionally left untouched.
func SessionFiles(sf *checkpoint.SessionFiles, opts Options) *checkpoint.SessionFiles {
	if sf == nil || !opts.Enabled {
		return sf
	}
	sf.Prompt = Text(sf.Prompt, opts)
	sf.Context = Text(sf.Context, opts)
	sf.Diff = Text(sf.Diff, opts)
	sf.FullJSONL = Text(sf.FullJSONL, opts)
	sf.Plan = Text(sf.Plan, opts)
	return sf
}

// applyPatterns replaces all matches of known secret patterns with [REDACTED].
func applyPatterns(s string) string {
	for _, p := range patterns {
		// If the pattern has a subgroup, replace only that subgroup; otherwise
		// replace the full match. This preserves surrounding context (e.g. the
		// "api_key=" prefix) while still hiding the value.
		s = replaceSubgroups(p.re, s)
	}
	return s
}

// replaceSubgroups replaces all capturing subgroup matches (if any) within re
// with [REDACTED]. When the regex has no subgroups the full match is replaced.
func replaceSubgroups(re *regexp.Regexp, s string) string {
	names := re.SubexpNames()
	hasGroup := len(names) > 1 // index 0 is the whole match

	return re.ReplaceAllStringFunc(s, func(match string) string {
		if !hasGroup {
			return Redacted
		}
		// Replace each subgroup match within the full match string.
		submatches := re.FindStringSubmatchIndex(match)
		if submatches == nil {
			return Redacted
		}
		result := match
		// Walk subgroups in reverse so offsets remain valid as we substitute.
		for i := len(submatches)/2 - 1; i >= 1; i-- {
			start, end := submatches[i*2], submatches[i*2+1]
			if start < 0 {
				continue
			}
			result = result[:start] + Redacted + result[end:]
		}
		return result
	})
}

// tokenSplitter is a regexp that splits text into whitespace-delimited tokens.
var tokenSplitter = regexp.MustCompile(`\S+`)

// applyEntropy scans each whitespace-delimited token in s and replaces those
// whose Shannon entropy exceeds threshold (and whose length meets minLen) with
// [REDACTED], unless the token looks like a common non-secret (e.g. a URL, a
// file path, or an already-redacted placeholder).
func applyEntropy(s string, threshold float64, minLen int) string {
	if threshold <= 0 {
		return s
	}
	return tokenSplitter.ReplaceAllStringFunc(s, func(token string) string {
		if len(token) < minLen {
			return token
		}
		// Skip already-redacted placeholders.
		if token == Redacted {
			return token
		}
		// Skip URL-shaped tokens and common file paths – they're high-entropy
		// but almost never secrets.
		if looksLikeURL(token) || looksLikeFilePath(token) {
			return token
		}
		// Only consider tokens that look like they could be keys: no spaces
		// (already guaranteed by the splitter) and composed of base64/hex chars.
		if !looksLikeEncodedString(token) {
			return token
		}
		if shannonEntropy(token) >= threshold {
			return Redacted
		}
		return token
	})
}

func looksLikeURL(s string) bool {
	return strings.HasPrefix(s, "http://") ||
		strings.HasPrefix(s, "https://") ||
		strings.HasPrefix(s, "ftp://") ||
		strings.HasPrefix(s, "git@")
}

func looksLikeFilePath(s string) bool {
	return strings.HasPrefix(s, "/") || strings.HasPrefix(s, "./") || strings.HasPrefix(s, "../")
}

// looksLikeEncodedString returns true when s consists only of characters
// commonly found in base64, hex, or JWT-style strings. This guards against
// false-positives on natural-language tokens (which score low entropy anyway,
// but filtering here avoids the entropy calculation entirely).
func looksLikeEncodedString(s string) bool {
	// Strip common wrapper punctuation (quotes, trailing comma/semicolon).
	s = strings.Trim(s, `"';,`)
	if len(s) == 0 {
		return false
	}
	for _, r := range s {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) &&
			r != '+' && r != '/' && r != '=' && r != '-' && r != '_' && r != '.' {
			return false
		}
	}
	return true
}
