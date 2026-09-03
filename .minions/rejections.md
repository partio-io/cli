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

## changelog 0.10.3: subagent work persists into committed checkpoints

<!-- partio:rejection:v1 -->

- source: `entireio-cli changelog 0.10.3 (subagent work persists into committed checkpoints for Claude Code, Codex, Cursor, Copilot CLI, and Factory AI Droid)`
- reason: `irrelevant`
- note: Partio writes one checkpoint per git commit from the post-commit hook. It has no subagent tracking pipeline, no incremental turn checkpoints, and no multi-agent transcript aggregation. The feature is Entire-specific.

## changelog 0.10.3: agent commit linking via process ancestry matching

<!-- partio:rejection:v1 -->

- source: `entireio-cli changelog 0.10.3 (agent commit linking via process ancestry matching instead of worktree path)`
- reason: `irrelevant`
- note: Already filed for Partio as issue #673 and program `session-commit-link-by-pid.md`. Not re-filed to avoid duplicates.

## changelog 0.10.3: zombie session healing with 24-hour uncondensed data detection

<!-- partio:rejection:v1 -->

- source: `entireio-cli changelog 0.10.3 (automatic zombie session healing with 24-hour uncondensed data detection)`
- reason: `irrelevant`
- note: Already filed for Partio as issue #679. Not re-filed to avoid duplicates.

## changelog 0.10.3: configurable redaction engine selection

<!-- partio:rejection:v1 -->

- source: `entireio-cli changelog 0.10.3 (configurable redaction engine selection via .entire/settings.json)`
- reason: `irrelevant`
- note: Already filed for Partio as issue #691 (independent control of pattern-based vs entropy-based redaction layers). Not re-filed to avoid duplicates.

## changelog 0.10.3: trail update etag support, content-free telemetry, search TUI changes

<!-- partio:rejection:v1 -->

- source: `entireio-cli changelog 0.10.3 (trail update etag, --overwrite option; content-free telemetry for skills/commits/search; search TUI continuous scrolling; accurate search result counts)`
- reason: `irrelevant`
- note: Partio has no trail, telemetry, skill, or search subsystem.

## changelog 0.10.3: transcript redaction performance improved via sharding

<!-- partio:rejection:v1 -->

- source: `entireio-cli changelog 0.10.3 (transcript redaction performance improved 66s → 2.0s via sharding)`
- reason: `irrelevant`
- note: Partio's redaction operates on individual commit sessions, not on large-scale multi-session transcript corpora. The sharding optimization addresses a scale problem Partio does not have.

## changelog 0.10.3: pathological repo hook processes bounded by 20s budget

<!-- partio:rejection:v1 -->

- source: `entireio-cli changelog 0.10.3 (pathological repo hook processes bounded by 20s budget)`
- reason: `irrelevant`
- note: Already filed for Partio as issue #700 (post-commit hook has no total deadline). Not re-filed to avoid duplicates.

## changelog 0.10.3: session-end condensation checkpoint durability, external agent hook cleanup, Cursor/OpenCode fixes

<!-- partio:rejection:v1 -->

- source: `entireio-cli changelog 0.10.3 (session-end condensation checkpoint durability; external agent hook cleanup on uninstall; Cursor file attribution and session title formatting; OpenCode Desktop hooks support)`
- reason: `irrelevant`
- note: Partio has no condensation pipeline, no Cursor support, and no OpenCode support. External agent hook cleanup does not apply to Partio's enable/disable model.

## changelog 0.10.4: native repo cloning, jurisdiction-scoped dispatch, skill installation, async mirror

<!-- partio:rejection:v1 -->

- source: `entireio-cli changelog 0.10.4 (native repo cloning; jurisdiction-scoped cloud dispatch; real agent skill installation; async mirror request support; trail commands JSON output; reduced runner scaffolding)`
- reason: `irrelevant`
- note: Partio has no repo clone command, no cloud dispatch, no skill system, no mirror infrastructure, and no trail commands.

## changelog 0.10.4: .entire directory symlink validation and os.Root confinement

<!-- partio:rejection:v1 -->

- source: `entireio-cli changelog 0.10.4 (.entire directory validation now requires real directories, not symlinks; filesystem operations confined to os.Root anchors)`
- reason: `irrelevant`
- note: Partio reads and writes to `.partio/` within the repository. The os.Root API is a Go 1.24+ addition and Partio targets Go 1.25, but the threat model (symlink-based directory traversal to write outside the repo) requires an attacker who can already commit to the repository — which is a higher bar than Partio's current security model addresses. No specific path has been identified in Partio's code where symlink traversal produces meaningful harm beyond what git itself already mitigates.

