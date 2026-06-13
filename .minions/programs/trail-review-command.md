---
id: trail-review-command
target_repos:
  - cli
acceptance_criteria:
  - "`partio trail review <trail-id>` displays a synthesized review of all sessions in a trail"
  - "Review output includes: summary of changes across sessions, files modified, agents used, and key decision points"
  - "`partio trail review --comments` lists review comments associated with the trail"
  - "Review correlates sessions with their commits to show the full development arc"
  - "Works with both local checkpoint data and remote trail data when available"
pr_labels:
  - minion
---

# Add `partio trail review` command for trail-scoped session reviews

Add a `partio trail review` subcommand that provides a consolidated review of all sessions grouped within a trail, showing the complete development arc from start to finish.

## Context

Proposal #341 introduced session trails for grouping related sessions across commits. Once sessions are grouped into trails, the natural next step is reviewing a trail as a coherent unit — seeing the full story of a feature or bug fix across multiple agent sessions and commits. The Entire CLI introduced trail review in PR #1266 with dashboard integration and comment listing.

## Desired behavior

- `partio trail review <trail-id>` synthesizes a review across all sessions in the trail:
  - Lists sessions chronologically with their commits
  - Shows cumulative files changed and agents used
  - Highlights key prompts and decision points from each session
- `partio trail review --comments` shows review comments left on the trail (from dashboard or API).
- `partio trail review --summary` outputs a concise one-paragraph summary of the trail's development arc.
- Falls back gracefully when trail data is incomplete (e.g., some sessions lack transcripts).

## Why

Individual session reviews miss the big picture. A trail review shows how a feature evolved across multiple sessions — what was attempted, revised, and completed. This is valuable for code review, onboarding (understanding how a feature was built), and retrospectives.
