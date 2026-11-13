// Package entitydiff summarises a checkpoint's stored unified diff as
// structural changes to declarations: which functions, methods and types
// were added, removed, or had their signatures changed.
//
// The analysis reads only the patch text. It never consults git objects,
// so it works when the commit a checkpoint names is gone from the clone.
// A line-oriented scanner per language finds declarations in each hunk's
// rebuilt old and new sides; a declaration it cannot see in full is left
// out rather than reported wrongly.
package entitydiff
