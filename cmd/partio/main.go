package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/jcleira/ai-workflow-core/internal/config"
	"github.com/jcleira/ai-workflow-core/internal/git"
	plog "github.com/jcleira/ai-workflow-core/internal/log"
)

var (
	version     = "dev"
	cfgLogLevel string
	cfg         config.Config
)

func main() {
	root := newRootCmd()
	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func newRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "partio",
		Short: "Capture the why behind your code changes",
		Long:  `partio hooks into Git workflows to capture AI agent sessions, preserving the why behind code changes alongside the what that Git already tracks.`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			repoRoot := ""
			if r, err := git.RepoRoot(); err == nil {
				repoRoot = r
			}

			var err error
			cfg, err = config.Load(repoRoot)
			if err != nil {
				return fmt.Errorf("loading config: %w", err)
			}

			if cfgLogLevel != "" {
				cfg.LogLevel = cfgLogLevel
			}

			plog.Setup(cfg.LogLevel)
			return nil
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	root.PersistentFlags().StringVar(&cfgLogLevel, "log-level", "", "log level (debug, info, warn, error)")

	root.AddCommand(
		newVersionCmd(),
		newEnableCmd(),
		newDisableCmd(),
		newStatusCmd(),
		newHookCmd(),
		newDoctorCmd(),
		newResetCmd(),
		newCleanCmd(),
		newRewindCmd(),
	)

	return root
}
