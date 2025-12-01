package module

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/elmhuangyu/dotman/pkg/config"
)

// FileMapping represents a two-way mapping between source and target files
type FileMapping struct {
	// sourceToTarget maps source file paths to target file paths
	sourceToTarget map[string]string
	// targetToSource maps target file paths to source file paths
	targetToSource map[string]string
	// templates maps source template file paths to their target paths
	templates map[string]string
}

// ValidationResult contains the results of dry-run validation
type ValidationResult struct {
	IsValid    bool
	Mappings   *FileMapping
	Errors     []string
	Operations []FileOperation
}

// FileOperation represents a file operation that would be performed
type FileOperation struct {
	Type        OperationType
	Source      string
	Target      string
	Description string
}

// OperationType represents the type of file operation
type OperationType string

const (
	OperationCreateLink     OperationType = "create_link"
	OperationCreateTemplate OperationType = "create_template"
	OperationForceLink      OperationType = "force_link"
	OperationForceTemplate  OperationType = "force_template"
	OperationSkip           OperationType = "skip"
)

// NewFileMapping creates a new empty FileMapping
func NewFileMapping() *FileMapping {
	return &FileMapping{
		sourceToTarget: make(map[string]string),
		targetToSource: make(map[string]string),
		templates:      make(map[string]string),
	}
}

// AddMapping adds a source-target mapping to the FileMapping
func (fm *FileMapping) AddMapping(source, target string) {
	fm.sourceToTarget[source] = target
	fm.targetToSource[target] = source
}

// AddTemplateMapping adds a template source-target mapping to the FileMapping
func (fm *FileMapping) AddTemplateMapping(source, target string) {
	fm.AddMapping(source, target)
	fm.templates[source] = target
}

// GetTarget returns the target path for a given source path
func (fm *FileMapping) GetTarget(source string) (string, bool) {
	target, exists := fm.sourceToTarget[source]
	return target, exists
}

// GetSource returns the source path for a given target path
func (fm *FileMapping) GetSource(target string) (string, bool) {
	source, exists := fm.targetToSource[target]
	return source, exists
}

// GetAllMappings returns all source-target mappings
func (fm *FileMapping) GetAllMappings() map[string]string {
	result := make(map[string]string)
	for source, target := range fm.sourceToTarget {
		result[source] = target
	}
	return result
}

// GetTargetConflicts returns any duplicate target mappings
func (fm *FileMapping) GetTargetConflicts() map[string][]string {
	conflicts := make(map[string][]string)
	targetToSources := make(map[string][]string)

	// Build reverse mapping of targets to all sources
	for source, target := range fm.sourceToTarget {
		targetToSources[target] = append(targetToSources[target], source)
	}

	// Find targets with multiple sources
	for target, sources := range targetToSources {
		if len(sources) > 1 {
			conflicts[target] = sources
		}
	}

	return conflicts
}

// IsTemplate checks if a source file is a template
func (fm *FileMapping) IsTemplate(source string) bool {
	_, exists := fm.templates[source]
	return exists
}

// GetTemplateMappings returns all template source-target mappings
func (fm *FileMapping) GetTemplateMappings() map[string]string {
	result := make(map[string]string)
	for source, target := range fm.templates {
		result[source] = target
	}
	return result
}

// BuildFileMapping creates a FileMapping from all modules in the config
func BuildFileMapping(modules []config.ModuleConfig) (*FileMapping, error) {
	mapping := NewFileMapping()

	for _, module := range modules {
		moduleMapping, err := buildModuleMapping(module)
		if err != nil {
			return nil, fmt.Errorf("failed to build mapping for module %s: %w", module.Dir, err)
		}

		// Merge module mapping into main mapping
		for source, target := range moduleMapping.GetAllMappings() {
			if moduleMapping.IsTemplate(source) {
				mapping.AddTemplateMapping(source, target)
			} else {
				mapping.AddMapping(source, target)
			}
		}
	}

	return mapping, nil
}

// buildModuleMapping creates a FileMapping for a single module
func buildModuleMapping(module config.ModuleConfig) (*FileMapping, error) {
	mapping := NewFileMapping()

	// Get all files in module directory
	entries, err := os.ReadDir(module.Dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read module directory %s: %w", module.Dir, err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue // Skip subdirectories for now
		}

		sourceFile := filepath.Join(module.Dir, entry.Name())

		// Skip if file is in ignores list
		if isIgnored(entry.Name(), module.Ignores) {
			continue
		}

		// Skip Dotfile config file
		if entry.Name() == "Dotfile" {
			continue
		}

		// Calculate target path
		targetName := entry.Name()
		if isTemplateFile(entry.Name()) {
			// Remove .dot-tmpl extension for target filename
			targetName = strings.TrimSuffix(entry.Name(), ".dot-tmpl")
		}
		targetFile := filepath.Join(module.TargetDir, targetName)

		if isTemplateFile(entry.Name()) {
			mapping.AddTemplateMapping(sourceFile, targetFile)
		} else {
			mapping.AddMapping(sourceFile, targetFile)
		}
	}

	return mapping, nil
}

// isIgnored checks if a file should be ignored based on the ignore patterns
func isIgnored(filename string, ignores []string) bool {
	for _, pattern := range ignores {
		if strings.Contains(filename, pattern) {
			return true
		}
	}
	return false
}

// isTemplateFile checks if a file is a template file (.dot-tmpl extension)
func isTemplateFile(filename string) bool {
	return strings.HasSuffix(filename, ".dot-tmpl")
}
