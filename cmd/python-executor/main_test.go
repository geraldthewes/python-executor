package main

import (
	"os"
	"reflect"
	"testing"
)

func TestResolveEnvVars_ExplicitValue(t *testing.T) {
	result, err := resolveEnvVars([]string{"FOO=bar", "BAZ=qux"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []string{"FOO=bar", "BAZ=qux"}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("got %v, want %v", result, expected)
	}
}

func TestResolveEnvVars_FromEnvironment(t *testing.T) {
	os.Setenv("TEST_VAR_12345", "test_value")
	defer os.Unsetenv("TEST_VAR_12345")

	result, err := resolveEnvVars([]string{"TEST_VAR_12345"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []string{"TEST_VAR_12345=test_value"}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("got %v, want %v", result, expected)
	}
}

func TestResolveEnvVars_MissingVariable(t *testing.T) {
	_, err := resolveEnvVars([]string{"NONEXISTENT_VAR_99999"})
	if err == nil {
		t.Error("expected error for missing environment variable")
	}
}

func TestResolveEnvVars_Mixed(t *testing.T) {
	os.Setenv("EXISTING_VAR", "from_env")
	defer os.Unsetenv("EXISTING_VAR")

	result, err := resolveEnvVars([]string{"EXISTING_VAR", "EXPLICIT=value"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []string{"EXISTING_VAR=from_env", "EXPLICIT=value"}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("got %v, want %v", result, expected)
	}
}

func TestResolveEnvVars_Empty(t *testing.T) {
	result, err := resolveEnvVars(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 0 {
		t.Errorf("expected empty result, got %v", result)
	}
}

func TestResolveEnvVars_ValueWithEquals(t *testing.T) {
	// Test that VAR=a=b=c is handled correctly (only split on first =)
	result, err := resolveEnvVars([]string{"CONFIG=key=value"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []string{"CONFIG=key=value"}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("got %v, want %v", result, expected)
	}
}
