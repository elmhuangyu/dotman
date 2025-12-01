package filesystem

import (
	"os"
)

// PathResolver handles path resolution utilities
type PathResolver struct{}

// NewPathResolver creates a new PathResolver
func NewPathResolver() *PathResolver {
	return &PathResolver{}
}

// EnsureDirExists ensures a directory exists, creating it if necessary
func (pr *PathResolver) EnsureDirExists(path string) error {
	return os.MkdirAll(path, 0755)
}

// DirExists checks if a directory exists
func (pr *PathResolver) DirExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// FileExists checks if a file exists
func (pr *PathResolver) FileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}
