package review

import (
	"regexp"
	"strings"
)

// excerptLen is about how many characters of a body the excerpt keeps.
const excerptLen = 300

// mdEscape is a Markdown backslash escape of a punctuation mark.
var mdEscape = regexp.MustCompile(`\\([!-/:-@\[-` + "`" + `{-~])`)

// mdHeading is an ATX heading line: up to three spaces, one to six
// hash marks, then a space or the end of the line. A line such as
// "#123 is the same idea" is text, not a heading.
var mdHeading = regexp.MustCompile(`^ {0,3}#{1,6}(?:[ \t\r]|$)`)

// mdComment is an HTML comment, such as a "<!-- partio:premise:v1 -->"
// marker, which is not text for a reader.
var mdComment = regexp.MustCompile(`(?s)<!--.*?-->`)

// excerpt returns the start of a proposal body as plain text: heading
// lines and HTML comments dropped, backslash escapes undone, runs of
// white space joined, and cut at the end of a word near excerptLen
// characters.
func excerpt(body string) string {
	var lines []string
	for _, line := range strings.Split(mdComment.ReplaceAllString(body, ""), "\n") {
		if mdHeading.MatchString(line) {
			continue
		}
		lines = append(lines, line)
	}
	text := strings.Join(strings.Fields(mdEscape.ReplaceAllString(strings.Join(lines, "\n"), "$1")), " ")
	runes := []rune(text)
	if len(runes) <= excerptLen {
		return text
	}
	cut := string(runes[:excerptLen+1])
	if i := strings.LastIndex(cut, " "); i > 0 {
		return cut[:i]
	}
	return string(runes[:excerptLen])
}
