package entitydiff

import "sort"

// sideEntity is an entity found on one side of a file, with the patch
// position of its first line.
type sideEntity struct {
	Entity
	pos int
}

// positionedChange is a change and the position it sorts by.
type positionedChange struct {
	Change
	pos int
}

// classify compares a file's old and new entities, keyed by kind and
// qualified name, and returns the changes in source order. Keys listed in
// partial were cut off at a hunk boundary on some side and are dropped.
//
// For each key, entities with an identical signature on both sides cancel
// out: they are unchanged or had body-only edits. One remaining entity on
// each side is a signature change; anything else remaining is removed or
// added.
func classify(oldEnts, newEnts []sideEntity, partial map[entityKey]bool) []Change {
	var order []entityKey
	oldBy := map[entityKey][]sideEntity{}
	newBy := map[entityKey][]sideEntity{}
	for _, e := range oldEnts {
		k := e.key()
		if _, seen := oldBy[k]; !seen {
			if _, seenNew := newBy[k]; !seenNew {
				order = append(order, k)
			}
		}
		oldBy[k] = append(oldBy[k], e)
	}
	for _, e := range newEnts {
		k := e.key()
		if _, seen := newBy[k]; !seen {
			if _, seenOld := oldBy[k]; !seenOld {
				order = append(order, k)
			}
		}
		newBy[k] = append(newBy[k], e)
	}

	var changes []positionedChange
	for _, k := range order {
		if partial[k] {
			continue
		}
		olds, news := cancelSame(oldBy[k], newBy[k])
		if len(olds) == 1 && len(news) == 1 {
			changes = append(changes, positionedChange{
				Change: Change{Kind: SignatureChanged, Old: olds[0].Entity, New: news[0].Entity},
				pos:    news[0].pos,
			})
			continue
		}
		for _, e := range olds {
			changes = append(changes, positionedChange{Change: Change{Kind: Removed, Old: e.Entity}, pos: e.pos})
		}
		for _, e := range news {
			changes = append(changes, positionedChange{Change: Change{Kind: Added, New: e.Entity}, pos: e.pos})
		}
	}

	sort.SliceStable(changes, func(i, j int) bool { return changes[i].pos < changes[j].pos })
	out := make([]Change, len(changes))
	for i, c := range changes {
		out[i] = c.Change
	}
	return out
}

// cancelSame removes pairs of entities with identical signatures from the
// two sides and returns what is left of each.
func cancelSame(olds, news []sideEntity) ([]sideEntity, []sideEntity) {
	if len(olds) == 0 || len(news) == 0 {
		return olds, news
	}
	used := make([]bool, len(news))
	var leftOld []sideEntity
	for _, o := range olds {
		matched := false
		for j, n := range news {
			if !used[j] && n.Signature == o.Signature {
				used[j] = true
				matched = true
				break
			}
		}
		if !matched {
			leftOld = append(leftOld, o)
		}
	}
	var leftNew []sideEntity
	for j, n := range news {
		if !used[j] {
			leftNew = append(leftNew, n)
		}
	}
	return leftOld, leftNew
}
