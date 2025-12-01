package module

import (
	"github.com/elmhuangyu/dotman/pkg/config"
	"github.com/elmhuangyu/dotman/pkg/module/filesystem"
	"github.com/elmhuangyu/dotman/pkg/module/state"
	"github.com/elmhuangyu/dotman/pkg/module/template"
)

// InstallResult contains the results of an installation
type InstallResult struct {
	IsSuccess        bool
	Summary          string
	Errors           []string
	CreatedLinks     []FileOperation
	CreatedTemplates []FileOperation
	SkippedLinks     []FileOperation
}

// Install performs the actual installation of dotfiles by creating symlinks and generating template files
func Install(modules []config.ModuleConfig, rootVars map[string]string, mkdir bool, force bool, dotfilesDir string) (*InstallResult, error) {
	config := &InstallConfig{
		Mkdir:     mkdir,
		Force:     force,
		DryRun:    false,
		Vars:      rootVars,
		StatePath: dotfilesDir,
	}
	return InstallWithConfig(modules, config)
}

// InstallWithConfig performs installation using the provided configuration
func InstallWithConfig(modules []config.ModuleConfig, config *InstallConfig) (*InstallResult, error) {
	// Initialize dependencies
	fileOp := filesystem.NewOperator()
	templateRenderer := template.NewRenderer()
	stateMgr := state.NewStateManager()

	// Create installer
	installer := NewInstaller(fileOp, templateRenderer, stateMgr)

	// Create install request
	req := &InstallRequest{
		Modules:     modules,
		RootVars:    config.Vars,
		Mkdir:       config.Mkdir,
		Force:       config.Force,
		DotfilesDir: config.StatePath,
	}

	// Perform installation
	return installer.Install(req)
}
