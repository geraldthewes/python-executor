package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds the application configuration
type Config struct {
	Server  ServerConfig
	Docker  DockerConfig
	Defaults DefaultsConfig
	Consul  ConsulConfig
	Cleanup CleanupConfig
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Host     string
	Port     string
	LogLevel string
}

// DockerConfig holds Docker client configuration
type DockerConfig struct {
	Socket string
}

// DefaultsConfig holds default execution parameters
type DefaultsConfig struct {
	Timeout      int
	MemoryMB     int
	DiskMB       int
	CPUShares    int
	DockerImage  string
}

// ConsulConfig holds Consul configuration
type ConsulConfig struct {
	Address   string
	Token     string
	KeyPrefix string
	Enabled   bool
}

// CleanupConfig holds cleanup configuration
type CleanupConfig struct {
	TTL time.Duration
}

// Load reads configuration from environment variables
func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Host:     getEnv("PYEXEC_HOST", "0.0.0.0"),
			Port:     getEnv("PYEXEC_PORT", "8080"),
			LogLevel: getEnv("PYEXEC_LOG_LEVEL", "info"),
		},
		Docker: DockerConfig{
			Socket: getEnv("PYEXEC_DOCKER_SOCKET", "/var/run/docker.sock"),
		},
		Defaults: DefaultsConfig{
			Timeout:     getEnvInt("PYEXEC_DEFAULT_TIMEOUT", 300),
			MemoryMB:    getEnvInt("PYEXEC_DEFAULT_MEMORY_MB", 1024),
			DiskMB:      getEnvInt("PYEXEC_DEFAULT_DISK_MB", 2048),
			CPUShares:   getEnvInt("PYEXEC_DEFAULT_CPU_SHARES", 1024),
			DockerImage: getEnv("PYEXEC_DEFAULT_IMAGE", "python:3.12-slim"),
		},
		Consul: ConsulConfig{
			Address:   getEnv("PYEXEC_CONSUL_ADDR", "localhost:8500"),
			Token:     getEnv("PYEXEC_CONSUL_TOKEN", ""),
			KeyPrefix: getEnv("PYEXEC_CONSUL_PREFIX", "python-executor"),
			Enabled:   getEnv("PYEXEC_CONSUL_ADDR", "") != "",
		},
		Cleanup: CleanupConfig{
			TTL: time.Duration(getEnvInt("PYEXEC_CLEANUP_TTL", 300)) * time.Second,
		},
	}
}

// getEnv retrieves an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvInt retrieves an environment variable as int or returns a default value
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}
