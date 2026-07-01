---
id: externalize-session-images
target_repos:
  - cli
acceptance_criteria:
  - Agent session images (screenshots, pasted images) are extracted from JSONL transcript data during checkpoint creation
  - Extracted images are stored as separate git blob objects in a per-session assets/ subtree within the checkpoint tree
  - An assets/manifest.json file is written alongside the blobs, mapping each asset to its original transcript reference
  - The feature is opt-in via a config setting (e.g., capture_images or externalize_assets)
  - When disabled (default), no images are extracted and checkpoint creation behaves identically to today
  - Transcript entries that referenced inline images are updated to reference the asset path instead
  - Checkpoint tree size is measurably smaller for sessions containing images when the feature is enabled
pr_labels:
  - minion
---

# Externalize images from agent session transcripts

## What

Add opt-in extraction of images (screenshots, pasted images) from Claude Code session transcripts during checkpoint creation. Instead of storing base64-encoded image data inline in the JSONL transcript blob, extract each image as a separate git blob object and store it in a per-session `assets/` subtree within the checkpoint tree, alongside an `assets/manifest.json` that maps each asset to its transcript reference.

## Why

Claude Code sessions can include screenshots and pasted images (e.g., from `Read` tool on image files, user-provided screenshots). These are stored as base64-encoded data in the JSONL transcript, which:

1. **Bloats checkpoint storage** — a single screenshot can add 500KB+ of base64 text to the transcript blob, and sessions with multiple images compound this quickly on the checkpoint branch.
2. **Makes transcripts harder to parse** — tools processing JSONL transcripts must handle arbitrarily large inline image data.
3. **Prevents efficient deduplication** — git can deduplicate identical blobs across checkpoints, but can't deduplicate images embedded within larger JSONL blobs.

Externalizing images as separate blobs lets git's object storage handle them efficiently and keeps transcripts lean.

## How

During checkpoint creation in `internal/checkpoint/`, after parsing the session JSONL:

1. Scan transcript entries for image content (base64-encoded data in tool results or user messages)
2. For each image, write it as a separate git blob via `hash-object`
3. Build an `assets/manifest.json` mapping asset IDs to blob hashes and original transcript positions
4. Replace inline image data in the transcript with asset references
5. Include the assets subtree in the checkpoint tree via `mktree`

Gate the feature behind a config setting (default: off) so existing behavior is preserved.

## Source

Inspired by entireio/cli#1589 and changelog 0.7.8 — Entire added opt-in image externalization for Claude Code, Codex, and Cursor sessions with per-session `assets/` folders and `assets/manifest.json`.
