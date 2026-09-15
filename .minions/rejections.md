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

## changelog 0.10.3: subagent work in committed checkpoints

<!-- partio:rejection:v1 -->

- source: `entireio-cli changelog 0.10.3 (subagent work survives into committed checkpoints for Claude Code, Codex, Cursor, Copilot CLI, Factory AI Droid)`
- reason: `irrelevant`
- note: Partio writes one checkpoint per git commit and has no subagent tracking concept. There is no task_records ledger, no subagent transcript materialisation, and no SubagentStop hook registration in this codebase.

## changelog 0.10.3: process-identity-based commit-to-session linking

<!-- partio:rejection:v1 -->

- source: `entireio-cli changelog 0.10.3 (link commits to sessions by process ancestry instead of worktree path)`
- reason: `irrelevant`
- note: Already filed for Partio as issue #673. Not re-filed to avoid duplicates.

## changelog 0.10.3: configurable secret-scanner engines

<!-- partio:rejection:v1 -->

- source: `entireio-cli changelog 0.10.3 (redaction.betterleaks.enabled and redaction.goredact.enabled toggles)`
- reason: `irrelevant`
- note: Already filed for Partio as issue #691. Not re-filed to avoid duplicates.

## changelog 0.10.3: zombie-session self-heal from session-start

<!-- partio:rejection:v1 -->

- source: `entireio-cli changelog 0.10.3 (zombie sessions self-heal via detached sweep from session-start)`
- reason: `irrelevant`
- note: Already filed for Partio as issue #679. Not re-filed to avoid duplicates.

## changelog 0.10.4: repo clone, dispatch, enable --search-skill, trail --json, async mirror

<!-- partio:rejection:v1 -->

- source: `entireio-cli changelog 0.10.4 (entire repo clone /et/…, entire dispatch --jurisdiction, entire enable --search-skill, trail list/show --json, async mirror creation)`
- reason: `irrelevant`
- note: Partio has no repo clone, dispatch, search-skill, trail, or mirror concepts. These are cloud-platform features of Entire that have no analogue in Partio's git hook-based checkpoint architecture.

## changelog 0.10.4: .entire symlink and FIFO protection

<!-- partio:rejection:v1 -->

- source: `entireio-cli changelog 0.10.4 (.entire must be a real directory; settings never read through a link)`
- reason: `irrelevant`
- note: Already filed for Partio as issue #707. Not re-filed to avoid duplicates.

## changelog 0.10.5: full.jsonl released after condensation

<!-- partio:rejection:v1 -->

- source: `entireio-cli changelog 0.10.5 (session full.jsonl released from .entire/metadata once condensed)`
- reason: `irrelevant`
- note: Partio has no JSONL staging buffer in the working tree. It reads session files directly from the Claude config directory (~/.claude/…) and writes only to the orphan branch; no accumulation in .partio/ occurs.

## changelog 0.10.5: remove Read(./.entire/metadata/**) deny rule

<!-- partio:rejection:v1 -->

- source: `entireio-cli changelog 0.10.5 (Entire no longer installs a Read(./.entire/metadata/**) deny rule into Claude Code's permission config)`
- reason: `irrelevant`
- note: Partio does not install any deny rules into .claude/settings.json. The enable.go code adds entries only to .gitignore, not to Claude's permissions config.

## changelog 0.10.5: doctor reports waiting for session lock

<!-- partio:rejection:v1 -->

- source: `entireio-cli changelog 0.10.5 (entire doctor says it is waiting for a session lock instead of stalling)`
- reason: `irrelevant`
- note: Partio's doctor (cmd/partio/doctor.go) performs only filesystem stat checks and prints results immediately. It acquires no locks and has no blocking wait path.

## changelog 0.10.6: Windows PowerShell installer

<!-- partio:rejection:v1 -->

- source: `entireio-cli changelog 0.10.6 (scripts/install.ps1 for Windows PowerShell 5.1 and 7)`
- reason: `irrelevant`
- note: Partio has no install scripts. Distribution is handled outside the repo; adding a PowerShell installer is not in scope for this project.

