package premise

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestClaimSourcePrefersTheBlock checks which claims a stage verifies when a
// proposal offers both. A proposal that states its premise has already done the
// work of naming what it depends on and what settles it; re-deriving that from
// the surrounding paragraphs would let a stage check something the proposal
// never claimed, and would make the block optional in practice.
func TestClaimSourcePrefersTheBlock(t *testing.T) {
	prose, err := os.ReadFile(filepath.Join("testdata", "issue-30-body.md"))
	if err != nil {
		t.Fatalf("read the issue #30 body fixture: %v", err)
	}

	block, err := Render(Block{Claims: []Claim{{
		Statement: "The post-commit hook names changed files with a git diff.",
		Evidence:  "internal/hooks/postcommit.go",
	}}})
	if err != nil {
		t.Fatalf("render a premise block: %v", err)
	}

	body := string(prose) + "\n" + block
	if got := ClaimSource(body); got != SourceBlock {
		t.Fatalf("a proposal that carries a block takes its claims from the %s, want the %s", got, SourceBlock)
	}

	parsed, err := Parse(body)
	if err != nil {
		t.Fatalf("the block does not parse out of the body: %v", err)
	}
	if len(parsed.Claims) != 1 {
		t.Fatalf("the block states 1 claim, %d were taken", len(parsed.Claims))
	}
	if strings.Contains(parsed.Claims[0].Statement, "full tree walk") {
		t.Error("a claim from the prose reached the verifier even though the proposal carries a block")
	}

	// The agent applies the same preference, so extraction has to be gated on
	// the marker being absent rather than offered as an alternative reading.
	noBlock, ok := section(readRepoFile(t, verifierDoc), noBlockHeading)
	if !ok {
		t.Fatalf("the premise verifier has no %q section", noBlockHeading)
	}
	if !strings.Contains(noBlock, Marker) {
		t.Errorf("the %q section does not name %s as what makes it apply, so it reads as a second way to check any proposal", noBlockHeading, Marker)
	}
}

// TestProseWithNoCheckableClaimProceeds guards the direction that would do the
// most damage quietly. Most of the backlog describes what to build rather than
// what is true, so reading "I found nothing to check" as a failure would block
// hundreds of proposals at once — the opposite of the intent, and indis-
// tinguishable from the gate working. An empty block stays an error, because a
// proposal that opens a premise section and states nothing in it has made a
// mistake; a proposal that never claimed anything has not.
func TestProseWithNoCheckableClaimProceeds(t *testing.T) {
	const wishes = "## Description\n\nAdd a --json flag to partio status.\n\n" +
		"## Acceptance Criteria\n\n- [ ] The flag prints machine-readable output\n"

	if got := ClaimSource(wishes); got != SourceProse {
		t.Fatalf("a proposal with no block takes its claims from the %s, want the %s", got, SourceProse)
	}
	if _, err := Parse("## Premise\n\n" + Marker + "\n"); err == nil {
		t.Error("a premise section that states nothing parses, so an empty block is indistinguishable from prose with no claim")
	}

	body, ok := section(readRepoFile(t, verifierDoc), noBlockHeading)
	if !ok {
		t.Fatalf("the premise verifier has no %q section", noBlockHeading)
	}
	lower := flat(body)
	for _, want := range []string{"no checkable claim", "continue", "unresolved", "say so"} {
		if !strings.Contains(lower, want) {
			t.Errorf("the %q section never says %q, so a proposal that claims nothing has no stated outcome", noBlockHeading, want)
		}
	}
}

// TestClaimSourceReadsTheMarkerNotTheProse covers what counts as a block. The
// marker is the whole test: a heading a proposal happened to write, or a
// paragraph that talks about premises, is prose, and prose is extracted from
// rather than parsed.
func TestClaimSourceReadsTheMarkerNotTheProse(t *testing.T) {
	tests := []struct {
		name string
		body string
		want Source
	}{
		{
			name: "a section carrying the marker is a block",
			body: "## Premise\n\n" + Marker + "\n\n- Attribution is binary. [evidence: `internal/attribution/calculate.go`]\n",
			want: SourceBlock,
		},
		{
			name: "a premise heading without the marker is not a block",
			body: "## Premise\n\n- Attribution is binary. [evidence: `internal/attribution/calculate.go`]\n",
			want: SourceProse,
		},
		{
			name: "prose that mentions a premise block is not one",
			body: "This proposal has no premise block yet.\n\nThe hook walks the tree.\n",
			want: SourceProse,
		},
		{
			name: "an empty body has no block",
			body: "",
			want: SourceProse,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ClaimSource(tt.body); got != tt.want {
				t.Errorf("ClaimSource() = %s, want %s", got, tt.want)
			}
		})
	}
}
