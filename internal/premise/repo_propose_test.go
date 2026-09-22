package premise

import (
	"os"
	"strings"
	"testing"
)

// proposeProgram is this repository's propose program, relative to the package
// directory that go test runs in.
const proposeProgram = "../../.minions/programs/propose.md"

// TestProposeProgramTemplateParses checks that the premise template the propose
// program shows its agent is the same format this package parses. If the
// program and the parser drift apart, every later stage is back to reading
// prose, which is the failure this whole feature exists to stop.
func TestProposeProgramTemplateParses(t *testing.T) {
	src, err := os.ReadFile(proposeProgram)
	if err != nil {
		t.Fatalf("read propose program: %v", err)
	}

	tmpl, ok := fencedBlockContaining(string(src), Marker)
	if !ok {
		t.Fatalf("propose program shows no fenced premise template containing %q", Marker)
	}

	block, err := Parse(tmpl)
	if err != nil {
		t.Fatalf("premise template in the propose program does not parse: %v", err)
	}
	if len(block.Claims) == 0 {
		t.Fatal("premise template in the propose program carries no claims")
	}
	for i, c := range block.Claims {
		if c.Evidence == "" {
			t.Errorf("claim %d in the propose program template has no evidence", i)
		}
	}
}

// TestProposeProgramFilesPremiseWithEveryProposal checks that the issue the
// program creates carries the premise section. Showing the format somewhere in
// the program's prose is not enough: the section has to reach the issue body.
func TestProposeProgramFilesPremiseWithEveryProposal(t *testing.T) {
	src, err := os.ReadFile(proposeProgram)
	if err != nil {
		t.Fatalf("read propose program: %v", err)
	}

	create, ok := lineContaining(string(src), "gh issue create")
	if !ok {
		t.Fatal("propose program gives no gh issue create instruction")
	}
	if !strings.Contains(create, "premise section") {
		t.Errorf("the gh issue create instruction leaves the premise section out of the issue body:\n%s", create)
	}
}

// TestProposeProgramKeepsItsExistingProposalBehaviour guards the parts of the
// proposal this slice must not disturb. The premise section is an addition to
// the issue body; the title, the labels and the duplicate check stay as they
// were.
func TestProposeProgramKeepsItsExistingProposalBehaviour(t *testing.T) {
	raw, err := os.ReadFile(proposeProgram)
	if err != nil {
		t.Fatalf("read propose program: %v", err)
	}
	src := string(raw)

	create, ok := lineContaining(src, "gh issue create")
	if !ok {
		t.Fatal("propose program gives no gh issue create instruction")
	}
	for _, want := range []string{
		"--repo <this-repo>",
		"--label minion-proposal",
		`--title "<title>"`,
		"<description",
	} {
		if !strings.Contains(create, want) {
			t.Errorf("the gh issue create instruction lost %q:\n%s", want, create)
		}
	}

	if want := `gh issue list --repo <this-repo> --label minion-proposal --search "<feature-id>" --limit 1`; !strings.Contains(src, want) {
		t.Errorf("propose program lost %q", want)
	}
}

// TestProposeProgramDropsAnIdeaWhosePremiseFails checks the order the proposer
// works in. Verifying after the issue exists changes nothing: the proposal is
// already filed and the operator is already reading it. The check has to come
// first, and a failed check has to leave no issue behind.
func TestProposeProgramDropsAnIdeaWhosePremiseFails(t *testing.T) {
	raw, err := os.ReadFile(proposeProgram)
	if err != nil {
		t.Fatalf("read propose program: %v", err)
	}
	agents, ok := section(string(raw), "## Agents")
	if !ok {
		t.Fatal("propose program has no agents section")
	}

	steps := []struct {
		what  string
		token string
	}{
		{"apply the verifier", VerifierPath},
		// Not merely "drop the idea": the premise-format bullet already
		// tells the agent to drop an idea it can find no evidence for, and
		// that is a different event from a premise that was checked and
		// failed.
		{"drop an idea whose premise does not hold", "does not hold, drop the idea"},
		{"create the issue", "gh issue create"},
	}
	at := make([]int, len(steps))
	for i, s := range steps {
		at[i] = strings.Index(agents, s.token)
		if at[i] < 0 {
			t.Fatalf("the proposer is never told to %s", s.what)
		}
		if i > 0 && at[i-1] > at[i] {
			t.Errorf("the proposer is told to %s before it is told to %s", steps[i].what, steps[i-1].what)
		}
	}

	drop, _, _ := strings.Cut(agents[at[1]:], "\n   - ")
	if !strings.Contains(drop, "no issue") {
		t.Errorf("dropping an idea leaves %q unsaid, so a failed check still files something:\n%s", "no issue", drop)
	}
}

// TestProposeProgramGrantsTheToolsVerificationNeeds pins the proposer's tool
// grant. Verification is reading the tree and running commands against it, so
// an agent without those tools cannot verify and will answer from the source
// item instead — the same wrong answer as before, now with a verdict attached
// to it. The grant already covered this before the slice; the test is here so
// a later edit cannot narrow it without saying so.
func TestProposeProgramGrantsTheToolsVerificationNeeds(t *testing.T) {
	raw, err := os.ReadFile(proposeProgram)
	if err != nil {
		t.Fatalf("read propose program: %v", err)
	}

	grant, ok := fencedBlockContaining(string(raw), "tools:")
	if !ok {
		t.Fatal("the proposer declares no tools")
	}
	for tool, why := range map[string]string{
		"Read": "read a file the evidence names",
		"Glob": "locate a path the evidence names",
		"Grep": "find a symbol the evidence names",
		"Bash": "run a command the evidence names",
	} {
		if !strings.Contains(grant, "- "+tool+"\n") {
			t.Errorf("the proposer cannot %s: %s is not granted", why, tool)
		}
	}
}

