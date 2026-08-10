package repairround

// DefaultMaxRounds is the repair budget one check gets on one pull
// request. Each check counts separately, so a dead-code repair never
// spends the budget a failing test would need.
const DefaultMaxRounds = 3

// Decision answers "may another repair round start for this check, and
// which round is it?".
type Decision struct {
	// Allowed reports whether another round may start.
	Allowed bool
	// Round is the round this run would be; 1 is the first attempt.
	// It is reported even when Allowed is false, where it names the
	// round that was declined.
	Round int
	// Prior is the number of rounds for this check already on the
	// branch.
	Prior int
}

// Decide counts the repair rounds for check already present in
// subjects — the commit subjects on the pull request's branch — and
// reports whether another may start under maxRounds.
//
// Only commits carrying this check's marker count. Rounds spent by
// another check, and commits a human wrote, are ignored.
func Decide(subjects []string, check string, maxRounds int) Decision {
	prior := 0
	for _, s := range subjects {
		if c, ok := CheckOf(s); ok && c == check {
			prior++
		}
	}
	return Decision{
		Allowed: prior < maxRounds,
		Round:   prior + 1,
		Prior:   prior,
	}
}
