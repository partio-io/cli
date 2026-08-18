# 08 — Older proposals are not exempt

**Source PRD**: [../prd.md](../prd.md)
**Blocked by**: [04 — The proposer checks before it files](./04-proposer-checks-before-filing.md)

## What to build

Apply the gate to proposals that were filed before the premise block
existed.

There are 534 open proposals in this repository and 56 of them already
carry the approved label, meaning they will build when something
touches them. None has a premise block. Without this slice the gate
protects only new work, and every existing proposal keeps its exemption
until someone rewrites it by hand.

When a stage meets a proposal that has no premise block, it extracts
the factual claims from the issue prose and verifies those instead,
with the same stop behaviour on failure. The proposal is not rewritten
and not backfilled — extraction happens at check time, so nothing needs
a bulk edit.

The operator's decision was explicit: the backlog is not swept ahead of
time and not pruned. Each proposal is checked when a stage next touches
it, which means the 56 approved ones are covered without any special
pass over them.

## User stories covered

PRD stories 16, 17, 18.

## Acceptance criteria

- [x] A proposal with no premise block has its factual claims extracted from the
      issue prose — `premise.ClaimSource` routes a body with no
      `<!-- partio:premise:v1 -->` marker to `SourceProse`, and the verifier's
      `## When there is no block` section says what to extract (what the code
      does today) and what to leave (wishes, acceptance criteria, the sibling
      product). Both were red first: the selector did not exist, and the
      verifier had no such section.
- [x] Extracted claims are verified by the same behaviour as block-carried
      claims — the verifier's procedure now reads "For each claim, in order,
      whether it came from a block or from the prose", and the extraction
      section routes into it rather than restating it. A test counts the
      gathering description and fails at anything but one. In Go, the fixture
      table gathers and decides in one loop for both inputs. Red first on two
      counts: the procedure said "in the block", and the section claimed no
      sameness.
- [x] A failed extracted claim triggers the same stop, label and comment as a
      failed block claim — the verifier's caller section now says "Only a
      premise that `holds`" instead of "Only a block", and the gate's comment
      covers "every claim you checked ... whether it came from a block or was
      extracted from the issue prose". The extraction section sits above the
      caller section, which is what makes its "nothing below this section
      changes" true; a test pins that order. Both were red.
- [x] The proposal body is not rewritten and no premise block is backfilled into
      it — the extraction section states it, and the gate's passing path now
      refreshes the block "only if the issue carries one", with an explicit "an
      issue with no block gets none". That second edit is in `stage-gate.md`,
      outside this slice's modules; see the note below.
- [x] A proposal that carries a block uses the block and does not fall back to
      prose extraction — `premise.ClaimSource` returns `SourceBlock` whenever
      `Find` locates the marker, so precedence is one decision, not a rule each
      stage repeats. A body carrying issue #30's prose *and* a block yields the
      block's claim only. This passed on first run, since the tracer bullet's
      selector already keyed on the marker, so a mutation check confirmed the
      assertions bite: forcing `ClaimSource` to always answer `prose` fails both
      tests.
- [x] A proposal whose prose contains no checkable factual claim proceeds rather
      than stopping, and says so in the run output — the verifier states that
      this is not `unresolved` ("an unresolved claim is one you could not
      settle; here there was no claim to settle") and to say so in the report.
      An empty *block* stays an error, so the two cases remain distinct. Red on
      all four phrases.
- [x] Issue #30's original body is used as a fixture, since its prose carries a
      claim already proven false — `testdata/issue-30-body.md` is the body as
      filed, fetched with `gh issue view 30`, machine block and all. It routes
      to `SourceProse`, and its prose claims verify as `fails` against today's
      tree on the same evidence the block fixture uses. The test re-reads the
      quoted prose from the fixture, so an extracted claim cannot drift into one
      the proposal never made.
- [x] No pass is made over the 534 existing proposals, and none is closed or
      relabelled by this work — the verifier says the backlog is not swept and a
      proposal is checked when a stage next touches it. A test fails if either
      description reaches for `gh issue list` or `gh issue close`, or if a
      `gh issue edit` line names anything but `$MINION_ISSUE_NUMBER`. No issue
      was listed, closed or relabelled while building this slice; the only `gh`
      call made was a read of issue #30.
- [x] `make test` and `make lint` pass — 18 packages ok with no failures, and
      `golangci-lint` reports 0 issues. `minions run
      .minions/programs/research.md --dry-run` still renders, and names
      `premise-verifier.md` and `stage-gate.md`, so the edited documents still
      reach the model.

## Modules touched

`premise-verifier`, `premise-block`.

`stage-gate` was also edited, with the operator's approval during the run. Two
of its lines were written for a proposal that carries a block: the blocked-stage
comment reported "every claim in the block", and the passing path refreshed the
premise block unconditionally. On an old proposal the first leaves an extracted
claim unreported, and the second reads as an instruction to write a block into
an issue that never had one — the opposite of this slice's fourth criterion.
Both are one-line conditionals now.

## Test prior art

- Slice 4's verifier fixtures. This slice adds a prose-input case
  beside the existing block-input cases rather than building a separate
  test surface.
- Real issue bodies from this repository's open proposals make better
  fixtures than invented prose, because the failure mode being guarded
  against is specifically how these were written.

## Out of scope

- Sweeping, closing, relabelling or pruning any existing proposal. The
  PRD's Out of Scope section names all of these.
- Backfilling premise blocks onto old proposals.
- Deciding whether an old proposal is still worth doing. The gate rules
  on whether its premise holds, not on whether the idea has value.
- Ranking or scoring the backlog.

## Verification note

A proposal with no checkable claim must proceed rather than stop.
Treating "I found nothing to check" as a failure would block most of
the backlog at once, which is the opposite of the intent.
