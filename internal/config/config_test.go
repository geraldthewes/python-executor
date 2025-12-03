package config

import (
	"os"
	"reflect"
	"testing"
)

func TestGetEnvStringSlice(t *testing.T) {
	tests := []struct {
		name         string
		envValue     string
		defaultValue []string
		expected     []string
	}{
		{
			name:         "default when env not set",
			envValue:     "",
			defaultValue: []string{"8.8.8.8", "8.8.4.4"},
			expected:     []string{"8.8.8.8", "8.8.4.4"},
		},
		{
			name:         "single value",
			envValue:     "1.1.1.1",
			defaultValue: []string{"8.8.8.8"},
			expected:     []string{"1.1.1.1"},
		},
		{
			name:         "multiple values",
			envValue:     "1.1.1.1,8.8.8.8,9.9.9.9",
			defaultValue: []string{"default"},
			expected:     []string{"1.1.1.1", "8.8.8.8", "9.9.9.9"},
		},
		{
			name:         "values with spaces",
			envValue:     "1.1.1.1, 8.8.8.8 , 9.9.9.9",
			defaultValue: []string{"default"},
			expected:     []string{"1.1.1.1", "8.8.8.8", "9.9.9.9"},
		},
		{
			name:         "empty values filtered",
			envValue:     "1.1.1.1,,8.8.8.8",
			defaultValue: []string{"default"},
			expected:     []string{"1.1.1.1", "8.8.8.8"},
		},
		{
			name:         "all empty returns default",
			envValue:     ",,,",
			defaultValue: []string{"default"},
			expected:     []string{"default"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := "TEST_STRING_SLICE"
			if tt.envValue != "" {
				os.Setenv(key, tt.envValue)
				defer os.Unsetenv(key)
			} else {
				os.Unsetenv(key)
			}

			result := getEnvStringSlice(key, tt.defaultValue)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("getEnvStringSlice() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestLoad_DNSServers(t *testing.T) {
	// Clean up any existing env vars
	os.Unsetenv("PYEXEC_DNS_SERVERS")
	defer os.Unsetenv("PYEXEC_DNS_SERVERS")

	// Test default DNS servers
	cfg := Load()
	expected := []string{"8.8.8.8", "8.8.4.4"}
	if !reflect.DeepEqual(cfg.Docker.DNSServers, expected) {
		t.Errorf("Default DNSServers = %v, want %v", cfg.Docker.DNSServers, expected)
	}

	// Test custom DNS servers
	os.Setenv("PYEXEC_DNS_SERVERS", "1.1.1.1,1.0.0.1")
	cfg = Load()
	expected = []string{"1.1.1.1", "1.0.0.1"}
	if !reflect.DeepEqual(cfg.Docker.DNSServers, expected) {
		t.Errorf("Custom DNSServers = %v, want %v", cfg.Docker.DNSServers, expected)
	}
}

func TestLoad_NetworkMode(t *testing.T) {
	// Clean up any existing env vars
	os.Unsetenv("PYEXEC_NETWORK_MODE")
	defer os.Unsetenv("PYEXEC_NETWORK_MODE")

	// Test default network mode
	cfg := Load()
	if cfg.Docker.NetworkMode != "host" {
		t.Errorf("Default NetworkMode = %q, want %q", cfg.Docker.NetworkMode, "host")
	}

	// Test custom network mode
	os.Setenv("PYEXEC_NETWORK_MODE", "bridge")
	cfg = Load()
	if cfg.Docker.NetworkMode != "bridge" {
		t.Errorf("Custom NetworkMode = %q, want %q", cfg.Docker.NetworkMode, "bridge")
	}
}
