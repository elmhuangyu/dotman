package state

import (
	"github.com/elmhuangyu/dotman/pkg/state"
)

// StateManager interface for state persistence operations
type StateManager interface {
	Load(path string) (*state.StateFile, error)
	Save(path string, stateFile *state.StateFile) error
	AddMapping(stateFile *state.StateFile, source, target, fileType string) error
	RemoveMappings(stateFile *state.StateFile, targets []string) error
}

// DefaultStateManager implements the StateManager interface
type DefaultStateManager struct{}

// NewStateManager creates a new StateManager instance
func NewStateManager() StateManager {
	return &DefaultStateManager{}
}

// Load loads a state file from the given path
func (sm *DefaultStateManager) Load(path string) (*state.StateFile, error) {
	return state.LoadStateFile(path)
}

// Save saves a state file to the given path
func (sm *DefaultStateManager) Save(path string, stateFile *state.StateFile) error {
	return state.SaveStateFile(path, stateFile)
}

// AddMapping adds a file mapping to the state file
func (sm *DefaultStateManager) AddMapping(stateFile *state.StateFile, source, target, fileType string) error {
	stateFile.AddFileMapping(source, target, fileType)
	return nil
}

// RemoveMappings removes file mappings from the state file by target paths
func (sm *DefaultStateManager) RemoveMappings(stateFile *state.StateFile, targets []string) error {
	if len(targets) == 0 {
		return nil
	}

	// Create a set of targets to remove for efficient lookup
	targetSet := make(map[string]bool)
	for _, target := range targets {
		targetSet[target] = true
	}

	// Filter out mappings that match the targets
	var remainingFiles []state.FileMapping
	for _, mapping := range stateFile.Files {
		if !targetSet[mapping.Target] {
			remainingFiles = append(remainingFiles, mapping)
		}
	}

	stateFile.Files = remainingFiles
	return nil
}
