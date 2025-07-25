package cmd

import (
	"github.com/garymjr/git-worktree-manager/pkg/log"
	"github.com/garymjr/git-worktree-manager/pkg/state"
	"github.com/spf13/cobra"
)

var cleanupCmd = &cobra.Command{
	Use:   "cleanup",
	Short: "Clean up stale worktree entries from state",
	Long: `Remove entries for worktrees that no longer exist on disk.
This helps keep the worktree state file clean and accurate.`,
	Aliases: []string{"clean"},
	Run: func(cmd *cobra.Command, args []string) {
		// Initialize state manager
		stateManager, err := state.NewStateManager()
		if err != nil {
			log.Errorf("initializing state manager: %v", err)
			return
		}

		// Get worktrees before cleanup for comparison
		beforeCount := len(stateManager.ListWorktrees())

		// Clean up stale entries
		err = stateManager.CleanupStaleEntries()
		if err != nil {
			log.Errorf("cleaning up stale entries: %v", err)
			return
		}

		// Get worktrees after cleanup
		afterCount := len(stateManager.ListWorktrees())
		removedCount := beforeCount - afterCount

		if removedCount > 0 {
			log.Infof("Cleaned up %d stale worktree entries\n", removedCount)
		} else {
			log.Info("No stale entries found\n")
		}
	},
}

func init() {
	// Add cleanup command to root
	rootCmd.AddCommand(cleanupCmd)
}
