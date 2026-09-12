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

## changelog 0.10.3: subagent tracking (Claude Code SubagentStop, Codex, Cursor, Factory AI)

<!-- partio:rejection:v1 -->

- source: `entireio-cli changelog 0.10.3 (subagent work survives into committed checkpoints for Claude Code, Codex, Cursor, Copilot CLI, Factory AI)`
- reason: `irrelevant`
- note: Partio has no subagent tracking for any agent. It writes one checkpoint per git commit. There is no SubagentStop hook, no per-task transcript ledger, and no multi-agent session architecture.

## changelog 0.10.3: zombie session self-heal from session-start hook

<!-- partio:rejection:v1 -->

- source: `entireio-cli changelog 0.10.3 (Zombie sessions self-heal instead of waiting for doctor, PR #2029)`
- reason: `irrelevant`
- note: Already filed for Partio as issue #679. Not re-filed to avoid duplicates.

## changelog 0.10.3: commit links to session by process ancestry

<!-- partio:rejection:v1 -->

- source: `entireio-cli changelog 0.10.3 (commit links to session by who made it, not worktree path, PR #2013)`
- reason: `irrelevant`
- note: Already filed for Partio as issue #673. Not re-filed to avoid duplicates.

## changelog 0.10.3: configurable secret-scanner engines (betterleaks/goredact)

<!-- partio:rejection:v1 -->

- source: `entireio-cli changelog 0.10.3 (Teams can choose redaction.betterleaks.enabled / redaction.goredact.enabled)`
- reason: `irrelevant`
- note: Partio has no configurable redaction pipeline. Its redact package uses entropy-based heuristics with no named engine registry.

## changelog 0.10.3: doctor trace --summary aggregation

<!-- partio:rejection:v1 -->

- source: `entireio-cli changelog 0.10.3 (entire doctor trace --summary aggregates per-hook p50/p90/max)`
- reason: `irrelevant`
- note: Partio's doctor command checks hook installation and session state; it has no hook-trace collection or latency aggregation.

## changelog 0.10.3: redaction diagnostics and per-process log summary

<!-- partio:rejection:v1 -->

- source: `entireio-cli changelog 0.10.3 (Redaction diagnostics reach .entire/logs, pack/rule/inline count INFO summary)`
- reason: `irrelevant`
- note: Partio has no redaction pipeline with pack loading, rule compilation, or PII/OPF tracking.

## changelog 0.10.3: git status --no-optional-locks

<!-- partio:rejection:v1 -->

- source: `entireio-cli changelog 0.10.3 (every git status Entire runs now passes --no-optional-locks, PR #2143)`
- reason: `irrelevant`
- note: Partio does not run git status. It uses git plumbing commands (hash-object, commit-tree, update-ref, rev-parse) and git diff. The index-corruption race on virtiofs/FUSE mounts described in the changelog does not apply.

## changelog 0.10.3: O(N²) redaction performance fix

<!-- partio:rejection:v1 -->

- source: `entireio-cli changelog 0.10.3 (Stop hook no longer re-redacts whole transcript; sharded goroutines + prefix reuse, PR #2002, #2107)`
- reason: `irrelevant`
- note: Partio has no transcript redaction pipeline. Session JSONL is stored as-is.

## changelog 0.10.3: agent-specific fixes (OpenCode, Cursor, Codex, search, telemetry)

<!-- partio:rejection:v1 -->

- source: `entireio-cli changelog 0.10.3 (OpenCode Desktop Bun→Node fix, OpenCode session attach for untracked sessions, Cursor file attribution, Cursor timestamp wrapper, Codex hooks in linked worktrees, search routing dedup, telemetry payload caching, CRLF unquoted path fix, User-Agent stamping)`
- reason: `irrelevant`
- note: Partio supports only Claude Code and Codex via git hooks. It has no OpenCode integration, no search backend, no telemetry, and no CRLF/User-Agent concerns in its hook path.

## changelog 0.10.4: concurrent checkpoint CAS

<!-- partio:rejection:v1 -->

