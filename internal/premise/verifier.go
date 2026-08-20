package premise

// Verification is a single described behaviour, not a copy per caller: given a
// premise block and a checked-out tree, gather the evidence each claim names
// and return a verdict plus what was found. The description lives in one file,
// and the propose, research and implement programs reference it. The constants
// below are the contract that file and its callers share.
const (
	// VerifierPath is where verification is described, relative to the
	// repository root.
	VerifierPath = ".minions/premise-verifier.md"

	// VerifierMarker names the description and its format version. Exactly
	// one file carries it, which is what makes "described once" checkable.
	VerifierMarker = "<!-- partio:premise-verifier:v1 -->"

	// NoBlockSection is where the description covers a proposal filed before
	// the block format existed. Its claims come from the issue prose. Every
	// proposal open when the gate shipped was in that state, so a stage that
	// treats a blockless issue as out of scope exempts the entire backlog.
	NoBlockSection = "## When there is no block"
)

// Verdict is the outcome of checking one claim against the tree.
type Verdict string

const (
	// Holds reports a claim the gathered evidence confirms.
	Holds Verdict = "holds"

	// Fails reports a claim the gathered evidence contradicts. One failed
	// claim fails the block, and the calling stage stops.
	Fails Verdict = "fails"

	// Unresolved reports a claim the named evidence does not settle, because
	// the path is missing or the command produced nothing to read. It is not
	// a pass: a claim nothing settles has not been checked.
	Unresolved Verdict = "unresolved"
)

// Verdicts is the closed set a caller may act on.
var Verdicts = []Verdict{Holds, Fails, Unresolved}
