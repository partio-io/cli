package premise

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// researchProgram is this repository's research program, relative to the package
// directory that go test runs in.
const researchProgram = "../../.minions/programs/research.md"

// gateDoc is this repository's stage gate, relative to the package directory
// that go test runs in.
const gateDoc = "../../" + GatePath

// The premise verdict passes between the research program's agents through a
// file, because each agent is a separate session in a worktree that is
// discarded when it finishes. These two are that channel: the path every agent
// reads, and the only first line that lets a run continue.
const (
	premiseStatePath = `PREMISE="/tmp/minion-research-premise-${MINION_ISSUE_NUMBER:-0}.md"`
	premiseOK        = "PREMISE_OK"
)

// TestResearchStageVerifiesBeforeItPlans is the end-to-end path this slice
// builds. A premise checked at filing time says nothing about the tree a
// research run reads months later: issue #30 was filed on 2026-03-26 and built
// on 2026-08-10. So research rechecks, and it rechecks first — a verdict that
// arrives after the PRD is written has not stopped anything.
func TestResearchStageVerifiesBeforeItPlans(t *testing.T) {
	gate := readRepoFile(t, gateDoc)
	if !strings.Contains(gate, GateMarker) {
		t.Errorf("the stage gate does not carry %q", GateMarker)
	}

	src := readRepoFile(t, researchProgram)

	context, ok := section(src, "## Context")
	if !ok {
		t.Fatal("research program has no context section")
	}
	for _, path := range []string{VerifierPath, GatePath} {
		if !strings.Contains(context, path) {
			t.Errorf("the research program does not put %s in front of its agents", path)
		}
	}

	agents, ok := section(src, "## Agents")
	if !ok {
		t.Fatal("research program has no agents section")
	}

	verify := strings.Index(agents, VerifierPath)
	if verify < 0 {
		t.Fatalf("the research stage is never told to apply %s", VerifierPath)
	}
	// Every artifact a research run produces. The check is worthless unless it
	// precedes all of them.
	for what, token := range map[string]string{
		"asks its first question": "### researcher",
		"writes the PRD":          "### prd-writer",
		"slices the plan":         "### slicer",
		"publishes anything":      "### publisher",
	} {
		at := strings.Index(agents, token)
		if at < 0 {
			t.Fatalf("research program has no %s agent", token)
		}
		if verify > at {
			t.Errorf("the research stage verifies the premise only after it %s", what)
		}
	}
}

// TestResearchRestatesNeitherVerificationNorGating checks that the research
// stage reaches the two descriptions rather than carrying its own copy. A
// pasted copy is free to drift: the stage keeps working, keeps producing
// verdicts, and stops meaning the same thing as the stage beside it. The
// markers are the anchor — a program that carries one has taken a copy.
func TestResearchRestatesNeitherVerificationNorGating(t *testing.T) {
	src := readRepoFile(t, researchProgram)

	for marker, doc := range map[string]string{
		VerifierMarker: VerifierPath,
		GateMarker:     GatePath,
	} {
		if strings.Contains(src, marker) {
			t.Errorf("the research program carries %s rather than referencing %s", marker, doc)
		}
	}

	checker, ok := section(src, "### premise-checker")
	if !ok {
		t.Fatal("research program has no premise-checker agent")
	}
	for path, what := range map[string]string{
		VerifierPath: "how to verify a claim",
		GatePath:     "what to do with the verdict",
	} {
		if !containsPhrase(checker, path) {
			t.Errorf("the premise-checker decides %s for itself: it never reaches %s", what, path)
		}
	}
	if !containsPhrase(checker, "as written") {
		t.Error("the premise-checker is not told to follow the descriptions as written, " +
			"so it is free to paraphrase them into a different behaviour")
	}
}

// TestGatingIsDescribedOnce mirrors TestVerificationIsDescribedOnce for the
// gate. Slices 7 and 8 add callers, and a caller that pastes the procedure into
// its own program stops its own way — which leaves the operator unable to tell,
// from a blocked issue, which stage blocked it or why.
func TestGatingIsDescribedOnce(t *testing.T) {
	carriers := markerCarriers(t, GateMarker)

	if len(carriers) != 1 {
		t.Fatalf("gating is described in %d files, want exactly one: %v", len(carriers), carriers)
	}
	if want := filepath.ToSlash(filepath.Join(repoRoot, GatePath)); carriers[0] != want {
		t.Errorf("gating is described in %s, want %s", carriers[0], want)
	}
}