## changelog 0.10.4: OAuth cross-host redirect refusal, plugin resolver absolute-path restriction, shadow-branch suffix validation, investigation run-ID verification

<!-- partio:rejection:v1 -->

- source: `entireio-cli changelog 0.10.4 (OAuth token exchange no longer follows cross-host redirects; plugin resolver restricts fallback path scanning to absolute entries; shadow-branch cleanup requires worktree-hash suffix validation; investigation findings resolved through validated run IDs)`
- reason: `irrelevant`
- note: Partio has no OAuth token exchange, no plugin resolver, no shadow branches, and no investigation findings. These are Entire cloud-infrastructure fixes.

## changelog 0.10.4: concurrent checkpoint writers protected with per-ref locking and git CAS

<!-- partio:rejection:v1 -->

- source: `entireio-cli changelog 0.10.4 (concurrent checkpoint writers protected with per-ref locking and Git CAS)`
- reason: `irrelevant`
- note: Already considered and rejected in a prior run (source: entireio-cli-issues #1917). Partio writes checkpoints from git hooks, which git serialises per-commit. The concurrent-writer race is Entire-specific.

## changelog 0.10.4: git isolation standardized across test runs

<!-- partio:rejection:v1 -->

- source: `entireio-cli changelog 0.10.4 (git isolation standardized across test runs)`
- reason: `irrelevant`
- note: Already filed for Partio as issue #696 (Isolate git test commands from the developer's global git config). Not re-filed to avoid duplicates.

## issues #2257: settings layer contamination when writing external_agents flag

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2257 (adding an external agent writes merged settings back to one scope, leaking values between settings.json and settings.local.json)`
- reason: `irrelevant`
- note: Partio's `SaveRepoSetting` (internal/config/save.go) writes a single key into the scope-level JSON file by name, not the whole merged config object. The load-merged-then-save-whole pattern that caused the Entire bug does not exist in Partio's settings write path.

## issues #2256: AppendCheckpointTrailer grammar mismatch with final-block parser

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2256 (AppendCheckpointTrailer can emit a trailer the final-block parser and git reject)`
- reason: `irrelevant`
- note: Partio adds trailers via `git.AmendTrailers` (internal/git/amend_trailers.go), which appends a separator and trailer lines to the commit message, then calls `git commit --amend -m`. Partio has no custom trailer-block parser that re-reads commit messages to identify checkpoints; the only reader is `git show` on the orphan branch, not the commit body. The grammar-mismatch attack surface Entire fixed does not exist in Partio's architecture.

## issues #2255: forged Entire-Checkpoint lines in commit bodies can mutate unrelated checkpoints

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2255 (forged Entire-Checkpoint lines in commit bodies can select and mutate unrelated checkpoints)`
- reason: `irrelevant`
- note: Partio's checkpoint resolution reads checkpoints from the orphan branch (`partio/checkpoints/v1`) by ID, using `git show`. It does not scan commit bodies for `Partio-Checkpoint:` lines to decide which checkpoint to operate on. A forged trailer in a commit body cannot redirect Partio's checkpoint operations.

## issues #2249: stale configure --agent hints in investigate and review error messages

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2249 (stale entire configure --agent hints in investigate and review error messages)`
- reason: `irrelevant`
- note: Partio has no investigate or review commands. Agent configuration uses `PARTIO_AGENT` env var or the `agent` config field, not a dedicated CLI subcommand.

## issues #2237: re-install discards another tool's newer hook and chains to stale backup

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2237 (re-install discards another tool's newer hook and chains to a stale .pre-entire backup)`
- reason: `irrelevant`
- note: Partio's hook installer (`internal/git/hooks/install.go`) does rename the current hook to `.partio-backup` if it is not a Partio hook, and does overwrite an existing backup with the current file. If another tool reinstalls its hook between two `partio enable` calls, the result is that the Partio hook chains to that tool's latest version — no stale copy survives. The specific failure mode (chaining to a stale backup after a reinstall) does not reproduce in Partio's install path.

## issues #2218: Windows binary reports version 0.0.0.0 in PE metadata

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2218 (Windows: entire.exe reports version 0.0.0.0 in PE metadata)`
- reason: `irrelevant`
- note: Already filed for Partio as issue #695 (Add Windows binary release artifacts and PowerShell installer script). The PE metadata concern is embedded in that scope. Not re-filed.

## issues #2215: Claude Code SubagentStop is dropped — hook does not send tool_use_id

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2215 (Claude Code SubagentStop is dropped: Entire correlates on tool_use_id, which the hook does not send)`
- reason: `irrelevant`
- note: Partio has no UserPromptSubmit hook and no subagent correlation mechanism. Session capture happens at git-commit time via standard git hooks, not via Claude Code's hook API.

## issues #2204, #2203, #2201, #2197, #2196: internal race conditions, test infrastructure, lock-file accumulation

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2204, #2203, #2201, #2197, #2196 (OPF CAS race; shellcheck gaps; auth-go lock dir in tests; lock-file accumulation; MigrateBranchToRefs performance)`
- reason: `irrelevant`
- note: OPF and go-git CAS races are Entire-specific (no OPF in Partio). Shellcheck and auth-go test isolation are internal Entire concerns. Persistent-ref lock-file accumulation and branch-migration performance both require Entire's per-checkpoint ref update mechanism, which Partio does not use.

## issues #2160: Goose and Qwen Code agent integrations

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2160 (feature request: Goose and Qwen Code agent integrations)`
- reason: `irrelevant`
- note: Already filed for Partio as issues #681 (Goose) and #684 (Qwen Code). Not re-filed to avoid duplicates.

## issues #2157: Grok reasoning block encrypted_content not covered by signature skip rule

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2157 (redaction corrupts Grok reasoning blocks: encrypted_content is not covered by the signature skip rule)`
- reason: `irrelevant`
- note: Partio has no Grok Build agent support and no reasoning-block redaction rules. Partio's redaction uses entropy-based heuristics, not per-agent field skip lists.

## issues #2148: finalizeAllTurnCheckpoints has no total deadline

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2148 (finalizeAllTurnCheckpoints has no total deadline: a slow remote can outlast the agent's hook timeout)`
- reason: `irrelevant`
- note: Already filed for Partio as issue #700 (Post-commit hook has no total deadline). Not re-filed to avoid duplicates.

## issues #2137: OpenCode Desktop does not work; #2126: agent picker hides last options; #2121: semantic search 401 for EU accounts

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2137, #2126, #2121 (OpenCode Desktop; agent picker scroll; EU semantic search 401)`
- reason: `irrelevant`
- note: Partio has no OpenCode support, no agent-picker UI, and no semantic search feature.

## issues #2115: hook timeouts — flat 30s is wrong in both directions

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2115 (hook timeouts: a flat 30s is wrong in both directions)`
- reason: `irrelevant`
- note: Already filed for Partio as issue #700 (Post-commit hook has no total deadline). Not re-filed to avoid duplicates.

## issues #2111: index emptied between staging and commit recording empty tree

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2111 (index emptied between staging and commit resulting in git commit recording an empty tree)`
- reason: `irrelevant`
- note: Already filed for Partio as issue #693 (AmendTrailers consumes the git index, silently including staged files in the amended commit). Not re-filed to avoid duplicates.

## issues #2098, #2091, #2089, #2087: performance regression, hook latency, checkpoint explain scan limit, OPF 9th layer

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2098, #2091, #2089, #2087 (CLI performance regression; hook latency remaining work; checkpoint explain false negative on truncation; OPF 9th layer never runs on git-refs backend)`
- reason: `irrelevant`
- note: Performance regression and hook latency are addressed by Partio issue #700. Checkpoint explain scan-limit false negative requires a `checkpoint explain` command that Partio does not have. OPF is Entire-specific.

## PRs #2074–#2254: cloud infrastructure, security patches, search, Cursor/Codex/trail fixes, internal refactors

<!-- partio:rejection:v1 -->

- source: `entireio-cli-pulls #2074–#2254 (async mirror, OAuth security, os.Root confinement, trail/shadow-branch fixes, search TUI/skill, Cursor ToolInvocationScanner, Codex subagent tracking, checkpoint refactor for gitrepo centralization, repo clone, various changelog PRs)`
- reason: `irrelevant`
- note: The bulk of these PRs address cloud infrastructure (mirror creation, jurisdiction routing, cross-host OAuth), security fixes for Entire's plugin resolver and shadow-branch deletion, search TUI and skill installation, Cursor and Codex agent internals, and checkpoint-subsystem refactors that move Entire's metadata resolution into a canonical gitrepo owner. None map to Partio's domain of git hook-based session capture and orphan-branch checkpoint storage. One PR in this range (feat(tokens): billing-class breakdown, #2210) inspired the filed proposal; it is recorded there, not here.
