package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "git-worktree-manager",
	Short: "A CLI tool for managing Git worktrees",
	Long: `A CLI tool for managing Git worktrees.

This tool helps you create and manage Git worktrees more efficiently.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