// addLabel matches a gh call that puts a label on the issue.
var addLabel = regexp.MustCompile(`--add-label\s+(\S+)`)

// TestBlockedResearchRunLabelsTheIssue checks the blocking label. The approval
// program already skips on do-not-build, so this label needs no new pipeline
// wiring — and a second, invented label would need all of it while looking, on
// the issue, exactly like a working block.
func TestBlockedResearchRunLabelsTheIssue(t *testing.T) {
	blocked := blockedSection(t)

	labels := addLabel.FindAllStringSubmatch(blocked, -1)
	if len(labels) == 0 {
		t.Fatal("a blocked stage adds no label, so nothing downstream knows the premise failed")
	}
	for _, m := range labels {
		if m[1] != BlockingLabel {
			t.Errorf("a blocked stage adds %q; the pipeline only understands %q", m[1], BlockingLabel)
		}
	}
}

// TestBlockedResearchRunSaysWhatItCheckedAndFound checks the comment. A label
// alone tells the operator a stage stopped and nothing about why; the operator
// then has to redo the verification by hand to decide whether to overrule it.
// The comment is the account, so it carries the claims and the evidence behind
// each verdict — including for claims that held.
func TestBlockedResearchRunSaysWhatItCheckedAndFound(t *testing.T) {
	blocked := blockedSection(t)

	if !containsPhrase(blocked, GateCommentMarker) {
		t.Errorf("the blocked-run comment carries no %s, so nothing can find it later", GateCommentMarker)
	}
	for what, token := range map[string]string{
		"the claim it checked":                "the claim, word for word",
		"the evidence that claim named":       "the evidence the claim named",
		"the verdict it reached":              "the verdict",
		"the excerpt behind that verdict":     "the excerpt that produced it",
		"the claims that failed, by name":     "the claims that failed",
		"the evidence that contradicted them": "the evidence that contradicted them",
	} {
		if !containsPhrase(blocked, token) {
			t.Errorf("the blocked-run comment leaves out %s", what)
		}
	}
}

// agentHeading matches the heading that opens one agent in a program.
var agentHeading = regexp.MustCompile(`(?m)^### (\S+)\s*$`)

// TestBlockedResearchRunProducesNoPlan checks that a failed premise stops every
// agent that could produce something, not merely the next one. Each agent is
// its own one-shot session in a discarded worktree, so there is no early return
// to take: a stopped stage is every later agent reading the verdict and
// declining to work. An agent added later without that read resumes the run
// halfway through, and publishes a PRD for a premise the tree contradicts.
func TestBlockedResearchRunProducesNoPlan(t *testing.T) {
	src := readRepoFile(t, researchProgram)

	agents, ok := section(src, "## Agents")
	if !ok {
		t.Fatal("research program has no agents section")
	}

	names := agentHeading.FindAllStringSubmatch(agents, -1)
	if len(names) < 2 {
		t.Fatalf("research program declares %d agents, want the checker plus the pipeline", len(names))
	}

	for _, m := range names {
		name := m[1]
		if name == "premise-checker" {
			continue
		}

		body, ok := section(agents, "### "+name)
		if !ok {
			t.Fatalf("cannot read the %s agent", name)
		}
		for what, token := range map[string]string{
			"read the premise verdict":           premiseStatePath,
			"stop unless the verdict passed":     premiseOK,
			"do it before it does anything else": "Before anything else",
		} {
			if !containsPhrase(body, token) {
				t.Errorf("the %s agent does not %s, so a blocked run still produces its artifact", name, what)
			}
		}
	}

	// The gate says the same thing once, for every stage that applies it.
	blocked := blockedSection(t)
	for _, want := range []string{"Stop before you produce anything", "No PRD, no slice plan"} {
		if !containsPhrase(blocked, want) {
			t.Errorf("the stage gate does not say %q, so what \"stop\" means is left to each caller", want)
		}
	}
}

