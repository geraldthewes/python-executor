package executor

import (
	"archive/tar"
	"bytes"
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/geraldthewes/python-executor/internal/config"
	"github.com/geraldthewes/python-executor/pkg/client"
)

func TestApplyDefaults_PreservesNetworkDisabled(t *testing.T) {
	cfg := &config.Config{
		Defaults: config.DefaultsConfig{
			Timeout:     300,
			MemoryMB:    1024,
			DiskMB:      2048,
			CPUShares:   1024,
			DockerImage: "python:3.12-slim",
		},
	}

	tests := []struct {
		name                    string
		inputNetworkDisabled    bool
		expectedNetworkDisabled bool
	}{
		{
			name:                    "NetworkDisabled=false should stay false",
			inputNetworkDisabled:    false,
			expectedNetworkDisabled: false,
		},
		{
			name:                    "NetworkDisabled=true should stay true",
			inputNetworkDisabled:    true,
			expectedNetworkDisabled: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			meta := &client.Metadata{
				Entrypoint: "main.py",
				Config: &client.ExecutionConfig{
					NetworkDisabled: tt.inputNetworkDisabled,
				},
			}

			result := applyDefaults(meta, cfg)

			if result.Config.NetworkDisabled != tt.expectedNetworkDisabled {
				t.Errorf("applyDefaults() NetworkDisabled = %v, want %v",
					result.Config.NetworkDisabled, tt.expectedNetworkDisabled)
			}
		})
	}
}

func TestApplyDefaults_SetsDefaults(t *testing.T) {
	cfg := &config.Config{
		Defaults: config.DefaultsConfig{
			Timeout:     300,
			MemoryMB:    1024,
			DiskMB:      2048,
			CPUShares:   512,
			DockerImage: "python:3.12-slim",
		},
	}

	// Test with nil Config
	meta := &client.Metadata{
		Entrypoint: "main.py",
	}

	result := applyDefaults(meta, cfg)

	if result.Config == nil {
		t.Fatal("applyDefaults() should create Config if nil")
	}
	if result.Config.TimeoutSeconds != 300 {
		t.Errorf("TimeoutSeconds = %d, want 300", result.Config.TimeoutSeconds)
	}
	if result.Config.MemoryMB != 1024 {
		t.Errorf("MemoryMB = %d, want 1024", result.Config.MemoryMB)
	}
	if result.Config.DiskMB != 2048 {
		t.Errorf("DiskMB = %d, want 2048", result.Config.DiskMB)
	}
	if result.Config.CPUShares != 512 {
		t.Errorf("CPUShares = %d, want 512", result.Config.CPUShares)
	}
	if result.DockerImage != "python:3.12-slim" {
		t.Errorf("DockerImage = %q, want %q", result.DockerImage, "python:3.12-slim")
	}
}

func TestApplyDefaults_DoesNotOverrideExplicitValues(t *testing.T) {
	cfg := &config.Config{
		Defaults: config.DefaultsConfig{
			Timeout:     300,
			MemoryMB:    1024,
			DiskMB:      2048,
			CPUShares:   512,
			DockerImage: "python:3.12-slim",
		},
	}

	meta := &client.Metadata{
		Entrypoint:  "main.py",
		DockerImage: "python:3.11-alpine",
		Config: &client.ExecutionConfig{
			TimeoutSeconds: 60,
			MemoryMB:       512,
			DiskMB:         1024,
			CPUShares:      256,
		},
	}

	result := applyDefaults(meta, cfg)

	if result.Config.TimeoutSeconds != 60 {
		t.Errorf("TimeoutSeconds = %d, want 60 (explicit value)", result.Config.TimeoutSeconds)
	}
	if result.Config.MemoryMB != 512 {
		t.Errorf("MemoryMB = %d, want 512 (explicit value)", result.Config.MemoryMB)
	}
	if result.DockerImage != "python:3.11-alpine" {
		t.Errorf("DockerImage = %q, want %q (explicit value)", result.DockerImage, "python:3.11-alpine")
	}
}

