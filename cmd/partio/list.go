package main

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/partio-io/cli/internal/checkpoint"
	"github.com/partio-io/cli/internal/git"
)

func newListCmd() *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List captured checkpoints",
		Long:  `List all captured checkpoints sorted by creation time (newest first).`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(jsonOutput)
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "output as JSON")

	return cmd
}

// checkpointListOutput is the JSON envelope for checkpoint listings.
type checkpointListOutput struct {
	Checkpoints []checkpoint.Metadata `json:"checkpoints"`
}

func runList(jsonOutput bool) error {
	_, err := git.RepoRoot()
	if err != nil {
		return fmt.Errorf("must be run inside a git repository")
	}

	checkpoints, err := checkpoint.List()
	if err != nil {
		if err == checkpoint.ErrNoBranch {
			if jsonOutput {
				return json.NewEncoder(os.Stdout).Encode(checkpointListOutput{Checkpoints: []checkpoint.Metadata{}})
			}
			fmt.Println("No checkpoints found. Run 'partio enable' and make some commits to capture checkpoints.")
			return nil
		}
		return err
	}

	if jsonOutput {
		if checkpoints == nil {
			checkpoints = []checkpoint.Metadata{}
		}
		return json.NewEncoder(os.Stdout).Encode(checkpointListOutput{Checkpoints: checkpoints})
	}

	if len(checkpoints) == 0 {
		fmt.Println("No checkpoints found.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(w, "ID\tCOMMIT\tAGENT\tATTRIBUTION\tCREATED\tBRANCH")
	for _, cp := range checkpoints {
		id := cp.ID
		if len(id) > 12 {
			id = id[:12]
		}
		commit := cp.CommitHash
		if len(commit) > 7 {
			commit = commit[:7]
		}
		attribution := fmt.Sprintf("%d%%", cp.AgentPercent)
		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
			id, commit, cp.Agent, attribution, cp.CreatedAt, cp.Branch)
	}
	_ = w.Flush()

	return nil
}
