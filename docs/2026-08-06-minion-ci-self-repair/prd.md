# Minion CI Self-Repair

## Problem Statement

The minion pipeline builds pull requests unattended, but the moment one
of its checks goes red the work stops and lands on the operator. That
is the exact toil the pipeline exists to remove, and today it happens
for three separate reasons.

**The repair round declines on its own guardrail.** The dead-code audit
already has a fix round, and it fires correctly. But the fix program
carries a hard rule — never change or remove an exported identifier,
even when proven inert, because an exported contract is a design
decision for a human. On PR #622 the audit proved that a newly-added
`Model` field on the session type is written by nothing and read by
nothing. The fix session read the finding, matched the rule, exported
no patch, and the gate failed the PR. The rule guards a public API this
module does not have: every Go file lives under `internal/` or `cmd/`,
so nothing outside the module can import any of it. The rule therefore
fires on nearly every finding the audit can produce, and the escalation
it triggers is always to a human who has no contract to protect.

**One attempt, ever.** The workflow takes exactly one repair round and
the loop guard actively prevents a second: any run whose head commit
begins with the repair prefix is skipped outright. A fix that lands
one edit short is indistinguishable from a fix that was never tried.

**The largest failure class has no check at all.** There is no
`make test` or `make lint` gate on pull requests. Both run inside the
minion session and inside the audit's apply step, but neither produces
a PR check, so a broken build never turns anything red and never
reaches the repair machinery. The operator finds it by reading the
diff or by merging it.

The net effect is a pipeline that is unattended right up until it isn't,
with no way to tell "the machine gave up" from "the machine never
tried".

## Solution

Give every blocking check the same repair contract, and let it try more
than once.

- **One verdict shape for every gate.** The lint/test gate emits the
  same verdict artifact the audits already emit — a status plus a list
  of findings, each with a location and the reasoning behind it. The
  existing verdict gate, the PR comment surface, the round counter, and
  the patch applier then work on all three checks without knowing which
  one produced the verdict.
- **Repair rounds are counted, not forbidden.** The loop guard stops
  being a hard skip and becomes an accounting question: how many repair
  commits for this check already sit on the PR branch, and is another
  round allowed? The cap is three per check, counted independently, so
  a dead-code repair never consumes a test repair's budget.
- **Repair is deterministic where it must be.** A repair session only
  ever produces a patch. Applying it to the true PR head, proving it
  with the repo's own checks, committing it and pushing it stays
  machine-owned, in one shared implementation rather than shell copied
  into each workflow. A patch that does not apply, or applies but fails
  the checks, pushes nothing and costs one round.
- **The fixer's reach matches the repo.** The exported-identifier rule
  comes out. Nothing in this module is importable from outside it, so
  there is no contract for the rule to protect and no reason for the
  fixer to stop at a capital letter.
- **Failure terminates visibly.** After the third round the PR stays
  red and gets one comment naming the findings that survived and how
  many rounds ran. The pipeline never rebuilds the PR on its own and
  never cycles.
- **Signal for everyone, commits for minions only.** The lint/test gate
  runs on every pull request, so hand-written branches get the same
  red/green signal. Automatic repair commits are pushed only to
  minion-labeled PRs whose head branch lives in this repository.

The e2e-need audit is unchanged. Its findings become `minion-proposal`
issues and never fail a PR, so it needs no repair round; its only red
state is an infrastructure failure, which retries once and then reports.

## User Stories

1. As the operator, I want a minion PR that fails an audit to repair
   itself, so that I am not hand-editing code a machine wrote.
2. As the operator, I want the dead-code fixer to remove an inert
   exported field, so that a finding it has fully proven does not
   escalate to me for a decision there is nothing to decide.
3. As the operator, I want the fixer's reach justified by the repo's
   actual import surface, so that the guardrail reflects real risk
   rather than a naming convention.
4. As the operator, I want a repair that lands one edit short to get
   another attempt, so that a near-miss is not treated as a refusal.
5. As the operator, I want repair attempts capped, so that a confused
   fixer cannot occupy my single self-hosted runner indefinitely.
6. As the operator, I want each check to have its own repair budget, so
   that a stubborn dead-code finding does not exhaust the attempts a
   failing test needed.
7. As the operator, I want `make test` to run on every pull request, so
   that a broken build is visible before review rather than after merge.
8. As the operator, I want `make lint` to run on every pull request, so
   that lint failures surface on the same terms as test failures.
9. As the operator, I want the lint/test gate on my own branches too,
   so that I get the same signal the minions get.
10. As the operator, I want no bot commits pushed to branches I wrote,
    so that my own work stays mine even while minion PRs self-repair.
