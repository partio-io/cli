# 02 — Catch assumed facts in premise checks

**Source PRD**: [../prd.md](../prd.md)
**Blocked by**: None — can start immediately

## What to build

Two rules let the false premise of issue #31 through. Fix both, so
that the current gates catch this class of error at once, before the
review exists.

**Rule 1: the verifier skips facts hidden in a request.** The
verifier's `## When there is no block` section extracts only sentences
that describe today's code ("what the code does today", "how it
behaves"). It tells the checker to leave "what the proposal wants to
build, and its acceptance criteria". #31 asked for "a merge strategy
appropriate for append-only checkpoint data". That phrase sits in a
request, so the checker skipped it. But it asserts a fact about
today's data, and the fact is false: `partio prune` removes
checkpoints, and `partio reset` deletes and recreates the checkpoint
branch.

Change the verifier's extraction rule:

- Add a third kind of claim to extract: a fact that a requested
  behaviour, a design instruction or an acceptance criterion takes for
  granted about today's code or data.
- The checker quotes the phrase that carries the fact and states it
  again as a checkable claim, with evidence. Use #31 as the worked
  example: "appropriate for append-only checkpoint data" → "nothing in
  Partio deletes or rewrites checkpoint data".
- A sweeping claim ("append-only", "nothing deletes", "never changes")
  is settled by a command that lists every path that changes the data,
  and one counterexample fails it.
- The "leave these" rule skips only the behaviour that the proposal
  asks for, not the facts that the behaviour rests on.

Give the ingest prompt the same rule. Today it says "a claim about
Partio's current behavior is a premise; a statement about what Partio
should do next is not". Add that an assumption about today's code or
data inside a "should" statement is still a premise and goes into
`premise`.

**Rule 2: research passes an issue without a premise block.** The
research program's `premise-checker` agent says that an issue with no
`## Premise` section "is out of scope for this program" and writes
`PREMISE_OK`. The August work removed this escape from the
premise-gate program and left it here. Route such an issue to the
verifier's `## When there is no block` extraction, as the premise-gate
program does.

Do not remove the verifier's "The backlog is not swept" and "Do not
rewrite the issue body" lines in this slice. Slice 04 removes them
when the review program arrives.

Program text rules: the minions runtime keeps only the title prose,
`## Context`, the planner section and `## Agents`. It drops a
`## Steps` section and `####` headings. Keep instructions as prose in
the agent's section, with bold labels.

## User stories covered

38, 39, 40, 41, 42, 54.

## Acceptance criteria

- [x] The verifier's no-block rule lists, as a kind of claim to extract, a fact that a requested behaviour, a design instruction or an acceptance criterion takes for granted about today's code or data.
- [x] The verifier tells the checker to quote the phrase that carries such a fact and to state it again as a checkable claim with evidence.
- [x] The verifier carries #31's phrase as its worked example of an assumed fact.
- [x] The verifier tells the checker to settle a sweeping claim with a command that lists every path that changes the data, and says that one counterexample fails it.
- [x] The verifier's "leave these" rule skips the behaviour a proposal asks for, but not the facts that behaviour rests on.
- [x] The ingest prompt's premise rule says that an assumption about today's code or data inside a "should" statement is a premise.
- [x] The research program's premise checker no longer declares an issue without a premise block out of scope, and it extracts claims from the prose through the verifier's no-block section.
- [x] A repo test fails if the research program writes a pass for an issue with no premise block, like the test that already guards the premise-gate program.
- [x] Repo tests pin the assumed-fact rule in the verifier and in the ingest prompt by their contract phrases.
- [x] The program-shape test passes for every changed program.
- [x] `make test` and `make lint` pass.

## Modules touched

- verifier
- ingest prompt
- research program
- premise repo tests

## Test prior art

- `internal/premise/repo_*_test.go` — the verifier, research, build and
  ingest-prompt tests that read program text and pin contract phrases;
  the build test for a proposal with no block is the model for the new
  research guard
- `internal/premise/verifier.go` and `source.go` — the Go side of the
  verifier contract
- `internal/programshape/*_test.go` — the shape test

## Out of scope

- The verifier's "not swept" and "not rewritten" lines (slice 04).
- The review program (slice 04) and the proposer's duplicate check
  (slice 11).
