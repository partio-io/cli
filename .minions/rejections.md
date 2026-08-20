# Rejections

Ideas the proposer did not file, and why. Each run appends; nothing here
is ever rewritten. The cursor has already moved past every item below, so
this file is the only record that the idea was seen at all.

## Protect against duplicate checkpoints when post-commit is killed mid-execution

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2035 (Checkpoint duplicated when a session-end hook is killed between the v1 write and the state save)`
- reason: `premise-failed`
- claim: Partio's post-commit hook can write a duplicate checkpoint for the same commit if the hook is killed between writing the checkpoint and saving the commit cache (`internal/hooks/postcommit.go:201–216`) [evidence: `internal/hooks/postcommit.go`]
- verdict: `fails`
- found: The pre-commit state file is deleted at line 38 of `internal/hooks/postcommit.go` — before any checkpoint is written (line 201). If the hook is killed after writing the checkpoint, the state file is already gone. Any subsequent re-invocation (e.g. from `git commit --amend`) finds no state file and returns nil at line 35 — no duplicate. The amend triggers pre-commit anew, which creates a new state file and a new commit hash; post-commit then writes a checkpoint for that new hash, which is a different commit, not a duplicate. Partio's design already prevents the duplicate-checkpoint scenario that Entire's issue described.

## Cursor: Cursor transcript <timestamp> wrapper leaking into prompt fields

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2065 (Cursor transcript-derived prompts leak <timestamp> wrapper tags)`
- reason: `irrelevant`
- note: Partio has no Cursor agent support. The only agents registered are claude-code and codex; Cursor integration does not exist in this codebase.

## Codex last_assistant_message discarded on Stop/SubagentStop

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2064 (Codex last_assistant_message is parsed then discarded on Stop/SubagentStop)`
- reason: `irrelevant`
- note: Partio's Codex session parser (`internal/agent/codex/parse_jsonl.go`) has no `stopRaw` or `subagentStopRaw` event types and does not track turn-end summaries. The bug described is specific to Entire's Codex event pipeline, which Partio does not replicate.

## Subagent transcripts skip image externalization before size cap

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2063 (Subagent transcripts skip image externalization before size cap / redaction)`
- reason: `irrelevant`
- note: Partio has no subagent transcript pipeline and no image externalization. The session capture writes a single JSONL file per commit; there is no multi-agent or image-embedding path.

## Claude Code parallel Task subagents misattribute TodoWrite incremental checkpoints

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2062 (Claude Code parallel Task subagents misattribute TodoWrite incremental checkpoints)`
- reason: `irrelevant`
- note: Partio writes one checkpoint per git commit. It has no incremental TodoWrite checkpoints and no parallel subagent tracking. This class of attribution bug does not exist in Partio's architecture.

## Cursor subagentStop has no subagent_id; correlation key goes empty

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2061 (Cursor subagentStop has no subagent_id; Entire correlation key goes empty)`
- reason: `irrelevant`
- note: Partio has no Cursor support and no subagent correlation logic.

## Codex input_image base64 silently destroyed by JSONL redaction

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2060 (Codex input_image base64 silently destroyed by JSONL redaction)`
- reason: `irrelevant`
- note: Partio's redaction (`internal/redact/`) operates on text tokens and entropy-based heuristics; it has no special handling for Codex image payloads and no image types in its Codex parser. The redaction path described in the issue is Entire-specific.

## Subagent data does not survive condensation

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2058 (Subagent data does not survive condensation — durable storage for subagent transcripts)`
- reason: `irrelevant`
- note: Partio has no subagent tracking, no condensation pipeline, and no durable subagent transcript storage. The feature request is entirely specific to Entire's architecture.

## --checkpoint-remote supports only the github provider

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2033 (--checkpoint-remote supports only the github provider — no self-hosted/generic git remote option)`
- reason: `irrelevant`
- note: Partio stores checkpoints on a local orphan branch (`partio/checkpoints/v1`) using git plumbing commands. There is no configurable checkpoint remote provider; the remote is whatever `origin` the repo already has.

## Codex sessions not captured in linked worktrees (hooks written to worktree-local .codex/hooks.json)

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2020 (Codex sessions are not captured in linked worktrees because hooks are installed in an ignored worktree-local location)`
- reason: `irrelevant`
- note: Partio uses git hooks (pre-commit, post-commit, pre-push) to capture Codex sessions, not Codex's own `.codex/hooks.json` mechanism. Git hooks are installed to `git rev-parse --git-common-dir`, which is shared across worktrees. Partio does not write to `.codex/hooks.json` at all.

