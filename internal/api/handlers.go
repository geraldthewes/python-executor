package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/geraldthewes/python-executor/internal/executor"
	"github.com/geraldthewes/python-executor/internal/storage"
	"github.com/geraldthewes/python-executor/pkg/client"
)

// Server holds the API dependencies
type Server struct {
	storage  storage.Storage
	executor executor.Executor
}

// NewServer creates a new API server
func NewServer(storage storage.Storage, exec executor.Executor) *Server {
	return &Server{
		storage:  storage,
		executor: exec,
	}
}

// ExecuteSync handles synchronous execution
// @Summary Execute code synchronously
// @Description Execute Python code and wait for result.
// @Description
// @Description IMPORTANT: Use the client libraries instead of calling this directly.
// @Description The request must be multipart/form-data with a tar archive and metadata JSON.
// @Tags execution
// @Accept multipart/form-data
// @Produce json
// @Param tar formData file true "Uncompressed tar archive containing Python files"
// @Param metadata formData string true "Execution metadata as JSON: {\"entrypoint\":\"main.py\",\"config\":{\"timeout_seconds\":300}}"
// @Success 200 {object} client.ExecutionResult "Execution completed"
// @Failure 400 {object} gin.H "Invalid request format"
// @Failure 500 {object} gin.H "Execution failed"
// @Router /exec/sync [post]
func (s *Server) ExecuteSync(c *gin.Context) {
	// Parse multipart form
	tarData, metadata, err := s.parseRequest(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Generate execution ID
	execID := fmt.Sprintf("exe_%s", uuid.New().String())

	// Create execution record
	now := time.Now()
	exec := &storage.Execution{
		ID:        execID,
		Status:    client.StatusPending,
		Metadata:  metadata,
		CreatedAt: now,
	}

	if err := s.storage.Create(c.Request.Context(), exec); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create execution"})
		return
	}

	// Update to running
	exec.Status = client.StatusRunning
	exec.StartedAt = &now
	s.storage.Update(c.Request.Context(), exec)

	// Execute
	req := &executor.ExecutionRequest{
		ID:       execID,
		TarData:  tarData,
		Metadata: metadata,
	}

	output, err := s.executor.Execute(c.Request.Context(), req)

	// Update execution with result
	finishedAt := time.Now()
	exec.FinishedAt = &finishedAt

	if err != nil {
		exec.Status = client.StatusFailed
		exec.Error = err.Error()
	} else {
		exec.Status = client.StatusCompleted
		exec.Stdout = output.Stdout
		exec.Stderr = output.Stderr
		exec.ExitCode = output.ExitCode
		exec.DurationMs = output.DurationMs
	}

	s.storage.Update(c.Request.Context(), exec)

	// Return result
	c.JSON(http.StatusOK, exec.ToExecutionResult())
}

// ExecuteAsync handles asynchronous execution
// @Summary Execute code asynchronously
// @Description Submit code for execution and return immediately with an execution ID.
// @Description
// @Description IMPORTANT: Use the client libraries instead of calling this directly.
// @Description The request must be multipart/form-data with a tar archive and metadata JSON.
// @Tags execution
// @Accept multipart/form-data
// @Produce json
// @Param tar formData file true "Uncompressed tar archive containing Python files"
// @Param metadata formData string true "Execution metadata as JSON: {\"entrypoint\":\"main.py\"}"
// @Success 202 {object} client.AsyncResponse "Execution submitted"
// @Failure 400 {object} gin.H "Invalid request format"
// @Failure 500 {object} gin.H "Failed to create execution"
// @Router /exec/async [post]
func (s *Server) ExecuteAsync(c *gin.Context) {
	// Parse multipart form
	tarData, metadata, err := s.parseRequest(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Generate execution ID
	execID := fmt.Sprintf("exe_%s", uuid.New().String())

	// Create execution record
	exec := &storage.Execution{
		ID:        execID,
		Status:    client.StatusPending,
		Metadata:  metadata,
		CreatedAt: time.Now(),
	}

	if err := s.storage.Create(c.Request.Context(), exec); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create execution"})
		return
	}

	// Execute in background
	go s.executeAsync(execID, tarData, metadata)

	// Return execution ID immediately
	c.JSON(http.StatusAccepted, client.AsyncResponse{
		ExecutionID: execID,
	})
}

