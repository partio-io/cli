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
		list bool
		toID string
	)

	cmd := &cobra.Command{
		Use:   "rewind",
		Short: "List or restore checkpoints",
		Long:  `List all captured checkpoints or restore the repository state to a specific checkpoint.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if list {
				return runRewindList()
			}
			if toID != "" {
				return runRewindTo(toID)
			}
			return cmd.Help()
		},
	}

	cmd.Flags().BoolVar(&list, "list", false, "list all checkpoints")
	cmd.Flags().StringVar(&toID, "to", "", "restore to a specific checkpoint ID")

	return cmd
}

func runRewindList() error {
	_, err := git.RepoRoot()
	if err != nil {
		return fmt.Errorf("must be run inside a git repository")
	}

	const branch = "partio/checkpoints/v1"

	// Check branch exists
	_, err = git.ExecGit("rev-parse", "--verify", branch)
	if err != nil {
		return fmt.Errorf("no checkpoint branch found - run 'partio enable' first")
	}

	// List top-level shard directories
	shards, err := git.ExecGit("ls-tree", "--name-only", branch)
	if err != nil || shards == "" {
		fmt.Println("No checkpoints found.")
		return nil
	}

	fmt.Println("Checkpoints:")
	fmt.Println()

	for _, shard := range strings.Split(shards, "\n") {
		// List entries within each shard
		entries, err := git.ExecGit("ls-tree", "--name-only", branch+":"+shard)
		if err != nil {
			continue
		}
		for _, entry := range strings.Split(entries, "\n") {
			cpID := shard + entry
			// Try to read metadata
			metaJSON, err := git.ExecGit("show", branch+":"+shard+"/"+entry+"/metadata.json")
			if err != nil {
				fmt.Printf("  %s  (metadata unavailable)\n", cpID)
				continue
			}

			var meta checkpoint.Metadata
			if err := json.Unmarshal([]byte(metaJSON), &meta); err != nil {
				fmt.Printf("  %s  (invalid metadata)\n", cpID)
				continue
			}

			fmt.Printf("  %s  branch=%s  agent=%d%%  created=%s\n",
				cpID, meta.Branch, meta.AgentPercent, meta.CreatedAt)
		}
	}

	return nil
}

func runRewindTo(id string) error {
	_, err := git.RepoRoot()
	if err != nil {
		return fmt.Errorf("must be run inside a git repository")
	}

	if len(id) != 12 {
		return fmt.Errorf("checkpoint ID must be 12 characters (got %d)", len(id))
	}

	const branch = "partio/checkpoints/v1"
	shard := id[:2]
	rest := id[2:]

	// Read metadata
	metaJSON, err := git.ExecGit("show", branch+":"+shard+"/"+rest+"/metadata.json")
	if err != nil {
		return fmt.Errorf("checkpoint %s not found", id)
	}

	var meta checkpoint.Metadata
	if err := json.Unmarshal([]byte(metaJSON), &meta); err != nil {
		return fmt.Errorf("invalid checkpoint metadata: %w", err)
	}

	// Read session context
	context, _ := git.ExecGit("show", branch+":"+shard+"/"+rest+"/0/context.md")

	fmt.Printf("Rewinding to checkpoint %s\n", id)
	fmt.Printf("  Commit: %s\n", meta.CommitHash)
	fmt.Printf("  Branch: %s\n", meta.Branch)
	if context != "" {
		fmt.Printf("  Context:\n%s\n", context)
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
