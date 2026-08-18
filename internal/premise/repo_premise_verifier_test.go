package premise

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// verifierDoc is this repository's premise verifier, relative to the package
// directory that go test runs in.
const verifierDoc = "../../" + VerifierPath

// TestPremiseVerifierReachesTheProposer is the end-to-end path this slice
// builds: a verifier that describes verification once, and a proposer that
// reaches it before it files anything. A verifier nothing calls changes no
// behaviour, and a proposer that verifies after filing has already filed.
func TestPremiseVerifierReachesTheProposer(t *testing.T) {
	raw, err := os.ReadFile(verifierDoc)
	if err != nil {
		t.Fatalf("read the premise verifier: %v", err)
	}
	doc := string(raw)

	if !strings.Contains(doc, VerifierMarker) {
		t.Errorf("the premise verifier does not carry %q", VerifierMarker)
	}

	// The inputs and the outputs are the contract a caller keys off. The
	// verifier consumes the premise block, so it names that block's marker,
	// and it returns one of a fixed set of verdicts.
	if !strings.Contains(doc, Marker) {
		t.Errorf("the premise verifier does not name the premise block it consumes (%q)", Marker)
	}
	for _, v := range Verdicts {
		if !strings.Contains(doc, string(v)) {
			t.Errorf("the premise verifier does not define the %q verdict", v)
		}
	}

	src, err := os.ReadFile(proposeProgram)
	if err != nil {
		t.Fatalf("read propose program: %v", err)
	}

	context, ok := section(string(src), "## Context")
	if !ok {
		t.Fatal("propose program has no context section")
	}
	if !strings.Contains(context, VerifierPath) {
		t.Errorf("the propose program does not put %s in front of its agent", VerifierPath)
	}

	agents, ok := section(string(src), "## Agents")
	if !ok {
		t.Fatal("propose program has no agents section")
	}
	verify := strings.Index(agents, VerifierPath)
	file := strings.Index(agents, "gh issue create")
	switch {
	case verify < 0:
		t.Fatalf("the proposer is never told to apply %s", VerifierPath)
	case file < 0:
		t.Fatal("propose program gives no gh issue create instruction")
	case verify > file:
		t.Error("the proposer is told to verify the premise only after it has filed the issue")
	}
}

// TestEveryEvidenceKindIsGatheredBeforeTheVerdict checks that the proposer can
// actually gather whatever a claim names. The claim grammar admits a path, a
// symbol or a command; a procedure that only describes reading files leaves the
// other two ungathered, and a claim nobody gathered evidence for gets decided
// from the source material instead — which is the defect this slice exists to
// stop. The gathering is also ordered ahead of the verdict, so the description
// cannot be read as "decide, then look".
func TestEveryEvidenceKindIsGatheredBeforeTheVerdict(t *testing.T) {
	raw, err := os.ReadFile(verifierDoc)
	if err != nil {
		t.Fatalf("read the premise verifier: %v", err)
	}
	doc := string(raw)

	procedure, ok := section(doc, "## Procedure")
	if !ok {
		t.Fatal("the premise verifier describes no procedure")
	}
	for _, kind := range []string{"path", "symbol", "command"} {
		if !strings.Contains(procedure, kind) {
			t.Errorf("the procedure never says how to gather evidence named as a %s", kind)
		}
	}

	gather := strings.Index(doc, "## Procedure")
	decide := strings.Index(doc, "## Verdicts")
	if gather < 0 || decide < 0 {
		t.Fatal("the premise verifier does not separate the procedure from the verdicts")
	}
	if gather > decide {
		t.Error("the premise verifier puts the verdicts ahead of the gathering that settles them")
	}

	// The proposer applies the verifier against a tree, so it has to be told
	// there is a tree to read rather than left to answer from the source item.
	src, err := os.ReadFile(proposeProgram)
	if err != nil {
		t.Fatalf("read propose program: %v", err)
	}
	agents, ok := section(string(src), "## Agents")
	if !ok {
		t.Fatal("propose program has no agents section")
	}
	if !strings.Contains(agents, "checked out") {
		t.Error("the proposer is never told the tree it must verify against is checked out")
	}
}

