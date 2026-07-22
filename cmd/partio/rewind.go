package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/partio-io/cli/internal/checkpoint"
	"github.com/partio-io/cli/internal/git"
)

func newRewindCmd() *cobra.Command {
	var (
		list       bool
		toID       string
		branchName string
	)

	cmd := &cobra.Command{
		Use:   "rewind",
		Short: "List or restore checkpoints",
		Long: `List all captured checkpoints or restore the repository state to a specific checkpoint.

Use --branch <name> with --list to filter checkpoints by branch name. This is
the recommended approach for finding sessions whose branch was squash-merged.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if list {
				return runRewindList(branchName)
			}
			if toID != "" {
				return runRewindTo(toID)
			}
			return cmd.Help()
		},
	}

	cmd.Flags().BoolVar(&list, "list", false, "list all checkpoints")
	cmd.Flags().StringVar(&toID, "to", "", "restore to a specific checkpoint ID")
	cmd.Flags().StringVar(&branchName, "branch", "", "filter listed checkpoints by branch name (recommended for squash-merged branches)")

	return cmd
}

func runRewindList(branchFilter string) error {
	_, err := git.RepoRoot()
	if err != nil {
		return fmt.Errorf("must be run inside a git repository")
	}

	const cpBranch = "partio/checkpoints/v1"

	// Check branch exists
	_, err = git.ExecGit("rev-parse", "--verify", cpBranch)
	if err != nil {
		return fmt.Errorf("no checkpoint branch found - run 'partio enable' first")
	}

	// List top-level shard directories
	shards, err := git.ExecGit("ls-tree", "--name-only", cpBranch)
	if err != nil || shards == "" {
		fmt.Println("No checkpoints found.")
		return nil
	}

	fmt.Println("Checkpoints:")
	fmt.Println()

	for _, shard := range strings.Split(shards, "\n") {
		// List entries within each shard
		entries, err := git.ExecGit("ls-tree", "--name-only", cpBranch+":"+shard)
		if err != nil {
			continue
		}
		for _, entry := range strings.Split(entries, "\n") {
			cpID := shard + entry
			// Try to read metadata
			metaJSON, err := git.ExecGit("show", cpBranch+":"+shard+"/"+entry+"/metadata.json")
			if err != nil {
				if branchFilter == "" {
					fmt.Printf("  %s  (metadata unavailable)\n", cpID)
				}
				continue
			}

			var meta checkpoint.Metadata
			if err := json.Unmarshal([]byte(metaJSON), &meta); err != nil {
				if branchFilter == "" {
					fmt.Printf("  %s  (invalid metadata)\n", cpID)
				}
				continue
			}

			if branchFilter != "" && meta.Branch != branchFilter {
				continue
			}

			fmt.Printf("  %s  branch=%s  agent=%d%%  created=%s\n",
				cpID, meta.Branch, meta.AgentPercent, meta.CreatedAt)
		}
	}

	return nil
}

func runRewindTo(id string) error {
	repoRoot, err := git.RepoRoot()
	if err != nil {
		return fmt.Errorf("must be run inside a git repository")
	}

	if len(id) != 12 {
		return fmt.Errorf("checkpoint ID must be 12 characters (got %d)", len(id))
	}

	const cpBranch = "partio/checkpoints/v1"
	shard := id[:2]
	rest := id[2:]

	// Read metadata
	metaJSON, err := git.ExecGit("show", cpBranch+":"+shard+"/"+rest+"/metadata.json")
	if err != nil {
		return fmt.Errorf("checkpoint %s not found", id)
	}

	var meta checkpoint.Metadata
	if err := json.Unmarshal([]byte(metaJSON), &meta); err != nil {
		return fmt.Errorf("invalid checkpoint metadata: %w", err)
	}

	// Read session context
	context, _ := git.ExecGit("show", cpBranch+":"+shard+"/"+rest+"/0/context.md")

	fmt.Printf("Rewinding to checkpoint %s\n", id)
	fmt.Printf("  Commit: %s\n", meta.CommitHash)
	fmt.Printf("  Branch: %s\n", meta.Branch)
	if context != "" {
		fmt.Printf("  Context:\n%s\n", context)
	}

	// Check whether the original commit is still reachable (it may not be after
	// a squash-merge followed by branch deletion and/or gc).
	reachable, err := git.CommitReachable(repoRoot, meta.CommitHash)
	if err != nil {
		return fmt.Errorf("checking commit reachability: %w", err)
	}

	if !reachable {
		shortSHA := meta.CommitHash
		if len(shortSHA) > 8 {
			shortSHA = shortSHA[:8]
		}
		fmt.Printf("Warning: original commit %s is no longer reachable (likely squash-merged or gc'd); session context is still available but branch checkout is skipped.\n", shortSHA)
		return nil
	}

	// Create a new branch at the checkpoint's commit
	branchName := fmt.Sprintf("partio/rewind/%s", id)
	_, err = git.ExecGit("checkout", "-b", branchName, meta.CommitHash)
	if err != nil {
		return fmt.Errorf("creating rewind branch: %w", err)
	}

	fmt.Printf("  Created branch: %s\n", branchName)
	return nil
}