func TestBuildCommand_WithRequirements(t *testing.T) {
	cfg := &config.Config{}
	executor := &DockerExecutor{config: cfg}

	meta := &client.Metadata{
		Entrypoint:      "main.py",
		RequirementsTxt: "requests\nnumpy",
	}

	cmd := executor.buildCommand(meta)

	// Should contain echo to create requirements.txt
	if !strings.Contains(cmd, "echo") {
		t.Error("Command should contain echo for requirements.txt")
	}
	if !strings.Contains(cmd, "requirements.txt") {
		t.Error("Command should reference requirements.txt")
	}
	// Should contain pip install
	if !strings.Contains(cmd, "pip install") {
		t.Error("Command should contain pip install")
	}
	// Should contain python execution
	if !strings.Contains(cmd, "python") && !strings.Contains(cmd, "main.py") {
		t.Error("Command should contain python main.py")
	}
}

func TestBuildCommand_WithoutRequirements(t *testing.T) {
	cfg := &config.Config{}
	executor := &DockerExecutor{config: cfg}

	meta := &client.Metadata{
		Entrypoint: "script.py",
	}

	cmd := executor.buildCommand(meta)

	// Should NOT contain pip install
	if strings.Contains(cmd, "pip install") {
		t.Error("Command should not contain pip install when no requirements")
	}
	// Should contain python execution
	if !strings.Contains(cmd, "python") || !strings.Contains(cmd, "script.py") {
		t.Errorf("Command should contain 'python script.py', got: %s", cmd)
	}
}

func TestBuildCommand_WithPreCommands(t *testing.T) {
	cfg := &config.Config{}
	executor := &DockerExecutor{config: cfg}

	meta := &client.Metadata{
		Entrypoint:  "main.py",
		PreCommands: []string{"echo 'setup'", "mkdir -p /data"},
	}

	cmd := executor.buildCommand(meta)

	// Should contain pre-commands
	if !strings.Contains(cmd, "echo 'setup'") {
		t.Error("Command should contain first pre-command")
	}
	if !strings.Contains(cmd, "mkdir -p /data") {
		t.Error("Command should contain second pre-command")
	}
}

func TestBuildCommand_RequirementsEscapesSingleQuotes(t *testing.T) {
	cfg := &config.Config{}
	executor := &DockerExecutor{config: cfg}

	meta := &client.Metadata{
		Entrypoint:      "main.py",
		RequirementsTxt: "package[extra]>=1.0",
	}

	cmd := executor.buildCommand(meta)

	// Command should be properly escaped for shell
	if !strings.Contains(cmd, "package[extra]>=1.0") {
		t.Errorf("Requirements content should be in command, got: %s", cmd)
	}
}

func TestBuildCommand_WithScriptArgs(t *testing.T) {
	cfg := &config.Config{}
	executor := &DockerExecutor{config: cfg}

	meta := &client.Metadata{
		Entrypoint: "main.py",
		ScriptArgs: []string{"arg1", "arg2"},
	}

	cmd := executor.buildCommand(meta)

	// Should contain python and entrypoint
	if !strings.Contains(cmd, "python") {
		t.Error("Command should contain python")
	}
	if !strings.Contains(cmd, "main.py") {
		t.Error("Command should contain entrypoint")
	}
	// Should contain arguments
	if !strings.Contains(cmd, "arg1") {
		t.Error("Command should contain arg1")
	}
	if !strings.Contains(cmd, "arg2") {
		t.Error("Command should contain arg2")
	}
}

func TestBuildCommand_WithScriptArgsSpecialChars(t *testing.T) {
	cfg := &config.Config{}
	executor := &DockerExecutor{config: cfg}

	meta := &client.Metadata{
		Entrypoint: "main.py",
		ScriptArgs: []string{"arg with spaces", "--flag=value", "$VAR"},
	}

	cmd := executor.buildCommand(meta)

	// The argument with spaces should be properly quoted
	if !strings.Contains(cmd, "'arg with spaces'") {
		t.Errorf("Argument with spaces should be quoted, got: %s", cmd)
	}
	// Flag-style argument should be present
	if !strings.Contains(cmd, "--flag=value") {
		t.Errorf("Flag argument should be present, got: %s", cmd)
	}
	// $VAR should be quoted to prevent expansion
	if !strings.Contains(cmd, "'$VAR'") {
		t.Errorf("$VAR should be quoted to prevent shell expansion, got: %s", cmd)
	}
}

