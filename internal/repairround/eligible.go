package repairround

import (
	"fmt"
	"slices"
)

// RepairLabel marks a pull request the pipeline opened and may
// therefore repair. It is the same label the audit workflows guard on.
const RepairLabel = "minion"

// Eligibility answers the question that comes before the round count:
// may this pull request receive a repair commit at all?
//
// It is separate from Decision because the two refusals mean different
// things. A spent budget is a pull request the pipeline tried and gave
// up on; an ineligible one was never the pipeline's to touch.
type Eligibility struct {
	// Allowed reports whether a repair round may push to this pull
	// request.
	Allowed bool
	// Reason states the decision in one line. It is written to the
	// workflow log, which is the only place a skipped repair is
	// visible, so it is never empty.
	Reason string
}

// Eligible reports whether a repair round may push to a pull request
// carrying labels, whose head branch lives in headRepo, opened against
// baseRepo. Both repositories are "owner/name".
//
// Two conditions, both required:
//
//   - The pull request carries the repair label. Work a person wrote
//     gets the check's verdict and nothing else — a bot that commits
//     on top of a hand-written branch is a bot the operator has to
//     undo.
//   - The head branch lives in the base repository. A fork head cannot
//     be pushed to, so it is refused here, before anything is fetched.
func Eligible(labels []string, headRepo, baseRepo string) Eligibility {
	if !hasLabel(labels, RepairLabel) {
		return Eligibility{Allowed: false, Reason: fmt.Sprintf(
			"the pull request carries no %q label, so it is reported and not repaired", RepairLabel)}
	}
	if headRepo == "" {
		return Eligibility{Allowed: false, Reason: fmt.Sprintf(
			"the head repository is unknown, so a repair would push to a branch it cannot name; %q label ignored",
			RepairLabel)}
	}
	if headRepo != baseRepo {
		return Eligibility{Allowed: false, Reason: fmt.Sprintf(
			"the head branch lives in %s and not in %s, and this workflow cannot push to a fork",
			headRepo, baseRepo)}
	}
	return Eligibility{Allowed: true, Reason: fmt.Sprintf(
		"the pull request carries the %q label and its head branch lives in %s", RepairLabel, baseRepo)}
}

// hasLabel reports whether want is among labels. The match is exact:
// it replaces a workflow expression that compared the same way, and a
// label that merely looks like the marker is not the marker.
func hasLabel(labels []string, want string) bool {
	return slices.Contains(labels, want)
}