## changelog 0.10.6: entire runner setup rework

<!-- partio:rejection:v1 -->

- source: `entireio-cli changelog 0.10.6 (reworked entire runner setup flags and interaction flow)`
- reason: `irrelevant`
- note: Partio has no runner setup command. Runners are an Entire-specific concept for AI-driven code review and trail management.

## changelog 0.10.6: named pipe/FIFO hang fix

<!-- partio:rejection:v1 -->

- source: `entireio-cli changelog 0.10.6 (named pipe where Entire expects a config file no longer hangs the process)`
- reason: `irrelevant`
- note: Already filed for Partio as issue #707. Not re-filed to avoid duplicates.

## `entire doctor` stuck-session prompt bypasses interactive guard (issue #2416)

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2416 (entire doctor stuck-session prompt bypasses CanPromptInteractively())`
- reason: `irrelevant`
- note: Partio's doctor (cmd/partio/doctor.go) has no interactive prompts. It performs only filesystem stat checks and prints plain-text results. There is no huh/bubbletea form or session-fix loop to guard.

## `entire enable --force` writes hook wrapper through symlink (issue #2410)

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2410 (entire enable --force writes hook wrapper through symlink, clobbering target file)`
- reason: `irrelevant`
- note: Partio's enable command has no --force flag. The basic installHooks path renames a non-partio hook to .partio-backup before writing (removing the symlink, not writing through it), so the symlink write-through scenario from this issue does not arise.

## ULID checkpoint refs unreadable on case-insensitive filesystem (issues #2402, #2401)

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2402 and #2401 (ULID checkpoint refs permanently unreadable when shard case-collides on macOS/Windows)`
- reason: `irrelevant`
- note: Partio stores checkpoints on an orphan branch (partio/checkpoints/v1) using git plumbing, not a ULID-based git-refs store. The case-collision scenario described does not exist in Partio's checkpoint architecture.

## Checkpoint linking fails for non-ASCII filenames (issue #2398)

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2398 (getStagedFiles uses git diff --cached --name-only without -z, miscompares non-ASCII paths)`
- reason: `irrelevant`
- note: Partio does not compare staged filenames against session file paths in its commit-linking path. The pre-commit hook detects whether an agent is running; the post-commit hook links the session to the commit by session ID. No filename comparison exists.

## Documentation: Homebrew tap order (issue #2394)

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2394 (docs: brew trust must precede brew tap in install instructions)`
- reason: `irrelevant`
- note: Partio has no Homebrew tap or install documentation in this repository.

## git-refs push queue bookkeeping (issue #2393)

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2393 (git-refs writes can return success without durable push bookkeeping)`
- reason: `irrelevant`
- note: Partio has no git-refs checkpoint backend or push queue. Checkpoints are written to an orphan branch synchronously; there is no enqueueForPush path.

## Cleanup makes uncondensed checkpoint commits unreachable (issue #2378)

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2378 (cleanup can make uncondensed checkpoint commits unreachable after state expiry or ref read failure)`
- reason: `irrelevant`
- note: Partio has no session-state expiry, no shadow branches, and no cleanup sweep. Each checkpoint is a commit on the orphan branch written immediately at post-commit time; there is no deferred condensation path.

## Session-wide token total overwritten (issue #2368)

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2368 (applyBackfilledSessionTokenUsage overwrites session-wide token total with window delta)`
- reason: `irrelevant`
- note: Partio captures TotalTokens from the parsed session file but does not track per-checkpoint token windows or accumulate session-wide totals across commits.

## session resume writes to hardcoded ~/.claude (issue #2367)

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2367 (entire session resume restores transcript to hardcoded ~/.claude/projects/… ignoring CLAUDE_CONFIG_DIR)`
- reason: `irrelevant`
- note: Partio's resume (cmd/partio/resume.go) writes a context prompt to os.TempDir() and passes it as a CLI argument when launching claude. It does not write to ~/.claude or any session directory.