// TestBlockedResearchRunNeverClosesTheIssue checks that blocking is reversible.
// A wrong verdict on a closed issue costs the operator the issue; a wrong
// verdict on a labelled one costs a label removal. The stage reports what it
// found and hands the decision back — it does not make it.
func TestBlockedResearchRunNeverClosesTheIssue(t *testing.T) {
	src := readRepoFile(t, researchProgram)

	for what, token := range map[string]string{
		"closes the issue":            "gh issue close",
		"marks the issue as finished": "--add-label minion-done",
	} {
		if strings.Contains(src, token+"\n") || strings.Contains(src, token+" ") {
			t.Errorf("the research stage %s, which the operator cannot undo with a label removal", what)
		}
	}

	blocked := blockedSection(t)
	for _, want := range []string{"Do not close the issue", "minion-done"} {
		if !containsPhrase(blocked, want) {
			t.Errorf("the stage gate does not forbid %q on a blocked issue", want)
		}
	}
}

// TestOperatorOverrulesByRemovingTheLabel checks the escape hatch. Verification
// is a judgement made by a model against a tree, so it will be wrong sometimes,
// and a gate with no override turns each wrong verdict into a dead issue. The
// override works because no stage treats the label as a stored decision: every
// run re-gathers the evidence, so removing the label is enough to get a fresh
// verdict rather than the old one replayed.
func TestOperatorOverrulesByRemovingTheLabel(t *testing.T) {
	override, ok := section(readRepoFile(t, gateDoc), "## How the operator overrules")
	if !ok {
		t.Fatal("the stage gate gives the operator no way to overrule a verdict")
	}
	for what, token := range map[string]string{
		"name the label the operator removes":     BlockingLabel,
		"say the removal is the whole mechanism":  "removes",
		"say a later run reaches its own verdict": "re-verifies",
	} {
		if !containsPhrase(override, token) {
			t.Errorf("the override section does not %s", what)
		}
	}

	// A stage that skipped on the label would read the previous run's verdict
	// instead of reaching its own, and the operator's removal would be the only
	// thing that ever decided anything.
	src := readRepoFile(t, researchProgram)
	if strings.Contains(src, BlockingLabel) {
		t.Errorf("the research program names %q itself; the label belongs to %s, "+
			"and a program that reads it can skip on it instead of re-verifying",
			BlockingLabel, GatePath)
	}

	// Re-adding a label that is already present fires no trigger, so a repeat
	// block would be silent. The gate removes it first, every time.
	blocked := blockedSection(t)
	remove := strings.Index(blocked, "--remove-label "+BlockingLabel)
	add := strings.Index(blocked, "--add-label "+BlockingLabel)
	switch {
	case remove < 0:
		t.Errorf("a repeat block re-adds %q without removing it first, so the trigger never fires", BlockingLabel)
	case add < 0:
		t.Fatalf("a blocked stage never adds %q", BlockingLabel)
	case remove > add:
		t.Errorf("the gate removes %q after adding it, which leaves the issue unblocked", BlockingLabel)
	}
}

var (
	// shellUse matches a shell variable the program tells an agent to expand.
	shellUse = regexp.MustCompile(`\$\{?([A-Z_][A-Z0-9_]*)`)

	// shellAssign matches a shell variable the program tells an agent to set.
	shellAssign = regexp.MustCompile(`(?m)^\s*([A-Z_][A-Z0-9_]*)=`)
)

// environmentVars are supplied to every agent by the runner, so a program may
// expand them without assigning them.
var environmentVars = map[string]bool{
	"MINION_ISSUE_NUMBER": true,
	"GITHUB_RUN_ID":       true,
}

