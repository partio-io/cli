package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/partio-io/cli/internal/checkpoint"
)

func newExplainCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "explain <checkpoint-id>",
		Short: "Show details of a checkpoint including the AI model used",
		Args:  cobra.ExactArgs(1),
		RunE:  runExplain,
	}
}

func runExplain(cmd *cobra.Command, args []string) error {
	id := args[0]

	data, err := checkpoint.Read(id)
	if err != nil {
		return err
	}

	meta := data.Metadata

	model := meta.Model
	if model == "" {
		model = "unknown"
	}

	fmt.Printf("Checkpoint: %s\n", meta.ID)
	fmt.Printf("  Commit:     %s\n", meta.CommitHash)
	fmt.Printf("  Branch:     %s\n", meta.Branch)
	fmt.Printf("  Created:    %s\n", meta.CreatedAt)
	fmt.Printf("  Agent:      %s\n", meta.Agent)
	fmt.Printf("  Model:      %s\n", model)
	fmt.Printf("  Attribution: %d%% agent\n", meta.AgentPercent)

	if meta.SessionID != "" {
		fmt.Printf("  Session:    %s\n", meta.SessionID)
	}
	if meta.PlanSlug != "" {
		fmt.Printf("  Plan:       %s\n", meta.PlanSlug)
	}
	if data.Prompt != "" {
		fmt.Printf("\nPrompt:\n%s\n", data.Prompt)
	}
	if data.Context != "" {
		fmt.Printf("\nContext:\n%s\n", data.Context)
	}

	return nil
}
