package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

var commonWorktreeDir string

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

		fmt.Printf("Creating new branch '%s' and worktree at '%s'\n", branchName, worktreePath)

		// Create the new branch
		cmdBranch := exec.Command("git", "branch", branchName)
		cmdBranch.Dir = gitRoot // Ensure command runs in the git root
		if err := cmdBranch.Run(); err != nil {
			fmt.Printf("Error creating branch '%s': %v\n", branchName, err)
			return
		}

		// Create the new worktree
		cmdWorktree := exec.Command("git", "worktree", "add", "-b", branchName, worktreePath)
		cmdWorktree.Dir = gitRoot // Ensure command runs in the git root
		if err := cmdWorktree.Run(); err != nil {
			fmt.Printf("Error creating worktree at '%s': %v\n", worktreePath, err)
			return
		}

		fmt.Printf("Successfully created branch '%s' and worktree at '%s'\n", branchName, worktreePath)
	},
}

func init() {
	defaultWorktreeDir := getDefaultWorktreeDir()
	if envVar := os.Getenv("GIT_WORKTREE_MANAGER_DIR"); envVar != "" {
		defaultWorktreeDir = envVar
	}

	rootCmd.AddCommand(createCmd)
	createCmd.Flags().StringVarP(&commonWorktreeDir, "worktree-dir", "w", defaultWorktreeDir, "Base directory for new worktrees")
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

// getDefaultWorktreeDir returns the default worktree directory based on the operating system.
func getDefaultWorktreeDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// Fallback if home directory cannot be determined
		return "/tmp/git-worktrees"
	}

	switch runtime.GOOS {
	case "windows":
		appData := os.Getenv("LOCALAPPDATA")
		if appData == "" {
			return filepath.Join(homeDir, "AppData", "Local", "git-worktree-manager")
		}
		return filepath.Join(appData, "git-worktree-manager")
	case "darwin", "linux":
		return filepath.Join(homeDir, ".local", "git-worktree-manager")
	default:
		return filepath.Join(homeDir, ".git-worktree-manager")
	}
}