# 04 — One shared apply-prove-push

**Source PRD**: [../prd.md](../prd.md)
**Blocked by**: [02 — Count repair rounds instead of forbidding them](./02-count-repair-rounds.md)

## What to build

A repair session never commits or pushes. It produces a patch, and the
workflow owns everything after that: fetch the pull request's real
head, detach a worktree onto it, apply the patch, prove the result with
the repository's own lint and test targets, commit it with a marked
subject, and push it. That sequence currently lives as inline shell in
the audit workflow, and a second gate would mean copying it.

Move it into a tested package and switch the dead-code gate over. After
this slice there is one implementation of apply-prove-push, so a defect
found in it is fixed once for every gate rather than once per workflow.

Behavior that must be preserved exactly, because each piece was learned
the hard way:

- A patch that does not apply pushes nothing and reports that, rather
  than failing the job.
- A patch that applies but leaves the checks red pushes nothing.
- The push uses the personal access token explicitly. A push made with
  the default Actions token does not trigger the follow-up run that
  round accounting depends on, so the fix would appear to work and
  then silently stop the loop.
- A pull request whose head branch lives in a fork is detected and
  refused before any fetch is attempted, since the workflow cannot push
  there.
- The commit subject carries the marker slice
  [02](./02-count-repair-rounds.md) counts on, naming which check
  produced the repair.

## User stories covered

PRD user stories 13, 14, 26, 28.

## Acceptance criteria

- [ ] A new package owns fetch, detach, apply, prove, commit, and push
      for a repair patch, given the patch, the pull request reference,
      and the check name.
- [ ] The dead-code audit workflow uses the package instead of its
      inline shell, and its end-to-end repair behavior is unchanged.
- [ ] A patch that fails to apply results in nothing pushed, reported as
      a failed round rather than a job error.
- [ ] A patch that applies but fails `make lint` or `make test` results
      in nothing pushed.
- [ ] The push is made with the personal access token, not the default
      Actions token.
- [ ] A fork head is refused before any fetch, with a clear reason.
- [ ] The commit subject carries the per-check repair marker that slice
      [02](./02-count-repair-rounds.md) counts.
- [ ] Tests drive the package against temporary git repositories created
      by the test, pushing to a local bare repository so no test touches
      the network, and cover: a patch that applies and passes, one that
      does not apply, one that applies but fails the checks, and a fork
      head.

## Modules touched

`internal/patchapply` (new), and the dead-code audit workflow from the
PRD's Implementation Decisions ("Workflows"). Produces commit subjects
that `internal/repairround` reads.

## Test prior art

`internal/git/*_test.go` for tests that build real repositories in
`t.TempDir()` and drive git operations against them. `internal/checkpoint`
tests are the precedent for exercising git plumbing without a remote.

Assert on observable outcomes — was anything pushed, what does the
target repository contain, what does the commit subject say — not on
which git commands ran in what order.

## Out of scope

- The lint/test gate's repair path, which becomes the package's second
  consumer in slice [06](./06-repair-failing-checks.md).
- Changing how a repair session produces its patch. The session contract
  is unchanged: emit a patch, never commit, never push, write the skip
  marker.
- Concurrency between gates pushing to the same branch. Slice
  [07](./07-gate-concurrency-and-retry.md) covers it.
- The e2e-need audit, which produces no patches and never pushes.
