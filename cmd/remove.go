package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/garymjr/git-worktree-manager/pkg/state"
	"github.com/spf13/cobra"
)

var removeBranch bool
var forceRemove bool

var removeCmd = &cobra.Command{
	Use:     "remove [branch-name]",
	Short:   "Remove an existing worktree",
	Aliases: []string{"rm"},
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		branchName := args[0]

		// Get the current Git repository root
		gitRootBytes, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
		if err != nil {
			fmt.Printf("Error getting git repository root: %v\n", err)
			return
		}
		gitRoot := strings.TrimSpace(string(gitRootBytes))

		// Get the remote URL
		remoteURLBytes, err := exec.Command("git", "config", "--get", "remote.origin.url").Output()
		if err != nil {
			fmt.Printf("Error getting remote origin URL: %v\n", err)
			return
		}
		remoteURL := strings.TrimSpace(string(remoteURLBytes))

		// Parse organization/username and repo name from remote URL
		orgRepo := ParseRemoteURL(remoteURL)
		if orgRepo == "" {
			fmt.Printf("Could not parse organization/username and repository name from remote URL: %s\n", remoteURL)
			return
	}

		// Initialize state manager
		stateManager, err := state.NewStateManager()
		if err != nil {
			fmt.Printf("Error initializing state manager: %v\n", err)
			return
		}

		// Try to get worktree from state first
		entry, exists := stateManager.GetWorktree(orgRepo, branchName)
		if !exists {
			// Fall back to old behavior if not found in state
			fmt.Printf("Worktree for branch '%s' not registered\n", branchName)
			return
		}

		worktreePath := entry.Path

		// Remove from state
		err = stateManager.RemoveWorktree(orgRepo, branchName)
		if err != nil {
			fmt.Printf("Error removing worktree from state: %v\n", err)
		}

		// Check if the worktree directory exists before attempting to remove
		_, err = os.Stat(worktreePath)
		if os.IsNotExist(err) {
			fmt.Printf("Worktree for branch '%s' not found at '%s'\n", branchName, worktreePath)
			return
		} else if err != nil {
			fmt.Printf("Error checking worktree path '%s': %v\n", worktreePath, err)
			return
		}

		// Remove the worktree
		removeArgs := []string{"worktree", "remove"}
		if forceRemove {
			removeArgs = append(removeArgs, "--force")
		}
		removeArgs = append(removeArgs, worktreePath)

		cmdRemoveWorktree := exec.Command("git", removeArgs...)
		cmdRemoveWorktree.Dir = gitRoot // Ensure command runs in the git root
		out, err := cmdRemoveWorktree.CombinedOutput()
		if err != nil {
			fmt.Printf("Error removing worktree at '%s': %v\nOutput: %s\n", worktreePath, err, out)
			return
		}

		var successMsg string
		if removeBranch {
			branchRemoveArgs := []string{"branch"}
			if forceRemove {
				branchRemoveArgs = append(branchRemoveArgs, "-D") // Force delete branch
			} else {
				branchRemoveArgs = append(branchRemoveArgs, "-d") // Delete branch
			}
			branchRemoveArgs = append(branchRemoveArgs, branchName)

			cmdRemoveBranch := exec.Command("git", branchRemoveArgs...)
			cmdRemoveBranch.Dir = gitRoot // Ensure command runs in the git root
			out, err := cmdRemoveBranch.CombinedOutput()
			if err != nil {
				fmt.Printf("Error removing branch '%s': %v\nOutput: %s\n", branchName, err, out)
				return
			}
			successMsg = fmt.Sprintf("Successfully removed worktree at '%s' and branch '%s'", worktreePath, branchName)
		} else {
			successMsg = fmt.Sprintf("Successfully removed worktree at '%s'", worktreePath)
		}
		fmt.Println(successMsg)
	},
}

func init() {
	removeCmd.Flags().BoolVarP(&removeBranch, "remove-branch", "b", false, "Also remove the associated Git branch")
	removeCmd.Flags().BoolVarP(&forceRemove, "force", "f", false, "Force removal of the worktree and/or branch")

	// Add the worktree-dir flag to the remove command as well
	defaultWorktreeDir := GetDefaultWorktreeDir()
	if envVar := os.Getenv("GIT_WORKTREE_MANAGER_DIR"); envVar != "" {
		defaultWorktreeDir = envVar
	}
	removeCmd.Flags().StringVarP(&commonWorktreeDir, "worktree-dir", "w", defaultWorktreeDir, "Base directory for new worktrees")
}
