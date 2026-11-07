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

## Case-insensitive filesystem: session stays pending when agent spells path in another case

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #3054 (On macOS and Windows, a session stays pending after its commit when the agent spells a tracked path in another case)`
- reason: `irrelevant`
- note: Partio does not track which files an agent touched between commits. The post-commit hook captures the session JSONL and git diff; it does not maintain a per-session file list that is compared against commit contents. The pending-session / carry-forward mechanism that caused this bug does not exist in Partio's architecture.

## Windows: track agent process liveness (proclive)

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #3053 (Track agent process liveness on Windows (proclive))`
- reason: `irrelevant`
- note: Partio's pre-commit hook detects agent presence by scanning running processes (pgrep-style) at hook time. It does not keep a live session ownership record that requires OS-specific liveness polling between hooks. The absence of a Windows liveness implementation does not block Partio's simpler detection model.

## Windows: cancelling a subprocess at deadline leaves child processes running

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #3052 (Windows: cancelling a subprocess at its deadline leaves its child processes running)`
- reason: `irrelevant`
- note: Partio does not launch or manage agent subprocesses. Hooks call `partio _hook <name>` synchronously; there is no deadline-driven cancellation of agent processes in Partio's codebase.

## Hook state reads Current while a moved copy of Entire's hook is stale

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #3050 (Hook state reads Current while a moved copy of Entire's hook is stale, so agent turns never refresh it)`
- reason: `irrelevant`
- note: Partio has no "Current" live hook state channel. The only hook state is the pre-commit.json file written by pre-commit and deleted by post-commit. A stale hook script can call a stale binary, but there is no equivalent to Entire's real-time session-state socket that this issue describes.

## Hook backup copies always set 0755 on no-hard-link filesystems

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #3049 (Backing up a hook without hard links makes it executable (0755), so a hook switched off with chmod -x runs again)`
- reason: `irrelevant`
- note: Partio backs up existing hooks via `os.Rename` (`internal/git/hooks/install.go:41`), not via a file copy. A rename is an atomic inode move that preserves the original permissions. The 0755 copy-permissions bug only arises on the copy path; Partio does not use that path.

## Web app: default Markdown files to Preview in Trails → Changed Files

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #3032 (Web app: default Markdown files to Preview in Trails → Changed Files)`
- reason: `irrelevant`
- note: Partio has no web app.

## Overlapping commits in one worktree can consume each other's stamp records

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #3030 (Overlapping commits in one worktree can consume each other's stamp records)`
- reason: `irrelevant`
- note: Partio has no stamp-record mechanism for commits. The pre-commit state file is keyed to a single concurrent operation and is deleted immediately by post-commit. Multiple simultaneous commits in one worktree are not a scenario Partio's architecture creates.

## Commit keeps trailer naming unwritten checkpoint when subagent finishes during commit

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #3029 (Commit keeps a trailer naming an unwritten checkpoint when a session's subagent finishes, or its turn ends, during the commit)`
- reason: `irrelevant`
- note: Partio has no subagent tracking and no mechanism for injecting trailers for in-progress sub-sessions. Partio's trailer is written by post-commit for the commit that just finished; there is no background subagent path.

## `entire enable` hangs after agreeing to commit when a pre-commit hook opens less

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #3025 (entire enable hangs after agreeing to commit when a pre-commit hook opens less)`
- reason: `irrelevant`
- note: Partio's `enable` command creates the checkpoint branch using git plumbing commands (`hash-object`, `mktree`, `commit-tree`, `update-ref`) that bypass hooks entirely (`cmd/partio/enable.go:192–214`). No `git commit` is run, so no pre-commit hook fires and no pager can open.

## Checkpoint replay reports success after dropping another clone's session

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #3014 (Checkpoint replay reports success after dropping another clone's session)`
- reason: `irrelevant`
- note: Partio has no server-side checkpoint replay. Checkpoints are written locally by the post-commit hook; replay is not a Partio concept.

## Sessions in a shared checkout claim every dirty tracked file

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #3011 (Sessions in a shared checkout claim every dirty tracked file, so idle and ended sessions land on unrelated commits)`
- reason: `irrelevant`
- note: Partio does not maintain a per-session file list. It detects agent presence at pre-commit time and links the whole commit to the session. There is no cross-session dirty-file tracking or contention.

## Pushing to the default branch through a GitHub mirror fails with remote rejected

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #3000 (Pushing to the default branch through a GitHub mirror fails with remote rejected (protected branch))`
- reason: `irrelevant`
- note: Partio has no GitHub mirror infrastructure. Checkpoints are pushed to whatever `origin` the repo already has.

