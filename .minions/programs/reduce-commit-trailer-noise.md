---
id: reduce-commit-trailer-noise
target_repos:
  - cli
acceptance_criteria:
  - Checkpoint trailer value is shortened (e.g., short hash or compact reference instead of full metadata)
  - A partio config option controls trailer format (full, short, or none)
  - Short format is the new default and renders cleanly in GitHub commit UI and git log
  - Full checkpoint details remain accessible via partio rewind or the checkpoint branch
  - Existing checkpoints with old trailer format continue to work
pr_labels:
  - minion
---

# Reduce visual noise of checkpoint trailers in git log and GitHub UI

## What

Shorten the `Partio-Checkpoint` commit trailer value so it is less visually intrusive in `git log`, GitHub's commit list, and PR diff views. Provide a config option to control trailer verbosity (short hash, full reference, or disabled).

## Why

Community feedback from similar tools shows that long checkpoint trailers clutter the commit UI, especially in teams where every commit carries one. A shorter trailer preserves the link between commits and checkpoints without dominating the visual space. This is particularly important for adoption — noisy trailers discourage teams from enabling the tool.

## How

- Default trailer to a short format: e.g., `Partio-Checkpoint: <short-hash>` instead of a full reference
- Add `trailer_format` config option with values: `short` (default), `full`, `none`
- When `none`, skip trailer entirely (checkpoints still exist on the orphan branch, linked by commit SHA)
- Ensure `partio rewind` and other commands can resolve both short and full trailer formats
- Maintain backward compatibility with existing full-format trailers

## Source

Inspired by entireio/cli#868 — community feedback that checkpoint trailers are too noisy in GitHub commit UI.
