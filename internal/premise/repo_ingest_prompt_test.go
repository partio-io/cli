package premise

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

// ingestPrompt is this repository's ingest prompt, relative to the package
// directory that go test runs in.
const ingestPrompt = "../../.minions/ingest-prompt.md"

// TestIngestPromptGroundsClaimsInTheTree checks that the prompt asks for claims
// about this repository rather than claims about the product the idea came
// from. The schema alone does not get that: a well-formed claim inferred from a
// sibling product's changelog fills every field and is still wrong here, which
// is exactly how issue #30 was filed. Adaptation stays the goal; importing the
// source product's premises does not.
func TestIngestPromptGroundsClaimsInTheTree(t *testing.T) {
	raw, err := os.ReadFile(ingestPrompt)
	if err != nil {
		t.Fatalf("read ingest prompt: %v", err)
	}

	instructions, ok := section(string(raw), "## Instructions")
	if !ok {
		t.Fatal("ingest prompt has no instructions section")
	}
	for _, want := range []string{"checked-out tree", "source material"} {
		if !strings.Contains(instructions, want) {
			t.Errorf("the ingest prompt never mentions the %q a claim must be grounded in or against", want)
		}
	}
}

// TestIngestPromptSchemaCarriesPremise checks that the model producing feature
// ideas produces claims and evidence with them, rather than having a premise
// bolted on after the idea already exists.
func TestIngestPromptSchemaCarriesPremise(t *testing.T) {
	src, err := os.ReadFile(ingestPrompt)
	if err != nil {
		t.Fatalf("read ingest prompt: %v", err)
	}

	raw, ok := fencedBlockContaining(string(src), `"acceptance_criteria"`)
	if !ok {
		t.Fatal("ingest prompt shows no feature-idea output schema")
	}

	var idea struct {
		Premise []struct {
			Claim    string `json:"claim"`
			Evidence string `json:"evidence"`
		} `json:"premise"`
	}
	if err := json.Unmarshal([]byte(raw), &idea); err != nil {
		t.Fatalf("feature-idea output schema is not valid JSON: %v", err)
	}

	if len(idea.Premise) == 0 {
		t.Fatal("feature-idea output schema declares no premise field")
	}
	for i, c := range idea.Premise {
		if c.Claim == "" {
			t.Errorf("premise entry %d declares no claim field", i)
		}
		if c.Evidence == "" {
			t.Errorf("premise entry %d declares no evidence field", i)
		}
	}
}

// TestIngestPromptTreatsAnAssumptionInAShouldAsAPremise checks the ingest side
// of #31's gap. "A statement about what Partio should do next is not a premise"
// is true of the behaviour asked for, and false of the facts it takes for
// granted: "a merge strategy appropriate for append-only checkpoint data" is a
// should statement that assumes nothing deletes checkpoint data. That
// assumption goes into premise, where it gets evidence, not into description.
func TestIngestPromptTreatsAnAssumptionInAShouldAsAPremise(t *testing.T) {
	instructions, ok := section(readRepoFile(t, ingestPrompt), "## Instructions")
	if !ok {
		t.Fatal("ingest prompt has no instructions section")
	}
	lower := flat(instructions)
	for _, want := range []string{
		`an assumption about today's code or data inside a "should" statement is still a premise`,
		"goes into `premise`",
	} {
		if !strings.Contains(lower, want) {
			t.Errorf("the ingest prompt never says %q", want)
		}
	}
}
