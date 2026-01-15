package executor

import (
	"strings"
	"testing"

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
