---
id: shell-completions
target_repos:
  - cli
acceptance_criteria:
  - "partio completion bash/zsh/fish/powershell generates valid completion scripts"
  - "Completions include all commands and flags"
  - "Hidden completion command is registered in the root command"
  - "make test passes"
  - "make lint passes"
pr_labels:
  - minion
---

# Add shell completion generation command

Add a `partio completion` command that generates shell completion scripts for bash, zsh, fish, and powershell. Cobra has built-in support for this via `GenBashCompletion`, `GenZshCompletion`, etc.

The command should be hidden (not shown in help) but functional, following the standard pattern used by other Cobra-based CLIs.

## Context hints

- `cmd/partio/` — CLI command registration
- Cobra's built-in completion generation: `cobra.Command.GenBashCompletion()`, etc.

## Agents

### implement

```capabilities
max_turns: 100
checks: true
retry_on_fail: true
retry_max_turns: 20
```
