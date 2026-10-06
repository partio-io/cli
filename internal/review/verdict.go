// Package review is the deterministic half of the proposal review. The
// review program writes one verdict file per issue; this package loads
// it, checks it and fails closed, and records it in the tracking issue.
package review

// Outcomes a verdict can carry.
const (
	OutcomeKeep    = "keep"
	OutcomeRewrite = "rewrite"
	OutcomeClose   = "close"
)

// Reasons a close can carry.
const (
	ReasonBuilt          = "built"
	ReasonFalsePremise   = "false-premise"
	ReasonDoesNotApply   = "does-not-apply"
	ReasonDuplicate      = "duplicate"
	ReasonCouldNotVerify = "could-not-verify"
)

// Premise and claim verdicts. NoClaims is valid for the premise only.
const (
	Holds      = "holds"
	Fails      = "fails"
	Unresolved = "unresolved"
	NoClaims   = "no-claims"
)

// Verdict is the contract between the review program and the gate: one
// file per reviewed issue.
type Verdict struct {
	Issue       int         `json:"issue"`
	Outcome     string      `json:"outcome"`
	CloseReason string      `json:"close_reason,omitempty"`
	DuplicateOf int         `json:"duplicate_of,omitempty"`
	Premise     Premise     `json:"premise"`
	Fit         Fit         `json:"fit"`
	Built       Built       `json:"built"`
	Duplicates  []Candidate `json:"duplicates"`
	Rewrite     *Rewrite    `json:"rewrite,omitempty"`
}

// Premise is the check of the issue's factual claims.
type Premise struct {
	Verdict string  `json:"verdict"`
	Claims  []Claim `json:"claims"`
}

// Claim is one factual claim, checked against the repository.
type Claim struct {
	Claim    string `json:"claim"`
	Evidence string `json:"evidence"`
	Verdict  string `json:"verdict"`
	Excerpt  string `json:"excerpt"`
}

// Fit says whether the issue applies to this project.
type Fit struct {
	Applies bool   `json:"applies"`
	Reason  string `json:"reason"`
}

// Built says whether the issue is already built.
type Built struct {
	Built    bool   `json:"built"`
	Evidence string `json:"evidence"`
}

// Candidate is one possible duplicate the review considered.
type Candidate struct {
	Issue int    `json:"issue"`
	Same  bool   `json:"same"`
	Why   string `json:"why"`
}

// Rewrite is the proposed new title and body of the issue.
type Rewrite struct {
	Title   string   `json:"title"`
	Body    string   `json:"body"`
	Changes []string `json:"changes"`
}
