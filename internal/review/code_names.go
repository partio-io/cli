package review

import (
	"path"
	"regexp"
	"strings"
)

var (
	codeFence = regexp.MustCompile("(?s)```.*?```")
	codeSpan  = regexp.MustCompile("`([^`\n]+)`")
)

// commonCodeNames are spans that most proposals carry.
var commonCodeNames = map[string]bool{
	"partio": true, "true": true, "false": true, "nil": true, "git": true, "go": true,
}

// codeNames returns the set of the inline code spans of a body: escaped
// backticks count, fenced blocks do not. Each span is trimmed, in lower
// case, with a "()" at the end cut, and a path with no space becomes its
// base name. Short spans, spans of digits and common names are dropped.
func codeNames(body string) map[string]bool {
	body = strings.ReplaceAll(body, "\\`", "`")
	body = codeFence.ReplaceAllString(body, " ")
	names := map[string]bool{}
	for _, m := range codeSpan.FindAllStringSubmatch(body, -1) {
		n := strings.TrimSuffix(strings.ToLower(strings.TrimSpace(m[1])), "()")
		if strings.Contains(n, "/") && !strings.Contains(n, " ") {
			n = path.Base(strings.TrimRight(n, "/"))
		}
		if len(n) < 3 || commonCodeNames[n] || strings.Trim(n, "0123456789") == "" {
			continue
		}
		names[n] = true
	}
	return names
}