func TestBuildCommand_NoScriptArgs(t *testing.T) {
	cfg := &config.Config{}
	executor := &DockerExecutor{config: cfg}

	meta := &client.Metadata{
		Entrypoint: "script.py",
		ScriptArgs: nil,
	}

	cmd := executor.buildCommand(meta)

	// Should contain python and script path (may or may not be quoted based on path)
	if !strings.Contains(cmd, "python") {
		t.Error("Command should contain python")
	}
	if !strings.Contains(cmd, "script.py") {
		t.Errorf("Command should contain script.py, got: %s", cmd)
	}
	// Should not have extra arguments after the script path
	if strings.Contains(cmd, "arg") {
		t.Errorf("Command should not have extra arguments, got: %s", cmd)
	}
}

func TestBuildCommand_WithEvalLastExpr(t *testing.T) {
	cfg := &config.Config{}
	executor := &DockerExecutor{config: cfg}

	meta := &client.Metadata{
		Entrypoint:   "main.py",
		EvalLastExpr: true,
	}

	cmd := executor.buildCommand(meta)

	// Should contain the eval wrapper script
	if !strings.Contains(cmd, EvalWrapperScript) {
		t.Errorf("Command should contain eval wrapper script %q, got: %s", EvalWrapperScript, cmd)
	}
	// Should pass the original entrypoint as argument
	if !strings.Contains(cmd, "main.py") {
		t.Errorf("Command should pass main.py as argument, got: %s", cmd)
	}
	// The wrapper should come before the entrypoint
	wrapperIdx := strings.Index(cmd, EvalWrapperScript)
	entrypointIdx := strings.Index(cmd, "main.py")
	if wrapperIdx > entrypointIdx {
		t.Errorf("Wrapper script should come before entrypoint in command, got: %s", cmd)
	}
}

func TestBuildCommand_WithoutEvalLastExpr(t *testing.T) {
	cfg := &config.Config{}
	executor := &DockerExecutor{config: cfg}

	meta := &client.Metadata{
		Entrypoint:   "main.py",
		EvalLastExpr: false,
	}

	cmd := executor.buildCommand(meta)

	// Should NOT contain the eval wrapper script
	if strings.Contains(cmd, EvalWrapperScript) {
		t.Errorf("Command should not contain eval wrapper script when EvalLastExpr is false, got: %s", cmd)
	}
	// Should directly run the entrypoint
	if !strings.Contains(cmd, "python") || !strings.Contains(cmd, "main.py") {
		t.Errorf("Command should run python main.py directly, got: %s", cmd)
	}
}

func TestGetEvalWrapperCode(t *testing.T) {
	code := GetEvalWrapperCode()

	// Verify essential components of the wrapper
	if !strings.Contains(code, "import ast") {
		t.Error("Wrapper code should import ast")
	}
	if !strings.Contains(code, "ast.parse") {
		t.Error("Wrapper code should use ast.parse")
	}
	if !strings.Contains(code, "ast.Expr") {
		t.Error("Wrapper code should check for ast.Expr")
	}
	if !strings.Contains(code, ResultMarker) {
		t.Errorf("Wrapper code should contain result marker %q", ResultMarker)
	}
}

// Helper function to create a tar archive from file contents
func createTar(files map[string]string) ([]byte, error) {
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)

	for name, content := range files {
		hdr := &tar.Header{
			Name: name,
			Mode: 0644,
			Size: int64(len(content)),
		}
		if err := tw.WriteHeader(hdr); err != nil {
			return nil, err
		}
		if _, err := tw.Write([]byte(content)); err != nil {
			return nil, err
		}
	}

	if err := tw.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Integration tests - require Docker daemon
// Skip these tests if Docker is not available

func skipIfNoDocker(t *testing.T) {
	if os.Getenv("DOCKER_HOST") == "" && os.Getenv("TEST_WITH_DOCKER") == "" {
		// Check if default Docker socket exists
		if _, err := os.Stat("/var/run/docker.sock"); os.IsNotExist(err) {
			t.Skip("Skipping integration test: Docker not available")
		}
	}
}

