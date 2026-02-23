package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/partio-io/cli/internal/checkpoint"
	"github.com/partio-io/cli/internal/git"
)

func newResumeCmd() *cobra.Command {
	var (
		printFlag bool
		copyFlag  bool
		branchFlag bool
	)

	cmd := &cobra.Command{
		Use:   "resume <checkpoint-id>",
		Short: "Resume a session from a checkpoint",
		Long:  `Read checkpoint data from the orphan branch and launch a new Claude Code session with the previous context.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runResume(args[0], printFlag, copyFlag, branchFlag)
		},
	}

	cmd.Flags().BoolVar(&printFlag, "print", false, "print the composed context prompt to stdout")
	cmd.Flags().BoolVar(&copyFlag, "copy", false, "copy the context prompt to clipboard")
	cmd.Flags().BoolVar(&branchFlag, "branch", false, "create a branch at the checkpoint's commit before launching")

	return cmd
}

func runResume(id string, printFlag, copyFlag, branchFlag bool) error {
	_, err := git.RepoRoot()
	if err != nil {
		return fmt.Errorf("must be run inside a git repository")
	}

	data, err := checkpoint.Read(id)
	if err != nil {
		return err
	}

	if branchFlag {
		branchName := fmt.Sprintf("partio/resume/%s", id)
		_, err := git.ExecGit("checkout", "-b", branchName, data.Metadata.CommitHash)
		if err != nil {
			return fmt.Errorf("creating resume branch: %w", err)
		}
		fmt.Printf("Created branch: %s\n", branchName)
	}

	prompt := composePrompt(id, data)

	if printFlag {
		fmt.Print(prompt)
		return nil
	}

	if copyFlag {
		return copyToClipboard(prompt)
	}

	return launchClaude(id, prompt)
}

func composePrompt(id string, data *checkpoint.CheckpointData) string {
	meta := data.Metadata

	plan := data.Plan
	if plan == "" {
		plan = "No plan was recorded."
	}

	diff := data.Diff
	if diff == "" {
		diff = "No diff was recorded."
	}

	prompt := data.Prompt
	if prompt == "" && data.Context != "" {
		prompt = data.Context
	}
	if prompt == "" {
		prompt = "(No prompt was recorded.)"
	}

	return fmt.Sprintf(`# Previous Session Context

You are continuing work from a previous Partio session (checkpoint %s).

## Original Request

%s

## Plan

%s

## Changes Made

%s

## Session Info

- **Branch:** %s
- **Commit:** %s
- **Date:** %s
- **Agent:** %s (%d%%)

---

Please review the current state of the repository and continue this work.
`, id, prompt, plan, diff, meta.Branch, meta.CommitHash, meta.CreatedAt, meta.Agent, meta.AgentPercent)
}

func copyToClipboard(text string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("pbcopy")
	case "linux":
		cmd = exec.Command("xclip", "-selection", "clipboard")
	default:
		return fmt.Errorf("clipboard not supported on %s", runtime.GOOS)
	}

	cmd.Stdin = strings.NewReader(text)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("copying to clipboard: %w", err)
	}

	fmt.Println("Context prompt copied to clipboard.")
	return nil
}

func launchClaude(id, prompt string) error {
	claudePath, err := exec.LookPath("claude")
	if err != nil {
		fmt.Println("Claude Code not found in PATH. Printing context instead:")
		fmt.Println()
		fmt.Print(prompt)
		return nil
	}

	// Write context file to temp directory
	contextFile := filepath.Join(os.TempDir(), "partio-resume-"+id+".md")
	if err := os.WriteFile(contextFile, []byte(prompt), 0o644); err != nil {
		return fmt.Errorf("writing context file: %w", err)
	}

	fmt.Printf("Context written to %s\n", contextFile)
	fmt.Println("Launching Claude Code...")

	// Replace this process with claude
	args := []string{
		"claude",
		fmt.Sprintf("Read %s for full context on a previous session, then continue that work.", contextFile),
	}
	return syscall.Exec(claudePath, args, os.Environ())
}
