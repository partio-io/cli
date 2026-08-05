package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/partio-io/cli/internal/checkpoint"
	"github.com/partio-io/cli/internal/git"
)

func newResumeCmd() *cobra.Command {
	var (
		printFlag      bool
		copyFlag       bool
		branchFlag     bool
		branchNameFlag string
	)

	cmd := &cobra.Command{
		Use:   "resume [<checkpoint-id>]",
		Short: "Resume a session from a checkpoint",
		Long: `Read checkpoint data from the orphan branch and launch a new Claude Code session with the previous context.

Use --branch-name to resume the most recent session from a named branch. This is the recommended
approach for sessions whose branch was squash-merged into main.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 && branchNameFlag == "" {
				return fmt.Errorf("accepts 1 arg(s), received 0\nUsage: partio resume <checkpoint-id>\n       partio resume --branch-name <branch>")
			}
			if branchNameFlag != "" {
				return runResumeByBranch(branchNameFlag, printFlag, copyFlag, branchFlag)
			}
			return runResume(args[0], printFlag, copyFlag, branchFlag)
		},
	}

	cmd.Flags().BoolVar(&printFlag, "print", false, "print the composed context prompt to stdout")
	cmd.Flags().BoolVar(&copyFlag, "copy", false, "copy the context prompt to clipboard")
	cmd.Flags().BoolVar(&branchFlag, "branch", false, "create a branch at the checkpoint's commit before launching")
	cmd.Flags().StringVar(&branchNameFlag, "branch-name", "", "resume the most recent session from the named branch (recommended for squash-merged work)")

	return cmd
}

// runResumeByBranch finds the most recent checkpoint for the given branch and resumes it.
func runResumeByBranch(branchName string, printFlag, copyFlag, branchFlag bool) error {
	repoDir, err := git.RepoRoot()
	if err != nil {
		return fmt.Errorf("must be run inside a git repository")
	}

	checkpoints, err := checkpoint.FindByBranch(repoDir, branchName)
	if err != nil {
		return fmt.Errorf("looking up checkpoints for branch %q: %w", branchName, err)
	}

	cp, err := pickMostRecentCheckpoint(checkpoints)
	if err != nil {
		return fmt.Errorf("no checkpoints found for branch %q", branchName)
	}

	return runResume(cp.ID, printFlag, copyFlag, branchFlag)
}

// pickMostRecentCheckpoint returns the checkpoint with the latest CreatedAt from the slice.
// Returns an error if the slice is empty.
func pickMostRecentCheckpoint(checkpoints []checkpoint.Checkpoint) (checkpoint.Checkpoint, error) {
	if len(checkpoints) == 0 {
		return checkpoint.Checkpoint{}, fmt.Errorf("no checkpoints")
	}
	sort.SliceStable(checkpoints, func(i, j int) bool {
		return checkpoints[i].CreatedAt.After(checkpoints[j].CreatedAt)
	})
	return checkpoints[0], nil
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
