package cmd

import (
	"fmt"

	"github.com/garymjr/git-worktree-manager/pkg/state"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Show configuration and state information",
	Long:  `Display information about the worktree manager configuration and state storage.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Initialize state manager
		stateManager, err := state.NewStateManager()
		if err != nil {
			fmt.Printf("Error initializing state manager: %v\n", err)
			return
		}

		fmt.Println("Git Worktree Manager Configuration:")
		fmt.Printf("State file location: %s\n", stateManager.GetConfigPath())
		fmt.Printf("Total managed worktrees: %d\n", len(stateManager.ListWorktrees()))
		
		// Show fallback directory for legacy behavior
		defaultDir := GetDefaultWorktreeDir()
		fmt.Printf("Legacy default directory: %s\n", defaultDir)
	},
}

func init() {
	// Add config command to root
	rootCmd.AddCommand(configCmd)
}
