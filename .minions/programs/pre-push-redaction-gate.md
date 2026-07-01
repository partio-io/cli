---
id: pre-push-redaction-gate
target_repos:
  - cli
acceptance_criteria:
  - "Pre-push hook scans all unpushed checkpoint blobs for secrets before pushing"
  - "Detected secrets block the push with a clear error listing the offending checkpoint and match"
  - "Users can configure regex patterns in partio settings under a `redaction_patterns` key"
  - "A default set of patterns covers common secret formats (API keys, tokens, private keys)"
  - "A `--skip-redaction` flag allows bypassing the gate when the user explicitly opts out"
  - "make test passes with tests for pattern matching, push blocking, and bypass"
  - "make lint passes"
pr_labels:
  - minion
---

# Add pre-push redaction gate to block secrets in checkpoint transcripts

## Problem

Partio captures AI agent session transcripts and stores them as checkpoint data. These transcripts can inadvertently contain secrets (API keys, tokens, credentials) that the agent encountered or the user pasted during the session. Currently there is no safeguard preventing these secrets from being pushed to a remote repository via `partio push` or the pre-push hook.

## What to implement

Add a redaction scanning step to the pre-push hook that inspects all unpushed checkpoint blobs before they leave the machine:

1. **Enumerate unpushed checkpoints** — In the pre-push hook, identify checkpoint commits on `partio/checkpoints/v1` that haven't been pushed yet
2. **Scan transcript blobs** — Read the JSONL transcript content from each unpushed checkpoint and run regex pattern matching against it
3. **Block on match** — If any pattern matches, abort the push with a clear error message showing which checkpoint and what was matched (redacted)
4. **Configurable patterns** — Allow users to add custom regex patterns via `redaction_patterns` in partio settings (layered config)
5. **Default patterns** — Ship a built-in set covering common secret formats: AWS keys, GitHub tokens, private key headers, generic API key patterns
6. **Bypass flag** — Support `--skip-redaction` for cases where the user has reviewed and accepted the content

## Why pre-push

Running the filter at push time (not capture time) has key advantages:
- It's the last checkpoint before data leaves the local machine
- It can batch-process all unpushed checkpoints efficiently
- New patterns retroactively catch secrets in older unpushed checkpoints
- It doesn't slow down the commit workflow

## Context

- `internal/hooks/pre_push.go` — Current pre-push hook implementation
- `internal/checkpoint/` — Checkpoint storage and retrieval
- `internal/config/` — Layered configuration system
- Inspired by entireio/cli's OpenAI Privacy Filter with pre-push architecture (PR #1214) and user-defined redaction rules (changelog 0.6.2)
