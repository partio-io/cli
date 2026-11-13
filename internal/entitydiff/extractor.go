package entitydiff

import "strings"

// extractor finds the declarations in a run of source lines: one rebuilt
// side of a hunk. A whole-file parser could replace a scanner behind this
// interface.
type extractor interface {
	extract(lines []string) extraction
}

// foundEntity is an entity and the index of its first line.
type foundEntity struct {
	Entity
	line int
}

// extraction is what an extractor found. partial lists the declarations
// whose header ran past the end of the lines; they are not reported.
type extraction struct {
	entities []foundEntity
	partial  []entityKey
}

// extractorFor returns the extractor for path's language, or nil when the
// language is not supported.
func extractorFor(path string) extractor {
	if strings.HasSuffix(path, ".go") {
		return goExtractor{}
	}
	return nil
}
