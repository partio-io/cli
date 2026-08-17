# 02 — Propose instructions reach the model

**Source PRD**: [../prd.md](../prd.md)
**Blocked by**: [01 — Program shape check](./01-program-shape-check.md)

## What to build

Make the propose program actually run the workflow it describes.

Today the propose program carries seven numbered steps under a
`## Steps` heading and defines no agents. The parser discards that
section, so the executor synthesizes a single implicit agent whose
whole instruction set is the one sentence under the H1 — "Scan
monitored sources for new content and create feature proposals." — plus
the file paths listed under `## Context`. The steps that tell it to
apply the ingest prompt, to check whether a proposal already exists, to
advance the source cursor, and to commit the result never reach the
model at all. It reads the three context files and invents the rest.

Move those instructions into an `## Agents` section, which the format
does carry, so the program behaves as written. This is a faithful
move, not a redesign: the workflow stays what the steps already
describe. Later slices change what the workflow does.

Once the program is fixed, drop it from the accepted-exceptions list
that slice 1 established.

## User stories covered

PRD stories 19, 24.

## Acceptance criteria

- [x] The propose program defines an `## Agents` section carrying its full workflow
      — one agent, `proposer`.
- [x] Every instruction previously under `## Steps` survives the move, with no step silently lost
      — all seven steps kept verbatim, including every `gh` and `git`
      command; verified by 14 literal probes against the rendered prompt.
- [x] The program no longer contains an instruction-bearing dropped section
- [x] The propose program is removed from slice 1's accepted-exceptions list, and the shape test still passes
- [x] A dry run of the program renders a prompt that visibly contains the workflow instructions
- [x] The rendered prompt is materially longer than before the change, confirming the instructions now reach the model
      — `AGENTS: 0` → `AGENTS: 1`; total prompt 5307 → 7587 chars; the
      agent instruction component 68 → 1905 chars, a 28× increase.
- [x] The agent is granted the tools the workflow needs, including reading files and running shell commands
      — Read, Write, Edit, Glob, Grep, Bash, WebFetch; the engine joins
      the list into the session's `--allowedTools`.
- [x] No change is made to the minions engine, so no release, tag or version pin bump is required
      — the engine repository is clean; it was read, never edited.
- [x] `make test` and `make lint` pass

## Modules touched

`propose-program`, `program-shape-test`.

## Test prior art

- Slice 1's shape test is the automated guard for this change; it
  should go green for the propose program without an exception entry.
- Programs already written correctly in this repository — the research,
  implement, checks-fix and audit programs — are the reference for how
  an `## Agents` section is laid out, what an agent definition carries,
  and how tools and turn limits are declared.

## Out of scope

- Changing what the proposer does. Premise emission is slice 3,
  verification is slice 4, and the rejection log is slice 5.
- The ingest prompt's content. It is a template read as context, not a
  parsed program, so it is untouched here.
- The approval, ingest, documentation-update and readme-update
  programs. They share the defect and stay on the exceptions list.

## Verification note

A dry run renders the prompt without creating issues or pushing. Use it
rather than waiting for the twice-daily schedule, and read the rendered
prompt directly — the character count alone is the signal that
instructions are now included.
