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

## changelog 0.10.3–0.10.6: subagent task records, cloud features, security hardening, OpenCode fixes

<!-- partio:rejection:v1 -->

- source: `entireio-cli changelog 0.10.3–0.10.6 (subagent task_records ledger, sibling-worktree session linking via owner fingerprint, trail/mirror async improvements, security anchoring and symlink refusals, runner-setup rework, named-pipe guard, Pi/OpenCode fixes)`
- reason: `irrelevant`
- note: All new items across these four versions either belong to Entire's cloud/trail infrastructure (mirrors, dispatch, cells, grants), are subagent-specific (task_records ledger, finalize-all-turn deadlines), are agent integrations Partio does not support (OpenCode, Cursor, Pi), or are security hardening tied to Entire's server-side architecture (os.Root anchors, OAuth redirects, plugin resolver). The zombie-session self-heal item (0.10.3) was already captured as Partio issue #679 in a prior run.

## [OpenCode] Captures nothing on OpenCode 2: V1 plugin API rejected and V1 export command prints help

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2588`
- reason: `irrelevant`
- note: Partio has no OpenCode support. There is no OpenCode agent, plugin, or export command in this codebase.

## hooks: register Entire as a Git config-based hook on Git 2.54+ (issue)

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2580`
- reason: `irrelevant`
- note: Already filed for Partio as issue #726. Not re-filed to avoid duplicates.

## Claude Code: backgrounded subagent treated as foreground, task record dropped at launch

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2556`
- reason: `irrelevant`
- note: Partio has no subagent tracking. It writes one checkpoint per git commit regardless of subagent activity.

## `entire session resume` reports branch not found for Unicode-equivalent NFC/NFD branch name (issue)

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2551`
- reason: `irrelevant`
- note: Already filed for Partio as issue #718. Not re-filed to avoid duplicates.

## `entire enable --force` exits 0 and refreshes nothing without a TTY

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2535`
- reason: `irrelevant`
- note: Partio's `enable` command (`cmd/partio/enable.go`) has no `--force` flag and no interactive agent-selection prompt. The non-interactive bailout described in this issue does not exist in Partio.

## `repo grant list` mirror placement checks active credentials

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2518`
- reason: `irrelevant`
- note: Partio has no cloud mirror infrastructure, grant commands, or credential management.

## `entire doctor` stuck-session prompt blocks in scripts and CI

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2416`
- reason: `irrelevant`
- note: Partio's `doctor` command (`cmd/partio/doctor.go`) has no interactive prompts. It prints a summary of findings and suggests `partio enable` but never blocks waiting for input.

## `entire enable --force` writes hook wrapper through symlink, clobbering target

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2410`
- reason: `irrelevant`
- note: Partio's `enable` command has no `--force` flag. The symlink-clobbering path described in this issue does not exist in Partio's hook installer.

## ULID checkpoint refs permanently unreadable on case-insensitive filesystem (macOS/Windows)

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2402 and #2401`
- reason: `irrelevant`
- note: Partio's checkpoint IDs are 12-character lowercase hexadecimal strings generated by `hex.EncodeToString` (`internal/checkpoint/checkpoint.go:38-42`). All shard prefixes are lowercase hex; there is no ULID format and no uppercase/lowercase mixing. The case-collision scenario requires mixed-case IDs.

## Checkpoint linking fails for non-ASCII filenames

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2398`
- reason: `irrelevant`
- note: Partio's `DiffNameOnly` (`internal/git/diff_name_only.go`) is used only for debug logging in the post-commit hook. Partio does not compare staged file paths against session content to decide whether to link a checkpoint — it captures the session unconditionally when an agent is active.

## Documentation: Homebrew 6 tap before trust disallowed

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2394`
- reason: `irrelevant`
- note: Partio has no Homebrew formula.

## git-refs writes can return success without durable push bookkeeping

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2393`
- reason: `irrelevant`
- note: Partio's pre-push hook runs `git push` synchronously. There is no asynchronous push queue or bookkeeping layer.

## Cleanup can make uncondensed checkpoint commits unreachable after state expiry

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2378`
- reason: `irrelevant`
- note: Partio has no shadow branches, no per-session state expiry that drives cleanup, and no uncondensed-work tracking. Checkpoints are written directly to the orphan branch commit tree.

## applyBackfilledSessionTokenUsage overwrites session-wide token total with checkpoint window delta

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2368`
- reason: `irrelevant`
- note: Partio has no per-turn token tracking or backfill mechanism.

## session resume writes to ~/.claude instead of $CLAUDE_CONFIG_DIR (issue)

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2367`
- reason: `irrelevant`
- note: Already filed for Partio as issue #708 (FindSessionDir ignores CLAUDE_CONFIG_DIR). Not re-filed to avoid duplicates.

