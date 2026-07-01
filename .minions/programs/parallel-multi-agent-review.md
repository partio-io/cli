---
id: parallel-multi-agent-review
target_repos:
  - cli
acceptance_criteria:
  - "A review orchestration mechanism can invoke multiple configured review commands in parallel"
  - "Each review subprocess is independently cancellable without affecting others"
  - "Combined review output is collected and attached to checkpoint metadata"
  - "Graceful handling when one or more review agents fail (partial results preserved)"
  - "Table-driven tests cover parallel execution, partial failure, and cancellation scenarios"
pr_labels:
  - minion
---

# Support parallel multi-agent review orchestration

## Description

Add the ability to orchestrate multiple review agents (or review skills) running in parallel against the same code changes. When a review is triggered, Partio should launch configured review commands as parallel subprocesses, collect their outputs independently, and attach the combined review metadata to the next checkpoint.

Each subprocess should be independently cancellable via context/signal propagation, so that if one agent hangs or the user cancels, partial results from completed agents are preserved rather than discarded.

## Why

Code review benefits from diverse perspectives — different agents may catch different categories of issues (security, style, correctness, performance). Running them sequentially is slow and blocks the developer. Parallel orchestration with independent cancellation makes multi-perspective review practical without increasing wall-clock time.

## Source

Inspired by entireio/cli PR #1018 — `entire review` multi-agent support with parallel subprocess orchestration and signal/cancellation handling.

## Context hints

- `cmd/partio/` (CLI command definitions)
- `internal/checkpoint/` (checkpoint metadata attachment)
- `internal/agent/` (agent detection and execution interface)
