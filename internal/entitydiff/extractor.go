package entitydiff

import (
	"path"
	"strings"
)

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

// extractors maps a file extension to the extractor for its language.
// TypeScript declaration files (.d.ts, .d.mts, .d.cts) end in a supported
// extension and are analysed like any other TypeScript file.
var extractors = map[string]extractor{
	".go":  goExtractor{},
	".ts":  jsExtractor{},
	".tsx": jsExtractor{},
	".mts": jsExtractor{},
	".cts": jsExtractor{},
	".js":  jsExtractor{},
	".jsx": jsExtractor{},
	".mjs": jsExtractor{},
	".cjs": jsExtractor{},
}

// extractorFor returns the extractor for p's language, or nil when the
// language is not supported. Minified JavaScript is generated code and is
// not analysed.
func extractorFor(p string) extractor {
	if strings.HasSuffix(p, ".min.js") {
		return nil
	}
	return extractors[path.Ext(p)]
}