## `entire status` collapses multiple sessions per agent (issue #2362)

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2362 (entire status --json deduplicates active_sessions to one entry per agent)`
- reason: `irrelevant`
- note: Partio's status command (cmd/partio/status.go) does not enumerate or report individual sessions. It reports hooks, checkpoint branch, and config values only.

## checkpoint_push_remote election accepts pushurl-only remote (issue #2360)

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2360 (checkpoint_push_remote election accepts pushurl-only remote that status then reports as working)`
- reason: `irrelevant`
- note: Partio has no configurable checkpoint push remote. The pre-push hook pushes the orphan branch to origin directly; there is no remote election or push-URL-only validation.

## Post-push cleanup deletes shadow ref when state is malformed (issue #2350)

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2350 (post-push cleanup deletes uncondensed shadow ref when session state is malformed)`
- reason: `irrelevant`
- note: Partio has no shadow branches, no session-state list, and no post-push cleanup sweep.

## Add Devin agent support (issue #2328)

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2328 (Add Devin agent support)`
- reason: `irrelevant`
- note: Already filed for Partio as issue #585. Not re-filed to avoid duplicates.

## Install script hides GitHub API errors (issue #2318)

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2318 (install script hides GitHub API authentication and rate-limit errors)`
- reason: `irrelevant`
- note: Partio has no install script in this repository.

## Windows install script URL gives login page (issue #2276)

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2276 (Windows install script URL gives a login page rather than the script)`
- reason: `irrelevant`
- note: Partio has no install scripts.

## `entire disable` does not reach linked worktrees (issue #2274)

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2274 (entire disable writes settings.local.json per worktree, so a new git worktree add comes back enabled)`
- reason: `irrelevant`
- note: Partio's disable command (cmd/partio/disable.go) removes git hooks from git-common-dir, which is shared across all linked worktrees. It writes no per-worktree settings file; there is no re-enable-on-new-worktree scenario.

## Docs: security page lacks data-residency information (issue #2271)

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2271 (docs: security page has no data-residency information)`
- reason: `irrelevant`
- note: Partio is a local-only tool; checkpoints are stored on an orphan branch in the repository itself. There is no hosted backend, no data-residency consideration, and no equivalent docs page.

## `entire status` reports hooks installed when hooks not installed (issue #2264)

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2264 (entire status reports checkpoints sync to origin when git hooks are not installed)`
- reason: `irrelevant`
- note: Already filed for Partio as issue #704. Not re-filed to avoid duplicates.

## Lefthook classified as non-overwriting hook manager (issue #2263)

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2263 (Lefthook classified as not overwriting Entire hooks, so users get the weaker warning)`
- reason: `irrelevant`
- note: Already covered by Partio issue #433 (auto-recover hooks after external hook manager overwrites). Not re-filed to avoid duplicates.

