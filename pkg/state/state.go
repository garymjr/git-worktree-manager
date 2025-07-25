package state

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

// WorktreeEntry represents a single worktree registration
type WorktreeEntry struct {
	ID           string    `json:"id"`            // Unique identifier (orgRepo/branchName)
	Path         string    `json:"path"`          // Full path to the worktree
	GitRepo      string    `json:"git_repo"`      // Organization/repository name (e.g., "owner/repo")
	BranchName   string    `json:"branch_name"`   // Branch name
	RemoteURL    string    `json:"remote_url"`    // Git remote URL
	CreatedAt    time.Time `json:"created_at"`    // When the worktree was created
	LastAccessed time.Time `json:"last_accessed"` // When the worktree was last accessed
}

// State represents the persistent state of the application
type State struct {
	Version   string                   `json:"version"`
	Worktrees map[string]WorktreeEntry `json:"worktrees"` // Key is the ID (orgRepo/branchName)
}

// StateManager handles loading and saving of persistent state
type StateManager struct {
	configPath string
	state      *State
}

// NewStateManager creates a new state manager
func NewStateManager() (*StateManager, error) {
	configPath, err := getConfigPath()
	if err != nil {
		return nil, fmt.Errorf("failed to get config path: %w", err)
	}

	sm := &StateManager{
		configPath: configPath,
		state: &State{
			Version:   "1.0",
			Worktrees: make(map[string]WorktreeEntry),
		},
	}

	// Load existing state if it exists
	if err := sm.load(); err != nil {
		return nil, fmt.Errorf("failed to load state: %w", err)
	}

	return sm, nil
}

// getConfigPath returns the path to the configuration file
func getConfigPath() (string, error) {
	var newDir, oldDir string
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	switch runtime.GOOS {
	case "windows":
		localAppData := os.Getenv("LocalAppData")
		if localAppData == "" {
			localAppData, _ = os.UserCacheDir()
		}
		newDir = filepath.Join(localAppData, "git-worktree-manager")
		appData := os.Getenv("APPDATA")
		if appData == "" {
			appData = filepath.Join(homeDir, "AppData", "Roaming")
		}
		oldDir = filepath.Join(appData, "git-worktree-manager")
	case "darwin", "linux":
		newDir = filepath.Join(homeDir, ".local", "share", "git-worktree-manager")
		oldDir = filepath.Join(homeDir, ".config", "git-worktree-manager")
	default:
		newDir = filepath.Join(homeDir, ".local", "share", "git-worktree-manager")
		oldDir = filepath.Join(homeDir, ".config", "git-worktree-manager")
	}

	// Ensure new directory exists
	if err := os.MkdirAll(newDir, 0755); err != nil {
		return "", err
	}

	newPath := filepath.Join(newDir, "state.json")
	oldPath := filepath.Join(oldDir, "state.json")

	// Migration logic: if old file exists and new file does not, move it
	if _, errOld := os.Stat(oldPath); errOld == nil {
		if _, errNew := os.Stat(newPath); os.IsNotExist(errNew) {
			errMv := os.Rename(oldPath, newPath)
			if errMv == nil {
				fmt.Printf("State file migrated from %s to %s\n", oldPath, newPath)
			} else {
				fmt.Printf("[ERROR] Failed to migrate state file: %v\n", errMv)
			}
		}
	}

	return newPath, nil
}

// load reads the state from disk
func (sm *StateManager) load() error {
	if _, err := os.Stat(sm.configPath); os.IsNotExist(err) {
		// File doesn't exist, use default state
		return nil
	}

	data, err := os.ReadFile(sm.configPath)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, sm.state)
}

// save writes the state to disk
func (sm *StateManager) save() error {
	data, err := json.MarshalIndent(sm.state, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(sm.configPath, data, 0644)
}

// AddWorktree registers a new worktree
func (sm *StateManager) AddWorktree(path, gitRepo, branchName, remoteURL string) error {
	id := filepath.Join(gitRepo, branchName)

	entry := WorktreeEntry{
		ID:           id,
		Path:         path,
		GitRepo:      gitRepo,
		BranchName:   branchName,
		RemoteURL:    remoteURL,
		CreatedAt:    time.Now(),
		LastAccessed: time.Now(),
	}

	sm.state.Worktrees[id] = entry
	return sm.save()
}

// RemoveWorktree unregisters a worktree
func (sm *StateManager) RemoveWorktree(gitRepo, branchName string) error {
	id := filepath.Join(gitRepo, branchName)
	delete(sm.state.Worktrees, id)
	return sm.save()
}

// GetWorktree retrieves a worktree by git repo and branch name
func (sm *StateManager) GetWorktree(gitRepo, branchName string) (WorktreeEntry, bool) {
	id := filepath.Join(gitRepo, branchName)
	entry, exists := sm.state.Worktrees[id]
	if exists {
		// Update last accessed time
		entry.LastAccessed = time.Now()
		sm.state.Worktrees[id] = entry
		sm.save() // Save the updated access time
	}
	return entry, exists
}

// ListWorktrees returns all registered worktrees
func (sm *StateManager) ListWorktrees() []WorktreeEntry {
	worktrees := make([]WorktreeEntry, 0, len(sm.state.Worktrees))
	for _, entry := range sm.state.Worktrees {
		worktrees = append(worktrees, entry)
	}
	return worktrees
}

// ListWorktreesByRepo returns all worktrees for a specific git repository
func (sm *StateManager) ListWorktreesByRepo(gitRepo string) []WorktreeEntry {
	worktrees := make([]WorktreeEntry, 0)
	for _, entry := range sm.state.Worktrees {
		if entry.GitRepo == gitRepo {
			worktrees = append(worktrees, entry)
		}
	}
	return worktrees
}

// CleanupStaleEntries removes entries for worktrees that no longer exist on disk
func (sm *StateManager) CleanupStaleEntries() error {
	toRemove := make([]string, 0)

	for id, entry := range sm.state.Worktrees {
		if _, err := os.Stat(entry.Path); os.IsNotExist(err) {
			toRemove = append(toRemove, id)
		}
	}

	for _, id := range toRemove {
		delete(sm.state.Worktrees, id)
	}

	if len(toRemove) > 0 {
		return sm.save()
	}

	return nil
}

// GetConfigPath returns the path to the configuration file (for external use)
func (sm *StateManager) GetConfigPath() string {
	return sm.configPath
}
