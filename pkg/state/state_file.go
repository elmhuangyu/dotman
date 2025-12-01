package state

import (
	"crypto/sha1"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const (
	// version of the state file, not the program, it used to check complibilty.
	version = "1.0.0"

	TypeLink      = "link"
	TypeGenerated = "generated"
)

type FileMapping struct {
	Source string `yaml:"source"`
	Target string `yaml:"target"`
	Type   string `yaml:"type"`           // link, generated
	SHA1   string `yaml:"sha1,omitempty"` // only for generated file
}

type StateFile struct {
	Version string        `yaml:"version"`
	Files   []FileMapping `yaml:"files"`
}

// LoadStateFile loads the state file from the given path
func LoadStateFile(path string) (*StateFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // State file doesn't exist, return nil
		}
		return nil, fmt.Errorf("failed to read state file: %w", err)
	}

	var stateFile StateFile
	if err := yaml.Unmarshal(data, &stateFile); err != nil {
		return nil, fmt.Errorf("failed to parse state file: %w", err)
	}

	return &stateFile, nil
}

// SaveStateFile saves the state file to the given path atomically
func SaveStateFile(path string, stateFile *StateFile) error {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create state directory: %w", err)
	}

	// Marshal to YAML
	data, err := yaml.Marshal(stateFile)
	if err != nil {
		return fmt.Errorf("failed to marshal state file: %w", err)
	}

	// Write to temporary file first
	tempPath := path + ".tmp"
	if err := os.WriteFile(tempPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write temporary state file: %w", err)
	}

	// Atomic rename
	if err := os.Rename(tempPath, path); err != nil {
		os.Remove(tempPath) // Clean up temp file
		return fmt.Errorf("failed to rename state file: %w", err)
	}

	return nil
}

// NewStateFile creates a new state file with the current version
func NewStateFile() *StateFile {
	return &StateFile{
		Version: version,
		Files:   []FileMapping{},
	}
}

// AddFileMapping adds a file mapping to the state file
func (sf *StateFile) AddFileMapping(source, target, fileType string) {
	// Convert to absolute paths
	absSource, err := filepath.Abs(source)
	if err != nil {
		absSource = source // fallback to original if conversion fails
	}

	absTarget, err := filepath.Abs(target)
	if err != nil {
		absTarget = target // fallback to original if conversion fails
	}

	// Check if mapping already exists
	for _, existing := range sf.Files {
		if existing.Source == absSource && existing.Target == absTarget {
			return // Already exists, don't add duplicate
		}
	}

	mapping := FileMapping{
		Source: absSource,
		Target: absTarget,
		Type:   fileType,
	}

	// Calculate SHA1 for generated files
	if fileType == TypeGenerated {
		if sha1, err := calculateSHA1(absTarget); err != nil {
			// Log warning but continue - SHA1 failure shouldn't break installation
			fmt.Printf("Warning: failed to calculate SHA1 for %s: %v\n", absTarget, err)
		} else {
			mapping.SHA1 = sha1
		}
	}

	sf.Files = append(sf.Files, mapping)
}

// calculateSHA1 computes the SHA1 hash of a file's content
func calculateSHA1(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file for SHA1 calculation: %w", err)
	}
	defer file.Close()

	hasher := sha1.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", fmt.Errorf("failed to read file for SHA1 calculation: %w", err)
	}

	hash := hasher.Sum(nil)
	return fmt.Sprintf("%x", hash), nil
}
