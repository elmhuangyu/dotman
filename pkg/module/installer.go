package module

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/elmhuangyu/dotman/pkg/config"
	"github.com/elmhuangyu/dotman/pkg/logger"
	"github.com/elmhuangyu/dotman/pkg/module/filesystem"
	"github.com/elmhuangyu/dotman/pkg/module/state"
	"github.com/elmhuangyu/dotman/pkg/module/template"
	dotmanState "github.com/elmhuangyu/dotman/pkg/state"
)

// InstallRequest contains the parameters for an installation request
type InstallRequest struct {
	Modules     []config.ModuleConfig
	RootVars    map[string]string
	Mkdir       bool
	Force       bool
	DotfilesDir string
}

// Installer handles the installation of dotfiles
type Installer struct {
	fileOp   filesystem.FileOperator
	template template.TemplateRenderer
	stateMgr state.StateManager
}

// NewInstaller creates a new Installer instance
func NewInstaller(fileOp filesystem.FileOperator, templateRenderer template.TemplateRenderer, stateMgr state.StateManager) *Installer {
	return &Installer{
		fileOp:   fileOp,
		template: templateRenderer,
		stateMgr: stateMgr,
	}
}

// Install performs the installation of dotfiles
func (i *Installer) Install(req *InstallRequest) (*InstallResult, error) {
	log := logger.GetLogger()

	// Initialize filesystem operators
	symlinkMgr := filesystem.NewSymlinkManager(i.fileOp)
	backupMgr := filesystem.NewBackupManager(i.fileOp)

	log.Info().Int("modules", len(req.Modules)).Msg("Starting installation")

	// Initialize state file
	var stateFile *dotmanState.StateFile
	var statePath string
	var err error

	if req.DotfilesDir != "" {
		statePath = filepath.Join(req.DotfilesDir, "state.yaml")
		stateFile, err = i.stateMgr.Load(statePath)
		if err != nil {
			log.Warn().Err(err).Msg("Failed to load state file, continuing without state logging")
			stateFile = nil
		}
		if stateFile == nil {
			stateFile = dotmanState.NewStateFile()
		}
	}

	// First validate the installation
	validation, err := Validate(req.Modules, req.RootVars, req.Mkdir, req.Force)
	if err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	result := &InstallResult{
		IsSuccess: true,
		Errors:    []string{},
	}

	// Check for validation errors or conflicts - if any exist, fail the installation
	if len(validation.Errors) > 0 {
		result.IsSuccess = false
		result.Errors = validation.Errors
		result.Summary = fmt.Sprintf("Installation failed: %d validation errors", len(validation.Errors))
		return result, nil
	}

	// Check for conflicts in the operations
	forceOps := len(validation.ForceLinkOperations) + len(validation.ForceTemplateOps)
	if forceOps > 0 && !req.Force {
		result.IsSuccess = false
		result.Errors = append(result.Errors, "conflicts detected - installation would overwrite existing files")
		result.Summary = "Installation failed: conflicts detected"
		return result, nil
	}

	result.SkippedLinks = validation.SkipOperations

	// Record skipped files in state file
	for _, operation := range validation.SkipOperations {
		if stateFile != nil {
			if err := i.stateMgr.AddMapping(stateFile, operation.Source, operation.Target, dotmanState.TypeLink); err != nil {
				log.Warn().Err(err).Msg("Failed to add mapping to state file for skipped operation")
			}
			if err := i.stateMgr.Save(statePath, stateFile); err != nil {
				log.Warn().Err(err).Msg("Failed to save state file for skipped operation")
			}
		}
		log.Info().Str("source", operation.Source).Str("target", operation.Target).Msg("Skipped (correct symlink already exists)")
	}

	// Perform the installation of symlinks
	if err := i.installSymlinks(validation.CreateOperations, symlinkMgr, req.Mkdir, stateFile, statePath, result); err != nil {
		return result, err
	}

	// Perform template file generation
	if err := i.installTemplates(validation.CreateTemplateOps, req.RootVars, req.Mkdir, stateFile, statePath, result); err != nil {
		return result, err
	}

	// Handle force operations (both links and templates)
	if req.Force {
		if err := i.handleForceOperations(validation.ForceLinkOperations, validation.ForceTemplateOps, symlinkMgr, backupMgr, req.RootVars, req.Mkdir, stateFile, statePath, result); err != nil {
			return result, err
		}
	}

	// Generate summary
	if result.IsSuccess {
		result.Summary = fmt.Sprintf("Installation successful: %d symlinks created, %d template files generated, %d skipped", len(result.CreatedLinks), len(result.CreatedTemplates), len(result.SkippedLinks))
	} else {
		result.Summary = fmt.Sprintf("Installation failed: %d errors", len(result.Errors))
	}

	log.Info().Bool("success", result.IsSuccess).Msg("Installation completed")

	return result, nil
}

// installSymlinks installs regular symlinks
func (i *Installer) installSymlinks(ops []FileOperation, symlinkMgr *filesystem.SymlinkManager, mkdir bool, stateFile *dotmanState.StateFile, statePath string, result *InstallResult) error {
	log := logger.GetLogger()

	for _, operation := range ops {

		if err := symlinkMgr.CreateSymlinkWithMkdir(operation.Source, operation.Target, mkdir); err != nil {
			result.IsSuccess = false
			result.Errors = append(result.Errors, fmt.Sprintf("failed to create symlink %s -> %s: %v", operation.Source, operation.Target, err))
		} else {
			// Record successful symlink in state file
			if stateFile != nil {
				if err := i.stateMgr.AddMapping(stateFile, operation.Source, operation.Target, dotmanState.TypeLink); err != nil {
					log.Warn().Err(err).Msg("Failed to add mapping to state file")
				}
				if err := i.stateMgr.Save(statePath, stateFile); err != nil {
					log.Warn().Err(err).Msg("Failed to save state file")
				}
			}
		}
		result.CreatedLinks = append(result.CreatedLinks, operation)
		log.Info().Str("source", operation.Source).Str("target", operation.Target).Msg("Created symlink")

		if !result.IsSuccess {
			break
		}
	}

	return nil
}

