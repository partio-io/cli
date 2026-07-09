package redact

import "regexp"

// secretPattern pairs a human-readable name with a compiled regex.
// The regex must contain a named capture group "secret" for the value to redact,
// or the entire match is replaced when no such group exists.
type secretPattern struct {
	name string
	re   *regexp.Regexp
}

// patterns is the list of known secret formats, modelled after gitleaks rules.
var patterns = []secretPattern{
	{
		name: "aws-access-key-id",
		re:   regexp.MustCompile(`(?i)(AKIA|ASIA|AROA|AIDA|ANPA|ANVA|APKA)[A-Z0-9]{16}`),
	},
	{
		name: "github-pat",
		re:   regexp.MustCompile(`ghp_[A-Za-z0-9]{36}`),
	},
	{
		name: "github-oauth",
		re:   regexp.MustCompile(`gho_[A-Za-z0-9]{36}`),
	},
	{
		name: "github-app-token",
		re:   regexp.MustCompile(`(ghu_|ghs_|ghr_)[A-Za-z0-9]{36}`),
	},
	{
		name: "github-fine-grained-pat",
		re:   regexp.MustCompile(`github_pat_[A-Za-z0-9_]{82}`),
	},
	{
		name: "generic-api-key",
		// Matches key/token/secret/password assignments in common formats.
		re: regexp.MustCompile(`(?i)(?:api[_-]?key|api[_-]?token|auth[_-]?token|access[_-]?token|secret[_-]?key|secret[_-]?token)\s*[:=]\s*['"]?([A-Za-z0-9\-_.+/]{20,})['"]?`),
	},
	{
		name: "bearer-token",
		re:   regexp.MustCompile(`(?i)bearer\s+([A-Za-z0-9\-_.+/]{20,})`),
	},
	{
		name: "private-key-header",
		re:   regexp.MustCompile(`-----BEGIN (?:RSA |EC |DSA |OPENSSH )?PRIVATE KEY-----`),
	},
	{
		name: "slack-token",
		re:   regexp.MustCompile(`xox[baprs]-[0-9A-Za-z\-]{10,}`),
	},
	{
		name: "stripe-key",
		re:   regexp.MustCompile(`(?:r|s)k_(?:live|test)_[0-9A-Za-z]{24}`),
	},
	{
		name: "sendgrid-key",
		re:   regexp.MustCompile(`SG\.[0-9A-Za-z\-_]{22}\.[0-9A-Za-z\-_]{43}`),
	},
	{
		name: "twilio-key",
		re:   regexp.MustCompile(`SK[0-9a-fA-F]{32}`),
	},
	{
		name: "npm-token",
		re:   regexp.MustCompile(`npm_[A-Za-z0-9]{36}`),
	},
	{
		name: "pypi-token",
		re:   regexp.MustCompile(`pypi-[A-Za-z0-9\-_]{50,}`),
	},
	{
		name: "google-api-key",
		re:   regexp.MustCompile(`AIza[0-9A-Za-z\-_]{35}`),
	},
}
