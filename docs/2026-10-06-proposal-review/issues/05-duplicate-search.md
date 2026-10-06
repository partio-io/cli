# 05 — Duplicate search

**Source PRD**: [../prd.md](../prd.md)
**Blocked by**: [04 — Review program and dry-run dispatch](./04-review-program-and-dry-run-dispatch.md)

## What to build

One deterministic way to find the same idea in the backlog, used by
the review now and by the proposer in slice 11.

**The search** (`review` package, `dupes` subcommand of
`minion-review`). Input: one or more source references, a title, and
optionally the issue number under review, which the search excludes.
The search lists every `minion-proposal` issue, open and closed, with
pagination. It does not use GitHub's search index. It returns the
candidates that match by either reason:

- **Same source item.** Extract the source items from each body:
  `owner/repo#N` shorthand, GitHub issue and pull request URLs,
  `Origin:` lines, and the `source:` line of the legacy
  `<!-- minion-task` block. Normalize each to `owner/repo#N`. Most
  proposals cite entireio/cli.
- **Strong title overlap.** Normalize titles (lower case, no
  punctuation, no stop words) and match on a high token overlap.

Each candidate carries its number, title, state, state reason, labels
and match reason. Source matches come before title matches. The
command prints the candidates as JSON:
`minion-review dupes --source <ref> [--source <ref>…] --title <title>
[--exclude N]`.

**The review program** calls the command with the issue's source
references and title, decides for each candidate whether it is the
same idea, and writes every candidate into the verdict's `duplicates`
list with `same` and `why`. It applies the duplicate rules (the first
issue reviewed stays):

- An open candidate that already carries `minion-reviewed` stays, and
  this issue closes as its duplicate (`duplicate_of`).
- An open candidate without `minion-reviewed` lets this issue stay. The
  other one closes when its own review runs.
- A candidate that the review closed (state reason not planned, label
  `minion-reviewed`) makes this issue its duplicate, because the idea
  already has a verdict.
- A candidate closed as completed makes this issue built only when the
  tree shows the work. A build can close an issue while its pull
  request never merges (#31 is such a case), so the closed state alone
  proves nothing.

## User stories covered

3, 7, 44 (search side), 53.

## Acceptance criteria

- [ ] The search lists every `minion-proposal` issue, open and closed, with pagination, through the shared GitHub client.
- [ ] It extracts source items in each form (shorthand, issue URL, pull request URL, `Origin:` line, legacy `source:` line) and normalizes them to `owner/repo#N`.
- [ ] A candidate with the same source item matches, whether it is open or closed.
- [ ] A candidate with a strongly overlapping normalized title matches, and an unrelated idea returns no candidate.
- [ ] The issue under review never appears among its own candidates.
- [ ] Each candidate carries number, title, state, state reason, labels and match reason, with source matches first.
- [ ] Test fixtures include at least one confirmed duplicate pair from the backlog (#710/#736, #714/#720 or #681/#688), and the pair matches.
- [ ] `minion-review dupes` prints the candidates as JSON.
- [ ] The review program calls the command and writes every candidate it considered into the verdict, with `same` and `why`.
- [ ] The review program applies the duplicate rules above and names the issue that stays in a duplicate close.
- [ ] Search tests run against a fake GitHub server, and a repo test pins that the review program calls the command.
- [ ] `make test` and `make lint` pass.

## Modules touched

- `review` package
- `minion-review` command
- review program
- premise and program-shape repo tests

## Test prior art

- `internal/auditgate/*_test.go` — fake GitHub over `httptest`, with
  paginated listing
- `internal/premise/source.go` and its tests — reads an issue body and
  classifies it; a model for the body parsers here
- `internal/premise/repo_*_test.go` — tests that pin what a program
  tells the model

## Out of scope

- The proposer's use of the command (slice 11).
- The gate's action on a duplicate close (slice 06).