// installTemplates installs template files
func (i *Installer) installTemplates(ops []FileOperation, vars map[string]string, mkdir bool, stateFile *dotmanState.StateFile, statePath string, result *InstallResult) error {
	log := logger.GetLogger()

	for _, operation := range ops {
		if err := i.createTemplateFile(operation.Source, operation.Target, vars, mkdir); err != nil {
			result.IsSuccess = false
			result.Errors = append(result.Errors, fmt.Sprintf("failed to create template file %s -> %s: %v", operation.Source, operation.Target, err))
		} else {
			// Record successful template generation in state file
			if stateFile != nil {
				if err := i.stateMgr.AddMapping(stateFile, operation.Source, operation.Target, dotmanState.TypeGenerated); err != nil {
					log.Warn().Err(err).Msg("Failed to add mapping to state file for template")
				}
				if err := i.stateMgr.Save(statePath, stateFile); err != nil {
					log.Warn().Err(err).Msg("Failed to save state file for template")
				}
			}
			result.CreatedTemplates = append(result.CreatedTemplates, operation)
			log.Info().Str("source", operation.Source).Str("target", operation.Target).Msg("Created template file")
		}

		if !result.IsSuccess {
			break
		}
	}

	return nil
}

// handleForceOperations handles force operations for both links and templates
func (i *Installer) handleForceOperations(forceLinkOps, forceTemplateOps []FileOperation, symlinkMgr *filesystem.SymlinkManager, backupMgr *filesystem.BackupManager, vars map[string]string, mkdir bool, stateFile *dotmanState.StateFile, statePath string, result *InstallResult) error {
	log := logger.GetLogger()

	// Handle force link operations
	for _, operation := range forceLinkOps {

		_, err := backupMgr.BackupAndReplace(operation.Target, func() error {
			return symlinkMgr.CreateSymlinkWithMkdir(operation.Source, operation.Target, mkdir)
		})
		if err != nil {
			result.IsSuccess = false
			result.Errors = append(result.Errors, fmt.Sprintf("failed to backup and create symlink %s -> %s: %v", operation.Source, operation.Target, err))
		} else {
			// Record successful symlink in state file
			if stateFile != nil {
				if err := i.stateMgr.AddMapping(stateFile, operation.Source, operation.Target, dotmanState.TypeLink); err != nil {
					log.Warn().Err(err).Msg("Failed to add mapping to state file")
				}
				if err := i.stateMgr.Save(statePath, stateFile); err != nil {
					log.Warn().Err(err).Msg("Failed to save state file")
				}
			}
			result.CreatedLinks = append(result.CreatedLinks, operation)
			log.Warn().Str("source", operation.Source).Str("target", operation.Target).Msg("Backed up existing file and created symlink")
		}

		if !result.IsSuccess {
			break
		}
	}

	// Handle force template operations
	for _, operation := range forceTemplateOps {
		_, err := backupMgr.BackupAndReplace(operation.Target, func() error {
			return i.createTemplateFile(operation.Source, operation.Target, vars, mkdir)
		})
		if err != nil {
			result.IsSuccess = false
			result.Errors = append(result.Errors, fmt.Sprintf("failed to backup and create template file %s -> %s: %v", operation.Source, operation.Target, err))
		} else {
			// Record successful template generation in state file
			if stateFile != nil {
				if err := i.stateMgr.AddMapping(stateFile, operation.Source, operation.Target, dotmanState.TypeGenerated); err != nil {
					log.Warn().Err(err).Msg("Failed to add mapping to state file for template")
				}
				if err := i.stateMgr.Save(statePath, stateFile); err != nil {
					log.Warn().Err(err).Msg("Failed to save state file for template")
				}
			}
			result.CreatedTemplates = append(result.CreatedTemplates, operation)
			log.Warn().Str("source", operation.Source).Str("target", operation.Target).Msg("Backed up existing file and created template file")
		}

		if !result.IsSuccess {
			break
		}
	}

	return nil
}

// createTemplateFile creates a template file by rendering the template and writing to target
func (i *Installer) createTemplateFile(source, target string, vars map[string]string, mkdir bool) error {

	// Ensure target directory exists
	targetDir := filepath.Dir(target)
	if !i.fileOp.FileExists(targetDir) {
		if mkdir {
			if err := i.fileOp.EnsureDirectory(targetDir); err != nil {
				return fmt.Errorf("failed to create target directory %s: %w", targetDir, err)
			}
		} else {
			return fmt.Errorf("target directory does not exist: %s", targetDir)
		}
	}

	// Render the template
	content, err := i.template.Render(source, vars)
	if err != nil {
		return fmt.Errorf("failed to render template: %w", err)
	}

	// Write the rendered content to the target file
	if err := os.WriteFile(target, content, 0644); err != nil {
		return fmt.Errorf("failed to write template file: %w", err)
	}

	return nil
}
