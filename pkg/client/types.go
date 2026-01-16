package client

import "time"

// ExecutionStatus represents the status of an execution
type ExecutionStatus string

const (
	StatusPending   ExecutionStatus = "pending"
	StatusRunning   ExecutionStatus = "running"
	StatusCompleted ExecutionStatus = "completed"
	StatusFailed    ExecutionStatus = "failed"
	StatusKilled    ExecutionStatus = "killed"
)

// Metadata contains execution parameters
type Metadata struct {
	Entrypoint      string           `json:"entrypoint"`
	DockerImage     string           `json:"docker_image,omitempty"`
	RequirementsTxt string           `json:"requirements_txt,omitempty"`
	PreCommands     []string         `json:"pre_commands,omitempty"`
	Stdin           string           `json:"stdin,omitempty"`
	Config          *ExecutionConfig `json:"config,omitempty"`
	EnvVars         []string         `json:"env_vars,omitempty"`
	ScriptArgs      []string         `json:"script_args,omitempty"`
}

// ExecutionConfig holds resource limits and settings
type ExecutionConfig struct {
	TimeoutSeconds  int  `json:"timeout_seconds,omitempty"`
	NetworkDisabled bool `json:"network_disabled,omitempty"`
	MemoryMB        int  `json:"memory_mb,omitempty"`
	DiskMB          int  `json:"disk_mb,omitempty"`
	CPUShares       int  `json:"cpu_shares,omitempty"`
}

// ExecutionResult represents the result of a code execution
type ExecutionResult struct {
	ExecutionID string          `json:"execution_id"`
	Status      ExecutionStatus `json:"status"`
	Stdout      string          `json:"stdout,omitempty"`
	Stderr      string          `json:"stderr,omitempty"`
	ExitCode    int             `json:"exit_code,omitempty"`
	Error       string          `json:"error,omitempty"`
	ErrorType   string          `json:"error_type,omitempty"`  // Python error type (e.g., "SyntaxError", "NameError")
	ErrorLine   int             `json:"error_line,omitempty"`  // Line number where error occurred
	StartedAt   *time.Time      `json:"started_at,omitempty"`
	FinishedAt  *time.Time      `json:"finished_at,omitempty"`
	DurationMs  int64           `json:"duration_ms,omitempty"`
}

// AsyncResponse is returned when submitting async execution
type AsyncResponse struct {
	ExecutionID string `json:"execution_id"`
}

// KillResponse is returned when killing an execution
type KillResponse struct {
	Status string `json:"status"`
}

// SimpleExecRequest is the JSON-only execution request format
// Compatible with Replit/Piston-style APIs for simpler integrations
type SimpleExecRequest struct {
	// Code is the source code to execute (for single-file execution)
	// If provided, creates a main.py with this content
	Code string `json:"code,omitempty"`

	// Files allows multiple files to be provided (Piston-compatible)
	// Takes precedence over Code if both are provided
	Files []CodeFile `json:"files,omitempty"`

	// Entrypoint is the file to execute (defaults to "main.py")
	Entrypoint string `json:"entrypoint,omitempty"`

	// Stdin is the standard input to provide to the script
	Stdin string `json:"stdin,omitempty"`

	// Config contains execution resource limits
	Config *ExecutionConfig `json:"config,omitempty"`

	// PythonVersion specifies the Python version to use (e.g., "3.10", "3.11", "3.12", "3.13")
	// If not specified, uses the server default (typically 3.12)
	PythonVersion string `json:"python_version,omitempty"`
}

// CodeFile represents a single file with its content
type CodeFile struct {
	Name    string `json:"name"`    // filename (e.g., "main.py")
	Content string `json:"content"` // file content
}
