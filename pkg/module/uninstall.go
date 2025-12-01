package module

import (
	"github.com/elmhuangyu/dotman/pkg/module/filesystem"
	"github.com/elmhuangyu/dotman/pkg/state"
)

// UninstallResult contains the results of an uninstallation
type UninstallResult struct {
	IsSuccess         bool
	Summary           string
	Errors            []string
	RemovedLinks      []FileOperation
	SkippedLinks      []OperationResult
	RemovedGenerated  []FileOperation
	SkippedGenerated  []OperationResult
	BackedUpGenerated []OperationResult
	FailedRemovals    []OperationResult
}

// Uninstall performs the uninstallation of dotfiles using the state file
func Uninstall(dotfilesDir string) (*UninstallResult, error) {
	config := &UninstallConfig{
		BackupModified: true, // Default to backing up modified files
		StatePath:      dotfilesDir,
	}
	return UninstallWithConfig(config)
}

// UninstallWithConfig performs uninstallation using the provided configuration
func UninstallWithConfig(config *UninstallConfig) (*UninstallResult, error) {
	// Initialize dependencies
	fileOp := filesystem.NewOperator()
	stateMgr := &stateManagerAdapter{} // Use adapter to maintain compatibility

	// Create uninstaller
	uninstaller := NewUninstaller(fileOp, stateMgr)

	// Create request
	req := &UninstallRequest{
		DotfilesDir:    config.StatePath,
		BackupModified: config.BackupModified,
	}

	// Perform uninstallation
	return uninstaller.Uninstall(req)
}

// stateManagerAdapter adapts the existing state functions to StateManager interface
type stateManagerAdapter struct{}

func (s *stateManagerAdapter) Load(path string) (*state.StateFile, error) {
	return state.LoadStateFile(path)
}

func (s *stateManagerAdapter) Save(path string, stateFile *state.StateFile) error {
	return state.SaveStateFile(path, stateFile)
}

func (s *stateManagerAdapter) AddMapping(stateFile *state.StateFile, source, target, fileType string) error {
	return state.AddMapping(stateFile, source, target, fileType)
}

func (s *stateManagerAdapter) RemoveMappings(stateFile *state.StateFile, targets []string) error {
	return state.RemoveMappings(stateFile, targets)
}
