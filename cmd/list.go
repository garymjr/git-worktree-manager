package cmd

import (
	"fmt"
	"os/exec"
	"sort"
	"strings"

	"github.com/garymjr/git-worktree-manager/pkg/state"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Short:   "List all git worktrees and their branches",
	Aliases: []string{"ls"},
	Run: func(cmd *cobra.Command, args []string) {
		// Initialize state manager
		stateManager, err := state.NewStateManager()
		if err != nil {
			fmt.Printf("Error initializing state manager: %v\n", err)
			return
		}

		// Get the current working directory to identify the active worktree
		currentDir, err := exec.Command("pwd").Output()
		if err != nil {
			fmt.Printf("Error getting current directory: %v\n", err)
			return
		}
		currentDirPath := strings.TrimSpace(string(currentDir))

		// Get managed worktrees from state
		managedWorktrees := stateManager.ListWorktrees()
		
		// Execute 'git worktree list --porcelain' to get actual git worktrees
		gitWorktreeListCmd := exec.Command("git", "worktree", "list", "--porcelain")
		output, err := gitWorktreeListCmd.Output()
		if err != nil {
			fmt.Printf("Error listing worktrees: %v\n", err)
			return
		}

		lines := strings.Split(string(output), "\n")
		gitWorktrees := make(map[string]string) // path -> branch

		for i := 0; i < len(lines); i++ {
			line := strings.TrimSpace(lines[i])
			if line == "" {
				continue
			}

			if strings.HasPrefix(line, "worktree ") {
				worktreePath := strings.TrimPrefix(line, "worktree ")

				// Read the next lines for branch and HEAD
				var branch string
				var isHead bool
				for j := i + 1; j < len(lines); j++ {
					subLine := strings.TrimSpace(lines[j])
					if strings.HasPrefix(subLine, "branch ") {
						branch = strings.TrimPrefix(subLine, "branch ")
						branch = strings.TrimPrefix(branch, "refs/heads/")
					} else if strings.HasPrefix(subLine, "HEAD ") {
						isHead = true
					} else if subLine == "" {
						i = j // Move main loop index past this worktree's details
						break
					}
					if j == len(lines)-1 { // End of output
						i = j
					}
				}

				if branch != "" {
					gitWorktrees[worktreePath] = branch
				} else if isHead {
					gitWorktrees[worktreePath] = "detached HEAD"
				}
			}
		}

		// Sort managed worktrees by branch name for consistent output
		sort.Slice(managedWorktrees, func(i, j int) bool {
			return managedWorktrees[i].BranchName < managedWorktrees[j].BranchName
		})

		fmt.Println("Managed Worktrees:")
		for _, entry := range managedWorktrees {
			indicator := "  "
			if strings.HasPrefix(currentDirPath, entry.Path) {
				indicator = "* " // Indicate active worktree
			}

			status := "✓" // Exists in git
			if _, exists := gitWorktrees[entry.Path]; !exists {
				status = "✗" // Not found in git (stale)
			}

			fmt.Printf("%s%s (%s) [%s] %s\n", indicator, entry.Path, entry.BranchName, entry.GitRepo, status)
			delete(gitWorktrees, entry.Path) // Remove from map to find unmanaged
		}

		// Show any git worktrees not managed by our tool
		if len(gitWorktrees) > 0 {
			fmt.Println("\nUnmanaged Git Worktrees:")
			for path, branch := range gitWorktrees {
				indicator := "  "
				if strings.HasPrefix(currentDirPath, path) {
					indicator = "* "
				}
				fmt.Printf("%s%s (%s) [unmanaged]\n", indicator, path, branch)
			}
		}
	},
}
