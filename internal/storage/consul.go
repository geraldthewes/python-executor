package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	consulapi "github.com/hashicorp/consul/api"
	"github.com/geraldthewes/python-executor/pkg/client"
)

// ConsulStorage implements storage using Consul KV
type ConsulStorage struct {
	client    *consulapi.Client
	keyPrefix string
}

// NewConsulStorage creates a new Consul-backed storage
func NewConsulStorage(address, token, keyPrefix string) (*ConsulStorage, error) {
	config := consulapi.DefaultConfig()
	config.Address = address
	if token != "" {
		config.Token = token
	}

	client, err := consulapi.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("creating consul client: %w", err)
	}

	return &ConsulStorage{
		client:    client,
		keyPrefix: keyPrefix,
	}, nil
}

// Create creates a new execution record
func (c *ConsulStorage) Create(ctx context.Context, exec *Execution) error {
	key := c.executionKey(exec.ID)

	// Check if exists
	kv := c.client.KV()
	existing, _, err := kv.Get(key, nil)
	if err != nil {
		return fmt.Errorf("checking existing key: %w", err)
	}
	if existing != nil {
		return fmt.Errorf("execution %s already exists", exec.ID)
	}

	// Serialize and store
	data, err := json.Marshal(exec)
	if err != nil {
		return fmt.Errorf("marshaling execution: %w", err)
	}

	p := &consulapi.KVPair{
		Key:   key,
		Value: data,
	}

	_, err = kv.Put(p, nil)
	if err != nil {
		return fmt.Errorf("storing execution: %w", err)
	}

	return nil
}

// Get retrieves an execution by ID
func (c *ConsulStorage) Get(ctx context.Context, id string) (*Execution, error) {
	key := c.executionKey(id)

	kv := c.client.KV()
	pair, _, err := kv.Get(key, nil)
	if err != nil {
		return nil, fmt.Errorf("getting key: %w", err)
	}
	if pair == nil {
		return nil, fmt.Errorf("execution %s not found", id)
	}

	var exec Execution
	if err := json.Unmarshal(pair.Value, &exec); err != nil {
		return nil, fmt.Errorf("unmarshaling execution: %w", err)
	}

	return &exec, nil
}

// Update updates an existing execution
func (c *ConsulStorage) Update(ctx context.Context, exec *Execution) error {
	key := c.executionKey(exec.ID)

	data, err := json.Marshal(exec)
	if err != nil {
		return fmt.Errorf("marshaling execution: %w", err)
	}

	p := &consulapi.KVPair{
		Key:   key,
		Value: data,
	}

	kv := c.client.KV()
	_, err = kv.Put(p, nil)
	if err != nil {
		return fmt.Errorf("updating execution: %w", err)
	}

	return nil
}

// Delete removes an execution
func (c *ConsulStorage) Delete(ctx context.Context, id string) error {
	key := c.executionKey(id)

	kv := c.client.KV()
	_, err := kv.Delete(key, nil)
	if err != nil {
		return fmt.Errorf("deleting execution: %w", err)
	}

	return nil
}

// List returns all executions (optionally filtered by status)
func (c *ConsulStorage) List(ctx context.Context, status *client.ExecutionStatus) ([]*Execution, error) {
	prefix := c.keyPrefix + "/executions/"

	kv := c.client.KV()
	pairs, _, err := kv.List(prefix, nil)
	if err != nil {
		return nil, fmt.Errorf("listing executions: %w", err)
	}

	var result []*Execution

	for _, pair := range pairs {
		var exec Execution
		if err := json.Unmarshal(pair.Value, &exec); err != nil {
			continue // Skip malformed entries
		}

		if status == nil || exec.Status == *status {
			result = append(result, &exec)
		}
	}

	return result, nil
}

// Cleanup removes executions older than the given duration
func (c *ConsulStorage) Cleanup(ctx context.Context, olderThan time.Duration) error {
	cutoff := time.Now().Add(-olderThan)

	executions, err := c.List(ctx, nil)
	if err != nil {
		return err
	}

	for _, exec := range executions {
		// Only cleanup completed/failed/killed executions
		if exec.Status == client.StatusCompleted ||
			exec.Status == client.StatusFailed ||
			exec.Status == client.StatusKilled {

			if exec.CreatedAt.Before(cutoff) {
				if err := c.Delete(ctx, exec.ID); err != nil {
					// Log error but continue cleanup
					continue
				}
			}
		}
	}

	return nil
}

// Close closes the Consul client
func (c *ConsulStorage) Close() error {
	return nil // Consul client doesn't need explicit closing
}

// executionKey generates the Consul key for an execution
func (c *ConsulStorage) executionKey(id string) string {
	return fmt.Sprintf("%s/executions/%s", c.keyPrefix, id)
}
