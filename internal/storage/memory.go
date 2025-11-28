package storage

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/geraldthewes/python-executor/pkg/client"
)

// MemoryStorage implements in-memory storage with mutex protection
type MemoryStorage struct {
	mu         sync.RWMutex
	executions map[string]*Execution
}

// NewMemoryStorage creates a new in-memory storage backend
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		executions: make(map[string]*Execution),
	}
}

// Create creates a new execution record
func (m *MemoryStorage) Create(ctx context.Context, exec *Execution) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.executions[exec.ID]; exists {
		return fmt.Errorf("execution %s already exists", exec.ID)
	}

	m.executions[exec.ID] = exec
	return nil
}

// Get retrieves an execution by ID
func (m *MemoryStorage) Get(ctx context.Context, id string) (*Execution, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	exec, exists := m.executions[id]
	if !exists {
		return nil, fmt.Errorf("execution %s not found", id)
	}

	return exec, nil
}

// Update updates an existing execution
func (m *MemoryStorage) Update(ctx context.Context, exec *Execution) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.executions[exec.ID]; !exists {
		return fmt.Errorf("execution %s not found", exec.ID)
	}

	m.executions[exec.ID] = exec
	return nil
}

// Delete removes an execution
func (m *MemoryStorage) Delete(ctx context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.executions, id)
	return nil
}

// List returns all executions (optionally filtered by status)
func (m *MemoryStorage) List(ctx context.Context, status *client.ExecutionStatus) ([]*Execution, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*Execution

	for _, exec := range m.executions {
		if status == nil || exec.Status == *status {
			result = append(result, exec)
		}
	}

	return result, nil
}

// Cleanup removes executions older than the given duration
func (m *MemoryStorage) Cleanup(ctx context.Context, olderThan time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	cutoff := time.Now().Add(-olderThan)

	for id, exec := range m.executions {
		// Only cleanup completed/failed/killed executions
		if exec.Status == client.StatusCompleted ||
			exec.Status == client.StatusFailed ||
			exec.Status == client.StatusKilled {

			if exec.CreatedAt.Before(cutoff) {
				delete(m.executions, id)
			}
		}
	}

	return nil
}

// Close is a no-op for memory storage
func (m *MemoryStorage) Close() error {
	return nil
}
