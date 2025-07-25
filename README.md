# git-worktree-manager

A command-line tool to manage git worktrees.

## Features

- **Persistent State Management**: Worktree information is stored and tracked across sessions.
- **Command Enhancements**:
  - **Create**: Register new worktrees in state, ensuring easy management and switching.
  - **Switch**: Seamlessly switch to registered worktrees.
  - **Remove**: Unregister and delete a worktree from the state.
  - **List**: Display managed and unmanaged worktrees, highlighting the active one.
  - **Cleanup**: Remove stale worktree entries from the state.
  - **Config**: Show path of state file and count of managed worktrees.

## Usage

This tool provides commands to create, list, remove, switch between, configure, and clean up git worktrees.

### Initialize Configuration

Use the config command to see where the state file is stored and the number of managed worktrees.

```bash
git-worktree-manager config
```

### Create a New Worktree

```bash
git-worktree-manager create <branch_name>
```

To create a new branch and worktree in one step, use the `-b`/`--create-branch` flag:

```bash
git-worktree-manager create -b <branch_name>
```

### List Worktrees

Displays both managed and unmanaged worktrees, indicating which is active.

```bash
git-worktree-manager list
```

### Remove a Worktree

Removes both the worktree and its Git branch if specified.

```bash
git-worktree-manager remove <branch_name> --remove-branch
```

### Switch to a Worktree

Switches to the specified worktree and opens a shell in its directory.

```bash
git-worktree-manager switch <branch_name>
```

### Cleanup Stale Entries

Removes entries for worktrees that no longer exist.

```bash
git-worktree-manager cleanup
```