- source: `entireio-cli changelog 0.10.4 (Concurrent checkpoint writers no longer lose each other's work: per-ref cross-process lock and Git CAS, PR #1926, #2200)`
- reason: `irrelevant`
- note: Already filed for Partio as issue #629. Not re-filed to avoid duplicates.

## changelog 0.10.4: .entire directory security (symlink/non-regular file guard)

<!-- partio:rejection:v1 -->

- source: `entireio-cli changelog 0.10.4 (.entire must be a real directory; settings never read through a link, PR #2154, #2156)`
- reason: `irrelevant`
- note: Already filed for Partio as issue #707 (validates .partio/ directory and file types). Not re-filed to avoid duplicates.

## changelog 0.10.4: trail, mirror, enable --search-skill, and other Entire-specific features

<!-- partio:rejection:v1 -->

- source: `entireio-cli changelog 0.10.4 (trail list/show --json, async mirror creation, entire enable --search-skill, runner setup changes, various security fixes for OAuth redirect, plugin resolver path, shadow-branch cleanup, investigate fix, config dir permissions, filesystem os.Root anchors)`
- reason: `irrelevant`
- note: Partio has no trail, mirror, search skill, runner, OAuth, plugin, shadow branch, or cloud dispatch concepts. The os.Root anchor pattern is already being pursued separately via issue #707.

## changelog 0.10.5: session JSONL released after condensation

<!-- partio:rejection:v1 -->

- source: `entireio-cli changelog 0.10.5 (A session's full.jsonl is released from .entire/metadata once its content has been condensed into a checkpoint, PR #2258)`
- reason: `irrelevant`
- note: Partio stores session JSONL directly on the orphan branch during commit; there is no staging buffer in .partio/metadata/ that accumulates files after condensation.

## changelog 0.10.5: async mirror now default, cross-jurisdiction auth, checkpoint explain security

<!-- partio:rejection:v1 -->

- source: `entireio-cli changelog 0.10.5 (async_mirror_requests defaults on, git clone cross-federation fix, entire doctor session lock wait message, cluster well-known auth hardening, checkpoint explain response verification)`
- reason: `irrelevant`
- note: Partio has no mirror infrastructure, federated auth, or server-side checkpoint explanation. These are all cloud platform features.

## changelog 0.10.6: named pipe guard blocking open calls

<!-- partio:rejection:v1 -->

- source: `entireio-cli changelog 0.10.6 (named pipe where Entire expects a config file no longer hangs the process, PR #2308)`
- reason: `irrelevant`
- note: Already filed for Partio as issue #707 (which covers FIFO hang and symlink redirect for .partio/ files). Not re-filed to avoid duplicates.

## changelog 0.10.6: Windows PowerShell installer, runner setup rework, trail native-repo fixes

<!-- partio:rejection:v1 -->

- source: `entireio-cli changelog 0.10.6 (Windows PowerShell install.ps1, entire runner setup flags rework, entire trail create pre-push hooks, trail resolution for native repos, PR #2152, #2217, #2272, #2306, #2239, #2245)`
- reason: `irrelevant`
- note: Partio has no Windows installer, runner setup, or trail concept. These are all Entire cloud platform features.

## changelog 0.10.6: git worktree metadata centralization

<!-- partio:rejection:v1 -->

- source: `entireio-cli changelog 0.10.6 (git worktree metadata has one owner, gitrepo.ResolveWorktreeMetadata; checkpoint duplicate traversal deleted, PR #2241, #2254)`
- reason: `irrelevant`
- note: Partio resolves the common git dir via a single git rev-parse call in the hooks package; it has no duplicate traversal or caching layer. The architectural concern doesn't apply.

## changelog 0.10.6: inode revalidation after acquiring file lock

<!-- partio:rejection:v1 -->

- source: `entireio-cli changelog 0.10.6 (Session and checkpoint file locks revalidate the locked file's inode after acquiring, PR #2273)`
- reason: `irrelevant`
- note: Partio uses a delete-before-write pattern for the pre-commit state file rather than file locking. There is no inode revalidation concern in Partio's hook flow.

