package cmd

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var createCmd = &cobra.Command{
	Use:   "create [branch-name]",
	Short: "Create a new branch and a new worktree",
	Args:  cobra.ExactArgs(1),
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
		orgRepo := parseRemoteURL(remoteURL)
		if orgRepo == "" {
			fmt.Printf("Could not parse organization/username and repository name from remote URL: %s\n", remoteURL)
			return
		}

		// Construct the worktree path
		worktreePath := filepath.Join(commonWorktreeDir, orgRepo, branchName)

		

		

		// Create the new worktree
        cmdWorktree := exec.Command("git", "worktree", "add", "-b", branchName, worktreePath)
        cmdWorktree.Dir = gitRoot // Ensure command runs in the git root
        out, err := cmdWorktree.CombinedOutput()
        if err != nil {
            fmt.Printf("Error creating worktree at '%s': %v\nOutput: %s\n", worktreePath, err, out)
            return
        }

		fmt.Printf("Successfully created branch '%s' and worktree at '%s'\n", branchName, worktreePath)
	},
}

func init() {
	rootCmd.AddCommand(createCmd)
}

// parseRemoteURL parses the remote URL to extract the organization/username and repository name.
// It handles both HTTPS and SSH URLs.
// Examples:
//   https://github.com/owner/repo.git -> owner/repo
//   git@github.com:owner/repo.git -> owner/repo
func parseRemoteURL(url string) string {
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