## `entire review` from a git worktree: "worktree path mismatch"

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2990 (entire review (output: local) does not persist findings when run from a git worktree)`
- reason: `irrelevant`
- note: Partio has no review command.

## Owner access for auto-created orgs

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2989 (Owner access for auto-created orgs misty-step and r90group)`
- reason: `irrelevant`
- note: Administrative/billing issue unrelated to Partio's domain.

## Windows: batch checkpoint ref push fails once queue exceeds ~340 refs

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2986 (Windows: the batch checkpoint ref push fails once the queue passes ~340 refs)`
- reason: `irrelevant`
- note: Partio pushes a single checkpoint branch (`partio/checkpoints/v1`) with one `git push` call. It has no batch ref push queue or Windows-specific push path.

## git-remote-entire: push uses ~4× pack size in memory; replica cache grows without bound

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2985 (git-remote-entire: push uses ~4× the pack size in memory; replica cache grows without bound)`
- reason: `irrelevant`
- note: Partio has no git remote helper. Checkpoints are pushed with a plain `git push` call.

## First preset Agent run completed but no promotional credit was granted

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2981 (First preset Agent run completed but no $100 promotional credit was granted)`
- reason: `irrelevant`
- note: Billing/promotional issue unrelated to Partio.

## Post-push session attach publishes linked_commits but checkpoint remains absent

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2980 (Post-push session attach publishes linked_commits but checkpoint remains absent from activity and Dispatch)`
- reason: `irrelevant`
- note: Partio has no session attach command and no cloud activity/dispatch system.

## Sibling-worktree commit linking picks a session idle for days

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2979 (Sibling-worktree commit linking picks a session idle for days; every commit from another worktree then condenses its multi-GB transcript)`
- reason: `irrelevant`
- note: Partio's session-to-commit linking is based on detecting any running agent process at pre-commit time, not on cross-worktree session ownership records. The idle-session hijacking and transcript condensation described are specific to Entire's complex multi-session state machine.

## PII address redaction misses European street addresses

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2977 (PII address redaction misses European street addresses)`
- reason: `irrelevant`
- note: Partio has no address PII redaction. The `internal/redact/` package uses entropy-based heuristics and known secret patterns; it has no PII address detection at all.

## PII phone redaction misses every non-North American number

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2976 (PII phone redaction misses every non-North American number)`
- reason: `irrelevant`
- note: Partio has no phone PII redaction. The `internal/redact/` package uses entropy and known secret patterns; it has no phone number detection.

## `session attach` drops the signature of every commit it rewrites

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2974 (session attach drops the signature of every commit it rewrites)`
- reason: `irrelevant`
- note: Partio has no session attach command.

## Re-wrapping a pre-push hook that chains through Entire's moved copy rotates the user's hook out

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2960 (Re-wrapping a pre-push hook that chains through Entire's moved copy rotates the user's own hook out of the chain)`
- reason: `irrelevant`
- note: Partio's hook installation backs up the original hook with `os.Rename` and writes a Partio wrapper. There is no second-level re-wrap cycle or entire-brain style chaining. The re-wrapping race condition described does not exist in Partio's simpler install model.

## `entire disable --uninstall` leaves Entire's pre-push hook running behind entire-brain's wrapper

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2957 (\`entire disable --uninstall\` leaves Entire's pre-push hook running behind entire-brain's wrapper)`
- reason: `irrelevant`
- note: Partio has no entire-brain plugin. `partio disable` uninstalls Partio's hooks and restores backups directly; there is no secondary wrapper layer to leave behind.

## Safer defaults for public remotes: checkpoint push on by default; redaction gaps

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2951 (Safer defaults for public remotes: checkpoint push is on by default; Bearer headers and binary transcripts slip past redaction)`
- reason: `irrelevant`
- note: Partio's `push_sessions` defaults to `true` and the concern about public remotes is already filed as Partio issue #802. The Bearer-header and binary-transcript redaction gaps are specific to Entire's backend transport; Partio has no Bearer-authenticated remote helper or binary transcript format.

## Doctor conflict with entire-brain setup

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2949 (Entite doctor conflict with brain setup)`
- reason: `irrelevant`
- note: Partio has no entire-brain plugin. `partio doctor` checks hook installation and settings; there is no brain-integration path to conflict with.

