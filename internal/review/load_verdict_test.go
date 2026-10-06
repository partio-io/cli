package review

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
)

// validKeep is a keep verdict for issue 12 that passes every check.
// Each case below changes one part of it.
func validKeep() Verdict {
	return Verdict{
		Issue:   12,
		Outcome: OutcomeKeep,
		Premise: Premise{Verdict: Holds, Claims: []Claim{{
			Claim: "the pre-push hook has no retry", Evidence: "internal/hooks/prepush.go",
			Verdict: Holds, Excerpt: "return push(remote)",
		}}},
		Fit:        Fit{Applies: true, Reason: "pre-push is in scope"},
		Duplicates: []Candidate{{Issue: 7, Same: false, Why: "different hook"}},
	}
}

func asClose(reason string, edit func(*Verdict)) func(*Verdict) {
	return func(v *Verdict) {
		v.Outcome, v.CloseReason = OutcomeClose, reason
		edit(v)
	}
}

func asRewrite(v *Verdict) {
	v.Outcome = OutcomeRewrite
	v.Rewrite = &Rewrite{Title: "Retry the pre-push", Body: "## What\nRetry once.", Changes: []string{"narrowed scope"}}
}

func writeJSON(t *testing.T, v Verdict) string {
	t.Helper()
	data, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal verdict: %v", err)
	}
	return writeVerdict(t, string(data))
}

// Every case the gate must read as "no verdict", one per row. The cause
// must name what is wrong, so the tracking row says why.
func TestLoadVerdictNoVerdict(t *testing.T) {
	tests := []struct {
		name      string
		edit      func(*Verdict)
		raw       string // used instead of edit when set
		missing   bool
		wantCause string
	}{
		{name: "missing file", missing: true, wantCause: "read verdict"},
		{name: "malformed JSON", raw: `{"issue": 12, "outcome": `, wantCause: "malformed verdict"},
		{name: "wrong JSON type", raw: `{"issue": "twelve"}`, wantCause: "malformed verdict"},
		{name: "other issue", edit: func(v *Verdict) { v.Issue = 13 }, wantCause: "issue 13, not 12"},
		{name: "unknown outcome", edit: func(v *Verdict) { v.Outcome = "park" }, wantCause: `unknown outcome "park"`},
		{name: "empty outcome", edit: func(v *Verdict) { v.Outcome = "" }, wantCause: `unknown outcome ""`},
		{name: "unknown close reason", edit: asClose("stale", func(*Verdict) {}), wantCause: `unknown close reason "stale"`},
		{name: "close without reason", edit: asClose("", func(*Verdict) {}), wantCause: `unknown close reason ""`},
		{name: "unknown premise verdict", edit: func(v *Verdict) { v.Premise.Verdict = "maybe" }, wantCause: `unknown premise verdict "maybe"`},
		{name: "unknown claim verdict", edit: func(v *Verdict) { v.Premise.Claims[0].Verdict = "no-claims" }, wantCause: `claim 1: unknown verdict "no-claims"`},
		{name: "claim without evidence", edit: func(v *Verdict) { v.Premise.Claims[0].Evidence = "" }, wantCause: "claim 1: no evidence"},
		{name: "claim without excerpt", edit: func(v *Verdict) { v.Premise.Claims[0].Excerpt = "" }, wantCause: "claim 1: no excerpt"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var path string
			switch {
			case tt.missing:
				path = filepath.Join(t.TempDir(), "absent.json")
			case tt.raw != "":
				path = writeVerdict(t, tt.raw)
			default:
				v := validKeep()
				tt.edit(&v)
				path = writeJSON(t, v)
			}
			_, err := LoadVerdict(path, 12)
			if err == nil {
				t.Fatal("LoadVerdict accepted the verdict")
			}
			if !strings.Contains(err.Error(), tt.wantCause) {
				t.Errorf("cause = %q, want it to contain %q", err, tt.wantCause)
			}
		})
	}
}

func failingClaim(v *Verdict) {
	v.Premise.Verdict = Fails
	v.Premise.Claims[0].Verdict = Fails
}

