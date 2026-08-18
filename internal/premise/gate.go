package premise

// A stage gate is what a stage does with a verdict. Verification decides
// whether a claim holds; the gate decides what happens next, and it is one
// described behaviour rather than a rule each stage invents. A stage that
// gates its own way stops differently from the stage beside it, and an
// operator reading a blocked issue cannot tell which of them stopped it.
const (
	// GatePath is where the gate is described, relative to the repository
	// root.
	GatePath = ".minions/stage-gate.md"

	// GateMarker names the description and its format version. Exactly one
	// file carries it, which is what makes "described once" checkable.
	GateMarker = "<!-- partio:stage-gate:v1 -->"
)

const (
	// BlockingLabel is the label a stage adds when a premise fails. The
	// approval program already treats it as a skip condition, so blocking
	// needs no new label and no new pipeline wiring. Removing it is how the
	// operator overrules a verdict.
	BlockingLabel = "do-not-build"

	// GateCommentMarker opens the comment a blocked stage posts. The comment
	// is the operator's whole account of why the stage stopped, so it is
	// findable by marker rather than by matching the prose around it.
	GateCommentMarker = "<!-- partio:premise-gate:v1 -->"
)
