# 02 — Dead-code audit, read-only

**Source PRD**: [../prd.md](../prd.md)
**Blocked by**: None — can start immediately

## What to build

The complete trigger-to-verdict path for the dead-code audit, without
the repair loop yet: a GitHub Action that fires on minion-labeled PRs,
runs a minions prompt program on the self-hosted minion runner, and
turns the program's verdict into the PR check's green/red plus at most
one comment.

"Semantically dead code" means code that is still invoked but provably
inert — the canonical exemplar (from the PR #586 review) is a callee
whose first branch short-circuits on a parameter that every remaining
call site hardcodes, leaving the rest of the block unreachable. No
deterministic tool flags this; the program's job is to reason about
the PR diff and catch it.

Shared contract established here (index lists the agreed names): the
program's final act is writing a verdict file — status `pass` or
`fail` plus a findings list with location and reasoning per finding —
and a deterministic workflow step maps it to the check outcome.
Missing or malformed verdict fails the run (fail-closed).

## User stories covered

1, 4, 5, 6, 7, 15, 17, 18, 19, 20, 21, 22.

## Acceptance criteria

- [x] Workflow triggers on pull-request opened/labeled/synchronize,
      proceeds only when the PR carries the `minion` label, and skips
      when the `no-audit` label is present.
- [x] A guard step exits early when the PR head commit subject starts
      with the `audit:` prefix (loop guard — consumed for real in
      issue 03), and concurrent runs on the same PR cancel superseded
      ones.
- [x] The audit session reviews the PR diff (fetching surrounding
      context as it judges necessary), and writes the verdict file as
      its final act.
- [x] Gate step behavior, proven with fixture verdicts: `pass` →
      green; `fail` → red; file absent or malformed → red.
- [x] On red, exactly one PR comment prefixed `Minion audit —`, each
      finding phrased with location and reasoning; re-runs update that
      comment in place rather than adding another.
- [x] On pass, no comment at all.
- [x] `workflow_dispatch` dry-run renders the prompt (minions
      dry-run path) and touches neither the PR nor the repo.
- [x] Workflow installs minions at the same version pin as the
      sibling workflows and runs on the same self-hosted runner.
- [x] Staged proof, then cleanup: a scratch PR with planted
      semantically-dead code goes red with the comment; a clean
      scratch PR stays green and silent; an unlabeled PR triggers no
      audit; a `no-audit`-labeled one skips.

## Modules touched

- deadcode-audit program
- deadcode-audit workflow
- audit verdict gate

## Test prior art

- Workflow shape: the minion/propose workflows under
  `.github/workflows/` (runner label, minions install step, PAT env,
  dry-run input on propose).
- Program shape: existing prompt programs under `.minions/programs/`
  (frontmatter with id and target repos, numbered steps, gh commands
  inline).
- Staged verification: the per-slice staged smoke test run against the
  slice pipeline (scratch issue/PR, verify, clean up).

## Out of scope

- Fixing anything the audit finds — issue 03 adds the bounded repair
  round.
- The e2e-need audit — issue 04 (it reuses this slice's verdict gate
  and trigger gating).
