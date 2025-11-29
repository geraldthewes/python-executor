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

// Client is the Go client for python-executor
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// New creates a new client
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

// ExecuteSync executes code synchronously
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

// ExecuteAsync executes code asynchronously
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

// GetExecution retrieves execution status and result
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

// KillExecution terminates a running execution
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

// WaitForCompletion polls until execution completes
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
