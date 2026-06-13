---
id: embedded-agent-docs
target_repos:
  - cli
acceptance_criteria:
  - partio learn outputs markdown documentation describing all Partio commands and workflows
  - Output is formatted for consumption by AI coding agents (concise, structured, actionable)
  - Documentation is embedded in the binary at build time, not fetched from external sources
  - partio learn --short outputs a condensed single-paragraph summary
  - Output includes common workflows (enable, commit flow, rewind, resume) with example commands
pr_labels:
  - minion
---

# Add `partio learn` command with embedded agent-friendly documentation

## Summary

Add a `partio learn` command that outputs comprehensive, agent-friendly documentation embedded in the CLI binary. This allows AI coding agents to quickly understand how Partio works without users needing to manually paste instructions or link to external docs.

## Motivation

AI coding agents (Claude Code, Codex, etc.) frequently work in repositories where Partio is enabled. When an agent encounters Partio hooks, trailers, or the `.partio/` directory, it has no built-in way to understand what Partio does or how to interact with it. Users currently have to paste documentation or explain Partio manually. A self-documenting CLI command solves this — agents can run `partio learn` to get the context they need.

Inspired by `entire learn` added in entireio/cli#1146, which embeds markdown documentation and regenerates it at release time.

## Design

1. **Embedded docs**: Create `internal/docs/learn.md` with agent-oriented documentation covering:
   - What Partio does (one paragraph)
   - Available commands with brief descriptions
   - Common workflows (enable → commit → checkpoint flow)
   - Environment variables
   - How to interpret Partio-Checkpoint trailers
   - How to use `partio rewind` and `partio resume`

2. **Build-time embedding**: Use Go's `embed` package to include the markdown file in the binary.

3. **Command**: Add `cmd/partio/learn.go` that prints the embedded documentation to stdout.

4. **Flags**:
   - `--short` — output a condensed summary (first paragraph only)
   - Default — full documentation output

5. **Release workflow**: Regenerate the embedded docs from command help text during release builds to keep them in sync.
