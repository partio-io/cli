package entitydiff

// Analyze returns the entity-level changes in a unified diff, one entry
// per file that has any, in diff order. Files in unsupported languages,
// files with a line longer than 1000 bytes, and files with only body
// edits are left out. Analyze never fails: whatever cannot be read is
// logged at debug and skipped.
func Analyze(patch string) []FileChanges {
	if patch == "" {
		return nil
	}
	var out []FileChanges
	for _, u := range buildUnits(parsePatch(patch)) {
		if changes := analyzeUnit(u); len(changes) > 0 {
			out = append(out, FileChanges{OldPath: u.oldPath, NewPath: u.newPath, Changes: changes})
		}
	}
	return out
}

func analyzeUnit(u fileUnit) []Change {
	partial := map[entityKey]bool{}
	oldEnts := extractSides(u.ext, u.oldSides, partial)
	newEnts := extractSides(u.ext, u.newSides, partial)
	return classify(oldEnts, newEnts, partial)
}

// extractSides runs ext over each side separately, since a hunk is the
// largest run of contiguous lines the patch shows.
func extractSides(ext extractor, sides []side, partial map[entityKey]bool) []sideEntity {
	var out []sideEntity
	for _, s := range sides {
		ex := ext.extract(s.lines)
		for _, k := range ex.partial {
			partial[k] = true
		}
		for _, e := range ex.entities {
			out = append(out, sideEntity{Entity: e.Entity, pos: s.pos[e.line]})
		}
	}
	return out
}
