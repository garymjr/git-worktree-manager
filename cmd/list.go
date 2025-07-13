package cmd

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Short:   "List all git worktrees and their branches",
	Aliases: []string{"ls"},
	Run: func(cmd *cobra.Command, args []string) {
		// Get the current working directory to identify the active worktree
		currentDir, err := exec.Command("pwd").Output()
		if err != nil {
			fmt.Printf("Error getting current directory: %v\n", err)
			return
		}
		currentDirPath := strings.TrimSpace(string(currentDir))

		// Execute 'git worktree list --porcelain'
		gitWorktreeListCmd := exec.Command("git", "worktree", "list", "--porcelain")
		output, err := gitWorktreeListCmd.Output()
		if err != nil {
			fmt.Printf("Error listing worktrees: %v\n", err)
			return
		}

		lines := strings.Split(string(output), "\n")

		fmt.Println("Git Worktrees:")
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

				indicator := "  "
				if strings.HasPrefix(currentDirPath, worktreePath) {
					indicator = "* " // Indicate active worktree
				}

				branchInfo := ""
				if branch != "" {
					branchInfo = fmt.Sprintf(" (%s)", strings.TrimPrefix(branch, "refs/heads/"))
				} else if isHead {
					// If no branch but HEAD is present, it's likely a detached HEAD
					branchInfo = " (detached HEAD)"
				}

				fmt.Printf("%s%s%s\n", indicator, worktreePath, branchInfo)
			}
		}
	},
}