11. As the operator, I want a failing test the minion itself just wrote
    to be repairable, so that its own mistakes do not become my queue.
12. As the operator, I want my established test suite frozen against the
    fixer, so that a real assertion cannot quietly disappear on the way
    to green.
13. As the operator, I want every repair proven with the repo's own
    checks before it is pushed, so that a repair cannot make the branch
    worse than it found it.
14. As the operator, I want a repair patch that fails to apply to push
    nothing, so that a stale patch never corrupts the branch.
15. As the operator, I want a PR still red after the last round to say
    so plainly, so that I can tell "gave up" from "never ran".
16. As the operator, I want the give-up comment to name the surviving
    findings, so that I can act without opening workflow logs.
17. As the operator, I want the give-up comment to state how many rounds
    ran, so that I know the budget was actually spent.
18. As the operator, I want the pipeline never to rebuild a PR on its
    own, so that a bad issue cannot cycle build-and-fail unattended.
19. As the operator, I want repair commits clearly marked in history, so
    that I can see at a glance which changes a machine made after the
    build.
20. As the operator, I want two failing gates never to push to the same
    branch at once, so that concurrent repairs cannot clobber each other.
21. As the operator, I want a crashed audit session to retry once, so
    that infrastructure noise does not read as a code defect.
22. As the operator, I want a crashed session that fails twice to go
    red, so that a broken runner is never silently treated as a pass.
23. As the operator, I want an unreadable or missing verdict to fail the
    check closed, so that a gate cannot pass by producing nothing.
24. As the operator, I want the e2e-need audit left alone, so that
    coverage gaps keep flowing back as issues instead of blocking PRs.
25. As the operator, I want repair round accounting to be tested code
    rather than inline shell, so that its edge cases are provable.
26. As the operator, I want the apply-prove-push step written once, so
    that a fix to it fixes every gate at the same time.
27. As the operator, I want the lint/test gate to reuse the existing
    verdict and comment machinery, so that a fourth check later inherits
    repair for free.
28. As the operator, I want repair confined to PRs whose head branch is
    in this repository, so that a fork PR fails cleanly instead of
    erroring on a push it cannot make.
29. As the operator, I want the whole change to ship from this
    repository, so that I do not have to cut a minions release and bump
    a version pin to get it live.
30. As the operator, I want PR #622 to go green without me touching it,
    so that the first proof of this system is the case that motivated it.

## Implementation Decisions

**Verdict as the shared contract.** The lint/test gate produces the
same verdict artifact the audits produce: a status field and a findings
list, each finding carrying a location and the reasoning that justifies
it. This is the seam that lets one repair mechanism serve three checks.
The existing verdict gate command is unchanged in its contract — it
still exits non-zero for anything that is not a valid pass, including a
missing or malformed verdict.

**`internal/repairround` (new).** Owns the question "may another repair
round start for this check, and which round is it?" Its input is the
list of commit subjects on the PR branch plus the check name and the
cap; its output is a decision and a round number. This replaces the
inline shell loop guard, which today answers a cruder question with a
prefix match and no counting. Repair commits are distinguishable per
check so the counts do not interfere. Rounds are capped at three per
check.

**`internal/patchapply` (new).** Owns the deterministic half of a
repair round: fetch the PR head, detach a worktree onto it, apply the
exported patch, run the repo's lint and test targets in that worktree,
and on success commit with a check-specific marked subject and push
with the PAT. Any failure along that path is reported as "nothing
pushed" rather than an error that fails the job. The push must use the
PAT explicitly — a push made with the default Actions token does not
trigger the follow-up run the round accounting depends on. Fork heads
are detected and refused before any fetch.

**`internal/checksverdict` (new).** Runs the repo's lint and test
targets and converts failures into the shared verdict shape, one
finding per failing package or linter finding, with enough reasoning
for a repair session to act without re-running anything itself. A clean
run yields a pass verdict.

**`internal/auditgate` (modified).** Gains the rounds-exhausted comment:
when the budget is spent and findings survive, the comment names the
survivors and the number of rounds that ran. It keeps its existing
upsert behavior so a PR accumulates one comment per check rather than
one per round.

**Dead-code fix program (modified).** The exported-identifier hard rule
is removed, along with the reasoning that justified it. The remaining
restraints stay: repair only what a finding's reasoning proves, make no
opportunistic edits, keep each repair coherent, and export nothing when
the checks cannot be made green. The rationale recorded in the program
is that this module exposes no importable API — every package is
internal to it — so no repair the fixer can make breaks an external
contract.

