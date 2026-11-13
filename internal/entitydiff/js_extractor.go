package entitydiff

import "strings"

// jsExtractor finds TypeScript and JavaScript declarations: functions,
// classes, interfaces, type aliases, enums, and const/let/var bindings to
// an arrow function or a function expression. Each may carry export,
// default, declare, async, abstract or const modifiers.
//
// It relies on formatter layout: a top-level declaration starts in column
// 0, and a declaration exported from an indented namespace or module
// block starts with "export". Other indented lines (class methods,
// object-literal members, nested code) are never read, so a control-flow
// line such as "if (x) {" or a method such as "  run(x) {" is not
// mistaken for a declaration.
type jsExtractor struct{}

func (jsExtractor) extract(lines []string) extraction {
	text, starts := jsText(lines)
	stops := jsStatementStops(lines)
	var ex extraction
	for i := 0; i < len(lines); i++ {
		if !isJSDeclCandidate(lines[i]) {
			continue
		}
		end := len(text)
		if j := min(i+maxHeaderLines, stops[i]); j < len(lines) {
			end = starts[j] - 1
		}
		d := parseJSDecl(text[starts[i]:end])
		switch d.status {
		case jsComplete:
			ex.entities = append(ex.entities, foundEntity{
				Entity: Entity{Kind: d.key.kind, Name: d.key.name, Signature: normalizeSignature(d.header)},
				line:   i,
			})
			i += d.extraLines
		case jsMore:
			ex.partial = append(ex.partial, d.key)
		}
	}
	return ex
}

// isJSDeclCandidate reports whether line may start a declaration: it
// starts in column 0, or it is an indented "export".
func isJSDeclCandidate(line string) bool {
	if line == "" {
		return false
	}
	if line[0] == ' ' || line[0] == '\t' {
		return strings.HasPrefix(strings.TrimLeft(line, " \t"), "export ")
	}
	return true
}

// jsStatementStops returns, for each line, the index of the next line
// that starts a new top-level statement: one that begins in column 0 with
// an identifier character. A formatted header never runs into such a
// line, so a declaration is read no further, which keeps an unbalanced
// header from scanning far ahead. Lines with no such successor map to
// len(lines).
func jsStatementStops(lines []string) []int {
	stops := make([]int, len(lines))
	next := len(lines)
	for i := len(lines) - 1; i >= 0; i-- {
		stops[i] = next
		if l := lines[i]; l != "" && (l[0] == '$' || isIdentByte(l[0])) {
			next = i
		}
	}
	return stops
}

// jsText joins lines, stripped of comments, with newlines, and returns the
// offset of each line in the text. Each declaration is read from its own
// line onward in a single pass.
func jsText(lines []string) (string, []int) {
	var b strings.Builder
	starts := make([]int, len(lines))
	for i, l := range lines {
		if i > 0 {
			b.WriteByte('\n')
		}
		starts[i] = b.Len()
		b.WriteString(stripJSComments(l))
	}
	return b.String(), starts
}

// stripJSComments removes // and /* */ comments from line, leaving string
// literals alone. A block comment left open runs to the end of the line.
func stripJSComments(line string) string {
	if !strings.Contains(line, "/") {
		return line
	}
	var b strings.Builder
	for j := 0; j < len(line); j++ {
		switch c := line[j]; {
		case c == '"' || c == '\'' || c == '`':
			k := closingQuote(line, j)
			b.WriteString(line[j:k])
			j = k - 1
		case c == '/' && j+1 < len(line) && line[j+1] == '/':
			return b.String()
		case c == '/' && j+1 < len(line) && line[j+1] == '*':
			k := strings.Index(line[j+2:], "*/")
			if k < 0 {
				return b.String()
			}
			b.WriteByte(' ')
			j += k + 3
		default:
			b.WriteByte(c)
		}
	}
	return b.String()
}
