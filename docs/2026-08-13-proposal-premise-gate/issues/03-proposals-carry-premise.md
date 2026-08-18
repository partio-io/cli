# 03 — Proposals carry a premise

**Source PRD**: [../prd.md](../prd.md)
**Blocked by**: [02 — Propose instructions reach the model](./02-propose-instructions-reach-model.md)

## What to build

Give every filed proposal a machine-checkable statement of what it
believes about this repository.

A proposal today states its case in prose. Issue #30 opened with
"Partio's post-commit hook currently performs a full tree walk… O(N) in
the number of tracked files". That sentence is false here, and nothing
downstream could tell, because prose cannot be rechecked.

Add a premise section to every proposal the program files. It lists the
factual claims the proposal depends on, one per line, and pairs each
with the evidence that settles it — a path, a symbol, or a command
whose output decides the matter. A later stage needs nothing but this
block in order to recheck; it does not have to parse the surrounding
prose.

The ingest prompt's output schema grows the matching fields, so the
model producing feature ideas produces claims and evidence alongside
them rather than having them bolted on afterwards.

This slice establishes the format and emits it. Verifying the claims
is slice 4.

## User stories covered

PRD stories 1, 2, 26.

## Acceptance criteria

- [x] The ingest prompt's output schema carries the premise claims and the evidence for each
      — `.minions/ingest-prompt.md` grows a `premise` array of
      `{claim, evidence}`; the test unmarshals the schema and asserts both
      fields.
- [x] Every proposal the program files contains a premise section
      — the `gh issue create` body in `propose.md` now composes the premise
      section in; the guard test was confirmed red without that edit.
- [x] Each claim occupies one line and names the path or command that proves it
      — the claim line is ``- <statement> [evidence: `<path or command>`]``;
      a claim continued onto a second line, and prose inside the section,
      are both rejected.
- [x] A claim with no evidence attached is rejected rather than filed with an empty field
      — `Render` returns `ErrNoEvidence` and emits nothing; `Parse` returns
      `ErrNoEvidence` for a missing, empty or blank evidence field.
- [x] The section is placed and headed consistently, so a later stage can find it without guessing
      — `## Premise` plus the `<!-- partio:premise:v1 -->` marker; `Find`
      locates it anywhere in a body, stops at the next `##` heading, and
      refuses a hand-written heading that carries no marker.
- [x] A dry run shows the premise section in the issue body the program would create
      — `minions run .minions/programs/propose.md --dry-run`, run against the
      pinned `v0.0.13` that CI installs, renders `AGENTS: 1`, the premise
      template at lines 49-53, and the `gh issue create` body that composes
      the section in.
- [x] The premise of issue #30 is captured as a fixture, since it is a known-false claim with known evidence
      — `internal/premise/testdata/issue-30.md` holds its two claims; the
      test parses them and confirms each evidence path exists in this
      repository. Verifying the claims stays with slice 4.
- [x] Existing proposal behaviour is otherwise unchanged — same title, body and labels
      — a guard test pins the repo, label, title, description, program-file
      link and program marker on the `gh issue create` line, plus the
      duplicate check and the program-file write. It passed on arrival: the
      premise section is an addition, and nothing was removed.
- [x] `make test` and `make lint` pass
      — both exit 0; `golangci-lint` reports 0 issues.

## Modules touched

`premise-block`, `ingest-prompt`, `propose-program`.

## Test prior art

- Table-driven standard-library tests with `t.TempDir()`, the house
  pattern in this repository.
- Slice 1's shape test is the model for a test that reads repository
  files and asserts something about their structure; a premise-block
  test is the same shape applied to issue bodies rather than programs.

## Out of scope

- Verifying that a claim is true. That is slice 4.
- Backfilling a premise onto the 534 existing proposals. Slice 8
  handles pre-block proposals by reading their prose instead.
- Stopping or gating anything. No slice before 6 changes what happens
  when a premise is wrong.

## Verification note

Use a dry run to inspect the issue body the program would file. Do not
wait for the twice-daily schedule, and do not create real issues to
test the format.
