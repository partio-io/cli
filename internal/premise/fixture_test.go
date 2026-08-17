package premise

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// repoRoot is this repository's root, relative to the package directory that go
// test runs in.
const repoRoot = "../.."

// claimEvidence records what the tree said about one claim, and the verdict
// that evidence produced. Present and absent are the excerpts the verdict
// rests on, re-read from the tree on every run: if the code they describe
// changes, the recorded verdict stops being true and this test says so, rather
// than the record quietly rotting into a comment nobody rechecks.
//
// Prose and evidence are set only for a fixture that carries no premise block.
// They are the claim a stage extracts from the issue prose, and the path that
// settles it. Prose is quoted from the body word for word, and the test fails
// unless it is still there, so an extracted claim is grounded in what the
// proposal says rather than in what this record wishes it said.
type claimEvidence struct {
	prose    string
	evidence string
	present  []string
	absent   []string
	verdict  Verdict
	why      string
}

// TestPremiseFixturesAreVerifiedAgainstThisRepository runs the verifier's
// procedure over both fixtures: gather the evidence each claim names, then
// decide from what was gathered. It checks both directions on purpose. A gate
// that drops everything and a gate that works are indistinguishable until
// something is shown to pass it.
func TestPremiseFixturesAreVerifiedAgainstThisRepository(t *testing.T) {
	tests := []struct {
		name    string
		fixture string
		source  Source
		block   Verdict
		claims  []claimEvidence
	}{
		{
			name:    "the issue #30 premise fails against this repository",
			fixture: "issue-30.md",
			source:  SourceBlock,
			block:   Fails,
			claims: []claimEvidence{{
				// The hook names the changed files with a git diff, which
				// compares two trees and reads no directory.
				present: []string{"git.DiffNameOnly(commitHash)"},
				absent:  []string{"filepath.Walk", "filepath.WalkDir", "os.ReadDir", "ls-files"},
				verdict: Fails,
				why:     "the post-commit hook walks no tree; it asks git to diff two commits",
			}, {
				// The loop runs over the numstat output, which carries one
				// line per changed file, not one per tracked file.
				present: []string{"git.DiffNumstat(commitHash)", "strings.Split(numstat"},
				absent:  []string{"filepath.Walk", "ls-files"},
				verdict: Fails,
				why:     "the cost follows the number of changed files, not the number tracked",
			}},
		},
		{
			// The proposal that started this work, as it was actually filed:
			// prose, no premise block, and a claim already proven false. An
			// older proposal is not exempt, so the same evidence settles it
			// the same way whether it arrived in a block or in a paragraph.
			name:    "the issue #30 body fails on claims extracted from its prose",
			fixture: "issue-30-body.md",
			source:  SourceProse,
			block:   Fails,
			claims: []claimEvidence{{
				prose:    "performs a full tree walk to determine which files were changed in a commit",
				evidence: "internal/hooks/postcommit.go",
				present:  []string{"git.DiffNameOnly(commitHash)"},
				absent:   []string{"filepath.Walk", "filepath.WalkDir", "os.ReadDir", "ls-files"},
				verdict:  Fails,
				why:      "the post-commit hook walks no tree; it asks git to diff two commits",
			}, {
				prose:    "This is O(N) in the number of tracked files",
				evidence: "internal/attribution/calculate.go",
				present:  []string{"git.DiffNumstat(commitHash)", "strings.Split(numstat"},
				absent:   []string{"filepath.Walk", "ls-files"},
				verdict:  Fails,
				why:      "the cost follows the number of changed files, not the number tracked",
			}},
		},
		{
			name:    "a claim that is true of this repository holds",
			fixture: "holds.md",
			source:  SourceBlock,
			block:   Holds,
			claims: []claimEvidence{{
				present: []string{"if agentActive", "result.AgentPercent = 100"},
				verdict: Holds,
				why:     "an agent-active commit is attributed wholly to the agent",
			}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			src, err := os.ReadFile(filepath.Join("testdata", tt.fixture))
			if err != nil {
				t.Fatalf("read the %s fixture: %v", tt.fixture, err)
			}
			body := string(src)
			if got := ClaimSource(body); got != tt.source {
				t.Fatalf("the %s fixture takes its claims from the %s, want the %s", tt.fixture, got, tt.source)
			}

			var claims []Claim
			switch tt.source {
			case SourceBlock:
				block, err := Parse(body)
				if err != nil {
					t.Fatalf("the %s fixture does not parse: %v", tt.fixture, err)
				}
				claims = block.Claims
			case SourceProse:
				// Extraction is the agent's work at check time, so what this
				// test pins is that the claims it verifies are the ones the
				// prose actually makes.
				for _, record := range tt.claims {
					if !strings.Contains(body, record.prose) {
						t.Fatalf("the %s fixture no longer says %q, so the extracted claim is invented", tt.fixture, record.prose)
					}
					claims = append(claims, Claim{Statement: record.prose, Evidence: record.evidence})
				}
			}

			if len(claims) != len(tt.claims) {
				t.Fatalf("the %s fixture states %d claims, the record covers %d", tt.fixture, len(claims), len(tt.claims))
			}

			// One loop from here down, whichever way the claims arrived. An
			// extracted claim is verified by the same behaviour as one the
			// proposal carried, or "the same behaviour" is only a promise.
			worst := Holds
			for i, c := range claims {
				record := tt.claims[i]

				// Gather. A claim whose evidence cannot be read settles
				// nothing, and an unread claim is never a pass.
				raw, err := os.ReadFile(filepath.Join(repoRoot, c.Evidence))
				if err != nil {
					t.Errorf("claim %q: evidence %q could not be gathered: %v", c.Statement, c.Evidence, err)
					worst = Unresolved
					continue
				}
				got := string(raw)

				// Decide, from what was gathered and nothing else.
				for _, wanted := range record.present {
					if !strings.Contains(got, wanted) {
						t.Errorf("claim %q: %s no longer contains %q, so the recorded verdict %q is stale",
							c.Statement, c.Evidence, wanted, record.verdict)
					}
				}
				for _, unwanted := range record.absent {
					if strings.Contains(got, unwanted) {
						t.Errorf("claim %q: %s now contains %q, so the recorded verdict %q is stale",
							c.Statement, c.Evidence, unwanted, record.verdict)
					}
				}

				if record.verdict == Fails {
					worst = Fails
				}
				t.Logf("claim %q -> %s (%s)", c.Statement, record.verdict, record.why)
			}

			if worst != tt.block {
				t.Errorf("the %s fixture verifies as %q, want %q", tt.fixture, worst, tt.block)
			}
		})
	}
}

// TestIssueThirtyPremiseFixtureParses pins the premise of issue #30, the
// proposal that started this work. Its claims are false for this repository and
// the evidence that settles them is known, which makes it the case a later
// stage has to catch. Slice 4 verifies the claims; this only captures them.
func TestIssueThirtyPremiseFixtureParses(t *testing.T) {
	path := filepath.Join("testdata", "issue-30.md")

	src, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read the issue #30 fixture: %v", err)
	}

	block, err := Parse(string(src))
	if err != nil {
		t.Fatalf("the issue #30 fixture does not parse: %v", err)
	}
	if len(block.Claims) == 0 {
		t.Fatal("the issue #30 fixture captures no claims")
	}

	for _, c := range block.Claims {
		if c.Evidence == "" {
			t.Errorf("claim %q captures no evidence", c.Statement)
			continue
		}
		// The fixture names paths, not commands, so the evidence has to be
		// something a later stage can actually open.
		if strings.ContainsAny(c.Evidence, " ") {
			continue
		}
		if _, err := os.Stat(filepath.Join(repoRoot, c.Evidence)); err != nil {
			t.Errorf("claim %q names evidence that is not in this repository: %v", c.Statement, err)
		}
	}
}
