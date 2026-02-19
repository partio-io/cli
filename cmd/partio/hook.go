package main

import (
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"

	"github.com/partio-io/cli/internal/hooks"
)

func newHookCmd() *cobra.Command {
	return &cobra.Command{
		Use:    "_hook",
		Short:  "Internal: called by git hooks",
		Hidden: true,
		Args:   cobra.MinimumNArgs(1),
		RunE:   runHook,
	}
}

func runHook(cmd *cobra.Command, args []string) error {
	hookName := args[0]
	slog.Debug("hook invoked", "name", hookName)

	if !cfg.Enabled {
		slog.Debug("partio disabled, skipping hook")
		return nil
	}

	runner, err := hooks.NewRunner(cfg)
	if err != nil {
		return fmt.Errorf("initializing hook runner: %w", err)
	}

	switch hookName {
	case "pre-commit":
		return runner.PreCommit()
	case "post-commit":
		return runner.PostCommit()
	case "pre-push":
		return runner.PrePush()
	default:
		return fmt.Errorf("unknown hook: %s", hookName)
	}
}
