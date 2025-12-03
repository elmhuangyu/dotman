package module

import (
	dotmanState "github.com/elmhuangyu/dotman/pkg/state"
	"os"
)

// MockFileOperator is a mock implementation of filesystem.FileOperator
type MockFileOperator struct {
	CreateSymlinkFunc   func(source, target string) error
	RemoveFileFunc      func(path string) error
	CreateBackupFunc    func(path string) (string, error)
	EnsureDirectoryFunc func(path string) error
	CopyFileFunc        func(src, dst string) error
	FileExistsFunc      func(path string) bool
	IsSymlinkFunc       func(path string) bool
	ReadlinkFunc        func(path string) (string, error)
	WriteFileFunc       func(path string, data []byte, perm os.FileMode) error
}

func (m *MockFileOperator) CreateSymlink(source, target string) error {
	if m.CreateSymlinkFunc != nil {
		return m.CreateSymlinkFunc(source, target)
	}
	return nil
}

func (m *MockFileOperator) RemoveFile(path string) error {
	if m.RemoveFileFunc != nil {
		return m.RemoveFileFunc(path)
	}
	return nil
}

func (m *MockFileOperator) CreateBackup(path string) (string, error) {
	if m.CreateBackupFunc != nil {
		return m.CreateBackupFunc(path)
	}
	return path + ".bak", nil
}

func (m *MockFileOperator) EnsureDirectory(path string) error {
	if m.EnsureDirectoryFunc != nil {
		return m.EnsureDirectoryFunc(path)
	}
	return nil
}

func (m *MockFileOperator) CopyFile(src, dst string) error {
	if m.CopyFileFunc != nil {
		return m.CopyFileFunc(src, dst)
	}
	return nil
}

func (m *MockFileOperator) FileExists(path string) bool {
	if m.FileExistsFunc != nil {
		return m.FileExistsFunc(path)
	}
	return false
}

func (m *MockFileOperator) IsSymlink(path string) bool {
	if m.IsSymlinkFunc != nil {
		return m.IsSymlinkFunc(path)
	}
	return false
}

func (m *MockFileOperator) Readlink(path string) (string, error) {
	if m.ReadlinkFunc != nil {
		return m.ReadlinkFunc(path)
	}
	return "", nil
}

func (m *MockFileOperator) WriteFile(path string, data []byte, perm os.FileMode) error {
	if m.WriteFileFunc != nil {
		return m.WriteFileFunc(path, data, perm)
	}
	return nil
}

// MockTemplateRenderer is a mock implementation of template.TemplateRenderer
type MockTemplateRenderer struct {
	RenderFunc   func(templatePath string, vars map[string]string) ([]byte, error)
	ValidateFunc func(templatePath string, vars map[string]string) error
}

func (m *MockTemplateRenderer) Render(templatePath string, vars map[string]string) ([]byte, error) {
	if m.RenderFunc != nil {
		return m.RenderFunc(templatePath, vars)
	}
	return []byte("rendered content"), nil
}

func (m *MockTemplateRenderer) Validate(templatePath string, vars map[string]string) error {
	if m.ValidateFunc != nil {
		return m.ValidateFunc(templatePath, vars)
	}
	return nil
}

// MockStateManager is a mock implementation of state.StateManager
type MockStateManager struct {
	LoadFunc           func(path string) (*dotmanState.StateFile, error)
	SaveFunc           func(path string, stateFile *dotmanState.StateFile) error
	AddMappingFunc     func(stateFile *dotmanState.StateFile, source, target, fileType string) error
	RemoveMappingsFunc func(stateFile *dotmanState.StateFile, targets []string) error
}

func (m *MockStateManager) Load(path string) (*dotmanState.StateFile, error) {
	if m.LoadFunc != nil {
		return m.LoadFunc(path)
	}
	return dotmanState.NewStateFile(), nil
}

func (m *MockStateManager) Save(path string, stateFile *dotmanState.StateFile) error {
	if m.SaveFunc != nil {
		return m.SaveFunc(path, stateFile)
	}
	return nil
}

func (m *MockStateManager) AddMapping(stateFile *dotmanState.StateFile, source, target, fileType string) error {
	if m.AddMappingFunc != nil {
		return m.AddMappingFunc(stateFile, source, target, fileType)
	}
	stateFile.AddFileMapping(source, target, fileType)
	return nil
}

func (m *MockStateManager) RemoveMappings(stateFile *dotmanState.StateFile, targets []string) error {
	if m.RemoveMappingsFunc != nil {
		return m.RemoveMappingsFunc(stateFile, targets)
	}
	// Default implementation: remove all mappings with matching targets
	var remainingFiles []dotmanState.FileMapping
	targetSet := make(map[string]bool)
	for _, target := range targets {
		targetSet[target] = true
	}

	for _, mapping := range stateFile.Files {
		if !targetSet[mapping.Target] {
			remainingFiles = append(remainingFiles, mapping)
		}
	}
	stateFile.Files = remainingFiles
	return nil
}
