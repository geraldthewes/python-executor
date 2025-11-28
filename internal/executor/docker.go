package executor

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/geraldthewes/python-executor/internal/config"
	internalttar "github.com/geraldthewes/python-executor/internal/tar"
	clientpkg "github.com/geraldthewes/python-executor/pkg/client"
)

// DockerExecutor implements the Executor interface using Docker
type DockerExecutor struct {
	client  *client.Client
	config  *config.Config
}

// NewDockerExecutor creates a new Docker-based executor
func NewDockerExecutor(cfg *config.Config) (*DockerExecutor, error) {
	cli, err := client.NewClientWithOpts(
		client.FromEnv,
		client.WithHost("unix://"+cfg.Docker.Socket),
	)
	if err != nil {
		return nil, fmt.Errorf("creating docker client: %w", err)
	}

	return &DockerExecutor{
		client: cli,
		config: cfg,
	}, nil
}

// Execute runs code in a Docker container
func (e *DockerExecutor) Execute(ctx context.Context, req *ExecutionRequest) (*ExecutionOutput, error) {
	startTime := time.Now()

	// Apply defaults
	meta := applyDefaults(req.Metadata, e.config)

	// Set timeout
	timeout := time.Duration(meta.Config.TimeoutSeconds) * time.Second
	execCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Create temporary work directory
	workDir, err := os.MkdirTemp("", fmt.Sprintf("pyexec-%s-*", req.ID))
	if err != nil {
		return nil, fmt.Errorf("creating work directory: %w", err)
	}
	defer os.RemoveAll(workDir)

	// Extract tar archive
	if err := internalttar.ExtractToDir(req.TarData, workDir); err != nil {
		return nil, fmt.Errorf("extracting tar: %w", err)
	}

	// Pull Docker image if needed
	if err := e.ensureImage(execCtx, meta.DockerImage); err != nil {
		return nil, fmt.Errorf("ensuring image: %w", err)
	}

	// Create container
	containerID, err := e.createContainer(execCtx, meta, workDir)
	if err != nil {
		return nil, fmt.Errorf("creating container: %w", err)
	}
	defer e.client.ContainerRemove(context.Background(), containerID, container.RemoveOptions{Force: true})

	// Start container
	if err := e.client.ContainerStart(execCtx, containerID, container.StartOptions{}); err != nil {
		return nil, fmt.Errorf("starting container: %w", err)
	}

	// Wait for container to finish
	statusCh, errCh := e.client.ContainerWait(execCtx, containerID, container.WaitConditionNotRunning)

	var exitCode int64
	select {
	case err := <-errCh:
		if err != nil {
			return nil, fmt.Errorf("waiting for container: %w", err)
		}
	case status := <-statusCh:
		exitCode = status.StatusCode
	case <-execCtx.Done():
		// Timeout - kill container
		e.client.ContainerKill(context.Background(), containerID, "SIGKILL")
		return nil, fmt.Errorf("execution timeout after %v", timeout)
	}

	// Get logs
	stdout, stderr, err := e.getLogs(context.Background(), containerID)
	if err != nil {
		return nil, fmt.Errorf("getting logs: %w", err)
	}

	duration := time.Since(startTime)

	return &ExecutionOutput{
		Stdout:     stdout,
		Stderr:     stderr,
		ExitCode:   int(exitCode),
		DurationMs: duration.Milliseconds(),
	}, nil
}

// Kill terminates a running container
func (e *DockerExecutor) Kill(ctx context.Context, containerID string) error {
	return e.client.ContainerKill(ctx, containerID, "SIGKILL")
}

// Close closes the Docker client
func (e *DockerExecutor) Close() error {
	return e.client.Close()
}

// ensureImage pulls the Docker image if it doesn't exist
func (e *DockerExecutor) ensureImage(ctx context.Context, imageName string) error {
	_, _, err := e.client.ImageInspectWithRaw(ctx, imageName)
	if err == nil {
		return nil // Image exists
	}

	// Pull image
	out, err := e.client.ImagePull(ctx, imageName, image.PullOptions{})
	if err != nil {
		return err
	}
	defer out.Close()

	// Wait for pull to complete
	_, err = io.Copy(io.Discard, out)
	return err
}

