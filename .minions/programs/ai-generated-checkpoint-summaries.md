---
id: ai-generated-checkpoint-summaries
target_repos:
  - cli
acceptance_criteria:
  - "partio explain --generate produces an AI-generated natural language summary of the checkpoint's session transcript"
  - "The summary provider is configurable via partio settings (e.g., claude, local model)"
  - "Summary generation has a configurable timeout (default 30s) to prevent indefinite hangs"
  - "When no provider is configured, partio explain --generate prompts the user to select one"
  - "Generated summaries are cached in checkpoint metadata so repeated calls don't re-generate"
  - "partio explain without --generate continues to show the raw transcript as before"
pr_labels:
  - minion
---

# AI-generated checkpoint summaries with provider selection

## Summary

Add a `--generate` flag to `partio explain` that produces an AI-generated natural language summary of a checkpoint's session transcript. Support configurable summary providers and cache results to avoid redundant API calls.

## Motivation

Raw session transcripts can be long and hard to scan, especially for checkpoints with many turns or tool invocations. An AI-generated summary would:

- Give reviewers a quick "what happened and why" for any checkpoint
- Make checkpoint browsing more useful for team leads reviewing agent-assisted work
- Complement the existing `partio explain` output with a higher-level narrative
- Support `git log` integration where a one-line summary is more practical than a full transcript

## Implementation Notes

- Add `--generate` flag to the `explain` command in `cmd/partio/`
- Create a summary provider interface in a new `internal/summary/` package
- Implement a Claude Code provider that invokes the Claude CLI to generate summaries
- Add a `summary_provider` setting to the config layer (`internal/config/`)
- Apply a 30-second timeout (configurable) to summary generation to prevent hangs
- Cache generated summaries in the checkpoint metadata on the orphan branch
- When no provider is configured and `--generate` is used, prompt the user to select one
- Keep summary generation fully optional — `partio explain` without `--generate` is unchanged

## Source

Inspired by entireio/cli#887 and entireio/cli#876 — adds summary provider selection for `explain --generate` and timeout protection for AI summary generation.
