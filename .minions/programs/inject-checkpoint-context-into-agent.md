---
id: inject-checkpoint-context-into-agent
target_repos:
  - cli
acceptance_criteria:
  - "partio enable writes a context snippet (e.g., to CLAUDE.md or a partio-managed context file) describing how to use partio commands to access prior checkpoint history"
  - "The injected context includes the partio log/rewind commands available to the agent"
  - "partio disable removes the injected context snippet"
  - "The context injection is idempotent — running partio enable twice does not duplicate the snippet"
  - "The context file path is configurable via partio settings"
pr_labels:
  - minion
---

# Inject checkpoint context into agent sessions

When Partio is enabled in a repository, inject a brief context snippet into the agent's available context (e.g., appending to `CLAUDE.md` for Claude Code) that tells the agent about Partio's checkpoint history and how to query it.

## Why

Agents currently have no awareness that Partio is capturing their sessions or that prior checkpoint history exists. If the agent knew about `partio status`, `partio rewind`, or the checkpoint branch, it could reference past reasoning when making decisions — closing the loop between capture and retrieval.

## What to implement

- During `partio enable`, append a managed section to the agent's context file (e.g., `CLAUDE.md` for Claude Code) with a brief description of available Partio commands
- Mark the section with comment delimiters so `partio disable` can cleanly remove it
- Make the target file configurable (default: `CLAUDE.md` for claude-code agent)
- Keep the injected context minimal — just enough for the agent to know it can query checkpoint history

## Source

Inspired by entireio/cli PR#1435 ("inject trail context into the model when trails are enabled") and changelog 0.7.6.