// TestHoldingPremiseIsRefreshedInTheIssue checks the passing path. Verifying and
// then discarding what you found leaves the build stage rechecking against the
// proposer's months-old excerpts, so each stage repeats the last stage's work
// and the issue never gets any fresher. The refresh is also a real command with
// a real path: an agent told to write `--body-file "$SOMETHING"` that nothing
// ever sets writes the empty string, and `gh` overwrites the issue body with
// nothing.
func TestHoldingPremiseIsRefreshedInTheIssue(t *testing.T) {
	src := readRepoFile(t, researchProgram)

	checker, ok := section(src, "### premise-checker")
	if !ok {
		t.Fatal("research program has no premise-checker agent")
	}
	for what, token := range map[string]string{
		"write the refreshed body back to the issue": "gh issue edit",
		"keep the premise section it refreshes":      Marker,
		"leave the rest of the body alone":           "the rest of the body untouched",
	} {
		if !containsPhrase(checker, token) {
			t.Errorf("the premise-checker does not %s", what)
		}
	}

	assigned := map[string]bool{}
	for _, m := range shellAssign.FindAllStringSubmatch(src, -1) {
		assigned[m[1]] = true
	}
	for _, m := range shellUse.FindAllStringSubmatch(checker, -1) {
		if name := m[1]; !assigned[name] && !environmentVars[name] {
			t.Errorf("the premise-checker expands $%s, which nothing in the program sets — "+
				"it expands to the empty string at run time", name)
		}
	}

	// The gate says what refreshing means, once, for every stage that holds.
	holds := holdsSection(t)
	if !containsPhrase(holds, "replace the evidence excerpts") {
		t.Error("the stage gate does not say the refresh replaces the evidence with what this run read")
	}
}

// TestHoldingPremiseRecordsWhatWasVerified checks that a green run carries its
// evidence too. A verdict with nothing behind it is the prose assertion the
// premise block replaced — and a run that only records failures leaves the
// operator unable to tell a checked premise from an unchecked one.
func TestHoldingPremiseRecordsWhatWasVerified(t *testing.T) {
	checker, ok := section(readRepoFile(t, researchProgram), "### premise-checker")
	if !ok {
		t.Fatal("research program has no premise-checker agent")
	}
	for what, token := range map[string]string{
		"record every claim it checked":      "record every claim",
		"record the excerpt behind each one": "the excerpt that produced it",
		"do so for a block that holds":       "for a block that holds",
	} {
		if !containsPhrase(checker, token) {
			t.Errorf("a passing run does not %s", what)
		}
	}

	if !containsPhrase(holdsSection(t), "Record the verdict") {
		t.Error("the stage gate does not tell a passing stage to record its verdict")
	}
}

// blockedSection returns the part of the stage gate that describes what a stage
// does when the premise does not hold.
func blockedSection(t *testing.T) string {
	t.Helper()
	return gateSection(t, "## When the premise does not hold")
}

// holdsSection returns the part of the stage gate that describes what a stage
// does when the premise holds.
func holdsSection(t *testing.T) string {
	t.Helper()
	return gateSection(t, "## When the premise holds")
}

// gateSection returns one section of the stage gate.
func gateSection(t *testing.T, heading string) string {
	t.Helper()

	body, ok := section(readRepoFile(t, gateDoc), heading)
	if !ok {
		t.Fatalf("the stage gate has no %q section", heading)
	}
	return body
}

// markerCarriers lists every markdown file under .minions that contains marker.
// Exactly one carrier is what makes "described once" checkable, so both the
// verifier and the gate check their own marker through this walk.
func markerCarriers(t *testing.T, marker string) []string {
	t.Helper()

	var carriers []string
	err := filepath.WalkDir(filepath.Join(repoRoot, ".minions"), func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || filepath.Ext(path) != ".md" {
			return err
		}
		src, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if strings.Contains(string(src), marker) {
			carriers = append(carriers, filepath.ToSlash(path))
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk the minions tree: %v", err)
	}
	return carriers
}

// whitespace matches any run of spaces, tabs and newlines.
var whitespace = regexp.MustCompile(`\s+`)

// containsPhrase reports whether src states want, ignoring where the prose
// happens to wrap. These programs are hand-wrapped markdown, so a phrase moves
// across a line break whenever a sentence above it is edited. Matching the
// bytes would fail on a reflow that changed nothing, and a test that cries wolf
// on rewrapping gets deleted rather than fixed.
func containsPhrase(src, want string) bool {
	return strings.Contains(whitespace.ReplaceAllString(src, " "), whitespace.ReplaceAllString(want, " "))
}

// readRepoFile reads a file this repository ships, failing the test rather than
// returning an error a caller has to re-report at every call site.
func readRepoFile(t *testing.T, path string) string {
	t.Helper()

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(raw)
}
