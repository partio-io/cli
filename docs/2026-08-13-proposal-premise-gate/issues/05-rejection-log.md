# 05 — Rejected ideas leave a record

**Source PRD**: [../prd.md](../prd.md)
**Blocked by**: [04 — The proposer checks before it files](./04-proposer-checks-before-filing.md)

## What to build

Write down what the proposer threw away, and why.

Slice 4 makes the proposer drop ideas. The run that drops an idea also
advances the cursor that tracks how far it has read each source, so the
source item is never revisited. Without a record, a dropped idea is
gone with no trace, and a bar set too high is indistinguishable from a
source that has gone quiet.

Append each rejection to a log the run commits alongside the cursor
update it already performs. Each entry carries enough to judge the
decision later: what the idea was, where it came from, and the reason
it was dropped — a failed claim with the evidence that killed it, or
simple irrelevance.

This is the only surface that shows whether the gate is catching the
right things, so it is written for reading, not just for auditing.

## User stories covered

PRD stories 12, 13, 14, 34.

## Acceptance criteria

- [x] Each rejected idea appends an entry to a log file in this repository
      — `internal/rejection` defines `LogPath = .minions/rejections.md` and
      `Append`, which creates the log on the first rejection ever recorded and
      appends below it after that. The propose program gains a record step next
      to its existing drop rule.
- [x] An entry records the idea, its source, and the reason it was dropped
      — all three are required. `Append` refuses an entry missing any of them,
      or carrying a reason outside `Reasons`, and returns before it opens the
      log, so a refused entry leaves no half-written block behind.
- [x] A rejection caused by a failed claim records the claim and the evidence
      that contradicted it — a `premise-failed` entry must carry the claim, the
      evidence the claim named, a verdict that rejects, and what the evidence
      actually showed. Confirmed red first: `Render` accepted an empty claim and
      wrote an unreadable entry. A `holds` verdict is refused too, because a
      claim the tree confirmed did not cause the rejection.
- [x] A rejection caused by irrelevance is distinguishable from one caused by a
      failed premise — the difference is structural, not a label. An
      `irrelevant` entry must carry a note and must carry no claim, no verdict
      and no finding; `Append` refuses one that does. The proposer is shown both
      entry formats. Confirmed red first, twice: the first version of the test
      passed on the word "irrelevant" already sitting in the run summary, so it
      was tightened to the exact reason field the parser reads, which the
      program did not yet carry.
- [x] The log is committed by the same run that advances the source cursor, so
      a dropped idea is never lost silently — one `git add` stages the log, the
      cursor and the program files, and the run makes exactly one commit. The
      cursor advance is pinned ahead of that commit. This criterion was already
      satisfied by the tracer-bullet edit; the single-commit and ordering
      assertions are added guards, not new behaviour.
- [x] Entries append rather than overwrite, so history accumulates across runs
      — three successive appends, each asserting the previous log is still a
      byte prefix of the new one, plus the header appearing exactly once and
      entries reading back oldest first. Proved by mutation: swapping
      `os.O_APPEND` for `os.O_TRUNC` fails the test at run 2.
- [x] A run that rejects nothing leaves the log unchanged and still commits
      cleanly — the log is staged unconditionally, so it must exist:
      `git add` on a never-created path fails and takes the cursor commit down
      with it, on every run rather than only quiet ones. Confirmed red first —
      `.minions/rejections.md` did not exist. It now ships with its header, and
      the program tells the agent to commit the cursor as usual when it rejected
      nothing.
- [x] A dry run writes and commits nothing — every write, `git add`,
      `git commit` and `git push` is pinned inside the `## Agents` steps and
      absent from the program body, so the runner performs none of them.
      Verified for real: `minions run .minions/programs/propose.md --dry-run`
      exits 0 and renders `AGENTS: 1` plus both entry formats, with
      `git status --porcelain -uall` and `git rev-parse HEAD` identical either
      side.
- [x] `make test` and `make lint` pass
      — both exit 0. All 18 packages pass with no FAIL lines, and
      `golangci-lint` reports 0 issues. Two issues it did raise on the new
      package were fixed: an unchecked `f.Close` on the log write handle, which
      can lose the entry just written, and a regex that escaped twice.

## Modules touched

`rejection-log`, `propose-program`.

## Test prior art

- Table-driven standard-library tests with `t.TempDir()` for the
  append-and-read behaviour.
- The proposer already commits the source cursor file in the same run;
  that existing commit step is the pattern the log follows rather than
  introducing a second commit.

## Out of scope

- Any user interface over the log. It is a file, read directly.
- Re-proposing something previously rejected. The cursor still moves
  past it; recovering an idea from the log is a manual act.
- Pruning or rotating the log. It accumulates.
- Changing what gets rejected. That behaviour is slice 4's.

## Verification note

Run a dry run first to confirm nothing is written or committed, then a
real run to confirm the entry lands and the commit carries both the log
and the cursor.