## entire status collapses multiple sessions per agent and mutates session state while reporting

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2362`
- reason: `irrelevant`
- note: Partio's `status` command does not list individual sessions. It shows a single summary line and calls `mgr.CleanupStale()` to remove stale sessions, which is the intended behavior. There is no per-session display to collapse, and the stale-session cleanup on status read is deliberate in Partio's design.

## checkpoint_push_remote: election accepts pushurl-only remote status reports as working

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2360`
- reason: `irrelevant`
- note: Partio has no configurable checkpoint push remote. Checkpoints are pushed via the pre-push hook to whatever `origin` the repo uses.

## Post-push cleanup deletes uncondensed shadow ref when session state is malformed

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2350`
- reason: `irrelevant`
- note: Partio has no shadow branches or post-push cleanup of session state.

## Add Devin agent support

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2328`
- reason: `irrelevant`
- note: Adding new agent integrations requires a detector and session parser. No Devin session format is documented or present in this codebase to base a premise on.

## Install script hides GitHub API authentication and rate-limit errors

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2318`
- reason: `irrelevant`
- note: Partio has no install scripts.

## Windows install script URL gives a login page

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2276`
- reason: `irrelevant`
- note: Partio has no Windows install script.

## `entire disable` does not reach linked worktrees (settings.local.json is per-worktree)

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2274`
- reason: `irrelevant`
- note: Partio's `disable` command removes hooks via `githooks.Uninstall`, which targets the git common dir (shared across all worktrees). Partio does not use a per-worktree `settings.local.json` to gate enablement; disabling is hook-removal only.

## Docs: security page has no data-residency information

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2271`
- reason: `irrelevant`
- note: Docs-only update for Entire's cloud product. Partio is local-only with no cloud data residency.

## entire status reports "Checkpoints sync to: origin" when git hooks are not installed

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2264`
- reason: `irrelevant`
- note: Partio's `status` command does not report a checkpoint sync destination. It shows hook installation status and checkpoint branch existence only.

## Lefthook is classified as not overwriting Entire's hooks

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2263`
- reason: `irrelevant`
- note: Partio's `DetectHookManagers` gives all hook managers the same warning with integration instructions (`internal/git/hooks/detect_hook_managers.go`). There is no two-tier warning system to fix. The underlying problem (hook managers overwriting Partio's hooks) is addressed at the architecture level by issue #726 (git config-based hooks).

## Read(./.entire/metadata/**) deny rule blocks auto/unattended permission modes

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2260`
- reason: `irrelevant`
- note: Partio does not install deny rules into agent permission configs. There is no `.partio/metadata/` deny rule in Partio's setup.

## Adding external agent writes merged settings back to one scope, leaking between settings files

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2257`
- reason: `irrelevant`
- note: Partio has no `partio agent add` command. Agent selection is via `PARTIO_AGENT` env var or config field.

## `AppendCheckpointTrailer` can emit a trailer the final-block parser and git reject

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2256`
- reason: `irrelevant`
- note: Partio's `AmendTrailers` (`internal/git/amend_trailers.go:22-23`) always appends trailers as `\n\n<key>: <value>`, inserting a blank line before the trailer block unconditionally. This does not share Entire's grammar divergence between writer and reader.

## Forged `Entire-Checkpoint` lines in commit bodies can select and mutate unrelated checkpoints

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2255`
- reason: `irrelevant`
- note: Partio writes `Partio-Checkpoint` trailers to commits but never reads them back from commit messages to make decisions. Checkpoint lookup is by ID on the orphan branch directly. A forged trailer line in a commit body has no effect on Partio's behavior.

## Stale `entire configure --agent` hints in investigate and review error messages

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2249`
- reason: `irrelevant`
- note: Partio has no `configure --agent` command. Agent selection is via `PARTIO_AGENT` env var.

## Re-install discards another tool's newer hook and chains to a stale backup

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2237`
- reason: `irrelevant`
- note: Partio's `installHooks` (`internal/git/hooks/install.go:38-44`) calls `os.Rename(hookPath, backupPath)` unconditionally when a foreign hook is present. On Linux, `os.Rename` atomically replaces the backup with the current foreign hook every time, so the backup is always refreshed. The stale-backup bug described in this issue does not apply to Partio.

## Windows: entire.exe reports version 0.0.0.0 in PE metadata

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2218`
- reason: `irrelevant`
- note: Partio embeds version information via Go linker flags from git tags at build time. There is no PE metadata embedding step in Partio's build.

## Claude Code SubagentStop is dropped: Entire correlates on tool_use_id which hook does not send

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2215`
- reason: `irrelevant`
- note: Partio has no subagent tracking for any agent.

## OPF writers still CAS checkpoint refs via go-git, racing native update-ref

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2204`
- reason: `irrelevant`
- note: Partio has no OPF writers, no go-git dependency, and uses `git update-ref` directly via shell for checkpoint branch writes.

## scripts/install.sh is never shellchecked and install.ps1 has no linter

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2203`
- reason: `irrelevant`
- note: Partio has no install scripts.

