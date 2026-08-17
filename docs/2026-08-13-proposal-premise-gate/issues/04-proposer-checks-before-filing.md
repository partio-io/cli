# 04 — The proposer checks before it files

**Source PRD**: [../prd.md](../prd.md)
**Blocked by**: [03 — Proposals carry a premise](./03-proposals-carry-premise.md)

## What to build

Make the proposer read this repository before it describes this
repository, and file only ideas whose premise survives.

The source tree is already checked out when the proposer runs — the
workflow checks it out and then never looks at it. The proposer draws
feature ideas from a sibling product's changelog and issue list, and
asserts that product's problems as this product's problems. That is how
a proposal came to describe a tree walk this codebase has never had.

This slice defines verification once and uses it here: given a premise
block and a checked-out tree, gather the evidence each claim names and
produce a verdict plus what was found. Claims that hold let the idea
through. Claims that fail drop the idea before it becomes an issue.

Verification is described as a single behaviour, not copied per
caller — slices 6, 7 and 8 reuse it without re-deciding what
verification means.

The ingest prompt also gains its grounding rule here: a claim must be
grounded in the checked-out tree, not inferred from the source
material. Adaptation from a sibling product stays the goal; importing
its premises does not.

Volume falls as a result of this slice. That is the intended mechanism —
the operator chose to raise the bar at the source rather than cap
output or prune the backlog.

## User stories covered

PRD stories 3, 4, 15, 22, 23.

## Acceptance criteria

- [x] Verification is described once as a single behaviour: premise block plus checked-out tree in, verdict plus evidence out
      — `.minions/premise-verifier.md`, marked `<!-- partio:premise-verifier:v1 -->`,
      states the two inputs, the gather-then-decide procedure, and the closed
      verdict set. `internal/premise/verifier.go` holds the shared contract
      (`VerifierPath`, `VerifierMarker`, `Holds`/`Fails`/`Unresolved`). A test
      walks `.minions/` and fails if a second file carries the marker, so a
      caller cannot paste its own copy.
- [x] The proposer gathers the evidence each claim names before deciding
      — the verifier's procedure covers all three evidence kinds the claim
      grammar admits (path, symbol, command) and sits ahead of the verdict
      section; the proposer is told the tree is already checked out. That last
      assertion was red before the propose-program edit.
- [x] An idea whose premise fails is not filed as an issue
      — the proposer's steps are ordered verify → drop → write program file →
      `gh issue create`, and the drop step states no program file and no issue.
      The test was confirmed red on ordering before the token was made specific.
- [x] An idea whose premise holds is filed, carrying the evidence that was
      gathered — the `gh issue create` body now composes in `gathered evidence`
      alongside the premise section.
- [x] The ingest prompt requires each claim to be grounded in the checked-out
      tree rather than the source material — a grounding rule sits next to the
      existing `premise` rule under `## Instructions`. Confirmed red first: the
      prompt mentioned neither the tree nor the source material.
- [x] The issue #30 premise fixture is verified as false against this
      repository — both claims verify as `fails`. The hook names changed files
      with `git.DiffNameOnly(commitHash)` and contains no `filepath.Walk`,
      `filepath.WalkDir`, `os.ReadDir` or `ls-files`; `attribution.Calculate`
      iterates `git.DiffNumstat` output, which carries one line per changed
      file, not per tracked file. The test re-reads both files on every run, so
      the record cannot rot into a stale comment.
- [x] A claim that genuinely holds for this repository is verified as true, so
      the bar does not reject everything — `testdata/holds.md` claims
      attribution is binary; `internal/attribution/calculate.go` sets
      `result.AgentPercent = 100` under `if agentActive`, so it verifies as
      `holds`. Both fixtures run in one table, so the passing direction cannot
      be dropped without the failing one noticing.
- [x] A source item that is simply irrelevant is still skipped, and is
      distinguishable from one dropped for a false premise — the run summary
      now reports `filed`, `dropped` and `skipped` as three separate counts,
      and a dropped idea names its claim, evidence and verdict. The ingest
      prompt's existing skip rule is pinned so raising the premise bar cannot
      turn "not relevant" into "rejected". Confirmed red on all three counts.
- [x] The agent has the tools it needs to read files and run commands in the
      checked-out tree — `Read`, `Glob`, `Grep` and `Bash` were already
      granted; a test now pins them so a later edit cannot narrow the grant
      silently. `max_turns` was raised from 40 to 80, because verification adds
      several tool calls per idea and the old budget would truncate a run.
- [x] `make test` and `make lint` pass
      — both exit 0; `golangci-lint` reports 0 issues. The slice-1 shape test
      still passes, so the new propose-program prose reaches the model.
      `minions run .minions/programs/propose.md --dry-run` renders `AGENTS: 1`
      and the whole instruction set, including the verify step, the drop rule
      and the three-outcome summary.

## Modules touched

`premise-verifier`, `ingest-prompt`, `propose-program`.

## Test prior art

- Table-driven standard-library tests with `t.TempDir()` to build a
  small repository fixture and assert a verdict against it. This repo
  already exercises real git invocations against temporary
  repositories in its git package tests, which is the closest prior
  art for evidence-gathering that shells out.
- Slice 3's premise-block fixture supplies the known-false case.

## Out of scope

- Recording what was rejected. That is slice 5, and until it lands a
  dropped idea leaves no trace beyond the run output.
- Any gating in the research or build stages. Those are slices 6 and 7.
- Proposals filed before the block format existed. That is slice 8.
- Capping proposals per run or closing any of the 534 existing
  proposals. Both are named in the PRD's Out of Scope section.

## Verification note

Use a dry run against real source content so the verdicts are produced
from genuine material rather than a hand-written premise. Confirm both
directions — something correctly dropped and something correctly
filed — because a bar that rejects everything looks identical to a
working one until you check.
