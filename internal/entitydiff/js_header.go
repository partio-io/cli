package entitydiff

import "strings"

// jsStatus is the outcome of reading a possible TS/JS declaration.
type jsStatus int

const (
	// jsNone means the text is not a declaration partio reports.
	jsNone jsStatus = iota
	// jsComplete means the declaration header was read in full.
	jsComplete
	// jsMore means the header continues past the end of the text.
	jsMore
)

// jsDecl is a TS/JS declaration read from header text. header is the text
// from the first modifier or keyword up to, not including, the body
// opener: "{" for a function, class, interface or enum, "=" for a type
// alias, "=>" for an arrow function. extraLines is how many lines the
// header spans after its first.
type jsDecl struct {
	key        entityKey
	header     string
	status     jsStatus
	extraLines int
}

// jsModifiers may precede a declaration keyword. "const" before "enum" is
// handled with the binding keywords.
var jsModifiers = map[string]bool{
	"export":   true,
	"default":  true,
	"declare":  true,
	"async":    true,
	"abstract": true,
}

// parseJSDecl reads the declaration that starts s: source lines, stripped
// of comments and joined with newlines. Newlines count as spaces except
// where a line end can end a header, such as after a function's
// parameters. The text runs out with status jsMore.
func parseJSDecl(s string) jsDecl {
	c := &jsCursor{s: s}
	w := c.word()
	for jsModifiers[w] {
		w = c.word()
	}
	switch w {
	case "function":
		return c.function()
	case "class":
		return c.named(KindClass)
	case "interface":
		return c.named(KindInterface)
	case "enum":
		return c.named(KindEnum)
	case "type":
		return c.typeAlias()
	case "const":
		if c.peekWord() == "enum" {
			c.word()
			return c.named(KindEnum)
		}
		return c.binding()
	case "let", "var":
		return c.binding()
	}
	return jsDecl{}
}

// jsCursor walks header text. Its methods skip spaces and newlines
// before reading.
type jsCursor struct {
	s string
	i int
}

func (c *jsCursor) skipSpace() {
	for c.i < len(c.s) && (c.s[c.i] == ' ' || c.s[c.i] == '\t' || c.s[c.i] == '\n') {
		c.i++
	}
}

// atEOL reports whether only spaces remain on the cursor's line.
func (c *jsCursor) atEOL() bool {
	j := c.i
	for j < len(c.s) && (c.s[j] == ' ' || c.s[j] == '\t') {
		j++
	}
	return j >= len(c.s) || c.s[j] == '\n'
}

// peek returns the next non-space byte, or 0 at the end of the text.
func (c *jsCursor) peek() byte {
	c.skipSpace()
	if c.i >= len(c.s) {
		return 0
	}
	return c.s[c.i]
}

// at reports whether the text continues with tok.
func (c *jsCursor) at(tok string) bool {
	c.skipSpace()
	return strings.HasPrefix(c.s[c.i:], tok)
}

func (c *jsCursor) peekWord() string {
	c.skipSpace()
	return jsIdent(c.s[c.i:])
}

func (c *jsCursor) word() string {
	w := c.peekWord()
	c.i += len(w)
	return w
}

// atAssign reports whether the text continues with a plain "=", not "=>"
// or "==".
func (c *jsCursor) atAssign() bool {
	return c.peek() == '=' && !c.at("=>") && !c.at("==")
}

func (c *jsCursor) complete(key entityKey) jsDecl {
	h := strings.TrimRight(c.s[:c.i], " \t\n")
	return jsDecl{
		key:        key,
		header:     strings.TrimSpace(h),
		status:     jsComplete,
		extraLines: strings.Count(h, "\n"),
	}
}

func more(key entityKey) jsDecl {
	return jsDecl{key: key, status: jsMore}
}

// function reads "function name<T>(params): R" after the keyword.
func (c *jsCursor) function() jsDecl {
	if c.peek() == '*' {
		c.i++
	}
	name := c.word()
	if name == "" {
		return jsDecl{} // anonymous default export
	}
	return c.callable(entityKey{kind: KindFunction, name: name}, false)
}

