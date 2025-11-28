package storage

import (
	"context"
	"testing"
	"time"

	"github.com/geraldthewes/python-executor/pkg/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMemoryStorage_CreateAndGet(t *testing.T) {
	store := NewMemoryStorage()
	ctx := context.Background()

	exec := &Execution{
		ID:        "test-1",
		Status:    client.StatusPending,
		CreatedAt: time.Now(),
	}

	// Create
	err := store.Create(ctx, exec)
	require.NoError(t, err)

	// Get
	retrieved, err := store.Get(ctx, "test-1")
	require.NoError(t, err)
	assert.Equal(t, exec.ID, retrieved.ID)
	assert.Equal(t, exec.Status, retrieved.Status)
}

func TestMemoryStorage_CreateDuplicate(t *testing.T) {
	store := NewMemoryStorage()
	ctx := context.Background()

	exec := &Execution{
		ID:        "test-1",
		Status:    client.StatusPending,
		CreatedAt: time.Now(),
	}

	// Create first time
	err := store.Create(ctx, exec)
	require.NoError(t, err)

	// Try to create again
	err = store.Create(ctx, exec)
	assert.Error(t, err)
}

func TestMemoryStorage_Update(t *testing.T) {
	store := NewMemoryStorage()
	ctx := context.Background()

	exec := &Execution{
		ID:        "test-1",
		Status:    client.StatusPending,
		CreatedAt: time.Now(),
	}

	// Create
	require.NoError(t, store.Create(ctx, exec))

	// Update
	exec.Status = client.StatusRunning
	err := store.Update(ctx, exec)
	require.NoError(t, err)

	// Verify
	retrieved, err := store.Get(ctx, "test-1")
	require.NoError(t, err)
	assert.Equal(t, client.StatusRunning, retrieved.Status)
}

func TestMemoryStorage_Delete(t *testing.T) {
	store := NewMemoryStorage()
	ctx := context.Background()

	exec := &Execution{
		ID:        "test-1",
		Status:    client.StatusPending,
		CreatedAt: time.Now(),
	}

	// Create
	require.NoError(t, store.Create(ctx, exec))

	// Delete
	err := store.Delete(ctx, "test-1")
	require.NoError(t, err)

	// Verify deleted
	_, err = store.Get(ctx, "test-1")
	assert.Error(t, err)
}

func TestMemoryStorage_List(t *testing.T) {
	store := NewMemoryStorage()
	ctx := context.Background()

	// Create multiple executions
	execs := []*Execution{
		{ID: "test-1", Status: client.StatusPending, CreatedAt: time.Now()},
		{ID: "test-2", Status: client.StatusRunning, CreatedAt: time.Now()},
		{ID: "test-3", Status: client.StatusCompleted, CreatedAt: time.Now()},
	}

	for _, exec := range execs {
		require.NoError(t, store.Create(ctx, exec))
	}

	// List all
	all, err := store.List(ctx, nil)
	require.NoError(t, err)
	assert.Len(t, all, 3)

	// List by status
	pending := client.StatusPending
	filtered, err := store.List(ctx, &pending)
	require.NoError(t, err)
	assert.Len(t, filtered, 1)
	assert.Equal(t, "test-1", filtered[0].ID)
}

func TestMemoryStorage_Cleanup(t *testing.T) {
	store := NewMemoryStorage()
	ctx := context.Background()

	now := time.Now()

	// Create old completed execution
	old := &Execution{
		ID:        "old-1",
		Status:    client.StatusCompleted,
		CreatedAt: now.Add(-10 * time.Minute),
	}
	require.NoError(t, store.Create(ctx, old))

	// Create recent execution
	recent := &Execution{
		ID:        "recent-1",
		Status:    client.StatusCompleted,
		CreatedAt: now.Add(-1 * time.Minute),
	}
	require.NoError(t, store.Create(ctx, recent))

	// Create running execution (should not be cleaned)
	running := &Execution{
		ID:        "running-1",
		Status:    client.StatusRunning,
		CreatedAt: now.Add(-20 * time.Minute),
	}
	require.NoError(t, store.Create(ctx, running))

	// Cleanup executions older than 5 minutes
	err := store.Cleanup(ctx, 5*time.Minute)
	require.NoError(t, err)

	// Verify old completed is deleted
	_, err = store.Get(ctx, "old-1")
	assert.Error(t, err)

	// Verify recent is still there
	_, err = store.Get(ctx, "recent-1")
	assert.NoError(t, err)

	// Verify running is still there (not cleaned even if old)
	_, err = store.Get(ctx, "running-1")
	assert.NoError(t, err)
}
