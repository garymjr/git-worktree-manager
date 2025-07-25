package cmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/garymjr/git-worktree-manager/pkg/log"
	"github.com/garymjr/git-worktree-manager/pkg/state"
	"github.com/spf13/cobra"
)

var silent bool

var switchCmd = &cobra.Command{
	Use:     "switch [branch-name]",
	Short:   "Switch to an existing worktree",
	Aliases: []string{"s"},
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		branchName := args[0]

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
		if exists {
			SwitchToWorktreeByPath(entry.Path, silent)
			return
		}

		// Fallback to old behavior if not found in state
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

func SwitchToWorktreeByPath(worktreePath string, silent bool) {
	// Check if the worktree directory exists
	_, err := os.Stat(worktreePath)
	if os.IsNotExist(err) {
		log.Warnf("worktree not found at '%s'", worktreePath)
		return
	} else if err != nil {
		log.Errorf("checking worktree path '%s': %v", worktreePath, err)
		return
	}

	if !silent {
		log.Infof("Switching to worktree at '%s'\n", worktreePath)
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
		log.Errorf("starting shell in worktree: %v", err)
		return
	}
}

func SwitchToWorktree(branchName string, orgRepo string, worktreeDir string, silent bool) {
	worktreePath := filepath.Join(worktreeDir, orgRepo, branchName)

	// Check if the worktree directory exists
	_, err := os.Stat(worktreePath)
	if os.IsNotExist(err) {
		log.Warnf("worktree for branch '%s' not found at '%s'", branchName, worktreePath)
		return
	} else if err != nil {
		log.Errorf("checking worktree path '%s': %v", worktreePath, err)
		return
	}

	if !silent {
		log.Infof("Switching to worktree at '%s'\n", worktreePath)
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
		log.Errorf("starting shell in worktree: %v", err)
		return
	}
}
