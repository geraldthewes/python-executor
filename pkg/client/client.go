// Package client provides a Go client for the python-executor service.
//
// The python-executor service allows remote execution of Python code in
// isolated Docker containers. This client handles all the complexity of
// the API, including tar archive creation, multipart requests, and response parsing.
//
// # Quick Start
//
//	c := client.New("http://pyexec.cluster:9999/")
//
//	tarData, _ := client.TarFromMap(map[string]string{
//	    "main.py": `print("Hello from Go!")`,
//	})
//
//	result, err := c.ExecuteSync(context.Background(), tarData, &client.Metadata{
//	    Entrypoint: "main.py",
//	})
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	fmt.Println(result.Stdout)
//
// # Creating Tar Archives
//
// Several helper functions are provided for creating tar archives:
//
//   - TarFromMap: Create from a map of filename to content
//   - TarFromFiles: Create from a list of file paths
//   - TarFromDirectory: Create from a directory path
//   - TarFromReader: Create from an io.Reader (e.g., stdin)
package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"time"
)

// Client is the Go client for the python-executor service.
//
// Create a new client with [New] and use methods like [Client.ExecuteSync]
// and [Client.ExecuteAsync] to execute Python code remotely.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// New creates a new python-executor client.
//
// The baseURL should point to the python-executor server, e.g.,
// "http://pyexec.cluster:9999/" or "http://localhost:8080".
//
// Options can be used to customize the client:
//
//	c := client.New("http://localhost:8080",
//	    client.WithTimeout(60 * time.Second),
//	)
func New(baseURL string, opts ...Option) *Client {
	c := &Client{
		baseURL: strings.TrimSuffix(baseURL, "/"),
		httpClient: &http.Client{
			Timeout: 5 * time.Minute,
		},
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// ExecuteSync executes Python code and waits for the result.
//
// This method blocks until execution completes. Use [Client.ExecuteAsync]
// for long-running scripts.
//
// Example:
//
//	tarData, _ := client.TarFromMap(map[string]string{
//	    "main.py": `print("Hello!")`,
//	})
//
//	result, err := c.ExecuteSync(ctx, tarData, &client.Metadata{
//	    Entrypoint: "main.py",
//	})
//	if err != nil {
//	    return err
//	}
//
//	fmt.Printf("Exit code: %d\n", result.ExitCode)
//	fmt.Printf("Output: %s\n", result.Stdout)
func (c *Client) ExecuteSync(ctx context.Context, tarData []byte, metadata *Metadata) (*ExecutionResult, error) {
	body, contentType, err := c.buildMultipartRequest(tarData, metadata)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/v1/exec/sync", body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentType)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned %d", resp.StatusCode)
	}

	var result ExecutionResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

// ExecuteAsync submits Python code for asynchronous execution.
//
// Returns an execution ID immediately. Use [Client.GetExecution] to check
// status or [Client.WaitForCompletion] to wait for the result.
//
// Example:
//
//	execID, err := c.ExecuteAsync(ctx, tarData, &client.Metadata{
//	    Entrypoint: "main.py",
//	})
//	if err != nil {
//	    return err
//	}
//
//	// Later, wait for completion
//	result, err := c.WaitForCompletion(ctx, execID, 2*time.Second)
func (c *Client) ExecuteAsync(ctx context.Context, tarData []byte, metadata *Metadata) (string, error) {
	body, contentType, err := c.buildMultipartRequest(tarData, metadata)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/v1/exec/async", body)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", contentType)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		return "", fmt.Errorf("server returned %d", resp.StatusCode)
	}

	var asyncResp AsyncResponse
	if err := json.NewDecoder(resp.Body).Decode(&asyncResp); err != nil {
		return "", err
	}

	return asyncResp.ExecutionID, nil
}

// GetExecution retrieves the current status and result of an execution.
//
// Returns the execution status which may be pending, running, completed,
// failed, or killed. Once completed, the result includes stdout, stderr,
// and exit code.
func (c *Client) GetExecution(ctx context.Context, executionID string) (*ExecutionResult, error) {
	url := fmt.Sprintf("%s/api/v1/executions/%s", c.baseURL, executionID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("execution not found")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned %d", resp.StatusCode)
	}

	var result ExecutionResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

// KillExecution terminates a running execution.
//
// The Docker container running the Python code will be forcefully stopped.
func (c *Client) KillExecution(ctx context.Context, executionID string) error {
	url := fmt.Sprintf("%s/api/v1/executions/%s", c.baseURL, executionID)

	req, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned %d", resp.StatusCode)
	}

	return nil
}

// WaitForCompletion polls the server until the execution completes.
//
// The method polls at the specified interval until the execution reaches
// a terminal state (completed, failed, or killed).
//
// Example:
//
//	execID, _ := c.ExecuteAsync(ctx, tarData, metadata)
//
//	// Poll every 2 seconds
//	result, err := c.WaitForCompletion(ctx, execID, 2*time.Second)
//	if err != nil {
//	    return err
//	}
//
//	fmt.Println(result.Stdout)
func (c *Client) WaitForCompletion(ctx context.Context, executionID string, pollInterval time.Duration) (*ExecutionResult, error) {
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			result, err := c.GetExecution(ctx, executionID)
			if err != nil {
				return nil, err
			}

			// Check if finished
			if result.Status == StatusCompleted ||
			   result.Status == StatusFailed ||
			   result.Status == StatusKilled {
				return result, nil
			}
		}
	}
}

// Eval executes Python code using the simplified JSON API with REPL-style evaluation.
//
// This method uses the /api/v1/eval endpoint which accepts JSON instead of
// multipart/form-data with tar archives, making it ideal for simple code execution.
//
// When EvalLastExpr is true (the default for this endpoint), the last expression
// in the code will be evaluated and its value returned in the Result field.
//
// Example:
//
//	result, err := c.Eval(ctx, &client.SimpleExecRequest{
//	    Code: "x = 5\nx * 2",
//	    EvalLastExpr: true,
//	})
//	if err != nil {
//	    return err
//	}
//	fmt.Println(*result.Result)  // Output: 10
func (c *Client) Eval(ctx context.Context, req *SimpleExecRequest) (*ExecutionResult, error) {
	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshaling request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/v1/eval", bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadRequest {
		var errResp struct {
			Error string `json:"error"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
			return nil, fmt.Errorf("server returned 400: unable to parse error response")
		}
		return nil, fmt.Errorf("bad request: %s", errResp.Error)
	}

	if resp.StatusCode == http.StatusRequestEntityTooLarge {
		return nil, fmt.Errorf("code exceeds maximum size limit")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned %d", resp.StatusCode)
	}

	var result ExecutionResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

// buildMultipartRequest creates a multipart form request
func (c *Client) buildMultipartRequest(tarData []byte, metadata *Metadata) (io.Reader, string, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add tar file
	tarPart, err := writer.CreateFormFile("tar", "code.tar")
	if err != nil {
		return nil, "", err
	}
	if _, err := io.Copy(tarPart, bytes.NewReader(tarData)); err != nil {
		return nil, "", err
	}

	// Add metadata
	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return nil, "", err
	}

	if err := writer.WriteField("metadata", string(metadataJSON)); err != nil {
		return nil, "", err
	}

	if err := writer.Close(); err != nil {
		return nil, "", err
	}

	return body, writer.FormDataContentType(), nil
}
