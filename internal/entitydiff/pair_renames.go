package entitydiff

import (
	"sort"
	"strings"
)

// minRenameSimilarity is the share of lines a deleted file and an added
// file must have in common to be read as one renamed file. It matches
// git's default rename threshold of 50%.
const minRenameSimilarity = 0.5

// fileUnit is one file to classify: its paths, its extractor, and the old
// and new sides of its hunks. For a rename written as a delete plus an
// add, the old sides come from the deleted file and the new sides from
// the added one.
type fileUnit struct {
	oldPath, newPath string
	ext              extractor
	oldSides         []side
	newSides         []side
}

// buildUnits keeps the supported files of a patch, in diff order, and
// joins a deleted file and an added file into one renamed unit when their
// contents are similar. That is how a rename looks with diff.renames=false.
func buildUnits(files []*filePatch) []fileUnit {
	var units []*fileUnit
	var deleted, added []int
	for _, f := range files {
		path := f.newPath
		if path == "" {
			path = f.oldPath
		}
		ext := extractorFor(path)
		if ext == nil || f.longLine {
			continue
		}
		u := &fileUnit{oldPath: f.oldPath, newPath: f.newPath, ext: ext}
		for _, h := range f.hunks {
			u.oldSides = append(u.oldSides, h.old)
			u.newSides = append(u.newSides, h.new)
		}
		switch {
		case f.oldPath != "" && f.newPath == "":
			deleted = append(deleted, len(units))
		case f.oldPath == "" && f.newPath != "":
			added = append(added, len(units))
		}
		units = append(units, u)
	}

	merged := pairDeletedAdded(units, deleted, added)
	out := make([]fileUnit, 0, len(units))
	for i, u := range units {
		if !merged[i] {
			out = append(out, *u)
		}
	}
	return out
}

// pairDeletedAdded merges each chosen added unit into its deleted partner,
// highest similarity first, one-to-one. The merged unit keeps the earlier
// position in the diff. It returns the indexes that were absorbed.
func pairDeletedAdded(units []*fileUnit, deleted, added []int) map[int]bool {
	absorbed := map[int]bool{}
	if len(deleted) == 0 || len(added) == 0 {
		return absorbed
	}

	type candidate struct {
		del, add int
		score    float64
	}
	var candidates []candidate
	for _, d := range deleted {
		oldLines := countLines(units[d].oldSides)
		for _, a := range added {
			if units[a].ext != units[d].ext {
				continue
			}
			if s := similarity(oldLines, units[a].newSides); s >= minRenameSimilarity {
				candidates = append(candidates, candidate{del: d, add: a, score: s})
			}
		}
	}
	sort.SliceStable(candidates, func(i, j int) bool { return candidates[i].score > candidates[j].score })

	used := map[int]bool{}
	for _, c := range candidates {
		if used[c.del] || used[c.add] {
			continue
		}
		used[c.del], used[c.add] = true, true
		keep, drop := c.del, c.add
		if drop < keep {
			keep, drop = drop, keep
		}
		units[keep] = &fileUnit{
			oldPath:  units[c.del].oldPath,
			newPath:  units[c.add].newPath,
			ext:      units[c.del].ext,
			oldSides: units[c.del].oldSides,
			newSides: units[c.add].newSides,
		}
		absorbed[drop] = true
	}
	return absorbed
}

// countLines counts the significant lines of sides. Short lines such as
// "}" or "" say nothing about whether two files are the same file.
func countLines(sides []side) map[string]int {
	counts := map[string]int{}
	for _, s := range sides {
		for _, l := range s.lines {
			if l = strings.TrimSpace(l); len(l) >= 4 {
				counts[l]++
			}
		}
	}
	return counts
}

// similarity is the share of significant lines old and sides have in
// common, relative to the larger of the two.
func similarity(old map[string]int, sides []side) float64 {
	newLines := countLines(sides)
	oldTotal, newTotal, common := 0, 0, 0
	for _, n := range old {
		oldTotal += n
	}
	for l, n := range newLines {
		newTotal += n
		common += min(n, old[l])
	}
	total := max(oldTotal, newTotal)
	if total == 0 {
		return 0
	}
	return float64(common) / float64(total)
}
