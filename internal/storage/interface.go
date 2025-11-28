package storage

import (
	"context"
	"time"

	"github.com/geraldthewes/python-executor/pkg/client"
)

// Execution represents a stored execution state
type Execution struct {
	ID          string
	Status      client.ExecutionStatus
	Metadata    *client.Metadata
	Stdout      string
	Stderr      string
	ExitCode    int
	Error       string
	StartedAt   *time.Time
	FinishedAt  *time.Time
	DurationMs  int64
	ContainerID string // Docker container ID for running executions
	CreatedAt   time.Time
}

// Storage defines the interface for execution state storage
type Storage interface {
	// Create creates a new execution record
	Create(ctx context.Context, exec *Execution) error

	// Get retrieves an execution by ID
	Get(ctx context.Context, id string) (*Execution, error)

	// Update updates an existing execution
	Update(ctx context.Context, exec *Execution) error

	// Delete removes an execution
	Delete(ctx context.Context, id string) error

	// List returns all executions (optionally filtered by status)
	List(ctx context.Context, status *client.ExecutionStatus) ([]*Execution, error)

	// Cleanup removes executions older than the given duration
	Cleanup(ctx context.Context, olderThan time.Duration) error

	// Close closes the storage backend
	Close() error
}

// ToExecutionResult converts a storage Execution to a client ExecutionResult
func (e *Execution) ToExecutionResult() *client.ExecutionResult {
	return &client.ExecutionResult{
		ExecutionID: e.ID,
		Status:      e.Status,
		Stdout:      e.Stdout,
		Stderr:      e.Stderr,
		ExitCode:    e.ExitCode,
		Error:       e.Error,
		StartedAt:   e.StartedAt,
		FinishedAt:  e.FinishedAt,
		DurationMs:  e.DurationMs,
	}
}
