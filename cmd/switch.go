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

var silent bool

var switchCmd = &cobra.Command{
	Use:   "switch [branch-name]",
	Short: "Switch to an existing worktree",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		branchName := args[0]

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

		// Determine the common worktree directory (same logic as create command)
		defaultWorktreeDir := GetDefaultWorktreeDir()
		if envVar := os.Getenv("GIT_WORKTREE_MANAGER_DIR"); envVar != "" {
			defaultWorktreeDir = envVar
		}
		// If the flag was set, it overrides everything
		if cmd.Flags().Changed("worktree-dir") {
			defaultWorktreeDir = commonWorktreeDir // commonWorktreeDir is populated by the flag
		}

		SwitchToWorktree(branchName, orgRepo, defaultWorktreeDir, silent)
	},
}

func init() {
	// Add the worktree-dir flag to the switch command as well
	defaultWorktreeDir := GetDefaultWorktreeDir()
	if envVar := os.Getenv("GIT_WORKTREE_MANAGER_DIR"); envVar != "" {
		defaultWorktreeDir = envVar
	}
	switchCmd.Flags().StringVarP(&commonWorktreeDir, "worktree-dir", "w", defaultWorktreeDir, "Base directory for new worktrees")
	switchCmd.Flags().BoolVarP(&silent, "silent", "s", false, "Suppress output messages")
}

func SwitchToWorktree(branchName string, orgRepo string, worktreeDir string, silent bool) {
	worktreePath := filepath.Join(worktreeDir, orgRepo, branchName)

	// Check if the worktree directory exists
	_, err := os.Stat(worktreePath)
	if os.IsNotExist(err) {
		fmt.Printf("Worktree for branch '%s' not found at '%s'\n", branchName, worktreePath)
		return
	} else if err != nil {
		fmt.Printf("Error checking worktree path '%s': %v\n", worktreePath, err)
		return
	}

	if !silent {
		fmt.Printf("Switching to worktree at '%s'\n", worktreePath)
	}

	// Determine the user's shell
	shell := os.Getenv("SHELL")
	if shell == "" {
		// Fallback for Windows or if SHELL is not set
		if runtime.GOOS == "windows" {
			shell = "cmd.exe"
		} else {
			shell = "bash"
		}
	}

	// Execute a new shell in the worktree directory
	cmdShell := exec.Command(shell)
	cmdShell.Dir = worktreePath
	cmdShell.Stdin = os.Stdin
	cmdShell.Stdout = os.Stdout
	cmdShell.Stderr = os.Stderr

	if err := cmdShell.Run(); err != nil {
		fmt.Printf("Error starting shell in worktree: %v\n", err)
		return
	}
}
