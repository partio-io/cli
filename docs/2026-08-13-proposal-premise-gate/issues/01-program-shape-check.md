# 01 — Program shape check

**Source PRD**: [../prd.md](../prd.md)
**Blocked by**: None — can start immediately

## What to build

A test that fails when a minion program hides its instructions in a
section the program format discards.

Background you need, because it is not obvious from the files: the
minions program parser keeps only four things from a program's Markdown
body — the H1 heading and the prose directly under it, a `## Context`
section, a `## Planner` section, and a `## Agents` section with its H3
sub-sections. Every other heading and its body are dropped silently. If
a program defines no agents, the executor synthesizes one whose entire
instruction set is the prose under the H1.

The practical effect is that a program written with a `## Steps`
section runs on one or two sentences and improvises the rest. Four
programs in this repository are in that state today.

The test reads every program in the repository's minions programs
directory and fails any that carries instruction-bearing content in a
dropped section. Programs already known to be broken are recorded as
accepted so the suite is green on arrival; the point is to stop new
ones appearing and to give later slices a checklist to work down.

When it fails it must name the offending program and the heading that
was dropped, so the fix needs no bisecting of format rules.

## User stories covered

PRD stories 20, 21, 30.

## Acceptance criteria

- [x] A Go test reads every `.md` program in the minions programs directory of this repo
- [x] The test fails a program that has instruction content under a heading the parser drops and no `## Agents` section
- [x] The test passes a program that carries its instructions in an `## Agents` section
- [x] The test passes a program whose entire instruction set is the prose under the H1, since that text does reach the model
- [x] A failure message names the offending program file and the dropped heading
- [x] The four currently-broken programs are recorded as accepted exceptions in one obvious place, each with a one-line reason
      — the scan found **eight**, not four: the five `## Steps` programs (the
      four named in the PRD's Out of Scope, plus `propose.md`) and three
      `e2e-*` feature programs the PRD never counted. All eight are in
      `knownUnreachable` with a one-line reason.
- [x] Adding a new program with a `## Steps` section and no `## Agents` section turns the suite red
- [x] `make test` and `make lint` pass

## Modules touched

`program-shape-test`.

## Test prior art

- Table-driven tests using only the standard library `testing` package,
  with `t.TempDir()` for filesystem isolation — the house pattern
  described in this repo's `CLAUDE.md` and used throughout `internal/`.
- The minions engine repository has parser tests covering section
  handling in its program package; they are the closest model for how
  to structure cases around heading levels.

## Out of scope

- Fixing any of the four broken programs. Slice 2 fixes the propose
  program; the other three are named in the PRD's Out of Scope section
  and left for separate work.
- Changing the minions engine or its parser. This test encodes the
  parser's existing behaviour; it does not alter it.
- Validating anything about a program's content beyond where its
  instructions live.
