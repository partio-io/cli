---
id: validate-checkpoint-path-inputs
target_repos:
  - cli
pr_labels:
  - minion
  - security
acceptance_criteria:
  - All path inputs used in checkpoint storage (session IDs, agent names, file paths) are validated to reject path traversal sequences (../, absolute paths, etc.)
  - All path inputs used in session state persistence (.partio/state/) are validated similarly
  - Attempting to create a checkpoint with a crafted identifier containing path traversal sequences returns an error instead of writing to an arbitrary location
  - Unit tests cover path traversal attempts for each entry point that constructs filesystem paths from external input
---

# Validate checkpoint and session path inputs against path traversal

Audit and sanitize all path inputs in Partio's checkpoint storage, session state, and agent lifecycle code to prevent path-traversal / arbitrary-file-write attacks.

## Context

Entireio/cli v0.7.5 disclosed and fixed a path-traversal vulnerability where identifiers (session IDs, agent names, etc.) could be exploited to overwrite files outside the intended directories. Partio has analogous code paths: checkpoint storage writes to git object database using identifiers, session state is written to `.partio/state/`, and agent detection uses directory names derived from external input.

## What to implement

1. Add a path validation helper that rejects identifiers containing `../`, absolute path prefixes, null bytes, or other traversal sequences.
2. Apply the validation at every point where external input (session IDs, agent names, file names) is used to construct filesystem paths — particularly in `internal/checkpoint/`, `internal/session/`, and `internal/agent/`.
3. Ensure git plumbing operations (hash-object, mktree) also validate tree entry names before writing.
