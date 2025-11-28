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
	Entrypoint      string         `json:"entrypoint"`
	DockerImage     string         `json:"docker_image,omitempty"`
	RequirementsTxt string         `json:"requirements_txt,omitempty"`
	PreCommands     []string       `json:"pre_commands,omitempty"`
	Stdin           string         `json:"stdin,omitempty"`
	Config          *ExecutionConfig `json:"config,omitempty"`
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
