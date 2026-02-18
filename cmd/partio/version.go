package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version of partio",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("partio %s\n", version)
		},
	}
}
