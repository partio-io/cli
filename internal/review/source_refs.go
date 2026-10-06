package review

import (
	"regexp"
	"strings"
)

// ghOwner is a GitHub user or organization name: letters, digits and
// hyphens. It never starts with "_", so "_entireio/cli#1_" in italics
// still reads.
const ghOwner = `[A-Za-z0-9][A-Za-z0-9-]*`

// sourceToken finds, in one line, the parts of a source reference:
//
//   - url:   a GitHub issue or pull request URL
//   - short: owner/repo#N shorthand, or "owner/repo #N"
//   - repo:  owner/repo followed by a word that marks a repo, not a
//     path: "PR", "issue", "pull request", "changelog", "release" (in
//     any case), a version or "("
//   - bare:  #N
//
// A url, short or repo token names the repo of the line. A bare #N
// after it takes that repo, as in "entireio/cli PR #2075, #2077" or the
// legacy "source: 'entireio/cli#594 #601'". A bare #N with no repo
// before it on its line names no source item.
var sourceToken = regexp.MustCompile(
	`(?P<url>https?://(?:[\w-]+\.)*github\.com/([\w.-]+)/([\w.-]+)/(?:issues|pull)/(\d+))` +
		"|(?P<short>`?(" + ghOwner + ")/([\\w.-]+)`?\\s?#(\\d+))" +
		"|(?P<repo>`?(" + ghOwner + ")/([\\w.-]+)`?\\s+(?:(?i:PRs?|pull requests?|issues?|changelog|release)\\b|v?\\d+\\.\\d|\\())" +
		`|(?P<bare>#(\d+))`)

// SourceRefs returns the source items that text cites, each normalized
// to a lower-case owner/repo#N, in order of first appearance and
// without repeats.
func SourceRefs(text string) []string {
	var out []string
	seen := map[string]bool{}
	add := func(repo, n string) {
		ref := repo + "#" + n
		if !seen[ref] {
			seen[ref] = true
			out = append(out, ref)
		}
	}
	for _, line := range strings.Split(text, "\n") {
		repo := ""
		for _, m := range sourceToken.FindAllStringSubmatch(line, -1) {
			switch {
			case m[1] != "":
				repo = strings.ToLower(m[2] + "/" + m[3])
				add(repo, m[4])
			case m[5] != "":
				repo = strings.ToLower(m[6] + "/" + m[7])
				add(repo, m[8])
			case m[9] != "":
				repo = strings.ToLower(m[10] + "/" + m[11])
			case repo != "":
				add(repo, m[13])
			}
		}
	}
	return out
}