// A close is accepted only with the evidence its reason needs, and a
// keep or a rewrite is rejected when its own parts say "close".
func TestLoadVerdictCloseEvidenceAndContradictions(t *testing.T) {
	tests := []struct {
		name      string
		edit      func(*Verdict)
		wantCause string // empty means the verdict is accepted
	}{
		{name: "keep", edit: func(*Verdict) {}},
		{name: "keep with no claims", edit: func(v *Verdict) { v.Premise = Premise{Verdict: NoClaims} }},
		{name: "rewrite", edit: asRewrite},

		{name: "false premise with a failing claim", edit: asClose(ReasonFalsePremise, failingClaim)},
		{name: "false premise without a failing claim", edit: asClose(ReasonFalsePremise, func(*Verdict) {}),
			wantCause: "false-premise close without a failing claim"},
		{name: "could not verify with an unresolved claim", edit: asClose(ReasonCouldNotVerify, func(v *Verdict) {
			v.Premise.Verdict, v.Premise.Claims[0].Verdict = Unresolved, Unresolved
		})},
		{name: "could not verify without an unresolved claim", edit: asClose(ReasonCouldNotVerify, func(*Verdict) {}),
			wantCause: "could-not-verify close without an unresolved claim"},
		{name: "built with evidence", edit: asClose(ReasonBuilt, func(v *Verdict) {
			v.Built = Built{Built: true, Evidence: "internal/hooks/retry.go"}
		})},
		{name: "built without built: true", edit: asClose(ReasonBuilt, func(v *Verdict) {
			v.Built = Built{Evidence: "internal/hooks/retry.go"}
		}), wantCause: "built close without built evidence"},
		{name: "built without evidence", edit: asClose(ReasonBuilt, func(v *Verdict) { v.Built = Built{Built: true} }),
			wantCause: "built close without built evidence"},
		{name: "does not apply with a reason", edit: asClose(ReasonDoesNotApply, func(v *Verdict) {
			v.Fit = Fit{Applies: false, Reason: "partio has no web UI"}
		})},
		{name: "does not apply while it applies", edit: asClose(ReasonDoesNotApply, func(*Verdict) {}),
			wantCause: "does-not-apply close without applies: false and a reason"},
		{name: "does not apply without a reason", edit: asClose(ReasonDoesNotApply, func(v *Verdict) { v.Fit = Fit{} }),
			wantCause: "does-not-apply close without applies: false and a reason"},
		{name: "duplicate with a matching candidate", edit: asClose(ReasonDuplicate, func(v *Verdict) {
			v.DuplicateOf = 7
			v.Duplicates[0].Same = true
		})},
		{name: "duplicate without duplicate_of", edit: asClose(ReasonDuplicate, func(v *Verdict) { v.Duplicates[0].Same = true }),
			wantCause: "duplicate close without duplicate_of"},
		{name: "duplicate whose candidate is not the same", edit: asClose(ReasonDuplicate, func(v *Verdict) { v.DuplicateOf = 7 }),
			wantCause: "without a candidate #7 marked same"},
		{name: "duplicate of an issue it did not consider", edit: asClose(ReasonDuplicate, func(v *Verdict) {
			v.DuplicateOf = 8
			v.Duplicates[0].Same = true
		}), wantCause: "without a candidate #8 marked same"},
		{name: "duplicate of itself", edit: asClose(ReasonDuplicate, func(v *Verdict) {
			v.DuplicateOf = 12
			v.Duplicates = append(v.Duplicates, Candidate{Issue: 12, Same: true, Why: "same issue"})
		}), wantCause: "duplicate of itself"},

		{name: "keep with a failing premise", edit: failingClaim, wantCause: "keep with a failing premise"},
		{name: "keep with a failing claim only", edit: func(v *Verdict) { v.Premise.Claims[0].Verdict = Fails },
			wantCause: "keep with a failing premise"},
		{name: "keep that does not apply", edit: func(v *Verdict) { v.Fit.Applies = false }, wantCause: "keep with applies: false"},
		{name: "keep that is built", edit: func(v *Verdict) { v.Built.Built = true }, wantCause: "keep with built: true"},
		{name: "keep with a close reason", edit: func(v *Verdict) { v.CloseReason = ReasonBuilt }, wantCause: `keep carries close_reason "built"`},
		{name: "rewrite with a failing premise", edit: func(v *Verdict) { asRewrite(v); failingClaim(v) },
			wantCause: "rewrite with a failing premise"},
		{name: "rewrite that does not apply", edit: func(v *Verdict) { asRewrite(v); v.Fit.Applies = false },
			wantCause: "rewrite with applies: false"},
		{name: "rewrite that is built", edit: func(v *Verdict) { asRewrite(v); v.Built.Built = true },
			wantCause: "rewrite with built: true"},
		{name: "rewrite without a rewrite", edit: func(v *Verdict) { v.Outcome = OutcomeRewrite },
			wantCause: "rewrite without a title"},
		{name: "rewrite without a title", edit: func(v *Verdict) { asRewrite(v); v.Rewrite.Title = "" },
			wantCause: "rewrite without a title"},
		{name: "rewrite without a body", edit: func(v *Verdict) { asRewrite(v); v.Rewrite.Body = "" },
			wantCause: "rewrite without a body"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := validKeep()
			tt.edit(&v)
			got, err := LoadVerdict(writeJSON(t, v), 12)
			if tt.wantCause == "" {
				if err != nil {
					t.Fatalf("LoadVerdict rejected a valid verdict: %v", err)
				}
				if got.Outcome != v.Outcome || got.CloseReason != v.CloseReason {
					t.Errorf("loaded %s/%s, want %s/%s", got.Outcome, got.CloseReason, v.Outcome, v.CloseReason)
				}
				return
			}
			if err == nil {
				t.Fatalf("LoadVerdict accepted the verdict, want %q", tt.wantCause)
			}
			if !strings.Contains(err.Error(), tt.wantCause) {
				t.Errorf("cause = %q, want it to contain %q", err, tt.wantCause)
			}
		})
	}
}