**Checks fix program (new).** Repairs a failing lint/test verdict. It
may modify implementation code freely, and may modify a test only when
that test was added by the pull request under repair; tests that
predate the PR are out of bounds and a failure in one is a surviving
finding. It follows the same session contract as the dead-code fixer:
produce a patch, never commit, never push, always write the skip
marker.

**Workflows.** A new checks workflow runs the lint/test gate on every
pull request. Its repair path is guarded on the minion label and on the
head branch living in this repository, so unlabeled and fork PRs get
the signal without the commits. The two audit workflows adopt the round
counter in place of their skip guard. All three gates share a
concurrency group keyed on the PR number, so at most one repair pushes
to a branch at a time. The e2e workflow gains a single retry around its
audit session, covering a crashed session or an unwritten verdict.

**No minions release.** Every change lands in this repository. The
session capabilities the repair programs rely on — the skip marker and
the audit directory for patch export — already exist in the pinned
minions version, so no tag and no version-pin bump are required.

**Language.** Go, matching the repository. The new work deliberately
moves logic out of inline workflow shell and into packages, continuing
the direction the existing audit-gate command already established with
its thin `main` over a tested internal package.

## Testing Decisions

A good test here asserts external behavior only: what a module returns
for a given input, not how it got there. These modules are unusually
well suited to that, because each one is a pure decision over inputs
that a test can construct — a list of commit subjects, a verdict file,
a directory of command output. Tests must not assert on log lines,
internal helper names, or the order of unexported calls.

Every module named above is tested. Specifically:

- `internal/repairround` — table-driven over commit-subject lists:
  no prior rounds, one, at the cap, over the cap, rounds for a
  different check interleaved, human commits interleaved, an empty
  branch, and subjects that merely resemble a repair marker.
- `internal/patchapply` — driven against temporary git repositories
  created in the test, covering a patch that applies and passes, one
  that does not apply, one that applies but fails the repo's checks,
  and a head that belongs to a fork. The push target is a local bare
  repository, so no network is involved.
- `internal/checksverdict` — over captured command output fixtures:
  clean runs, a failing test, a lint finding, and unparseable output,
  asserting the resulting verdict shape and status.
- `internal/auditgate` — extends the existing tests with the
  rounds-exhausted comment body and the upsert path that replaces a
  prior comment rather than appending.

Prior art is `internal/auditgate/run_test.go`: table-driven, standard
library `testing` only, no external assertion frameworks, `t.TempDir()`
for filesystem isolation and a local `httptest` server standing in for
the GitHub API. The new tests follow that shape exactly, and the
repository's existing convention of one primary concern per file
applies to the new packages.

Two areas are deliberately left without unit tests, named here so the
gap is visible rather than silent. The workflow YAML itself is not
unit-testable — its correctness is established by a live run, with PR
#622 as the first case and a known-correct outcome. The repair prompt
programs are not unit-testable either, being model-driven by nature;
their contract is enforced downstream, since a session that produces a
bad patch fails the apply step's checks and pushes nothing.

## Out of Scope

- **Rebuilding a failed PR.** After the last round the pipeline stops
  and reports. Closing the PR and re-running the build stays a manual
  decision.
- **Merging.** No part of this system merges, approves, or marks a PR
  ready. A green PR is still reviewed by a human.
- **The e2e-need audit's judgment.** Its findings continue to become
  proposal issues and continue never to block a PR. Only its
  infrastructure retry changes.
- **Repair on fork pull requests.** Fork heads cannot be pushed to from
  this workflow; they receive the gate signal and nothing more.
- **The research and planning stages.** Nothing here changes how issues
  are researched, sliced, or approved.
- **Any change to the minions runtime.** Should a repair need a
  capability the pinned version lacks, that is a separate piece of work
  with its own release and version-pin bump.
- **Widening the checks gate beyond lint and test.** Other quality
  signals may earn a gate later; they would inherit the repair contract
  rather than extend this scope.

## Further Notes

The motivating case is PR #622, "Track and display AI model name in
session and checkpoint info". Its dead-code audit produced exactly one
finding: a `Model` field added to the session type with no writer and
no reader, while four sibling `Model` fields on other types all had
proven write paths. The fix round ran and declined on the exported
identifier rule. That PR is intentionally left red until this work
ships, at which point re-running its audit is the acceptance test — the
field should disappear without anyone touching it.

The decision to cap rounds at three rather than repairing until green
is a runner constraint, not a philosophical one. The minion runner is a
single self-hosted machine, and an unbounded repair loop starves every
other job queued behind it.

The exported-identifier rule was not wrong when it was written; it was
written for a general case and this repository is not the general case.
If a package here is ever promoted out of `internal/` and becomes
importable, the rule should come back scoped to that package rather
than to capitalization everywhere.