// TestProposeProgramSeparatesSkippedFromDropped checks that a run reports the
// two ways an idea can fail to become an issue as different events. An
// irrelevant source item was never an idea; a dropped idea was one, and the
// tree contradicted it. Folded into one count they are unreadable, and the
// cursor has already advanced past both — so the run output is the only place
// the difference survives.
func TestProposeProgramSeparatesSkippedFromDropped(t *testing.T) {
	raw, err := os.ReadFile(proposeProgram)
	if err != nil {
		t.Fatalf("read propose program: %v", err)
	}
	agents, ok := section(string(raw), "## Agents")
	if !ok {
		t.Fatal("propose program has no agents section")
	}

	at := strings.Index(agents, "Print summary")
	if at < 0 {
		t.Fatal("the proposer prints no summary of what a run did")
	}
	summary := agents[at:]

	for outcome, meaning := range map[string]string{
		"filed":   "an idea whose premise held",
		"dropped": "an idea whose premise did not hold",
		"skipped": "a source item that was never relevant",
	} {
		if !strings.Contains(summary, outcome) {
			t.Errorf("the run summary does not report %q — %s", outcome, meaning)
		}
	}

	// Skipping is the ingest prompt's job and predates this slice. Raising the
	// bar on premises must not quietly turn "not relevant" into "rejected".
	prompt, err := os.ReadFile(ingestPrompt)
	if err != nil {
		t.Fatalf("read ingest prompt: %v", err)
	}
	if !strings.Contains(string(prompt), "Skip features that don't apply") {
		t.Error("the ingest prompt no longer skips irrelevant source items")
	}
}

// TestProposeProgramFilesTheEvidenceItGathered checks that a proposal which
// passed the check carries the evidence that passed it. A green verdict with
// nothing behind it is the same prose assertion the premise block replaced.
func TestProposeProgramFilesTheEvidenceItGathered(t *testing.T) {
	raw, err := os.ReadFile(proposeProgram)
	if err != nil {
		t.Fatalf("read propose program: %v", err)
	}

	create, ok := lineContaining(string(raw), "gh issue create")
	if !ok {
		t.Fatal("propose program gives no gh issue create instruction")
	}
	if !strings.Contains(create, "gathered evidence") {
		t.Errorf("the issue body carries the premise but not the evidence gathered for it:\n%s", create)
	}
}

// TestProposeProgramKeepsTheWholeProposalInTheIssue checks that a proposal is
// one GitHub issue and nothing else. A build reads the issue and never a file in
// the tree, so a proposal split across the two loses to the build whatever sits
// in the file. On 2026-09-22, 490 of the 500 program files that open proposals
// linked to did not exist on main, and 171 of those proposals carried no
// acceptance criteria in the issue.
func TestProposeProgramKeepsTheWholeProposalInTheIssue(t *testing.T) {
	raw, err := os.ReadFile(proposeProgram)
	if err != nil {
		t.Fatalf("read propose program: %v", err)
	}
	agents, ok := section(string(raw), "## Agents")
	if !ok {
		t.Fatal("propose program has no agents section")
	}

	create, ok := lineContaining(agents, "gh issue create")
	if !ok {
		t.Fatal("propose program gives no gh issue create instruction")
	}
	for _, want := range []string{"what to build", "acceptance criteria", "proposal id"} {
		if !strings.Contains(create, want) {
			t.Errorf("the issue body leaves out %q, so no build ever sees it:\n%s", want, create)
		}
	}

	// The duplicate check searches the issues for the id. With no program
	// marker to carry it, the id line is what the search finds.
	if !strings.Contains(agents, "`Proposal id: <id>`") {
		t.Error("the proposer is never told to end the issue with its id, so the duplicate check has nothing to find")
	}

	for _, gone := range []string{"write a program file", "<!-- program:"} {
		if strings.Contains(agents, gone) {
			t.Errorf("the proposer is still told %q; a proposal is an issue, not a file", gone)
		}
	}
}

// lineContaining returns the first line of src that contains want.
func lineContaining(src, want string) (string, bool) {
	for _, line := range strings.Split(src, "\n") {
		if strings.Contains(line, want) {
			return line, true
		}
	}
	return "", false
}

// fencedBlockContaining returns the body of the first fenced code block in src
// that contains want.
func fencedBlockContaining(src, want string) (string, bool) {
	lines := strings.Split(src, "\n")
	for i := 0; i < len(lines); i++ {
		if !strings.HasPrefix(strings.TrimSpace(lines[i]), "```") {
			continue
		}
		for j := i + 1; j < len(lines); j++ {
			if !strings.HasPrefix(strings.TrimSpace(lines[j]), "```") {
				continue
			}
			body := strings.Join(lines[i+1:j], "\n")
			if strings.Contains(body, want) {
				return body, true
			}
			i = j
			break
		}
	}
	return "", false
}
