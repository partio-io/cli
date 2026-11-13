package entitydiff

import "strings"

// pairRenames replaces a removed and an added change with one renamed
// change when they are each other's only candidate: the kind matches and
// the signatures are identical apart from the declared name. When a
// removed or added entity has more than one candidate, both stay as they
// are; partio does not guess. The renamed change takes the added
// entity's position.
func pairRenames(changes []positionedChange) []positionedChange {
	type template struct {
		kind           EntityKind
		prefix, suffix string
	}
	type group struct{ removed, added []int }
	groups := map[template]*group{}
	var order []template
	for i, c := range changes {
		e := c.New
		if c.Kind == Removed {
			e = c.Old
		} else if c.Kind != Added {
			continue
		}
		prefix, suffix, ok := splitAtName(e)
		if !ok {
			continue
		}
		t := template{kind: e.Kind, prefix: prefix, suffix: suffix}
		g := groups[t]
		if g == nil {
			g = &group{}
			groups[t] = g
			order = append(order, t)
		}
		if c.Kind == Removed {
			g.removed = append(g.removed, i)
		} else {
			g.added = append(g.added, i)
		}
	}

	drop := make([]bool, len(changes))
	for _, t := range order {
		g := groups[t]
		if len(g.removed) != 1 || len(g.added) != 1 {
			continue
		}
		r, a := g.removed[0], g.added[0]
		changes[a] = positionedChange{
			Change: Change{Kind: Renamed, Old: changes[r].Old, New: changes[a].New},
			pos:    changes[a].pos,
		}
		drop[r] = true
	}

	out := changes[:0]
	for i, c := range changes {
		if !drop[i] {
			out = append(out, c)
		}
	}
	return out
}

// splitAtName splits e's signature around its declared name: the first
// occurrence of the unqualified name that stands alone as an identifier
// and is outside any parentheses, which skips a Go method's receiver. It
// reports false when the name is not found.
func splitAtName(e Entity) (prefix, suffix string, ok bool) {
	name := e.Name
	if i := strings.LastIndexByte(name, '.'); i >= 0 {
		name = name[i+1:]
	}
	sig := e.Signature
	if name == "" {
		return "", "", false
	}
	depth := 0
	for i := 0; i < len(sig); i++ {
		switch sig[i] {
		case '(':
			depth++
			continue
		case ')':
			depth--
			continue
		}
		if depth != 0 || !strings.HasPrefix(sig[i:], name) {
			continue
		}
		end := i + len(name)
		if (i > 0 && isIdentByte(sig[i-1])) || (end < len(sig) && isIdentByte(sig[end])) {
			continue
		}
		return sig[:i], sig[end:], true
	}
	return "", "", false
}