// createContainer creates a Docker container with security constraints
func (e *DockerExecutor) createContainer(ctx context.Context, meta *clientpkg.Metadata, workDir string) (string, error) {
	// Build command
	cmd := e.buildCommand(meta, workDir)

	// Network mode
	networkMode := "none"
	if !meta.Config.NetworkDisabled {
		networkMode = "bridge"
	}

	// Resource limits
	resources := container.Resources{
		Memory:    int64(meta.Config.MemoryMB) * 1024 * 1024,
		CPUShares: int64(meta.Config.CPUShares),
	}

	// Create container config
	containerConfig := &container.Config{
		Image:        meta.DockerImage,
		Cmd:          []string{"sh", "-c", cmd},
		User:         "1000:1000",
		WorkingDir:   "/work",
		AttachStdout: true,
		AttachStderr: true,
	}

	// Add stdin if provided
	if meta.Stdin != "" {
		containerConfig.OpenStdin = true
		containerConfig.StdinOnce = true
	}

	// Host config with security
	hostConfig := &container.HostConfig{
		NetworkMode: container.NetworkMode(networkMode),
		Resources:   resources,
		ReadonlyRootfs: true,
		Tmpfs: map[string]string{
			"/work": fmt.Sprintf("size=%dm", meta.Config.DiskMB),
			"/tmp":  "size=100m",
		},
		Binds: []string{
			fmt.Sprintf("%s:/work-init:ro", workDir),
		},
	}

	// Create container
	resp, err := e.client.ContainerCreate(ctx, containerConfig, hostConfig, nil, nil, "")
	if err != nil {
		return "", err
	}

	return resp.ID, nil
}

// buildCommand creates the shell command to run inside the container
func (e *DockerExecutor) buildCommand(meta *clientpkg.Metadata, workDir string) string {
	var parts []string

	// Copy files from read-only mount to tmpfs
	parts = append(parts, "cp -r /work-init/* /work/ 2>/dev/null || true")

	// Run pre-commands
	for _, cmd := range meta.PreCommands {
		parts = append(parts, cmd)
	}

	// Install requirements
	if meta.RequirementsTxt != "" {
		reqFile := filepath.Join("/work", "requirements.txt")
		parts = append(parts, fmt.Sprintf("echo '%s' > %s", strings.ReplaceAll(meta.RequirementsTxt, "'", "'\\''"), reqFile))
		parts = append(parts, fmt.Sprintf("pip install --no-cache-dir -r %s", reqFile))
	}

	// Run Python script
	scriptPath := filepath.Join("/work", meta.Entrypoint)
	parts = append(parts, fmt.Sprintf("python %s", scriptPath))

	return strings.Join(parts, " && ")
}

// getLogs retrieves stdout and stderr from a container
func (e *DockerExecutor) getLogs(ctx context.Context, containerID string) (string, string, error) {
	options := container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
	}

	logs, err := e.client.ContainerLogs(ctx, containerID, options)
	if err != nil {
		return "", "", err
	}
	defer logs.Close()

	// Docker multiplexes stdout/stderr - we need to demultiplex
	stdout, stderr, err := demuxLogs(logs)
	if err != nil {
		return "", "", err
	}

	return stdout, stderr, nil
}

// demuxLogs separates stdout and stderr from Docker's multiplexed stream
func demuxLogs(logs io.Reader) (string, string, error) {
	var stdoutBuf, stderrBuf strings.Builder

	// Docker uses an 8-byte header for each frame
	// [stream_type, 0, 0, 0, size1, size2, size3, size4]
	header := make([]byte, 8)

	for {
		_, err := io.ReadFull(logs, header)
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", "", err
		}

		// Parse size (big-endian uint32)
		size := uint32(header[4])<<24 | uint32(header[5])<<16 | uint32(header[6])<<8 | uint32(header[7])

		// Read payload
		payload := make([]byte, size)
		if _, err := io.ReadFull(logs, payload); err != nil {
			return "", "", err
		}

		// Stream type: 1=stdout, 2=stderr
		switch header[0] {
		case 1:
			stdoutBuf.Write(payload)
		case 2:
			stderrBuf.Write(payload)
		}
	}

	return stdoutBuf.String(), stderrBuf.String(), nil
}

// applyDefaults fills in missing configuration values
func applyDefaults(meta *clientpkg.Metadata, cfg *config.Config) *clientpkg.Metadata {
	if meta.Config == nil {
		meta.Config = &clientpkg.ExecutionConfig{}
	}

	if meta.DockerImage == "" {
		meta.DockerImage = cfg.Defaults.DockerImage
	}

	if meta.Config.TimeoutSeconds == 0 {
		meta.Config.TimeoutSeconds = cfg.Defaults.Timeout
	}
	if meta.Config.MemoryMB == 0 {
		meta.Config.MemoryMB = cfg.Defaults.MemoryMB
	}
	if meta.Config.DiskMB == 0 {
		meta.Config.DiskMB = cfg.Defaults.DiskMB
	}
	if meta.Config.CPUShares == 0 {
		meta.Config.CPUShares = cfg.Defaults.CPUShares
	}

	// Default to network disabled
	if meta.Config.NetworkDisabled == false && meta.Config.MemoryMB != 0 {
		// If config was explicitly provided but network not set, default to true
		meta.Config.NetworkDisabled = true
	}

	return meta
}
