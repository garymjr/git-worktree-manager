package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "git-worktree-manager",
	Short: "A CLI tool for managing Git worktrees",
	Long: `A CLI tool for managing Git worktrees.

This tool helps you create and manage Git worktrees more efficiently.`,
}

var commonWorktreeDir string

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	defaultWorktreeDir := GetDefaultWorktreeDir()
	if envVar := os.Getenv("GIT_WORKTREE_MANAGER_DIR"); envVar != "" {
		defaultWorktreeDir = envVar
	}

	rootCmd.AddCommand(createCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(removeCmd)
	rootCmd.AddCommand(switchCmd)
	createCmd.Flags().StringVarP(&commonWorktreeDir, "worktree-dir", "w", defaultWorktreeDir, "Base directory for new worktrees")
}

// getDefaultWorktreeDir returns the default worktree directory based on the operating system.
func GetDefaultWorktreeDir() string {
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
