package entitydiff

import "strings"

// maxHeaderLines bounds how many lines a declaration header may span.
// A longer one is treated as cut off and not reported.
const maxHeaderLines = 64

// goHeader joins the declaration header that starts at lines[start] until
// its brackets balance and its body opener is reached. It returns the
// header text without the body opener and the index of its last line.
// ok is false when the header runs past the end of lines.
//
// For a function, a "{" after "struct" or "interface" opens a type
// literal in the signature, not the body. For a type, the first "{"
// outside brackets opens the struct or interface body.
func goHeader(lines []string, start int, isFunc bool) (header string, end int, ok bool) {
	var b strings.Builder
	paren, bracket, brace := 0, 0, 0
	last := min(len(lines), start+maxHeaderLines)
	for i := start; i < last; i++ {
		line := lines[i]
		if i == start {
			line = strings.TrimLeft(line, " \t")
		} else {
			b.WriteByte(' ')
		}
		for j := 0; j < len(line); j++ {
			c := line[j]
			switch c {
			case '/':
				if j+1 < len(line) && line[j+1] == '/' {
					j = len(line)
					continue
				}
				if j+1 < len(line) && line[j+1] == '*' {
					if k := strings.Index(line[j+2:], "*/"); k >= 0 {
						j += k + 3
					} else {
						j = len(line)
					}
					continue
				}
			case '"', '\'', '`':
				k := closingQuote(line, j)
				b.WriteString(line[j:k])
				j = k - 1
				continue
			case '(':
				paren++
			case ')':
				paren--
			case '[':
				bracket++
			case ']':
				bracket--
			case '{':
				if paren <= 0 && bracket <= 0 && brace <= 0 && (!isFunc || !endsWithTypeLiteral(b.String())) {
					return b.String(), i, true
				}
				brace++
			case '}':
				brace--
			}
			b.WriteByte(c)
		}
		if paren <= 0 && bracket <= 0 && brace <= 0 {
			return b.String(), i, true
		}
	}
	return "", start, false
}

// closingQuote returns the index just past the string or rune literal
// that opens at line[j], or len(line) when it does not close on this line.
func closingQuote(line string, j int) int {
	q := line[j]
	for k := j + 1; k < len(line); k++ {
		switch {
		case line[k] == '\\' && q != '`':
			k++
		case line[k] == q:
			return k + 1
		}
	}
	return len(line)
}

// endsWithTypeLiteral reports whether s ends with the keyword "struct" or
// "interface", so that a following "{" opens a type literal.
func endsWithTypeLiteral(s string) bool {
	s = strings.TrimRight(s, " \t")
	for _, kw := range []string{"struct", "interface"} {
		if strings.HasSuffix(s, kw) {
			rest := s[:len(s)-len(kw)]
			if rest == "" || !isIdentByte(rest[len(rest)-1]) {
				return true
			}
		}
	}
	return false
}

func isIdentByte(c byte) bool {
	return c == '_' || c >= 0x80 ||
		('a' <= c && c <= 'z') || ('A' <= c && c <= 'Z') || ('0' <= c && c <= '9')
}
