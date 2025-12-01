package validation

import (
	"github.com/elmhuangyu/dotman/pkg/module"
)

// ValidationResult contains the results of dry-run validation
type ValidationResult struct {
	IsValid    bool
	Mappings   *module.FileMapping
	Errors     []string
	Operations []module.FileOperation
}
