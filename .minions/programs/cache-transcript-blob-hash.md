---
id: cache-transcript-blob-hash
target_repos:
  - cli
acceptance_criteria:
  - When writing consecutive checkpoints for the same session, full.jsonl blob hash is reused if content has not changed since the previous checkpoint
  - A content hash comparison (e.g. SHA-256 of the file content) is used to detect unchanged transcripts before calling git hash-object
  - Large transcript files (>1MB) do not cause noticeable latency during checkpoint writes in rapid succession
  - Existing checkpoint write behavior is preserved when cache is empty or content has changed
pr_labels:
  - minion
---

# Cache transcript blob hashes across consecutive checkpoint writes

## Problem

When Partio writes multiple checkpoints during a single agent session (e.g., one per commit in a multi-commit turn), the `full.jsonl` transcript file — which can grow to tens of megabytes — is re-hashed via `git hash-object` on every write, even when its content hasn't changed since the last checkpoint. This adds unnecessary latency to hook execution, which is time-sensitive since it runs inside `post-commit`.

## Desired behavior

The checkpoint `Store` should cache the blob hash of large session files (particularly `full.jsonl`) and reuse the cached hash when the content hasn't changed. This avoids redundant `git hash-object` calls for unchanged content.

## Implementation approach

Add an in-memory cache to `checkpoint.Store` that maps a content fingerprint (e.g., length + leading bytes, or a fast hash) to the git blob hash returned by `hash-object`. Before hashing `full.jsonl`, compare the content fingerprint against the cache. If it matches, reuse the stored blob hash; otherwise, call `hash-object` and update the cache.

The cache should be scoped to the lifetime of the `Store` instance (one hook invocation), so there is no persistence concern.

## Context hints

- `internal/checkpoint/write.go` — the `Write` method where `hashObject` is called for each session file
- `internal/checkpoint/store.go` — the `Store` type that could hold the cache
