---
id: preflight-knowledge-context
target_repos:
  - cli
acceptance_criteria:
  - Checkpoint metadata includes a knowledge_context field capturing structured context files present at commit time
  - At minimum, captures the presence and hash of CLAUDE.md, .cursorrules, and similar agent context files
  - The knowledge context is recorded during pre-commit detection alongside session state
  - Post-commit includes the knowledge context in the checkpoint metadata
  - The captured data answers "what context files were available to the agent when this code was written"
  - Unit tests verify knowledge context is captured and stored in checkpoint metadata
pr_labels:
  - minion
---

# Capture pre-flight knowledge context in checkpoint metadata

Record which agent context files (CLAUDE.md, .cursorrules, etc.) were present and their content hashes at checkpoint time, so reviewers can understand what knowledge the agent was operating from.

## Context

Checkpoints capture what the agent thought and decided — prompts, transcripts, tool calls. But they don't capture what the agent knew before execution: the structured knowledge context that was available at the time.

For audit and review purposes, the complete record should answer three questions:
1. What knowledge was the agent operating from? (knowledge available)
2. What did it decide and do? (reasoning applied — already captured by checkpoints)
3. What was the outcome? (code changes — already captured by git)

Currently only questions 2 and 3 are answered. Question 1 requires manually checking the git history of context files, which may have changed since the checkpoint was created.

## Approach

During pre-commit detection, scan for known agent context files in the repository:
- `CLAUDE.md` / `.claude/` directory
- `.cursorrules`
- `.gemini/` directory
- Any files matching common agent instruction patterns

For each file found, record its path and git blob hash. Store this as a `knowledge_context` field in the checkpoint metadata JSON. This is lightweight (just paths and hashes, not file contents) and provides a durable snapshot of what the agent had available.

The actual file contents can be retrieved later via `git cat-file -p <hash>` since the blob is already in the git object database.