// named reads the name and header of a class, interface or enum, up to
// its body's "{".
func (c *jsCursor) named(kind EntityKind) jsDecl {
	name := c.word()
	if name == "" || name == "extends" || name == "implements" {
		return jsDecl{}
	}
	if c.i < len(c.s) && !strings.ContainsRune(" \t\n<{", rune(c.s[c.i])) {
		return jsDecl{}
	}
	key := entityKey{kind: kind, name: name}
	var d jsDepth
	for c.i < len(c.s) {
		if c.skipToken() {
			continue
		}
		ch := c.s[c.i]
		if d.top() {
			switch ch {
			case '{':
				return c.complete(key)
			case ';', '=':
				return jsDecl{}
			}
		}
		if !d.step(ch) {
			return jsDecl{}
		}
		c.i++
	}
	return more(key)
}

// typeAlias reads "type Name<T>" up to its "=".
func (c *jsCursor) typeAlias() jsDecl {
	name := c.word()
	if name == "" {
		return jsDecl{}
	}
	key := entityKey{kind: KindType, name: name}
	if c.peek() == '<' && !c.skipGroup('<', '>') {
		return more(key)
	}
	switch {
	case c.peek() == 0:
		return more(key)
	case c.atAssign():
		return c.complete(key)
	}
	return jsDecl{}
}

// binding reads "const name: T = " followed by an arrow function or a
// function expression. Any other initializer is not an entity.
func (c *jsCursor) binding() jsDecl {
	name := c.word()
	if name == "" {
		return jsDecl{} // destructuring
	}
	key := entityKey{kind: KindFunction, name: name}
	if c.peek() == '!' {
		c.i++
	}
	if c.peek() == ':' {
		c.i++
		if st := c.skipAnnotation(); st != jsComplete {
			if st == jsMore {
				return more(key)
			}
			return jsDecl{}
		}
	}
	if !c.atAssign() {
		return jsDecl{}
	}
	c.i++
	if c.peek() == 0 {
		return more(key)
	}
	if c.peekWord() == "async" {
		c.word()
		if c.peek() == 0 {
			return more(key)
		}
	}
	switch {
	case c.peekWord() == "function":
		c.word()
		if c.peek() == '*' {
			c.i++
		}
		c.word() // optional name of the function expression
		return c.callable(key, false)
	case c.peek() == '(' || c.peek() == '<':
		return c.callable(key, true)
	case c.word() != "" && c.at("=>"):
		return c.complete(key) // single-parameter arrow: x => ...
	}
	return jsDecl{}
}

// skipAnnotation skips a binding's type annotation, stopping at its "=".
// It returns jsNone when the binding has no initializer.
func (c *jsCursor) skipAnnotation() jsStatus {
	var d jsDepth
	for c.i < len(c.s) {
		if c.skipToken() {
			continue
		}
		ch := c.s[c.i]
		if d.top() {
			switch {
			case ch == '\n' && !jsContinues(c.s[:c.i]):
				return jsNone // "let x: T" with no initializer
			case ch == '=' && !c.at("=="):
				return jsComplete
			case ch == ';' || ch == ',':
				return jsNone
			}
		}
		if !d.step(ch) {
			return jsNone
		}
		c.i++
	}
	return jsMore
}

// callable reads "<T>(params): R" and the end of the header. For an arrow
// function the header must end at "=>"; for a function it ends at the
// body's "{", at ";" for an overload or ambient declaration, or at the end
// of a line where nothing is left open.
func (c *jsCursor) callable(key entityKey, arrow bool) jsDecl {
	if c.peek() == '<' && !c.skipGroup('<', '>') {
		return more(key)
	}
	switch c.peek() {
	case 0:
		return more(key)
	case '(':
		if !c.skipGroup('(', ')') {
			return more(key)
		}
	default:
		return jsDecl{}
	}
	if c.atEOL() {
		if arrow {
			return jsDecl{} // a parenthesized expression, not "(params) =>"
		}
		return c.complete(key)
	}
	switch c.peek() {
	case ':':
		c.i++
		return c.returnType(key, arrow)
	case '=':
		if arrow && c.at("=>") {
			return c.complete(key)
		}
	case '{', ';':
		if !arrow {
			return c.complete(key)
		}
	}
	return jsDecl{}
}