// TestVerificationIsDescribedOnce checks that the description has one home.
// Slices 6, 7 and 8 add callers, and a caller that pastes the procedure into
// its own program is free to drift from it. The marker is the anchor: exactly
// one file carries it, so a second copy of the description is a test failure
// rather than something noticed months later.
func TestVerificationIsDescribedOnce(t *testing.T) {
	carriers := markerCarriers(t, VerifierMarker)

	if len(carriers) != 1 {
		t.Fatalf("verification is described in %d files, want exactly one: %v", len(carriers), carriers)
	}
	if want := filepath.ToSlash(filepath.Join(repoRoot, VerifierPath)); carriers[0] != want {
		t.Errorf("verification is described in %s, want %s", carriers[0], want)
	}
}

// noBlockHeading opens the part of the verifier that covers a proposal filed
// before the premise block existed.
const noBlockHeading = "## When there is no block"

// TestAProposalWithNoBlockIsCheckedFromItsProse covers the backlog this gate
// would otherwise exempt: 534 open proposals, 56 already approved, none
// carrying a block. If the verifier only knows how to read a block, every one
// of them builds unchecked, and the gate protects new work only.
func TestAProposalWithNoBlockIsCheckedFromItsProse(t *testing.T) {
	raw, err := os.ReadFile(verifierDoc)
	if err != nil {
		t.Fatalf("read the premise verifier: %v", err)
	}

	body, ok := section(string(raw), noBlockHeading)
	if !ok {
		t.Fatalf("the premise verifier has no %q section, so a proposal without a block is exempt", noBlockHeading)
	}
	for _, wanted := range []string{"extract", "prose"} {
		if !strings.Contains(flat(body), wanted) {
			t.Errorf("the %q section never says to %s the claims from the issue prose", noBlockHeading, wanted)
		}
	}
}

// TestExtractedClaimsAreVerifiedLikeBlockClaims checks that extraction feeds
// the procedure that already exists rather than growing one beside it. A second
// procedure is free to drift from the first, and then an old proposal is held
// to a different bar than a new one — which is the exemption this slice
// removes, reintroduced in a subtler form.
func TestExtractedClaimsAreVerifiedLikeBlockClaims(t *testing.T) {
	raw, err := os.ReadFile(verifierDoc)
	if err != nil {
		t.Fatalf("read the premise verifier: %v", err)
	}
	doc := string(raw)

	body, ok := section(doc, noBlockHeading)
	if !ok {
		t.Fatalf("the premise verifier has no %q section", noBlockHeading)
	}
	if !strings.Contains(flat(body), "same") {
		t.Errorf("the %q section never says extracted claims are verified the same way", noBlockHeading)
	}

	procedure, ok := section(doc, "## Procedure")
	if !ok {
		t.Fatal("the premise verifier describes no procedure")
	}
	if strings.Contains(procedure, "in the block") {
		t.Error("the procedure scopes itself to claims in the block, so an extracted claim falls outside it")
	}

	// One gathering, described once. A copy inside the extraction section
	// would be the second procedure this test exists to prevent.
	if n := strings.Count(doc, "Gather the evidence"); n != 1 {
		t.Errorf("the premise verifier describes the gathering %d times, want exactly once", n)
	}
}

// TestAFailedExtractedClaimStopsTheStageTheSameWay checks that a false claim
// costs an old proposal what it costs a new one. The extraction section says
// nothing below it changes; that sentence is only true while the caller's stop
// is in fact below it and is not written for a block alone. One stop, one
// label, one comment — an old proposal is not stopped by a quieter mechanism
// the operator has to learn separately.
func TestAFailedExtractedClaimStopsTheStageTheSameWay(t *testing.T) {
	doc := readRepoFile(t, verifierDoc)

	const callerHeading = "## What the caller does"
	extract, caller := strings.Index(doc, noBlockHeading), strings.Index(doc, callerHeading)
	switch {
	case extract < 0:
		t.Fatalf("the premise verifier has no %q section", noBlockHeading)
	case caller < 0:
		t.Fatalf("the premise verifier has no %q section", callerHeading)
	case extract > caller:
		t.Errorf("%q sits after %q, so what the caller does is not covered by it", noBlockHeading, callerHeading)
	}

	callerBody, ok := section(doc, callerHeading)
	if !ok {
		t.Fatalf("the premise verifier has no %q section", callerHeading)
	}
	if strings.Contains(callerBody, "A block that") {
		t.Error("the caller's stop is written for a block, so an extracted premise falls outside it")
	}

	gate := readRepoFile(t, gateDoc)
	stop, ok := section(gate, "## When the premise does not hold")
	if !ok {
		t.Fatal("the stage gate does not say what happens when the premise does not hold")
	}
	if strings.Contains(stop, "in the block") {
		t.Error("the blocked-stage comment reports only claims in the block, so an extracted claim goes unreported")
	}
	for _, want := range []string{BlockingLabel, GateCommentMarker} {
		if !strings.Contains(gate, want) {
			t.Errorf("the stage gate does not name %q, so the stop it describes is not the one the caller performs", want)
		}
	}
}