## Read(./.entire/metadata/**) deny rule breaks auto/unattended mode (issue #2260)

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2260 (Read(./.entire/metadata/**) deny rule makes ordinary commands ask for approval)`
- reason: `irrelevant`
- note: Partio's enable command (cmd/partio/enable.go) does not install any deny rules into .claude/settings.json. Only .gitignore entries are added; no Claude permission config is modified.

## External agent config scope leaks between settings files (issue #2257)

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2257 (adding an external agent writes merged settings back to one scope, leaking values)`
- reason: `irrelevant`
- note: Partio has no interactive agent management UI or external agent add command. Agent selection is via the PARTIO_AGENT env var or the agent config field.

## AppendCheckpointTrailer emits trailer the parser rejects (issue #2256)

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2256 (AppendCheckpointTrailer can emit a trailer the final-block parser rejects)`
- reason: `irrelevant`
- note: Partio's AmendTrailers (internal/git/amend_trailers.go) always prepends \n\n before the trailer lines and does not have a strict final-block parser. Commit trailers are decorative metadata; checkpoint lookup uses the orphan branch, not commit message scanning.

## Forged Entire-Checkpoint lines in commit body (issue #2255)

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2255 (forged Entire-Checkpoint lines in commit bodies can select and mutate unrelated checkpoints)`
- reason: `irrelevant`
- note: Partio does not use commit trailers to look up or mutate checkpoints. Checkpoints are located by ID on the orphan branch. A Partio-Checkpoint trailer is decorative; no code path reads it to select or modify checkpoint storage.

## Stale `entire configure --agent` hints in error messages (issue #2249)

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2249 (stale entire configure --agent hints in investigate and review error messages)`
- reason: `irrelevant`
- note: Partio has no investigate or review commands and no entire configure --agent CLI path.

## Re-install discards newer hook and chains to stale backup (issue #2237)

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2237 (Re-install discards another tool's newer hook and chains to a stale .pre-entire backup)`
- reason: `premise-failed`
- claim: `installHooks` overwrites the existing `.partio-backup` on reinstall, discarding the prior backup content [evidence: `internal/git/hooks/install.go`]
- verdict: `holds` — but the consequence is the opposite of Entire's bug. Partio's `os.Rename` atomically replaces the backup with the **current** foreign hook (e.g. the updated lefthook script), not a stale copy. The chain therefore calls the most recent foreign hook, which is correct behaviour. The Entire bug (chaining to a stale copy) does not reproduce here because Partio does not skip the rename when a backup exists. The related issue — that the **original** user hook (backed up on the very first install) is silently overwritten by a foreign hook on a subsequent enable — was filed separately as partio-io/cli#710.
- found: `internal/git/hooks/install.go` line 41: `os.Rename(hookPath, backupPath)` runs unconditionally. On Linux, rename(2) atomically replaces the destination, so the backup is always updated to the current foreign hook content. Entire's variant of the bug (skipping the rename when a backup exists, then chaining to the old backup) is not present in Partio. The distinct loss — original user hook overwritten — is the scenario filed as #710.

## Windows: entire.exe reports version 0.0.0.0 in PE metadata (issue #2218)

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2218 (Windows: entire.exe reports version 0.0.0.0 in PE metadata)`
- reason: `irrelevant`
- note: Partio embeds version via Go ldflags at build time (accessible via `partio version`). Windows PE VERSIONINFO resource embedding is not currently in scope; the project targets Linux/macOS as primary platforms.

## Claude Code SubagentStop dropped (issue #2215)

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2215 (Claude Code SubagentStop is dropped: Entire correlates on tool_use_id which the hook does not send)`
- reason: `irrelevant`
- note: Partio has no SubagentStop hook, no subagent correlation, and no per-agent hook event system. Session capture is git hook-based (pre-commit, post-commit, pre-push), not agent lifecycle hook-based.

## OPF writers still CAS checkpoint refs via go-git (issue #2204)

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2204 (OPF writers still CAS checkpoint refs via go-git, racing native update-ref)`
- reason: `irrelevant`
- note: Partio has no OPF (On-Push Finalization) layer, no go-git dependency, and no CAS checkpoint ref logic. Checkpoints are written using git plumbing commands directly.

## scripts/install.sh never shellchecked (issue #2203)

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2203 (scripts/install.sh is never shellchecked and install.ps1 has no linter)`
- reason: `irrelevant`
- note: Partio has no install scripts.

## Make test git isolation structural (issue #2202)

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2202 (Make test git isolation structural: runGit helper + forbidigo rule)`
- reason: `irrelevant`
- note: This is an internal testing infrastructure improvement specific to Entire's codebase. Partio's tests already use t.TempDir() and t.Setenv() for isolation.

## auth-go lock dir isolation doesn't cover spawned binaries (issue #2201)

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2201 (auth-go lock dir isolation doesn't cover spawned binaries)`
- reason: `irrelevant`
- note: Partio has no authentication layer or process lock management.

## Persistent-ref lock files accumulate (issue #2197)

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2197 (persistent-ref lock files accumulate one per checkpoint and nothing reaps them)`
- reason: `irrelevant`
- note: Partio has no persistent-ref backend. Checkpoints are written to an orphan branch using git plumbing; there are no per-checkpoint lock files.

## Goose and Qwen Code agent integrations request (issue #2160)

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2160 (Feature request: Goose and Qwen Code agent integrations)`
- reason: `irrelevant`
- note: Already filed for Partio as issues #681 (Goose) and #684 (Qwen Code). Not re-filed to avoid duplicates.

## Redaction corrupts Grok reasoning blocks (issue #2157)

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2157 (Redaction corrupts Grok reasoning blocks: encrypted_content not covered by signature skip rule)`
- reason: `irrelevant`
- note: Partio supports only claude-code and codex agents. There is no Grok agent or Grok-specific session format in this codebase.

## finalizeAllTurnCheckpoints has no total deadline (issue #2148)

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2148 (finalizeAllTurnCheckpoints has no total deadline: slow remote can outlast hook timeout)`
- reason: `irrelevant`
- note: Partio has no turn-by-turn checkpoint finalization. Post-commit runs synchronously and writes one checkpoint to the local orphan branch; there is no remote fetch, no turn loop, and no configurable hook timeout.

## OpenCode Desktop hooks never fire (issue #2137)

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2137 (It doesn't work with Opencode Desktop)`
- reason: `irrelevant`
- note: Partio has no OpenCode support.

## Agent picker hides last options in `entire enable` (issue #2126)

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2126 (Agent picker hides last options: OpenCode and Pi need scrolling in entire enable)`
- reason: `irrelevant`
- note: Partio has no interactive agent picker. Agent selection is via PARTIO_AGENT env var or config field.

## Semantic search 401 for EU account (issue #2121)

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2121 (semantic search returns 401 for EU-jurisdiction account)`
- reason: `irrelevant`
- note: Partio has no semantic search feature. Checkpoints are stored on a local orphan branch with no cloud search backend.

## Hook timeouts: flat 30s wrong in both directions (issue #2115)

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2115 (Hook timeouts: a flat 30s is wrong in both directions)`
- reason: `irrelevant`
- note: Partio does not configure timeouts for external agent hooks. The Entire issue is about timeout values Entire injects into Codex's hook configuration. Partio's git hooks (pre-commit, post-commit, pre-push) run to completion without a configured timeout.

## index emptied between staging and commit (issue #2111)

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2111 (index emptied between staging and commit resulting in git commit recording an empty tree)`
- reason: `irrelevant`
- note: Already filed for Partio as issue #693 (AmendTrailers consumes the git index). Not re-filed to avoid duplicates.

## CLI Performance Regression After Enabling (issue #2098)

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2098 (CLI Performance Regression After Enabling Full Repository)`
- reason: `irrelevant`
- note: This is a user report about Entire's condensation and hook latency. Partio's hooks are much simpler (no condensation, no complex session tracking) and do not exhibit the same performance profile.

## Hook latency: remaining work after #2002 (issue #2091)

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2091 (Hook latency: remaining work after #2002 — full.jsonl re-redacted from byte zero every checkpoint)`
- reason: `irrelevant`
- note: Partio's post-commit hook reads and redacts the session JSONL once per commit, not incrementally. There is no persistent blob cache or re-redaction from a shared buffer. The specific optimisation described is internal to Entire's condensation pipeline.

## checkpoint explain --session false negative when scan limit truncates (issue #2089)

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2089 (checkpoint explain --session reports No checkpoints found when scan limit truncates)`
- reason: `irrelevant`
- note: Partio's explain command (cmd/partio/explain.go) takes a checkpoint ID directly and reads it from the orphan branch. It has no scan limit and no --session flag.

## PRs #2074–#2481: shadow PRs, Windows fix, cluster list, test isolation

<!-- partio:rejection:v1 -->

- source: `entireio-cli-pulls #2074–#2481 (shadow/draft trail PRs #2432–#2478, PR #2479 doctor_test TTY isolation, PR #2480 entire cluster list, PR #2481 Windows double-Enter prompt fix)`
- reason: `irrelevant`
- note: The bulk of PRs in this range are Entire internal shadow/draft trail PRs with no code. The non-draft PRs are: #2479 (test env-var isolation for Entire's doctor TTY test — internal concern), #2480 (entire cluster list for cloud cluster catalog — no cluster concept in Partio), #2481 (Windows CONIN$ read cancellation for interactive prompts — Windows-specific teardown; Partio's promptCommitLinking uses /dev/tty which is Unix-only). None map to Partio's domain.