## Make test git isolation structural: runGit helper + forbidigo rule

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2202`
- reason: `irrelevant`
- note: Internal test-architecture improvement specific to Entire's codebase.

## auth-go lock dir isolation doesn't cover spawned binaries

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2201`
- reason: `irrelevant`
- note: Partio has no auth infrastructure or spawned-binary isolation requirements.

## Persistent-ref lock files accumulate one per checkpoint and nothing reaps them

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2197`
- reason: `irrelevant`
- note: Partio does not use persistent-ref lock files. Checkpoint writes use git plumbing commands directly.

## MigrateBranchToRefs re-runs take a flock and git subprocess per already-imported checkpoint

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2196`
- reason: `irrelevant`
- note: Partio has no branch-to-refs migration pipeline.

## Feature request: Goose and Qwen Code agent integrations

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2160`
- reason: `irrelevant`
- note: Adding new agent integrations requires a detector and session parser specific to that agent. No Goose or Qwen Code session format is documented or present in this codebase to base a premise on.

## Redaction corrupts Grok reasoning blocks: encrypted_content not covered by signature skip rule

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2157`
- reason: `irrelevant`
- note: Partio has no Grok agent support and no signature-skip rule in its redaction package.

## finalizeAllTurnCheckpoints has no total deadline: slow remote can outlast hook timeout

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2148`
- reason: `irrelevant`
- note: Partio has no turn-level checkpoint finalization pipeline or hook timeout management.

## OpenCode Desktop does not work / Agent picker hides last options

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2137 and #2126`
- reason: `irrelevant`
- note: Partio has no OpenCode support and no interactive agent picker.

## Semantic search returns 401 for EU-jurisdiction account

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2121`
- reason: `irrelevant`
- note: Partio has no semantic search feature.

## Hook timeouts: a flat 30s is wrong in both directions

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2115`
- reason: `irrelevant`
- note: Partio's hooks are plain bash scripts that call `partio _hook <name>`. There is no configurable hook timeout; hooks run to completion.

## index emptied between staging and commit resulting in empty tree

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2111`
- reason: `irrelevant`
- note: Partio's hooks do not touch the git index. The index-emptying described in this issue is caused by Entire's UserPromptSubmit hook, which Partio does not have.

## CLI Performance Regression After Enabling Full Repository

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2098`
- reason: `irrelevant`
- note: Partio has no full-repository indexing mode. Performance is not affected by repository size.

## Hook latency: remaining work after #2002

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2091`
- reason: `irrelevant`
- note: Partio's hooks run simple shell commands with no multi-step pipeline to optimize.

## PRs #2074–#2599: dependency bumps, CI/build fixes, cloud and trail features, OpenCode V2, Antigravity agent, squash/redo trailer inheritance

<!-- partio:rejection:v1 -->

- source: `entireio-cli-pulls #2074–#2599 (dependency bumps, CI fixes, changelog PRs, OpenCode V2 support, Antigravity CLI agent, cloud trail/mirror/cell changes, squash-inherit-checkpoint-trailers, redo-inherit-checkpoint-trailers, hooks-stop-deleting-idle-sessions, non-ASCII staged filenames, skipped-hook file carry-forward, Unicode branch name resume, go-git migration, external-agent improvements)`
- reason: `irrelevant`
- note: The bulk of PRs in this range are dependency version bumps, CI/build fixes, cloud-infrastructure changes (trail, mirror, cell, dispatch, auth), and agent-specific additions (OpenCode V2, Antigravity CLI) that have no equivalent in Partio. Three PRs produced relevant ideas that were already filed as Partio proposals in prior runs: squash-inherit-checkpoint-trailers (#722), unicode-branch-name-lookup (#718), and git-config-based-hooks (#726). Two others (non-ASCII staged filenames #2586 and skipped-hook file carry-forward #2577) depend on Entire's shadow-branch and staged-file-matching architecture that Partio does not have.

## changelog 0.11.0–0.11.3: control plane restructure, OpenCode V2, Antigravity, squash/redo trailers, zombie-session fix

<!-- partio:rejection:v1 -->

- source: `entireio-cli changelog 0.11.0–0.11.3 (breaking control-plane renames, enable simplified, OpenCode V2, Antigravity CLI agent, squash/redo checkpoint trailer inheritance, hooks stop deleting idle sessions, already-committed files not re-checkpointed, token totals fixed, trail/mirror/dispatch/auth changes)`
- reason: `irrelevant`
- note: All new items fall into one of: cloud/trail/dispatch/auth infrastructure not present in Partio; OpenCode or Antigravity agent integrations Partio does not support; or features whose relevant ideas were already filed in prior runs (squash trailers → #722; zombie self-heal → #679; git-config-based hooks → #726). The "already-committed files not re-checkpointed" item (0.11.0) refers to Entire's shadow-branch carry-forward mechanism, which Partio does not have.