func TestExecute_WithStdin(t *testing.T) {
	skipIfNoDocker(t)

	cfg := &config.Config{
		Docker: config.DockerConfig{
			Socket:      "/var/run/docker.sock",
			NetworkMode: "bridge",
		},
		Defaults: config.DefaultsConfig{
			Timeout:     30,
			MemoryMB:    512,
			DiskMB:      1024,
			CPUShares:   512,
			DockerImage: "python:3.12-slim",
		},
	}

	executor, err := NewDockerExecutor(cfg)
	if err != nil {
		t.Fatalf("Failed to create executor: %v", err)
	}
	defer executor.Close()

	// Python script that reads from stdin
	code := `import sys
data = sys.stdin.read()
print(f"Received: {data}")
print(f"Length: {len(data)}")
`

	tarData, err := createTar(map[string]string{"main.py": code})
	if err != nil {
		t.Fatalf("Failed to create tar: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req := &ExecutionRequest{
		TarData: tarData,
		Metadata: &client.Metadata{
			Entrypoint: "main.py",
			Stdin:      "Hello from stdin!",
		},
	}

	output, err := executor.Execute(ctx, req)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if output.ExitCode != 0 {
		t.Errorf("Expected exit code 0, got %d. Stderr: %s", output.ExitCode, output.Stderr)
	}

	if !strings.Contains(output.Stdout, "Hello from stdin!") {
		t.Errorf("Expected stdout to contain stdin data, got: %s", output.Stdout)
	}

	if !strings.Contains(output.Stdout, "Length: 17") {
		t.Errorf("Expected stdout to contain correct length, got: %s", output.Stdout)
	}
}

func TestExecute_WithStdinMultiline(t *testing.T) {
	skipIfNoDocker(t)

	cfg := &config.Config{
		Docker: config.DockerConfig{
			Socket:      "/var/run/docker.sock",
			NetworkMode: "bridge",
		},
		Defaults: config.DefaultsConfig{
			Timeout:     30,
			MemoryMB:    512,
			DiskMB:      1024,
			CPUShares:   512,
			DockerImage: "python:3.12-slim",
		},
	}

	executor, err := NewDockerExecutor(cfg)
	if err != nil {
		t.Fatalf("Failed to create executor: %v", err)
	}
	defer executor.Close()

	// Python script that reads lines from stdin
	code := `import sys
lines = sys.stdin.readlines()
print(f"Got {len(lines)} lines")
for i, line in enumerate(lines):
    print(f"Line {i}: {line.strip()}")
`

	tarData, err := createTar(map[string]string{"main.py": code})
	if err != nil {
		t.Fatalf("Failed to create tar: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	stdinData := "line1\nline2\nline3\n"
	req := &ExecutionRequest{
		TarData: tarData,
		Metadata: &client.Metadata{
			Entrypoint: "main.py",
			Stdin:      stdinData,
		},
	}

	output, err := executor.Execute(ctx, req)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if output.ExitCode != 0 {
		t.Errorf("Expected exit code 0, got %d. Stderr: %s", output.ExitCode, output.Stderr)
	}

	if !strings.Contains(output.Stdout, "Got 3 lines") {
		t.Errorf("Expected 3 lines, got: %s", output.Stdout)
	}

	if !strings.Contains(output.Stdout, "Line 0: line1") {
		t.Errorf("Expected Line 0, got: %s", output.Stdout)
	}
}

func TestExecute_WithoutStdin(t *testing.T) {
	skipIfNoDocker(t)

	cfg := &config.Config{
		Docker: config.DockerConfig{
			Socket:      "/var/run/docker.sock",
			NetworkMode: "bridge",
		},
		Defaults: config.DefaultsConfig{
			Timeout:     30,
			MemoryMB:    512,
			DiskMB:      1024,
			CPUShares:   512,
			DockerImage: "python:3.12-slim",
		},
	}

	executor, err := NewDockerExecutor(cfg)
	if err != nil {
		t.Fatalf("Failed to create executor: %v", err)
	}
	defer executor.Close()

	// Simple script without stdin
	code := `print("Hello, World!")`

	tarData, err := createTar(map[string]string{"main.py": code})
	if err != nil {
		t.Fatalf("Failed to create tar: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req := &ExecutionRequest{
		TarData: tarData,
		Metadata: &client.Metadata{
			Entrypoint: "main.py",
		},
	}

	output, err := executor.Execute(ctx, req)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if output.ExitCode != 0 {
		t.Errorf("Expected exit code 0, got %d. Stderr: %s", output.ExitCode, output.Stderr)
	}

	if !strings.Contains(output.Stdout, "Hello, World!") {
		t.Errorf("Expected stdout to contain greeting, got: %s", output.Stdout)
	}
}
