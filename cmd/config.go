package cmd

import (
	"github.com/garymjr/git-worktree-manager/pkg/log"
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
			log.Errorf("initializing state manager: %v", err)
			return
		}

		log.Infof("State file location: %s\n", stateManager.GetConfigPath())
		log.Infof("Total managed worktrees: %d\n", len(stateManager.ListWorktrees()))

		// Show fallback directory for legacy behavior
		defaultDir := GetDefaultWorktreeDir()
		log.Infof("Legacy default directory: %s\n", defaultDir)
	},
}

func init() {
	// Add config command to root
	rootCmd.AddCommand(configCmd)
}