## changelog 0.10.6: doctor agent directory scan improvements and e2e/test fixes

<!-- partio:rejection:v1 -->

- source: `entireio-cli changelog 0.10.6 (entire doctor reports non-directory at agent path; scans .claude/agents/, .codex/agents; entire enable shows all 8 agents; various e2e test fixes, PR #2220, #2290, #2266, #2309, #2311)`
- reason: `irrelevant`
- note: Partio's doctor checks partio-specific hook and config paths, not agent directory layouts. Agent directory scanning and the 8-agent picker are Entire-specific features.

## Issues #2089, #2091, #2098, #2111, #2115, #2148, #2197, #2201, #2202: Entire infrastructure and test concerns

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2089–#2202 (checkpoint explain scan limit false negative, hook latency follow-up, CLI performance regression, index emptied on virtiofs, hook timeout sizing, finalizeAllTurnCheckpoints deadline, persistent-ref lock file accumulation, auth-go lock dir, test git isolation)`
- reason: `irrelevant`
- note: These issues cover: Entire's checkpoint explain command (Partio has no such command), hook and redaction latency tracking (Partio has no redaction or hook trace infrastructure), a git index race on virtiofs (not caused by Partio), per-hook timeout configuration (Partio has no per-hook timeout settings), finalizeAllTurnCheckpoints (Entire-specific async turn pipeline), persistent-ref lock file cleanup (Partio has no persistent-ref lock directory), and test isolation concerns for Entire's own test suite.

## Issue #2215: Claude Code SubagentStop correlation bug

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2215 (Claude Code SubagentStop is dropped: Entire correlates on tool_use_id, which the hook does not send)`
- reason: `irrelevant`
- note: Partio has no SubagentStop hook and no subagent correlation logic. It captures sessions through git hooks (pre-commit, post-commit), not through agent lifecycle events.

## Issue #2218: Windows PE binary version 0.0.0.0

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2218 (Windows: entire.exe reports version 0.0.0.0 in PE metadata)`
- reason: `irrelevant`
- note: Partio has no Windows build or PE metadata. Its version is embedded via ldflags in a Linux/macOS binary.

