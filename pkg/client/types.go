package client

import "time"

// ExecutionStatus represents the status of a code execution.
type ExecutionStatus string

// Execution status constants.
const (
	// StatusPending indicates the execution is queued but not yet started.
	StatusPending ExecutionStatus = "pending"
	// StatusRunning indicates the execution is currently in progress.
	StatusRunning ExecutionStatus = "running"
	// StatusCompleted indicates the execution finished (check ExitCode for success).
	StatusCompleted ExecutionStatus = "completed"
	// StatusFailed indicates the execution failed due to an internal error.
	StatusFailed ExecutionStatus = "failed"
	// StatusKilled indicates the execution was terminated by the user.
	StatusKilled ExecutionStatus = "killed"
)

// Metadata contains execution parameters sent to the server.
//
// At minimum, Entrypoint must be specified. All other fields are optional.
//
// Example:
//
//	metadata := &client.Metadata{
//	    Entrypoint:      "main.py",
//	    RequirementsTxt: "requests\nnumpy",
//	    Config: &client.ExecutionConfig{
//	        NetworkDisabled: false,  // Allow network for pip
//	    },
//	}
type Metadata struct {
	// Entrypoint is the Python file to execute (required).
	Entrypoint string `json:"entrypoint"`
	// DockerImage is the Docker image to use (default: python:3.11-slim).
	DockerImage string `json:"docker_image,omitempty"`
	// RequirementsTxt is the contents of requirements.txt for pip install.
	RequirementsTxt string `json:"requirements_txt,omitempty"`
	// PreCommands are shell commands to run before Python execution.
	PreCommands []string `json:"pre_commands,omitempty"`
	// Stdin is data to provide on standard input.
	Stdin string `json:"stdin,omitempty"`
	// Config contains resource limits and settings.
	Config *ExecutionConfig `json:"config,omitempty"`
	// EnvVars are environment variables in "KEY=value" format.
	EnvVars []string `json:"env_vars,omitempty"`
	// ScriptArgs are arguments passed to the Python script (sys.argv).
	ScriptArgs []string `json:"script_args,omitempty"`
	// EvalLastExpr enables REPL-style behavior for simple execution.
	// Internal use only - set via SimpleExecRequest.EvalLastExpr.
	EvalLastExpr bool `json:"-"`
}

// ExecutionConfig holds resource limits and execution settings.
//
// Example:
//
//	config := &client.ExecutionConfig{
//	    TimeoutSeconds:  60,
//	    NetworkDisabled: false,
//	    MemoryMB:        2048,
//	}
type ExecutionConfig struct {
	// TimeoutSeconds is the maximum execution time (default: 300).
	TimeoutSeconds int `json:"timeout_seconds,omitempty"`
	// NetworkDisabled disables network access if true (default: true).
	NetworkDisabled bool `json:"network_disabled,omitempty"`
	// MemoryMB is the memory limit in megabytes (default: 1024).
	MemoryMB int `json:"memory_mb,omitempty"`
	// DiskMB is the disk space limit in megabytes (default: 2048).
	DiskMB int `json:"disk_mb,omitempty"`
	// CPUShares is the CPU shares (relative weight, default: 1024).
	CPUShares int `json:"cpu_shares,omitempty"`
}

// ExecutionResult contains the output and status of an execution.
type ExecutionResult struct {
	// ExecutionID is the unique identifier for this execution.
	ExecutionID string `json:"execution_id"`
	// Status is the current execution state.
	Status ExecutionStatus `json:"status"`
	// Stdout is the standard output from the Python script.
	Stdout string `json:"stdout,omitempty"`
	// Stderr is the standard error from the Python script.
	Stderr string `json:"stderr,omitempty"`
	// ExitCode is the process exit code (0 = success).
	ExitCode int `json:"exit_code"`
	// Error is an error message if the execution failed internally.
	Error string `json:"error,omitempty"`
	// ErrorType is the Python exception type (e.g., "SyntaxError", "NameError").
	ErrorType string `json:"error_type,omitempty"`
	// ErrorLine is the line number where the error occurred.
	ErrorLine int `json:"error_line,omitempty"`
	// StartedAt is when execution started (UTC).
	StartedAt *time.Time `json:"started_at,omitempty"`
	// FinishedAt is when execution finished (UTC).
	FinishedAt *time.Time `json:"finished_at,omitempty"`
	// DurationMs is the total execution time in milliseconds.
	DurationMs int64 `json:"duration_ms,omitempty"`
	// Result contains the value of the last expression when EvalLastExpr is true.
	// The value is the repr() of the Python object, or null if the last
	// statement was not an expression.
	Result *string `json:"result,omitempty"`
}

// AsyncResponse is returned when submitting async execution.
type AsyncResponse struct {
	ExecutionID string `json:"execution_id"`
}

// KillResponse is returned when killing an execution.
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

	// EvalLastExpr enables REPL-style behavior: if the last statement is an
	// expression, its value is captured and returned in the Result field.
	// Only applies to single-file code execution.
	EvalLastExpr bool `json:"eval_last_expr,omitempty"`
}

// CodeFile represents a single file with its content
type CodeFile struct {
	Name    string `json:"name"`    // filename (e.g., "main.py")
	Content string `json:"content"` // file content
}
