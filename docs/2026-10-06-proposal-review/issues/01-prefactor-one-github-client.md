# 01 — Prefactor: one GitHub client for the minion tools

**Source PRD**: [../prd.md](../prd.md)
**Blocked by**: None — can start immediately

## What to build

This slice adds no behavior. It reshapes current code so that the
review gate (slices 03 onward) is a small change.

Today three packages carry their own copy of the same GitHub REST
helper, `doGitHub(cfg Config, req *http.Request, out any) error`:

- the audit gate package (it also has `upsertComment` and `findComment`:
  find a comment on an issue or pull request by body prefix, with
  pagination, then update it in place or create it)
- the patch-apply package (reads a pull request head)
- the repair-round package (lists pull request commits)

The three copies are identical: they set `Accept:
application/vnd.github+json` and `Authorization: Bearer <token>`, use
the configured `*http.Client` or `http.DefaultClient`, reject a non-2xx
response with `github: <method> <path>: <status>: <up to 512 bytes of
body>`, and decode JSON into `out` when `out` is not nil.

Extract one shared GitHub client package that the minion tools use:

- one request helper with exactly that behavior and error text
- a helper for paginated GET requests (`per_page=100`, stop on a short
  page), as the audit gate's comment search does today
- comment upsert by body prefix on an issue or pull request: page
  through the comments, update the first comment whose body starts with
  the prefix, or create one. It never matches by author: minion
  comments post with a human token, so the author tells nothing.

Then move the audit gate, patch-apply and repair-round packages onto
it, and delete their private copies. Requests, error texts and comment
matching stay exactly as they are. Later slices add issue endpoints to
this client (labels, issue edits and closes, timeline, search).

## User stories covered

48, 56 (indirectly: every GitHub write of the review goes through one
tested client, and the change stays in this repository).

## Acceptance criteria

- [x] The full suite is green before the first edit and after the last: `make test` passes and `make lint` reports 0 issues.
- [x] One shared package holds the GitHub request helper. It sets the Accept and Authorization headers, rejects a non-2xx response with method, path, status and up to 512 bytes of body, and decodes JSON.
- [x] The shared package has a paginated GET helper with `per_page=100` that stops on a short page.
- [x] The shared package has comment upsert by body prefix on an issue or pull request. It pages through comments, updates a match in place, creates a comment otherwise, and never matches by author.
- [x] The audit gate, patch-apply and repair-round packages use the shared client, and none of them keeps its own request helper.
- [x] The current tests of the three packages pass with no change to their assertions.
- [x] The shared package has table-driven tests against an `httptest` server: headers, the non-2xx error text, pagination, and upsert that creates and upsert that updates.
- [x] The commands that wrap these packages keep their flags and their environment variables (`GITHUB_REPOSITORY`, `GH_TOKEN`, `GITHUB_API_URL`).

## Modules touched

- shared GitHub client (new)
- audit gate
- patch apply
- repair round

## Test prior art

- `internal/auditgate/*_test.go` — `fakeGitHub` over `httptest`, with
  comment upsert cases
- `internal/repairround/*_test.go` and `internal/patchapply/*_test.go`
  — their own fake servers
- House style (repo `CLAUDE.md`): one primary concern per file,
  table-driven tests, standard library `testing` only

## Out of scope

- Issue endpoints: labels, issue edit and close, timeline and search.
  Slices 03, 05, 06, 08 and 10 add the ones they need.
- Any change to the audit gate's comment texts or verdict rules.