## OpenCode Desktop app: hooks never fire (plugin uses Bun globals, desktop runs Node)

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2014 (OpenCode hooks never fire in the OpenCode Desktop app)`
- reason: `irrelevant`
- note: Partio has no OpenCode support. There is no OpenCode agent, plugin, or hook configuration in this codebase.

## git-refs push queue can drop a newer same-ref update during an in-flight push

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2008 (git-refs push queue can drop a newer same-ref update during an in-flight push)`
- reason: `irrelevant`
- note: Partio's pre-push hook pushes the checkpoint branch synchronously via a `git push` call. There is no asynchronous push queue; the race described does not apply.

## Replace obsolete `entire configure --agent` guidance

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2007 (Replace obsolete entire configure --agent guidance left by #1062)`
- reason: `irrelevant`
- note: Partio uses the `PARTIO_AGENT` environment variable and `agent` config field, not a dedicated `agent add/configure` command. The stale docs issue is specific to Entire's CLI command structure.

## OpenCode: preserve user prompts when message events arrive in sequence

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2001 (OpenCode: Preserve user prompts when message events arrive in sequence)`
- reason: `irrelevant`
- note: Partio has no OpenCode support.

## Discover untracked OpenCode sessions during session attach

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #1992 (discover untracked OpenCode sessions during session attach)`
- reason: `irrelevant`
- note: Partio has no OpenCode support and no `session attach` command.

## Warn when a coding agent commits in an enabled repo but no session is being recorded

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #1965 (warn when a coding agent commits in an enabled repo but no session is being recorded)`
- reason: `irrelevant`
- note: Already filed for Partio as issues #650 and #664. Not re-filed to avoid duplicates.

## Spam link

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #1962 (re-create-gamegambling.app)`
- reason: `irrelevant`
- note: Not a feature or bug report; link to a gambling site.

## `entire agent add/list` does not discover external agents

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #1928 (entire agent add/list does not discover external agents)`
- reason: `irrelevant`
- note: Partio has no `partio agent` subcommand. Agent selection is via `PARTIO_AGENT` env var or config field. The issue is specific to Entire's agent management UX.

## agent remove claude-code deletes unrelated .claude/settings.json keys

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #1924 (agent remove claude-code deletes unrelated .claude/settings.json keys)`
- reason: `irrelevant`
- note: Partio has no `partio agent remove` command. Hook installation and removal is handled by `partio enable` and `partio disable`.

## Document or support fallback when semantic search is unavailable in a repository region

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #1923 (Document or support fallback when semantic search is unavailable in a repository region)`
- reason: `irrelevant`
- note: Partio has no search feature. Checkpoints are stored on a local orphan branch with no cloud search backend.

## Concurrent git-refs checkpoint writes can silently lose completed data

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #1917 (Concurrent git-refs checkpoint writes can silently lose completed data)`
- reason: `irrelevant`
- note: Partio writes checkpoints from git hooks, which git serialises per-commit. Concurrent writes to the orphan branch do not occur in normal usage; the race condition described is specific to Entire's async checkpoint pipeline.

## Integrate Muse Code as a supported agent

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #1912 (Integrate Muse Code as a supported agent)`
- reason: `irrelevant`
- note: Partio supports claude-code and codex. Adding Muse Code would require a new agent detector and session parser. This is a new integration, not a feature inspired by the source material; no Muse Code session format is documented in this codebase to base a premise on.

## Docs: phone PII redaction is NANP-only but documented without locale qualifier

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #1909 (Docs: phone PII redaction is NANP-only in practice but documented with no locale qualifier)`
- reason: `irrelevant`
- note: Partio's redaction (`internal/redact/`) uses entropy-based heuristics, not named PII patterns. There is no phone number pattern in Partio's redaction configuration.

## Repo relocation silently loses commit trailers (findSessionsForWorktree has no fallback)

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #1890 (Repo relocation silently loses commit trailers on 0.9.0)`
- reason: `irrelevant`
- note: Already filed for Partio as issues #606 and #615. Not re-filed to avoid duplicates.

## dispatch returns 404 for ready EU mirror while regional recap sees checkpoints

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #1889 (dispatch returns 404 for ready EU mirror while regional recap sees checkpoints)`
- reason: `irrelevant`
- note: Partio has no dispatch command or cloud mirror infrastructure. Checkpoints are local.

## UserPromptSubmit hook blocks ~30s and times out on most prompts (per-session flock contention)

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #1887 (UserPromptSubmit hook blocks ~30s and times out on most prompts)`
- reason: `irrelevant`
- note: Partio has no UserPromptSubmit hook. Its hooks (pre-commit, post-commit, pre-push) are standard git hooks without per-session file locking.

## Checkpoint metadata.json exceeds GitHub's 100 MB limit (files_touched never deduplicated)

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #1927 (Checkpoint metadata.json exceeds GitHub's 100 MB limit: files_touched is never deduplicated, and nested git worktrees are walked)`
- reason: `irrelevant`
- note: Partio's `checkpoint.Metadata` type (`internal/checkpoint/checkpoint.go`) has no `files_touched` field. The checkpoint branch stores session content (JSONL, diff, context) but not a deduplicated file-path list. The size issue described does not exist in Partio's checkpoint schema.

## changelog 0.10.2: subagent hooks for Codex and Cursor (SubagentStart/SubagentStop)

<!-- partio:rejection:v1 -->

- source: `entireio-cli changelog 0.10.2 (better subagent handling for Codex and Cursor)`
- reason: `irrelevant`
- note: Partio has no subagent tracking for any agent. Adding SubagentStart/SubagentStop support would require a new hook mechanism that does not exist in Partio's architecture.

## changelog 0.10.2: --object-format flag for repo create (sha1/sha256)

<!-- partio:rejection:v1 -->

- source: `entireio-cli changelog 0.10.2 (entire repo create --object-format)`
- reason: `irrelevant`
- note: Partio has no repo create command. Checkpoints are written to an orphan branch in the existing repo using git plumbing.

## changelog 0.10.1: checkpoint restore error handling for out-of-boundary restores

<!-- partio:rejection:v1 -->

- source: `entireio-cli changelog 0.10.1 (entire checkpoint restore now returns an error when restoring outside a session boundary)`
- reason: `irrelevant`
- note: Partio has no checkpoint restore command. Rewinding is done via `partio rewind` which applies git operations, not Entire's server-side restore API.

## changelog 0.10.1: self-heal zombie sessions via detached sweep from session-start

<!-- partio:rejection:v1 -->

- source: `entireio-cli changelog 0.10.1 (feat: self-heal zombie sessions via detached sweep from session-start) / PR #2029`
- reason: `irrelevant`
- note: Filed as Partio issue #679. Not logged as an idea rejection — listed here for traceability only, as the analogous Partio proposal was filed.

## changelog 0.10.0: Rust-based git library for pull/checkout tracking

<!-- partio:rejection:v1 -->

- source: `entireio-cli changelog 0.10.0 (Entire now includes a full Rust-based git library)`
- reason: `irrelevant`
- note: Partio is a pure Go project with no Rust components. Pull/checkout tracking is not part of Partio's scope; it captures commits, not individual git operations.

## changelog 0.10.0: session cache-stats command

<!-- partio:rejection:v1 -->

- source: `entireio-cli changelog 0.10.0 (entire session cache-stats lists session caches with estimated memory size)`
- reason: `irrelevant`
- note: Partio has no concept of session caches or memory limits. Sessions are JSONL files on disk; there is no in-process cache to inspect.

## changelog 0.10.0: session list --where filter by model name

<!-- partio:rejection:v1 -->

- source: `entireio-cli changelog 0.10.0 (entire session list --where can now filter by model name)`
- reason: `irrelevant`
- note: Partio has no session list command. Model information is stored in checkpoints but not exposed through a queryable interface.

## changelog 0.10.0: SSH key signing for commits in Claude Code/Codex/Cursor

<!-- partio:rejection:v1 -->

- source: `entireio-cli changelog 0.10.0 (embeds reference implementation for signing commits with SSH keys)`
- reason: `irrelevant`
- note: Partio uses `git commit --amend` to add trailers; it defers commit signing to git's own configuration. Partio does not manage signing keys.

## changelog 0.10.0: link mid-task background-subagent commits to their session (PR #2034)

<!-- partio:rejection:v1 -->

- source: `entireio-cli changelog 0.10.0 / PR #2034 (feat: link mid-task background-subagent commits to their session)`
- reason: `irrelevant`
- note: Partio has no background subagent concept. Commit-to-session linking via process identity is already filed as Partio issue #673.

## changelog 0.9.5: session pause/resume flags

<!-- partio:rejection:v1 -->

- source: `entireio-cli changelog 0.9.5 (entire session update --pause and --resume flags)`
- reason: `irrelevant`
- note: Partio's session state machine (ACTIVE, IDLE, ENDED) does not include a paused state and there is no interactive session management UI. Partio is a background capture tool, not a session controller.

## changelog 0.9.5: global --timestamp flag across all commands

<!-- partio:rejection:v1 -->

- source: `entireio-cli changelog 0.9.5 (--timestamp flag moved to global; supports RFC3339, Unix, natural language)`
- reason: `irrelevant`
- note: Partio's commands (enable, disable, status, doctor, rewind, etc.) operate on a single repo's state and do not filter output by time range.

## changelog 0.9.3: entire repo copy command

<!-- partio:rejection:v1 -->

- source: `entireio-cli changelog 0.9.3 (new entire repo copy command for copying git repositories with full history)`
- reason: `irrelevant`
- note: Partio has no repo management commands. It hooks into git operations but does not copy repositories.

## changelog 0.9.3: session diagnose command with memory/CPU usage

<!-- partio:rejection:v1 -->

- source: `entireio-cli changelog 0.9.3 (new diagnostic command entire session diagnose)`
- reason: `irrelevant`
- note: Partio has `partio doctor` for health checks. `partio doctor` checks file existence and hook installation; it does not monitor process memory/CPU. Adding detailed session diagnostics would require a running server, which Partio does not have. A deeper doctor mode is already filed as issue #661.

## changelog 0.9.2: search --format=json and --context-lines flag

<!-- partio:rejection:v1 -->

- source: `entireio-cli changelog 0.9.2 (entire search --format=json, --context-lines flag)`
- reason: `irrelevant`
- note: Partio has no search command. Checkpoint content is readable via `partio rewind` and raw git commands, not a search interface.

## changelog 0.9.1: checkpoint explain streaming and partial output on cancel

<!-- partio:rejection:v1 -->

- source: `entireio-cli changelog 0.9.1 (entire checkpoint explain partial explanations, streaming)`
- reason: `irrelevant`
- note: Partio has no checkpoint explain command. The `partio explain` command (`cmd/partio/explain.go`) calls an external LLM API; it is not part of Partio's checkpoint storage architecture.

## PRs #1882–#2073: dependency bumps, CI fixes, platform infrastructure

<!-- partio:rejection:v1 -->

- source: `entireio-cli-pulls #1882–#2073 (dependency updates, CI hang fix, trail backend retirement, dispatch routing, telemetry, release channel, OpenCode plugin, Cursor/Codex subagent fixes, search TUI changes, changelog PRs)`
- reason: `irrelevant`
- note: The bulk of PRs in this range are dependency version bumps, CI configuration fixes (e.g. #2072 apt hang), platform routing changes (e.g. #2046–#2051 cell targets, dispatch), telemetry (#2023–#2024), trail backend retirement (#2021, #2037), OpenCode plugin changes (#2018, #2027, #2053), Cursor/Codex subagent bug fixes (#2066–#2071), search TUI updates (#2022, #2044), and changelog/credit PRs (#2039, #2050, #2073). None of these map to Partio's domain of git hook-based session capture and checkpoint storage.