// TestAnOldProposalIsNotRewritten checks the promise that makes this gate cheap
// enough to apply at all: extraction happens at check time and leaves nothing
// behind. Backfilling a block would rewrite the operator's words in hundreds of
// issues, one stage run at a time, and the gate would start editing the backlog
// it was only meant to check. The gate's passing path is where the risk sits —
// it refreshes the block it just verified, and on an old proposal there is no
// block to refresh.
func TestAnOldProposalIsNotRewritten(t *testing.T) {
	body, ok := section(readRepoFile(t, verifierDoc), noBlockHeading)
	if !ok {
		t.Fatalf("the premise verifier has no %q section", noBlockHeading)
	}
	for _, want := range []string{"rewrite", "backfill"} {
		if !strings.Contains(flat(body), want) {
			t.Errorf("the %q section never rules out a %s of the issue body", noBlockHeading, want)
		}
	}

	holds, ok := section(readRepoFile(t, gateDoc), "## When the premise holds")
	if !ok {
		t.Fatal("the stage gate does not say what happens when the premise holds")
	}
	if !strings.Contains(flat(holds), "no block") {
		t.Error("the stage gate refreshes the premise block unconditionally, so a proposal that carries none gets one written into it")
	}
}

// TestTheBacklogIsNotSwept checks that applying the gate to old proposals stays
// a per-run cost and never becomes a pass over the 534 open ones. The operator
// chose that explicitly: not swept ahead of time, not pruned. A sweep would
// relabel or close issues nobody asked about, in bulk, from a stage that was
// only meant to check the one in front of it — so the checking is lazy, and the
// only issue a run may edit is the one it was handed.
func TestTheBacklogIsNotSwept(t *testing.T) {
	body, ok := section(readRepoFile(t, verifierDoc), noBlockHeading)
	if !ok {
		t.Fatalf("the premise verifier has no %q section", noBlockHeading)
	}
	lower := flat(body)
	for _, want := range []string{"check time", "touches"} {
		if !strings.Contains(lower, want) {
			t.Errorf("the %q section never says %q, so it does not say when an old proposal is checked", noBlockHeading, want)
		}
	}

	// Neither description may reach for the backlog. A stage is handed one
	// issue; listing the others is the sweep this slice must not introduce.
	for _, doc := range []string{verifierDoc, gateDoc} {
		src := readRepoFile(t, doc)
		for _, banned := range []string{"gh issue list", "gh issue close"} {
			if strings.Contains(src, banned) {
				t.Errorf("%s runs %q, which reaches past the issue the stage was handed", doc, banned)
			}
		}
	}

	// Labelling is per-issue for the same reason: the issue number comes from
	// the run, never from a list the stage assembled itself.
	for _, line := range strings.Split(readRepoFile(t, gateDoc), "\n") {
		if strings.Contains(line, "gh issue edit") && !strings.Contains(line, "$MINION_ISSUE_NUMBER") {
			t.Errorf("the stage gate edits an issue it was not handed: %s", strings.TrimSpace(line))
		}
	}
}

// flat lowercases src and collapses each run of whitespace to one space. A
// phrase these tests look for then reads the same whether or not the paragraph
// carrying it was rewrapped: prose gets rewrapped, and the rule it states does
// not change when it does.
func flat(src string) string {
	return strings.ToLower(strings.Join(strings.Fields(src), " "))
}

// section returns the body of the section src opens with heading, up to the
// next heading of the same level or the end of src. It ignores fenced code
// blocks, because the propose program shows a `## Premise` template inside one
// and that template does not end the section it sits in.
func section(src, heading string) (string, bool) {
	lines := strings.Split(src, "\n")
	closes := strings.Repeat("#", strings.Count(heading, "#")) + " "

	start, fenced := -1, false
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "```") {
			fenced = !fenced
			continue
		}
		if fenced {
			continue
		}

		switch {
		case start < 0 && trimmed == heading:
			start = i + 1
		case start >= 0 && strings.HasPrefix(trimmed, closes):
			return strings.Join(lines[start:i], "\n"), true
		}
	}
	if start < 0 {
		return "", false
	}
	return strings.Join(lines[start:], "\n"), true
}
