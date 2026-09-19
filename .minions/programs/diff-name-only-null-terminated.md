---
id: diff-name-only-null-terminated
target_repos: [cli]
acceptance_criteria:
  - DiffNameOnly passes -z to git diff and splits output on \x00 instead of \n
  - DiffNameOnly returns the real filename (e.g. "café.go") for a commit that adds a non-ASCII path, not its C-escaped form ("caf\303\251.go")
  - A table-driven test covers a commit adding a file whose name contains non-ASCII bytes; DiffNameOnly returns the literal name
pr_labels: [bug]
---

`DiffNameOnly` (`internal/git/diff_name_only.go`) calls `git diff --name-only from to` without the `-z` flag and splits the output on `"\n"`. Git's default configuration (`core.quotePath=true`) C-escapes non-ASCII bytes in filenames and wraps them in double quotes, so a file named `café.go` appears in the output as `"caf\303\251.go"`. The current split-on-newline logic returns that escaped form rather than the real path.

The function is currently used only in a debug-log path (`internal/hooks/postcommit.go:94–105`, guarded by `slog.LevelDebug`), so today the corruption is confined to log output. The function is a public API, however, and any future caller that uses the returned paths for session matching, file attribution, or filtering would silently produce wrong results on repositories containing non-ASCII filenames.

The fix is to add `-z` to the command and split the output on `"\x00"` instead of `"\n"`. With `-z`, git emits each path as a raw null-terminated byte sequence regardless of `core.quotePath`.
