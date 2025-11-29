package client

import "testing"

func TestNewClient_TrailingSlash(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"with trailing slash", "http://example.com:8080/", "http://example.com:8080"},
		{"without trailing slash", "http://example.com:8080", "http://example.com:8080"},
		{"with path and trailing slash", "http://example.com:8080/api/", "http://example.com:8080/api"},
		{"with path no trailing slash", "http://example.com:8080/api", "http://example.com:8080/api"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := New(tt.input)
			if c.baseURL != tt.expected {
				t.Errorf("New(%q).baseURL = %q, want %q", tt.input, c.baseURL, tt.expected)
			}
		})
	}
}
