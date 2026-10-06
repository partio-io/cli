package review

import (
	"errors"
	"fmt"
	"slices"
	"strings"
)

// check rejects a verdict that is incomplete or that contradicts
// itself. It says nothing about whether the verdict is right.
func (v Verdict) check(issue int) error {
	if v.Issue != issue {
		return fmt.Errorf("verdict is for issue %d, not %d", v.Issue, issue)
	}
	if err := v.Premise.check(); err != nil {
		return err
	}
	switch v.Outcome {
	case OutcomeClose:
		return v.checkClose()
	case OutcomeKeep, OutcomeRewrite:
		if v.CloseReason != "" {
			return fmt.Errorf("%s carries close_reason %q", v.Outcome, v.CloseReason)
		}
		if err := v.checkNoContradiction(); err != nil {
			return err
		}
		if v.Outcome == OutcomeRewrite {
			return v.checkRewrite()
		}
		return nil
	default:
		return fmt.Errorf("unknown outcome %q", v.Outcome)
	}
}

func (p Premise) check() error {
	if !slices.Contains([]string{Holds, Fails, Unresolved, NoClaims}, p.Verdict) {
		return fmt.Errorf("unknown premise verdict %q", p.Verdict)
	}
	for i, c := range p.Claims {
		if !slices.Contains([]string{Holds, Fails, Unresolved}, c.Verdict) {
			return fmt.Errorf("claim %d: unknown verdict %q", i+1, c.Verdict)
		}
		if c.Evidence == "" {
			return fmt.Errorf("claim %d: no evidence", i+1)
		}
		if c.Excerpt == "" {
			return fmt.Errorf("claim %d: no excerpt", i+1)
		}
		if c.Correction != "" && c.Verdict != Fails {
			return fmt.Errorf("claim %d: a correction on a claim whose verdict is %s", i+1, c.Verdict)
		}
	}
	return nil
}

// hasClaim reports whether any claim carries verdict.
func (p Premise) hasClaim(verdict string) bool {
	return slices.ContainsFunc(p.Claims, func(c Claim) bool { return c.Verdict == verdict })
}

// checkClose accepts a close only with the evidence its reason needs.
func (v Verdict) checkClose() error {
	switch v.CloseReason {
	case ReasonFalsePremise:
		if !v.Premise.hasClaim(Fails) {
			return errors.New("false-premise close without a failing claim")
		}
	case ReasonCouldNotVerify:
		if !v.Premise.hasClaim(Unresolved) {
			return errors.New("could-not-verify close without an unresolved claim")
		}
	case ReasonBuilt:
		if !v.Built.Built || v.Built.Evidence == "" {
			return errors.New("built close without built evidence")
		}
	case ReasonDoesNotApply:
		if v.Fit.Applies || v.Fit.Reason == "" {
			return errors.New("does-not-apply close without applies: false and a reason")
		}
	case ReasonDuplicate:
		if v.DuplicateOf == 0 {
			return errors.New("duplicate close without duplicate_of")
		}
		if v.DuplicateOf == v.Issue {
			return fmt.Errorf("duplicate close of #%d as a duplicate of itself", v.Issue)
		}
		if !slices.ContainsFunc(v.Duplicates, func(c Candidate) bool { return c.Issue == v.DuplicateOf && c.Same }) {
			return fmt.Errorf("duplicate close without a candidate #%d marked same", v.DuplicateOf)
		}
	default:
		return fmt.Errorf("unknown close reason %q", v.CloseReason)
	}
	return nil
}

// checkNoContradiction rejects a keep or a rewrite whose own parts say
// the issue should close. A rewrite may carry a failing claim when it
// corrects that claim: the idea survives the false fact, and the new
// body rests on the correction. A keep never carries one.
func (v Verdict) checkNoContradiction() error {
	failing := v.Premise.Verdict == Fails || v.Premise.hasClaim(Fails)
	switch {
	case failing && v.Outcome == OutcomeKeep:
		return fmt.Errorf("%s with a failing premise", v.Outcome)
	case failing && !v.Premise.hasClaim(Fails):
		return fmt.Errorf("%s with a failing premise and no failing claim to correct", v.Outcome)
	case !v.Fit.Applies:
		return fmt.Errorf("%s with applies: false", v.Outcome)
	case v.Built.Built:
		return fmt.Errorf("%s with built: true", v.Outcome)
	}
	for i, c := range v.Premise.Claims {
		if c.Verdict == Fails && strings.TrimSpace(c.Correction) == "" {
			return fmt.Errorf("%s with failing claim %d and no correction", v.Outcome, i+1)
		}
	}
	return nil
}

func (v Verdict) checkRewrite() error {
	if v.Rewrite == nil || strings.TrimSpace(v.Rewrite.Title) == "" {
		return errors.New("rewrite without a title")
	}
	if v.Rewrite.Body == "" {
		return errors.New("rewrite without a body")
	}
	return nil
}
