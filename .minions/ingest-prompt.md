You are analyzing content to extract feature ideas for the Partio project.

Partio captures the reasoning behind code changes by hooking into Git workflows to preserve AI agent sessions alongside commits. It consists of:

- **cli** (Go): The core CLI tool — hooks into git, captures sessions, stores checkpoints
- **app** (Next.js): Dashboard for browsing checkpoint data
- **docs** (Mintlify): Documentation site
- **site** (Next.js): Marketing website
- **extension**: Browser extension

## Source Content

Type: {{SOURCE_TYPE}}
URL: {{SOURCE_URL}}

### Content
{{CONTENT}}

## Instructions

Analyze the content above and extract feature ideas that could be adapted for Partio. These are INSPIRATION — not direct copies. Partio and the source are related but independent products.

For each feature idea, output a JSON object with these fields:

```json
{
  "id": "kebab-case-id",
  "title": "Short descriptive title",
  "source": "reference to original (e.g., 'entireio/cli#373 (changelog 0.4.5)')",
  "description": "What Partio should implement, adapted for its own architecture and conventions. Be specific about the desired behavior.",
  "why": "Brief explanation of why this matters for Partio — what problem it solves or what value it adds for users.",
  "user_relevance": "Why this change is relevant to Partio users — how it improves their experience, workflow, or the value they get from Partio.",
  "target_repos": ["cli"],
  "context_hints": ["cli/internal/relevant/path/"],
  "acceptance_criteria": ["specific testable criterion 1", "specific testable criterion 2"]
}
```

Do NOT include `docs` in `target_repos`. Documentation updates are handled automatically by the doc minion after code PRs merge — speculative docs PRs created alongside code changes become stale during review and get orphaned if the code PR is rejected.

Output a JSON array of feature objects. Only include features that are genuinely relevant to Partio's domain (Git workflows, AI agent sessions, code attribution, checkpoints). Skip features that don't apply.

If no relevant features are found, output an empty array: `[]`

Output ONLY the JSON array, no other text.
