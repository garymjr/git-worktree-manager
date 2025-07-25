package cmd

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/garymjr/git-worktree-manager/pkg/state"
	"github.com/spf13/cobra"
)

var createBranch bool

func init() {
	createCmd.Flags().BoolVarP(&createBranch, "create-branch", "b", false, "Create branch if it does not exist")
}

var createCmd = &cobra.Command{
	Use:     "create [branch-name]",
	Short:   "Create a new worktree, optionally creating the branch if it does not exist",
	Aliases: []string{"n", "new"},
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

		// Construct the worktree path
		worktreePath := filepath.Join(commonWorktreeDir, orgRepo, branchName)

		// Add worktree to state
		err = stateManager.AddWorktree(worktreePath, orgRepo, branchName, remoteURL)
		if err != nil {
			fmt.Printf("Error adding worktree to state: %v\n", err)
			return
		}

		// Create the new worktree; only create the branch if requested
		var cmdArgs []string
		if createBranch {
			cmdArgs = []string{"worktree", "add", "-b", branchName, worktreePath}
		} else {
			cmdArgs = []string{"worktree", "add", worktreePath, branchName}
		}
		cmdWorktree := exec.Command("git", cmdArgs...)
		cmdWorktree.Dir = gitRoot
		out, err := cmdWorktree.CombinedOutput()
		if err != nil {
			fmt.Printf("Error creating worktree at '%s': %v\nOutput: %s\n", worktreePath, err, out)
			return
		}

		if createBranch {
			fmt.Printf("Successfully created branch '%s' and worktree at '%s'\n", branchName, worktreePath)
		} else {
			fmt.Printf("Successfully created worktree for branch '%s' at '%s'\n", branchName, worktreePath)
		}

		// Switch to the new worktree
		SwitchToWorktree(branchName, orgRepo, commonWorktreeDir, false)
	},
}

// parseRemoteURL parses the remote URL to extract the organization/username and repository name.
// It handles both HTTPS and SSH URLs.
// Examples:
//
//	https://github.com/owner/repo.git -> owner/repo
//	git@github.com:owner/repo.git -> owner/repo
func ParseRemoteURL(url string) string {
	// Remove .git suffix if present
	url = strings.TrimSuffix(url, ".git")

	// Handle HTTPS
	if strings.HasPrefix(url, "https://") {
		parts := strings.Split(url, "/")
		if len(parts) >= 2 {
			return strings.Join(parts[len(parts)-2:], "/")
		}
	} else if strings.HasPrefix(url, "git@") {
		// Handle SSH
		parts := strings.Split(url, ":")
		if len(parts) >= 2 {
			return strings.Join(strings.Split(parts[1], "/"), "/")
		}
	}

	return ""
}