// returnType reads a return type annotation up to the end of the header.
// In a function's return type, a "{" that starts a type, such as after
// ":" or "|", opens an object type literal; any other "{" at the top
// level opens the body.
func (c *jsCursor) returnType(key entityKey, arrow bool) jsDecl {
	var d jsDepth
	for c.i < len(c.s) {
		if d.top() && arrow && c.at("=>") {
			return c.complete(key)
		}
		if c.skipToken() {
			continue
		}
		ch := c.s[c.i]
		if d.top() {
			switch ch {
			case '\n':
				if !arrow && !jsContinues(c.s[:c.i]) {
					return c.complete(key)
				}
			case '{':
				if !arrow && !jsTypeLiteralStart(c.s[:c.i]) {
					return c.complete(key)
				}
			case ';':
				if arrow {
					return jsDecl{}
				}
				return c.complete(key)
			case ',', '=':
				return jsDecl{}
			}
		}
		if !d.step(ch) {
			return jsDecl{}
		}
		c.i++
	}
	if arrow || !d.top() || jsContinues(c.s) {
		return more(key)
	}
	return c.complete(key)
}

// skipGroup moves past the bracketed group that opens at the cursor and
// reports whether it closed within the text.
func (c *jsCursor) skipGroup(open, closer byte) bool {
	depth := 0
	for c.i < len(c.s) {
		if c.skipToken() {
			continue
		}
		switch c.s[c.i] {
		case open:
			depth++
		case closer:
			depth--
			if depth == 0 {
				c.i++
				return true
			}
		}
		c.i++
	}
	return false
}

// skipToken moves past a string literal or an "=>" at the cursor, whose
// characters must not count as brackets, and reports whether it did.
func (c *jsCursor) skipToken() bool {
	switch ch := c.s[c.i]; {
	case ch == '"' || ch == '\'' || ch == '`':
		c.i = closingQuote(c.s, c.i)
		return true
	case ch == '=' && strings.HasPrefix(c.s[c.i:], "=>"):
		c.i += 2
		return true
	}
	return false
}

// jsDepth tracks bracket nesting in a header. Angle brackets are counted
// so that a "{" inside a type argument such as Promise<{ a: T }> is not
// read as a body opener.
type jsDepth struct {
	paren, bracket, brace, angle int
}

func (d *jsDepth) top() bool {
	return d.paren == 0 && d.bracket == 0 && d.brace == 0 && d.angle == 0
}

// step updates the nesting for ch. It reports false for a closing bracket
// with no opener, which means the text is not a header.
func (d *jsDepth) step(ch byte) bool {
	switch ch {
	case '(':
		d.paren++
	case '[':
		d.bracket++
	case '{':
		d.brace++
	case '<':
		d.angle++
	case ')':
		d.paren--
		return d.paren >= 0
	case ']':
		d.bracket--
		return d.bracket >= 0
	case '}':
		d.brace--
		return d.brace >= 0
	case '>':
		if d.angle > 0 {
			d.angle--
		}
	}
	return true
}

// jsTypeLiteralStart reports whether a "{" following prefix begins a type
// rather than a function body.
func jsTypeLiteralStart(prefix string) bool {
	prefix = strings.TrimRight(prefix, " \t\n")
	if prefix == "" || strings.HasSuffix(prefix, "=>") {
		return true
	}
	return strings.IndexByte(":|&,(<[?", prefix[len(prefix)-1]) >= 0
}

// jsContinues reports whether s ends in a way that needs another line,
// such as a ":" or "|" with the type still to come.
func jsContinues(s string) bool {
	s = strings.TrimRight(s, " \t\n")
	if s == "" || strings.HasSuffix(s, "=>") {
		return true
	}
	return strings.IndexByte(":|&,.?", s[len(s)-1]) >= 0
}

// jsIdent returns the JS identifier at the start of s, or "".
func jsIdent(s string) string {
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == '$' || isIdentByte(c) {
			if i == 0 && '0' <= c && c <= '9' {
				return ""
			}
			continue
		}
		return s[:i]
	}
	return s
}
