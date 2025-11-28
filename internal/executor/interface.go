package executor

import (
	"context"

	"github.com/geraldthewes/python-executor/pkg/client"
)

// ExecutionRequest contains all data needed for execution
type ExecutionRequest struct {
	ID        string
	TarData   []byte
	Metadata  *client.Metadata
}

// ExecutionOutput contains the execution results
type ExecutionOutput struct {
	Stdout     string
	Stderr     string
	ExitCode   int
	DurationMs int64
}

// Executor defines the interface for code execution
type Executor interface {
	// Execute runs code in a sandboxed environment
	Execute(ctx context.Context, req *ExecutionRequest) (*ExecutionOutput, error)

	// Kill terminates a running execution
	Kill(ctx context.Context, containerID string) error

	// Close cleans up executor resources
	Close() error
}
