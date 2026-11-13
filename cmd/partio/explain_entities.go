package main

import (
	"fmt"

	"github.com/partio-io/cli/internal/entitydiff"
)

// maxSignatureLen is how many characters of a signature explain prints
// before cutting it off, so that each change stays on one terminal line.
const maxSignatureLen = 100

// printChangedEntities prints the "Changed entities:" section for a
// checkpoint diff. It prints nothing when no supported file has an
// entity-level change.
func printChangedEntities(diff string) {
	files := entitydiff.Analyze(diff)
	if len(files) == 0 {
		return
	}
	fmt.Printf("\nChanged entities:\n")
	for _, f := range files {
		fmt.Printf("  %s\n", f.DisplayPath())
		for _, c := range f.Changes {
			sig := c.New.Signature
			if c.Kind == entitydiff.Removed {
				sig = c.Old.Signature
			}
			fmt.Printf("    %-17s  %s\n", c.Kind, truncateSignature(sig))
		}
	}
}

// truncateSignature cuts sig to maxSignatureLen characters, marking the cut
// with an ellipsis.
func truncateSignature(sig string) string {
	runes := []rune(sig)
	if len(runes) <= maxSignatureLen {
		return sig
	}
	return string(runes[:maxSignatureLen-1]) + "…"
}
