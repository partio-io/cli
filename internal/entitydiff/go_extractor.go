package entitydiff

import (
	"strings"
	"unicode"
)

// goExtractor finds Go declarations: top-level functions, methods and
// type specs, including the specs of a grouped "type (...)" block.
// Top-level const and var declarations are not entities.
//
// It relies on gofmt layout: top-level declarations start in column 0 and
// the specs of a type group are indented by exactly one tab.
type goExtractor struct{}

func (goExtractor) extract(lines []string) extraction {
	var ex extraction
	inGroup := false
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		if inGroup {
			if strings.HasPrefix(line, ")") {
				inGroup = false
				continue
			}
			if len(line) > 1 && line[0] == '\t' && leadingIdent(line[1:]) != "" {
				i = ex.addGo(lines, i, goTypeKey(line[1:]), "type ", false)
			}
			continue
		}

		switch {
		case isGoTypeGroup(line):
			inGroup = true
		case strings.HasPrefix(line, "func ") || strings.HasPrefix(line, "func("):
			i = ex.addGo(lines, i, goFuncKey(line), "", true)
		case strings.HasPrefix(line, "type "):
			i = ex.addGo(lines, i, goTypeKey(strings.TrimPrefix(line, "type ")), "", false)
		}
	}
	return ex
}

// addGo records the declaration starting at lines[i] and returns the index
// of its last header line. A declaration with no key is not one partio
// understands and is skipped; one whose header is cut off is recorded as
// partial.
func (ex *extraction) addGo(lines []string, i int, key entityKey, prefix string, isFunc bool) int {
	if key.name == "" {
		return i
	}
	header, end, ok := goHeader(lines, i, isFunc)
	if !ok {
		ex.partial = append(ex.partial, key)
		return i
	}
	ex.entities = append(ex.entities, foundEntity{
		Entity: Entity{Kind: key.kind, Name: key.name, Signature: normalizeSignature(prefix + header)},
		line:   i,
	})
	return end
}

func isGoTypeGroup(line string) bool {
	if c := strings.Index(line, "//"); c >= 0 {
		line = line[:c]
	}
	line = strings.TrimSpace(line)
	return line == "type (" || line == "type("
}

// goFuncKey reads the key of "func Name(" or "func (r *T) Name(" from the
// first line of a declaration.
func goFuncKey(line string) entityKey {
	rest := strings.TrimLeft(strings.TrimPrefix(line, "func"), " \t")
	recv := ""
	if strings.HasPrefix(rest, "(") {
		end := matchingParen(rest)
		if end < 0 {
			return entityKey{}
		}
		recv = rest[1:end]
		rest = strings.TrimLeft(rest[end+1:], " \t")
	}
	name := leadingIdent(rest)
	if name == "" {
		return entityKey{}
	}
	after := strings.TrimLeft(rest[len(name):], " \t")
	if !strings.HasPrefix(after, "(") && !strings.HasPrefix(after, "[") {
		return entityKey{}
	}
	if recv == "" {
		return entityKey{kind: KindFunction, name: name}
	}
	typ := receiverType(recv)
	if typ == "" {
		return entityKey{}
	}
	return entityKey{kind: KindMethod, name: typ + "." + name}
}

// receiverType turns a receiver such as "r *T[K]" into "(*T)", or "s T"
// into "T".
func receiverType(recv string) string {
	if i := strings.IndexByte(recv, '['); i >= 0 {
		recv = recv[:i]
	}
	fields := strings.Fields(recv)
	if len(fields) == 0 {
		return ""
	}
	typ := fields[0]
	if len(fields) > 1 {
		typ = strings.Join(fields[1:], "")
	}
	typ = strings.Trim(typ, "()")
	pointer := strings.HasPrefix(typ, "*")
	typ = strings.TrimPrefix(typ, "*")
	if typ == "" || leadingIdent(typ) != typ {
		return ""
	}
	if pointer {
		return "(*" + typ + ")"
	}
	return typ
}

// goTypeKey reads the key of a type spec "Name ..." with "type" removed.
func goTypeKey(spec string) entityKey {
	spec = strings.TrimLeft(spec, " \t")
	name := leadingIdent(spec)
	if name == "" || strings.TrimSpace(spec[len(name):]) == "" {
		return entityKey{}
	}
	return entityKey{kind: KindType, name: name}
}

// matchingParen returns the index of the ")" closing the "(" at s[0], or
// -1 when it is not on this line.
func matchingParen(s string) int {
	depth := 0
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '(':
			depth++
		case ')':
			depth--
			if depth == 0 {
				return i
			}
		}
	}
	return -1
}

// leadingIdent returns the identifier at the start of s, or "".
func leadingIdent(s string) string {
	for i, r := range s {
		if !isIdentRune(r, i == 0) {
			return s[:i]
		}
	}
	return s
}

func isIdentRune(r rune, first bool) bool {
	if r == '_' || unicode.IsLetter(r) {
		return true
	}
	return !first && unicode.IsDigit(r)
}
