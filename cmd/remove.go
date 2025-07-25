package cmd

import (
	"os"
	"os/exec"
	"strings"

	"github.com/garymjr/git-worktree-manager/pkg/log"
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
			log.Errorf("getting git repository root: %v", err)
			return
		}
		gitRoot := strings.TrimSpace(string(gitRootBytes))

		// Get the remote URL
		remoteURLBytes, err := exec.Command("git", "config", "--get", "remote.origin.url").Output()
		if err != nil {
			log.Errorf("getting remote origin URL: %v", err)
			return
		}
		remoteURL := strings.TrimSpace(string(remoteURLBytes))

		// Parse organization/username and repo name from remote URL
		orgRepo := ParseRemoteURL(remoteURL)
		if orgRepo == "" {
			log.Errorf("could not parse organization/username and repository name from remote URL: %s", remoteURL)
			return
		}

		// Initialize state manager
		stateManager, err := state.NewStateManager()
		if err != nil {
			log.Errorf("initializing state manager: %v", err)
			return
		}

		// Try to get worktree from state first
		entry, exists := stateManager.GetWorktree(orgRepo, branchName)
		if !exists {
			// Fall back to old behavior if not found in state
			log.Warnf("worktree for branch '%s' not registered", branchName)
			return
		}

		worktreePath := entry.Path

		// Remove from state
		err = stateManager.RemoveWorktree(orgRepo, branchName)
		if err != nil {
			log.Errorf("removing worktree from state: %v", err)
		}

		// Check if the worktree directory exists before attempting to remove
		_, err = os.Stat(worktreePath)
		if os.IsNotExist(err) {
			log.Warnf("worktree for branch '%s' not found at '%s'", branchName, worktreePath)
			return
		} else if err != nil {
			log.Errorf("checking worktree path '%s': %v", worktreePath, err)
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
			log.Errorf("removing worktree at '%s': %v\nOutput: %s", worktreePath, err, out)
			return
		}

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
				log.Errorf("removing branch '%s': %v\nOutput: %s", branchName, err, out)
				return
			}
			log.Infof("Successfully removed worktree at '%s' and branch '%s'\n", worktreePath, branchName)
		} else {
			log.Infof("Successfully removed worktree at '%s'\n", worktreePath)
		}
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
