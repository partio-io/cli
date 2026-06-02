---
id: investigate-command-checkpoint-archaeology
target_repos:
  - cli
acceptance_criteria:
  - partio investigate <file> [--line N] queries checkpoint history to find sessions that touched the specified code
  - Output includes the session prompt, key decisions, and commit context for each relevant checkpoint
  - Results are ordered by relevance (most recent checkpoint touching the target code first)
  - Works offline using only the local checkpoint branch data
  - Hidden behind partio labs until stable
pr_labels:
  - minion
---

# Add `partio investigate` command for checkpoint-powered code archaeology

## Summary

Add a `partio investigate` command that queries checkpoint data to trace the reasoning history behind specific code, turning passively captured session data into an active debugging and understanding tool.

## Problem

Partio captures rich context about why code was written — prompts, transcripts, tool calls, reasoning. But this data is only accessible by browsing checkpoints chronologically. When a developer asks "why was this code written this way?", there's no direct way to query the checkpoint history for a specific file or line range.

## Solution

Implement `partio investigate <file> [--line N|N-M]` that:

1. Walks the checkpoint branch to find checkpoints whose associated commits touched the target file/lines
2. Extracts the relevant session context (prompt, key transcript excerpts, tool calls)
3. Presents a summary showing the reasoning chain behind the code

This works entirely offline using the local `partio/checkpoints/v1` branch data and git plumbing.

## Why this matters

This closes the loop on Partio's value proposition: capturing the "why" behind code changes is only valuable if developers can efficiently retrieve that context when they need it. An investigate command transforms checkpoint data from a passive archive into an active tool for code understanding and debugging.

## Source

Inspired by entireio/cli "entire investigate" labs command for multi-agent investigation loops (changelog 0.7.0, #1231).
