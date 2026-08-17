# 06 — Research rechecks and refreshes

**Source PRD**: [../prd.md](../prd.md)
**Blocked by**: [04 — The proposer checks before it files](./04-proposer-checks-before-filing.md)

## What to build

Make the research stage verify a proposal's premise before it plans
anything, and stop when the premise no longer holds.

Checking once at filing time is worthless by the time a build runs.
Issue #30 was filed on 2026-03-26 and built on 2026-08-10 — four and a
half months in which the codebase moved underneath it. A claim that was
true in March can be false in August, so every stage rechecks against
the tree in front of it.

Research reuses the verification behaviour defined in slice 4. It does
not carry its own idea of what verification means.

When the premise fails, the stage stops: it adds the blocking label the
pipeline already understands, comments the claims it checked and the
evidence that contradicted them, and ends without planning. It does not
close the issue — the operator decides what happens next, and can
overrule by removing the label. Nothing here changes state the operator
has not seen.

When the premise holds, research refreshes the block with the evidence
it just gathered, so the build stage checks against the newest facts
rather than the filing-time ones. A holding verdict is recorded too, so
a green plan carries its evidence.

## User stories covered

PRD stories 5, 7, 8, 9, 10, 11, 25, 26, 27, 29.

## Acceptance criteria

- [x] The research stage verifies the premise before producing any plan
      — a new `premise-checker` agent runs first in `.minions/programs/research.md`.
      The tracer-bullet test pins the verifier reference ahead of all four
      producing agents (`researcher`, `prd-writer`, `slicer`, `publisher`), so
      an agent reordered in front of the check fails the build. Confirmed red
      first: the gate contract did not compile.
- [x] Verification reuses the behaviour from slice 4 rather than restating it
      — the checker is told to apply `.minions/premise-verifier.md` "as
      written". A test fails if the research program carries either
      `VerifierMarker` or `GateMarker`, which is what a pasted copy looks like.
      Green on arrival, so proved by mutation: appending the gate marker to the
      program fails both this test and `TestGatingIsDescribedOnce`.
- [x] A failed premise adds the existing `do-not-build` label
      — `internal/premise/gate.go` holds `BlockingLabel`, and the test scans
      every `--add-label` in the gate's failing path and rejects any label that
      is not it. Proved by mutation: renaming the label to `premise-blocked`
      fails the test.
- [x] A failed premise posts a comment naming the claims checked and the
      evidence found — the comment opens with `GateCommentMarker` and carries,
      per claim, the claim word for word, the evidence it named, the verdict,
      and the excerpt behind it, including for claims that held. Proved by
      mutation: replacing "the evidence that contradicted them" with "the
      reason" fails the test.
- [x] A failed premise stops the stage, and no plan is produced
      — each agent is its own session in a discarded worktree, so there is no
      early return to take. The checker writes `PREMISE_BLOCKED` or
      `PREMISE_OK` to a stable `/tmp` path, and every later agent reads it
      first and stops. The test enumerates the agent headings rather than a
      fixed list, so an agent added later without the check fails: appending a
      `### summarizer` with no gate read produced three failures.
- [x] The stage never closes the issue
      — no `gh issue close` and no `--add-label minion-done` anywhere in the
      research program, and the gate forbids both. Proved by mutation:
      appending a `gh issue close` line fails the test.
- [x] Removing the label lets a subsequent run proceed, so the operator can
      overrule a wrong verdict — the override works because no stage stores a
      verdict: the program never names the label at all, so it cannot skip on
      it, and re-verifies against the tree on every run whatever labels the
      issue carries. Confirmed red first — the checker named `do-not-build`
      itself, which is exactly the read that would let a later edit turn the
      label into a skip condition. The gate also removes the label before
      adding it, because a label already present fires no fresh trigger and a
      repeat block would otherwise be silent.
- [x] A holding premise is refreshed in the issue with the newly gathered
      evidence — the checker rewrites the `## Premise` section with what this
      run read, keeps the marker and every claim, and leaves the rest of the
      body alone. Confirmed red first, and it caught a real defect: the refresh
      called `gh issue edit --body-file "$REFRESHED_BODY"` with nothing setting
      that variable, which at run time expands to the empty string and
      overwrites the issue body with nothing. The test now checks every shell
      variable the checker expands against the ones the program assigns.
- [x] A holding premise records what was verified, so a successful run carries
      its evidence — `$PREMISE` records every claim, its evidence, its verdict
      and the excerpt behind it for a block that holds as well as one that does
      not. Confirmed red first: the excerpt was not recorded on the passing
      path.
- [x] The research program's instructions live where the parser carries them,
      and the shape test stays green — every addition sits in `## Agents` prose,
      the H1 prose, or `## Context`; nothing was added under a `## Steps`
      heading, which the runner drops. `go test ./internal/programshape/` passes.
      Verified for real: `minions run .minions/programs/research.md --dry-run`
      renders `AGENTS: 6` with `premise-checker` first at 3425 chars of
      instructions, and carries both `cli/.minions/premise-verifier.md` (2758
      chars) and `cli/.minions/stage-gate.md` (3617 chars) into every agent.
- [x] `make test` and `make lint` pass
      — both exit 0. All 18 packages pass with no FAIL lines, and
      `golangci-lint` reports 0 issues.

## Modules touched

`stage-gate`, `premise-verifier`, `research-program`, `premise-block`.

## Test prior art

- Slice 4's verifier fixtures supply the known-false and known-true
  premises; this slice reuses them rather than inventing new ones.
- The research program already writes structured comments that a
  downstream parser reads strictly. Treat the comment this slice adds
  with the same care: a format change there has broken a build before.

## Out of scope

- The build stage's own check. That is slice 7, and it repeats this
  behaviour rather than depending on it.
- Proposals that carry no premise block. That is slice 8.
- Changing how research slices a plan, or the format of the slice plan
  comment.
- Closing, labelling or otherwise touching the 534 existing proposals
  ahead of time.

## Verification note

The label the gate applies is `do-not-build`, chosen because the
approval program already treats it as a skip condition. Do not
introduce a new label.

Be careful when re-running: a label already present on an issue does
not fire a fresh trigger, so a re-run needs the label removed and added
again rather than added twice.