// GetExecution retrieves execution status
// @Summary Get execution status
// @Description Retrieve the status and result of an execution.
// @Description Status values: pending, running, completed, failed, killed
// @Tags execution
// @Produce json
// @Param id path string true "Execution ID (e.g., exe_550e8400-e29b-41d4-a716-446655440000)"
// @Success 200 {object} client.ExecutionResult "Execution status and result"
// @Failure 404 {object} gin.H "Execution not found"
// @Router /executions/{id} [get]
func (s *Server) GetExecution(c *gin.Context) {
	id := c.Param("id")

	exec, err := s.storage.Get(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "execution not found"})
		return
	}

	c.JSON(http.StatusOK, exec.ToExecutionResult())
}

// KillExecution terminates a running execution
// @Summary Kill execution
// @Description Terminate a running execution.
// @Description If the execution is not running, returns the current status.
// @Tags execution
// @Produce json
// @Param id path string true "Execution ID (e.g., exe_550e8400-e29b-41d4-a716-446655440000)"
// @Success 200 {object} client.KillResponse "Execution killed or current status"
// @Failure 404 {object} gin.H "Execution not found"
// @Failure 500 {object} gin.H "Failed to kill execution"
// @Router /executions/{id} [delete]
func (s *Server) KillExecution(c *gin.Context) {
	id := c.Param("id")

	exec, err := s.storage.Get(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "execution not found"})
		return
	}

	// Only kill if running
	if exec.Status != client.StatusRunning {
		c.JSON(http.StatusOK, client.KillResponse{Status: string(exec.Status)})
		return
	}

	// Kill container
	if exec.ContainerID != "" {
		if err := s.executor.Kill(c.Request.Context(), exec.ContainerID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to kill container"})
			return
		}
	}

	// Update status
	exec.Status = client.StatusKilled
	s.storage.Update(c.Request.Context(), exec)

	c.JSON(http.StatusOK, client.KillResponse{Status: "killed"})
}

// parseRequest parses multipart form data
func (s *Server) parseRequest(c *gin.Context) ([]byte, *client.Metadata, error) {
	// Parse multipart form
	if err := c.Request.ParseMultipartForm(100 << 20); err != nil { // 100 MB max
		return nil, nil, fmt.Errorf("parsing form: %w", err)
	}

	// Get tar file
	tarFile, _, err := c.Request.FormFile("tar")
	if err != nil {
		return nil, nil, fmt.Errorf("missing tar file: %w", err)
	}
	defer tarFile.Close()

	tarData, err := io.ReadAll(tarFile)
	if err != nil {
		return nil, nil, fmt.Errorf("reading tar: %w", err)
	}

	// Get metadata
	metadataStr := c.Request.FormValue("metadata")
	if metadataStr == "" {
		return nil, nil, fmt.Errorf("missing metadata")
	}

	var metadata client.Metadata
	if err := json.Unmarshal([]byte(metadataStr), &metadata); err != nil {
		return nil, nil, fmt.Errorf("parsing metadata: %w", err)
	}

	return tarData, &metadata, nil
}

// executeAsync runs execution in background
func (s *Server) executeAsync(execID string, tarData []byte, metadata *client.Metadata) {
	ctx := context.Background()

	// Get execution
	exec, err := s.storage.Get(ctx, execID)
	if err != nil {
		return
	}

	// Update to running
	now := time.Now()
	exec.Status = client.StatusRunning
	exec.StartedAt = &now
	s.storage.Update(ctx, exec)

	// Execute
	req := &executor.ExecutionRequest{
		ID:       execID,
		TarData:  tarData,
		Metadata: metadata,
	}

	output, err := s.executor.Execute(ctx, req)

	// Update with result
	finishedAt := time.Now()
	exec.FinishedAt = &finishedAt

	if err != nil {
		exec.Status = client.StatusFailed
		exec.Error = err.Error()
	} else {
		exec.Status = client.StatusCompleted
		exec.Stdout = output.Stdout
		exec.Stderr = output.Stderr
		exec.ExitCode = output.ExitCode
		exec.DurationMs = output.DurationMs
	}

	s.storage.Update(ctx, exec)
}
