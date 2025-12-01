package module

import "fmt"

// OperationType represents the type of operation performed
type OperationType string

const (
	OperationCreateLink     OperationType = "create_link"
	OperationCreateTemplate OperationType = "create_template"
	OperationForceLink      OperationType = "force_link"
	OperationForceTemplate  OperationType = "force_template"
	OperationSkip           OperationType = "skip"
)

// OperationResult unified result type for all operations
type OperationResult struct {
	Type     OperationType          `json:"type"`
	Source   string                 `json:"source"`
	Target   string                 `json:"target"`
	Success  bool                   `json:"success"`
	Error    error                  `json:"error,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// ResultSummary for consistent reporting across operations
type ResultSummary struct {
	Total      int      `json:"total"`
	Successful int      `json:"successful"`
	Failed     int      `json:"failed"`
	Skipped    int      `json:"skipped"`
	Errors     []string `json:"errors,omitempty"`
}

// AddResult adds a result to the summary
func (rs *ResultSummary) AddResult(result OperationResult) {
	rs.Total++
	if result.Success {
		rs.Successful++
	} else {
		rs.Failed++
		if result.Error != nil {
			rs.Errors = append(rs.Errors, result.Error.Error())
		}
	}
}

// AddSkipped increments the skipped count
func (rs *ResultSummary) AddSkipped() {
	rs.Total++
	rs.Skipped++
}

// HasErrors returns true if there are any errors
func (rs *ResultSummary) HasErrors() bool {
	return rs.Failed > 0 || len(rs.Errors) > 0
}

// SuccessRate returns the success rate as a percentage
func (rs *ResultSummary) SuccessRate() float64 {
	if rs.Total == 0 {
		return 0
	}
	return float64(rs.Successful) / float64(rs.Total) * 100
}

// InstallationError represents an error that occurred during installation
type InstallationError struct {
	Operation string
	Source    string
	Target    string
	Cause     error
}

// Error implements the error interface
func (e *InstallationError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("installation error: %s operation from %s to %s failed: %v", e.Operation, e.Source, e.Target, e.Cause)
	}
	return fmt.Sprintf("installation error: %s operation from %s to %s failed", e.Operation, e.Source, e.Target)
}

// Unwrap returns the underlying cause
func (e *InstallationError) Unwrap() error {
	return e.Cause
}

// ValidationError represents an error that occurred during validation
type ValidationError struct {
	Type   string
	Path   string
	Reason string
}

// Error implements the error interface
func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error: %s at %s: %s", e.Type, e.Path, e.Reason)
}

// InstallConfig contains configuration for install operations
type InstallConfig struct {
	Mkdir     bool              `json:"mkdir"`
	Force     bool              `json:"force"`
	DryRun    bool              `json:"dry_run"`
	Vars      map[string]string `json:"vars,omitempty"`
	StatePath string            `json:"state_path"`
}

// UninstallConfig contains configuration for uninstall operations
type UninstallConfig struct {
	BackupModified bool   `json:"backup_modified"`
	StatePath      string `json:"state_path"`
}
