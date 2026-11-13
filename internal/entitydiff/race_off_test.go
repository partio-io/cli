//go:build !race

package entitydiff

// raceEnabled reports whether the race detector is on. Timing guards are
// skipped under it, since it slows code by an order of magnitude.
const raceEnabled = false
