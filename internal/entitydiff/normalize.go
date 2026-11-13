package entitydiff

import "strings"

// normalizeSignature makes a declaration header comparable across
// formatting: whitespace runs collapse to one space, no space follows an
// opening bracket or precedes a closing bracket or comma, every comma is
// followed by one space, and a trailing comma before a closing bracket is
// dropped. A gofmt reflow of a multi-line parameter list therefore
// normalizes to the same text as the one-line form.
func normalizeSignature(s string) string {
	fields := strings.Fields(s)
	s = strings.Join(fields, " ")

	var b strings.Builder
	b.Grow(len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		next := nextNonSpace(s, i+1)
		switch c {
		case ' ':
			prev := byte(0)
			if b.Len() > 0 {
				prev = b.String()[b.Len()-1]
			}
			if prev == '(' || prev == '[' || prev == ' ' || next == ')' || next == ']' || next == ',' || next == 0 {
				continue
			}
		case ',':
			if next == ')' || next == ']' {
				continue
			}
			b.WriteString(", ")
			for i+1 < len(s) && s[i+1] == ' ' {
				i++
			}
			continue
		}
		b.WriteByte(c)
	}
	return strings.TrimSpace(b.String())
}

// nextNonSpace returns the first non-space byte of s at or after i, or 0.
func nextNonSpace(s string, i int) byte {
	for ; i < len(s); i++ {
		if s[i] != ' ' {
			return s[i]
		}
	}
	return 0
}