## Issue #2237: Re-install discards another tool's newer hook

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2237 (Re-install discards another tool's newer hook and chains to a stale .pre-entire backup)`
- reason: `irrelevant`
- note: Partio's hook installer (`internal/git/hooks/install.go`) always calls `os.Rename(hookPath, backupPath)` when the existing hook is not Partio's, atomically replacing the backup. It does not gate the rename on whether a backup already exists, so the stale-backup bug described does not apply.

## Issue #2249: Stale `entire configure --agent` hints in error messages

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2249 (Stale entire configure --agent hints in investigate and review error messages)`
- reason: `irrelevant`
- note: Partio has no `partio configure --agent` or `partio agent add` command. Agent selection is via the `PARTIO_AGENT` environment variable and the `agent` config field.

## Issue #2255: Forged Entire-Checkpoint lines in commit bodies

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2255 (Forged Entire-Checkpoint lines in commit bodies can select and mutate unrelated checkpoints)`
- reason: `irrelevant`
- note: Partio writes `Partio-Checkpoint:` trailers via `git commit --amend` but never reads them back. The checkpoint store is keyed by checkpoint ID on the orphan branch, not resolved from commit message trailers. The mutation vector described does not exist in Partio's architecture.

## Issue #2256: AppendCheckpointTrailer grammar divergence

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2256 (AppendCheckpointTrailer can emit a trailer the final-block parser and git reject)`
- reason: `irrelevant`
- note: Partio uses `git interpret-trailers` (or equivalent) via a subprocess to append the `Partio-Checkpoint:` trailer; it does not have its own trailer writer or parser. There is no grammar divergence between a custom writer and git.

## Issue #2257: Adding external agent writes merged settings to wrong scope

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2257 (Adding an external agent writes merged settings back to one scope, leaking values between settings.json and settings.local.json)`
- reason: `irrelevant`
- note: Partio has no external-agent registration flow. Hook installation is done by `partio enable`; the command reads and writes to the intended scope file directly without loading a merged view.

## Issue #2260: Read deny rule breaks auto/unattended mode

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2260 (Read(./.entire/metadata/**) deny rule makes ordinary commands ask for approval, breaking auto/unattended permission modes)`
- reason: `irrelevant`
- note: Partio does not install any deny rules into `.claude/settings.json` or any agent permission config. It only installs git hooks.

## Issue #2263: Lefthook not classified as overwriting hooks

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2263 (Lefthook is classified as not overwriting Entire's hooks, so users get the weaker warning)`
- reason: `irrelevant`
- note: Already filed for Partio as issue #519 (Auto-configure hook managers during partio enable). Not re-filed to avoid duplicates.

## Issue #2264: entire status reports health when hooks not installed

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2264 (entire status reports "Checkpoints sync to: origin" when git hooks are not installed)`
- reason: `irrelevant`
- note: Partio's status command already checks for hook installation (`cmd/partio/status.go` lines 64–82) and reports "Hooks: missing (run 'partio enable' to reinstall)" when hooks are absent. The gap described does not exist in Partio.

## Issue #2271: Security docs lack data-residency information

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2271 (Docs: security page has no data-residency information)`
- reason: `irrelevant`
- note: Partio stores checkpoints locally on an orphan branch; there is no cloud residency, no jurisdiction selector, and no hosted server to describe.

## Issue #2274: `entire disable` does not reach linked worktrees

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2274 (entire disable does not reach linked worktrees: settings.local.json is per worktree, so a new git worktree add comes back enabled)`
- reason: `irrelevant`
- note: Partio's `partio disable` removes the installed git hooks via `githooks.Uninstall()` (`cmd/partio/disable.go:39`); it does not write an `"enabled": false` flag to `settings.local.json`. Since git hooks are installed to `git rev-parse --git-common-dir` (shared across worktrees), disabling in one checkout disables in all. The per-worktree settings file leak described does not apply.

## Issue #2318: Install script hides GitHub API authentication and rate-limit errors

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2318 (Install script hides GitHub API authentication and rate-limit errors)`
- reason: `irrelevant`
- note: Partio has no shell install script that fetches release metadata from GitHub. Installation is via `go install` or `make install`.

## Issues #2350, #2360, #2362, #2368, #2378, #2393: Entire checkpoint infrastructure bugs

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2350–#2393 (shadow ref cleanup on malformed state, checkpoint push remote election, status collapses sessions, token usage overwrite during condensation, uncondensed commits made unreachable, git-refs push queue partial failure)`
- reason: `irrelevant`
- note: Partio has no shadow branches, no configurable push remote election, no multi-session status display, no token usage tracking, no condensation pipeline, and no async push queue. These issues are all specific to Entire's git-refs checkpoint backend and cloud sync infrastructure.

## Issue #2367: session resume writes to ~/.claude instead of $CLAUDE_CONFIG_DIR

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2367 (entire session resume writes to ~/.claude instead of $CLAUDE_CONFIG_DIR)`
- reason: `irrelevant`
- note: Already filed for Partio as issue #708 (FindSessionDir ignores CLAUDE_CONFIG_DIR). Not re-filed to avoid duplicates.

## Issue #2394: Documentation taps before trusting (Homebrew)

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2394 (Documentation tries to tap before trusting which is not allowed in Homebrew 6)`
- reason: `irrelevant`
- note: Partio has no Homebrew tap or cask. Installation is via `go install` or `make install`.

## PRs #2074–#2330: Entire infrastructure, cloud, agent, and test work

<!-- partio:rejection:v1 -->

- source: `entireio-cli-pulls #2074–#2330 (trail backend, OpenCode 1/2 integration, Cursor/Codex/Pi/Copilot fixes, search TUI, mirror async, token usage, shadow cleanup, worktree metadata, git status --no-optional-locks, redaction performance, telemetry, dependency bumps, E2E test fixes, Windows hook wrappers, complexity tooling, various security hardening PRs)`
- reason: `irrelevant`
- note: These PRs cover Entire's cloud platform (trails, mirrors, dispatch, federation), agent integrations (OpenCode, Cursor, Pi, Copilot, Factory AI) Partio doesn't support, search infrastructure, telemetry, token tracking, redaction, and test infrastructure. None address Partio's core domain of git-hook-based session capture and orphan-branch checkpoint storage.

## PR #2331: Fix Lefthook overwrite detection and remediation guidance

<!-- partio:rejection:v1 -->

- source: `entireio-cli-pulls #2331 (Fix Lefthook overwrite detection and remediation guidance)`
- reason: `irrelevant`
- note: Already filed for Partio as issue #519 (Auto-configure hook managers during partio enable). Not re-filed to avoid duplicates.

## PR #2375: Honor CLAUDE_CONFIG_DIR for session directory resolution

<!-- partio:rejection:v1 -->

- source: `entireio-cli-pulls #2375 (Honor each agent's home relocation variable when resolving session dirs)`
- reason: `irrelevant`
- note: Already filed for Partio as issue #708 (FindSessionDir ignores CLAUDE_CONFIG_DIR). Not re-filed to avoid duplicates.

## PR #2383: Stop condensation from resetting session token totals

<!-- partio:rejection:v1 -->

- source: `entireio-cli-pulls #2383 (fix(strategy): stop condensation from resetting session token totals)`
- reason: `irrelevant`
- note: Partio has no token usage tracking or condensation strategy that accumulates token counts across checkpoints.

## PR #2385: Fail closed before deleting shadow history

<!-- partio:rejection:v1 -->

- source: `entireio-cli-pulls #2385 (fix(cleanup): fail closed before deleting shadow history)`
- reason: `irrelevant`
- note: Partio has no shadow branches. Checkpoint commits are made directly to the orphan branch; there is no per-session shadow ref to clean up.

## PR #2388: Register with Lefthook instead of owning hook files

<!-- partio:rejection:v1 -->

- source: `entireio-cli-pulls #2388 (fix(hooks): register Entire with Lefthook instead of owning its hook files)`
- reason: `irrelevant`
- note: Already filed for Partio as issue #519 (Auto-configure hook managers during partio enable). Not re-filed to avoid duplicates.

## PR #2395: Surface git-refs push queue failures

<!-- partio:rejection:v1 -->

- source: `entireio-cli-pulls #2395 (fix(checkpoint): surface git-refs push queue failures)`
- reason: `irrelevant`
- note: Partio's pre-push hook pushes the checkpoint branch synchronously via a single git push call. There is no async push queue; the partial-failure scenario does not apply.

## PR #2397: OpenCode 2 CLI and plugin API support

<!-- partio:rejection:v1 -->

- source: `entireio-cli-pulls #2397 (fix(opencode): support the OpenCode 2 CLI and plugin API)`
- reason: `irrelevant`
- note: Partio has no OpenCode integration.

## PRs #2332–#2397 (remaining): Windows hooks, E2E improvements, deps, trail, misc

<!-- partio:rejection:v1 -->

- source: `entireio-cli-pulls #2332–#2397 (session adopt, complexity CI gate, trail runner spike, checkpoint remote election, git metadata migration, GitHub install script errors, checkpoint HTTPS auth, git common dir strip, symbolic ref fix, empty-tree commit detection, differentiated hook timeouts, generic git provider for --checkpoint-remote, OPF progress, OpenCode session attach, Copilot/Pi fixes, warn on no active session, Windows plugin install, E2E improvements, cursor model pinning, E2E task regex, fetch URL requirement for sync remote, status strategy name drop, shadow PR stubs, dependency bumps)`
- reason: `irrelevant`
- note: These PRs cover Entire's cloud platform, Windows-specific hook wrappers, agent integrations Partio doesn't support (OpenCode, Copilot, Pi, Cursor), OPF/redaction pipeline, and test infrastructure improvements. The empty-tree commit detection (PR #2303) and "warn on no active session" (PR #2294) are analogous to ideas already filed as Partio issues #664 and #650 respectively.
