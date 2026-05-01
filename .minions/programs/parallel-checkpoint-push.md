---
id: parallel-checkpoint-push
target_repos:
  - cli
acceptance_criteria:
  - Checkpoint refs are pushed concurrently rather than sequentially during pre-push hook
  - Push output is aggregated into a single status line rather than per-ref progress
  - Failed pushes identify which specific ref could not be synced
  - No regression in push reliability when multiple refs exist
pr_labels:
  - minion
  - enhancement
---

# Push checkpoint refs in parallel

When `partio` pushes the checkpoint branch during the pre-push hook, it currently does so sequentially. For repositories with multiple checkpoint refs (e.g., after many sessions), this adds noticeable latency to `git push`.

## What to implement

Modify the pre-push hook's checkpoint push logic to:

1. Identify all checkpoint refs that need to be pushed
2. Push them concurrently using goroutines with a bounded worker pool
3. Collect results and report a single aggregated status message
4. On partial failure, clearly indicate which ref(s) failed while allowing successful pushes to complete

## Why this matters

Users experience the pre-push hook as added latency on every `git push`. Reducing this latency improves the developer experience and reduces the chance that users disable checkpoint pushing due to slowness.

## Source

Inspired by entireio/cli PR #1094 — parallel v2 checkpoint push with cleaner output.
