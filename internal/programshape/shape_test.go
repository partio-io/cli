package programshape

import (
	"strings"
	"testing"
)

func TestCheck(t *testing.T) {
	tests := []struct {
		name string
		src  string
		want []string // headings expected in the findings, in order
	}{
		{
			name: "steps section with no agents section is unreachable",
			src: `---
id: propose
---

# Propose features from monitored sources

Scan monitored sources and create feature proposals.

## Steps

1. **Read sources config** from ` + "`.minions/sources.yaml`" + `.
2. **For each source**, fetch new content since the cursor.
`,
			want: []string{"## Steps"},
		},
		{
			name: "instructions in an agents section reach the model",
			src: `# Repair the lint and test failures a check found

A check failed. Repair it.

## Steps

1. This section is dropped, but the agent below carries the procedure.

## Agents

### checks-fix

1. Read the failing check output.
2. Repair the failure and push the fix.
`,
			want: nil,
		},
		{
			name: "an entire instruction set under the H1 reaches the model",
			src: `# Warn when importing agent history while not authenticated

When ` + "`partio enable`" + ` runs and finds existing history to import, it
must check whether the user is authenticated. If not, import as normal and
warn that the checkpoints are local-only until the user authenticates.

## Why

Silent local-only imports create a deceptive success state.

## Source

entireio/cli issue #1773.
`,
			want: nil,
		},
		{
			name: "a dropped section holding only a command block is unreachable",
			src: `# Publish the release

## Release

` + "```bash\ngit push --tags\n```" + `
`,
			want: []string{"## Release"},
		},
		{
			name: "a kept section is never reported",
			src: `# Research minion

Do the research.

## Context

1. ` + "`.minions/sources.yaml`" + `
2. ` + "`.minions/project.yaml`" + `
`,
			want: nil,
		},
		{
			name: "a heading inside a fenced block is not a section",
			src: `# Show the program format

The format is described below, and nothing here is dropped.

## Example

` + "```markdown\n## Steps\n\n1. A step inside a fence.\n```" + `
`,
			want: []string{"## Example"},
		},
		{
			name: "every dropped section carrying instructions is reported",
			src: `# Build an end-to-end test

## Background

1. The trailer is written by the post-commit hook.

## What to build

1. Create a repository fixture.
2. Commit twice in rapid succession.
`,
			want: []string{"## Background", "## What to build"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			findings := Check(tt.src)

			if len(findings) != len(tt.want) {
				t.Fatalf("Check() returned %d findings, want %d: %v", len(findings), len(tt.want), findings)
			}
			for i, heading := range tt.want {
				if findings[i].Heading != heading {
					t.Errorf("finding %d heading = %q, want %q", i, findings[i].Heading, heading)
				}
				if strings.TrimSpace(findings[i].Reason) == "" {
					t.Errorf("finding %d for %q has an empty reason", i, heading)
				}
			}
		})
	}
}