## Show transcript images (screenshots) in the entire.com session timeline

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2942 (Show transcript images (screenshots) in the entire.com session timeline)`
- reason: `irrelevant`
- note: Partio has no web timeline or image display. Session transcripts are stored as JSONL on the orphan branch; there is no rendering layer.

## Windows: large Codex transcript saturates CPU during post-commit (12m43s)

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2933 (Windows: large Codex transcript saturates CPU during post-commit (12m43s), interrupted capture leaves checkpoint trailer without remote data)`
- reason: `irrelevant`
- note: Partio's Codex session parser is simpler than Entire's. The 12-minute CPU saturation was caused by Entire's transcript condensation pipeline, which Partio does not have. Partio reads and redacts the JSONL once per commit; it has no condensation or multi-pass rewriting step.

## Configure Trail gates as code (per repository and per organization)

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2932 (Configure Trail gates as code (per repository and per organization))`
- reason: `irrelevant`
- note: Partio has no trails or gates.

## Docs: Show Bash and PowerShell install commands on the CLI feature page

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2931 (docs: Show Bash and PowerShell install commands on the CLI feature page)`
- reason: `irrelevant`
- note: Documentation update for Entire's website; not applicable to Partio.

## Pi extension: `ENTIRE_PI_NESTED` survives /reload, /new and /resume

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2930 (Pi extension: \`ENTIRE_PI_NESTED\` survives /reload, /new and /resume; unhandled stdin \`error\` can crash Pi)`
- reason: `irrelevant`
- note: Partio has no Pi agent support.

## OpenCode v2 fails to load the Entire plugin: v1 plugin format

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2928 (OpenCode v2 fails to load the Entire plugin: v1 plugin format (no default export))`
- reason: `irrelevant`
- note: Partio has no OpenCode support.

## Windows: background Entire hooks repeatedly flash black windows and steal focus

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2926 (Windows: background Entire hooks repeatedly flash black windows and steal focus, blocking IDE use)`
- reason: `irrelevant`
- note: Partio's hooks are synchronous shell scripts that run in the terminal where git was invoked. Partio has no background hook infrastructure on Windows that would open separate windows.

## No CLI command returns why an Entire Gates check failed

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2924 (No CLI command returns why an Entire Gates check failed (trail show omits gate data))`
- reason: `irrelevant`
- note: Partio has no trails or gates.

## Capture every agent session without agent hooks

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2918 (Capture every agent session without agent hooks)`
- reason: `irrelevant`
- note: Partio's capture model is git-hook-based: pre-commit detects the agent, post-commit writes the checkpoint. Capturing sessions without git hooks would require a background daemon or agent-specific integration point that does not exist in Partio's codebase.

## `import claude-code`: turns that commit with "git commit -q" are anchored to the branch tip

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2908 (import claude-code: turns that commit with "git commit -q" are anchored to the branch tip)`
- reason: `irrelevant`
- note: Partio has no session import command.

## Subagent task record skipped when the subagent commits its own work

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2903 (Subagent task record skipped when the subagent commits its own work ("no file changes detected"))`
- reason: `irrelevant`
- note: Partio has no subagent tracking.

## `checkpoint explain`: 120s foreground fetch timeout turns sidecar fetch into infinitely repeated work

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2902 (checkpoint explain: 120s foreground fetch timeout turns the sidecar metadata fetch into infinitely repeated discarded work)`
- reason: `irrelevant`
- note: Partio's `partio explain` command calls an external LLM API; it does not perform git fetches or sidecar metadata lookups. There is no equivalent foreground-fetch timeout path.

## Post-commit background fetch leaves a 115 MB `tmp_pack_*` file after every commit

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2895 (Post-commit background fetch leaves a 115 MB \`tmp_pack_*\` file after every commit)`
- reason: `irrelevant`
- note: Partio's post-commit hook does not perform any git fetch operations. There is no background fetch and no partial pack file left behind.

## Documented way to create service accounts / API keys for non-interactive CLI auth

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2894 (Documented way to create service accounts / API keys for non-interactive CLI auth)`
- reason: `irrelevant`
- note: Partio has no cloud authentication. It operates entirely locally.

## Codex Windows hooks: emit native commandWindows and preserve it across managed rewrites

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2884 (Codex Windows hooks: emit native commandWindows and preserve it across managed rewrites)`
- reason: `irrelevant`
- note: Partio's Codex support uses standard git hooks (`pre-commit`, `post-commit`, `pre-push`), not Codex's own hook configuration system. Partio does not write to `.codex/hooks.json` or manage Windows-specific command formats.

## Add goose support

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2812 (Add goose support)`
- reason: `irrelevant`
- note: Partio supports Claude Code and Codex. Adding goose would require a new agent detector and JSONL parser for goose's transcript format. No goose session format is documented in this codebase to base a premise on; this would be a wholly new integration, not an adaptation of an existing feature.

## Support OpenCode 2.0

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2810 (Support OpenCode 2.0)`
- reason: `irrelevant`
- note: Partio has no OpenCode support. Adding it would require a new agent detector and session parser; nothing in the codebase handles OpenCode sessions.

## Bare AWS access key IDs are not redacted without a paired secret

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2808 (Bare AWS access key IDs are not redacted without a paired secret)`
- reason: `irrelevant`
- note: Partio's `aws-access-key-id` pattern in `internal/redact/patterns.go` matches the key-ID prefix (`AKIA|ASIA|AROA|AIDA|ANPA|ANVA|APKA`) followed by 16 uppercase base-32 characters, with no requirement for a paired secret access key. A bare key ID is caught by Partio's pattern without needing Entire's betterleaks component rule.

## OPF pre-push batching leaves CR intact, causing multiple JSON documents for CRLF input

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2806 (OPF pre-push batching leaves CR intact, causing multiple JSON documents for CRLF input)`
- reason: `irrelevant`
- note: Partio has no OPF (Off-Platform Federation) infrastructure.

## Remote helper: fall back to login when a credential helper returns a non-Entire JWT

<!-- partio:rejection:v1 -->

- source: `entireio-cli-issues #2748 (Remote helper: fall back to login when a credential helper returns a non-Entire JWT)`
- reason: `irrelevant`
- note: Partio has no git remote helper. Checkpoint pushes use a plain `git push` call without custom credential handling.

## Changelog 0.10.3–0.10.6: subagents, identity-based linking, search, telemetry, cloud repo commands, Windows installer

<!-- partio:rejection:v1 -->

- source: `entireio-cli changelog 0.10.3–0.10.6 (subagent checkpoint durable storage, identity-based commit linking via process ancestry, zombie session self-heal, redaction engine configuration, doctor trace summary, search TUI, telemetry signals, git-status --no-optional-locks, entire repo clone, dispatch jurisdiction, Windows PowerShell installer, runner setup, trail create pre-push hooks)`
- reason: `irrelevant`
- note: Subagent durable checkpoints, zombie-session healing, and identity-based commit linking via process fingerprinting are all specific to Entire's multi-session state machine; Partio has no equivalent. Redaction engine selection (betterleaks/goredact flags) requires a pluggable engine architecture not present in Partio. The search TUI, telemetry, cloud repo commands, Windows installer, and runner setup are Entire-specific features with no analogue in Partio's local-only CLI. The git-status --no-optional-locks fix applies to Entire's internal git-status calls; Partio does not call `git status` in its hooks.

## Changelog 0.11.0–0.11.6: control-plane rename, trails, shadow-branch removal, Gemini removal, Antigravity, checkpoint delete, attribution removal, review profiles, OpenCode session edits

<!-- partio:rejection:v1 -->

- source: `entireio-cli changelog 0.11.0–0.11.6 (control-plane command renames, enable no longer creates GitHub remote, trail commands, shadow-branch removal, Gemini CLI removal, Antigravity CLI integration, checkpoint delete command, line attribution removal, review profile per-reviewer config, OpenCode subagents as sub-sessions, session edits in another worktree, Claude Code StopFailure hook)`
- reason: `irrelevant`
- note: Control-plane commands and cloud-remote management are Entire-specific. Shadow branches are an Entire concept Partio never had. Gemini/Antigravity/OpenCode/Pi are agents Partio does not support. Checkpoint delete, the cloud-side review profile system, and the trail command suite have no equivalent in Partio's local orphan-branch storage model. Claude Code StopFailure hook is an Entire-specific agent hook event type that Partio's hook configuration does not use.

## Changelog 0.12.0: trails as default, checkpoint delete, attribution removal, session resume, session ID validation

<!-- partio:rejection:v1 -->

- source: `entireio-cli changelog 0.12.0 (project trails as default, checkpoint delete with --dry-run, line attribution removed, session resume across agent homes, session ID path validation, concurrent OPF ref update fix, squashed commit trailer inheritance, \`entire enable\` no longer pushes GitHub remote)`
- reason: `irrelevant`
- note: Project trails, checkpoint delete, and cloud session resume are Entire-specific. Line attribution removal affects a data model (AI/Human/Mixed labels) that Partio replaced with binary 0%/100% attribution from the start. Session ID path validation is specific to Entire's server-side session store. Squashed commit trailer inheritance and concurrent OPF ref fixes depend on Entire's OPF backend.

## Changelog 0.12.1: checkpoint list --outgoing, PII IP redaction, redaction of *_API_KEY env vars, checkpoint explain shows redaction counts

<!-- partio:rejection:v1 -->

- source: `entireio-cli changelog 0.12.1 (\`entire checkpoint list --outgoing\`, IP address PII redaction category, redaction of *_CLIENT_SECRET/*_API_KEY/*_TOKEN env var assignments and piped credential commands, checkpoint explain shows redaction counts, checkpoint_ref_list_timeout_seconds, Factory AI Droid token usage, entire brain install prompt)`
- reason: `irrelevant`
- note: `checkpoint list --outgoing` is for Entire's git-refs backend, which tracks per-ref push state. Partio pushes a single orphan branch with no per-ref queue. PII IP address redaction is a new PII category; Partio has no PII redaction at all. Redaction of `*_CLIENT_SECRET`/`*_API_KEY` env var assignments is already filed for Partio as issue #796. Factory AI Droid and entire-brain are agents/plugins Partio does not support. The checkpoint_ref_list_timeout_seconds setting and redaction-count metadata are backend-specific.

## Changelog 0.12.2: OpenCode image externalization, trail/review command changes, first-turn context injection, idle-session commit condensation

<!-- partio:rejection:v1 -->

- source: `entireio-cli changelog 0.12.2 (Pi images in assets/, first-turn context injected into every session, trail comment commands refuse agent sessions, idle session condensed at commit for stacked worktrees)`
- reason: `irrelevant`
- note: Pi image externalization requires a Pi agent and an image-externalisation pipeline; Partio has neither. First-turn context injection via `entire agent-help` is Entire's cloud-coaching feature; Partio injects nothing at session start. Trail comment restrictions and idle-session stacked-worktree condensation are Entire-specific.

## Changelog 0.13.1: OpenCode 2 support, trail commands, project agent runs, service accounts, checkpoint push policy, partial pack cleanup

<!-- partio:rejection:v1 -->

- source: `entireio-cli changelog 0.13.1 (OpenCode 2 plugin, new trail commands, project agent run list/show/cancel/retry, service account management, partial pack cleanup after failed checkpoint fetch, sibling-worktree session ownership fix, case-insensitive path fix, Trunk hook chaining, session liveness on Windows)`
- reason: `irrelevant`
- note: OpenCode 2 support, trail commands, and service accounts are Entire-specific. The partial pack cleanup after a failed checkpoint fetch applies to Entire's git-refs remote-fetch infrastructure; Partio has no background fetch. Sibling-worktree session ownership and case-insensitive path fixes address Entire's multi-session state machine. Trunk hook chaining and Windows liveness are Entire-specific infrastructure not present in Partio.

## PRs #2996–#3057: pre-push quarantine fix, case-insensitive path fix, shared-checkout session fix, Trunk chaining, service accounts, trail management, dependency bumps

<!-- partio:rejection:v1 -->

- source: `entireio-cli-pulls #2996–#3057 (fix pre-push quarantine race (#3044, fixes #2794), spell agent-reported paths on case-insensitive filesystems (#3056), stop shared-checkout sessions from claiming each other's files (#3055), fix Trunk launcher hook chaining (#3051), redact passwords on command line (#3048), back off between token refresh retries (#3047), keep user's pre-push hook in entire-brain's chain on re-wrap (#3046), offer to install Codex hooks in linked worktrees (#3045), trail watch (#3043), dependency bumps (#3042, #3041, #3027), fix setup commit hook-opens-pager hang (#3040), changelog PRs (#3039, #3038), git-remote-entire negotiation (#3037), service account management (#3036, #3035, #3022, #3020, #3007), trail and project commands (#3031, #3023, #3016, #3015, #3013, #3012, #3010, #3009, #3002, #3001), fix sibling-worktree idle session (#3005, #3034), fix hooks in entire-brain uninstall (#3003), and infrastructure PRs (#3033, #3028, #3026, #3024, #3021, #3019, #3018, #3006, #3004, #2999, #2998, #2997, #2996))`
- reason: `irrelevant`
- note: The pre-push quarantine fix (#3044) addresses the same race as Partio issue #778 (already filed). Password redaction (#3048) addresses the same gap as Partio issue #795 (already filed). The case-insensitive path fix, shared-checkout session fix, entire-brain hook chaining, and Trunk launcher are specific to Entire's multi-session state machine and plugin ecosystem. Service account management, trail commands, project agent commands, and cloud infrastructure PRs have no equivalent in Partio's local-only model. Dependency bumps, changelog PRs, and CI changes are Entire-internal housekeeping.
