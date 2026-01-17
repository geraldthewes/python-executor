package client_test

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/geraldthewes/python-executor/pkg/client"
)

// These examples are runnable tests that also serve as documentation.
// Run with: go test -v ./pkg/client/... -run Example
//
// Note: These require a running python-executor server.
// Set PYEXEC_SERVER environment variable or use default http://pyexec.cluster:9999/

func getServerURL() string {
	if url := os.Getenv("PYEXEC_SERVER"); url != "" {
		return url
	}
	return "http://pyexec.cluster:9999/"
}

// ExampleNew demonstrates creating a new client.
func ExampleNew() {
	// Create a client with default settings
	c := client.New("http://pyexec.cluster:9999/")
	_ = c

	// Create a client with custom timeout
	c2 := client.New("http://localhost:8080",
		client.WithTimeout(60*time.Second),
	)
	_ = c2

	fmt.Println("Clients created")
	// Output: Clients created
}

// ExampleTarFromMap demonstrates creating a tar archive from a map.
func ExampleTarFromMap() {
	files := map[string]string{
		"main.py":   `print("Hello from main!")`,
		"helper.py": `def greet(): return "Hi!"`,
	}

	tarData, err := client.TarFromMap(files)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Created tar archive: %d bytes\n", len(tarData))
	// Output: Created tar archive: 2048 bytes
}

// ExampleDetectEntrypoint demonstrates automatic entrypoint detection.
func ExampleDetectEntrypoint() {
	// Archive with main.py
	tarData, _ := client.TarFromMap(map[string]string{
		"main.py":   `print("main")`,
		"helper.py": `print("helper")`,
	})

	entrypoint, err := client.DetectEntrypoint(tarData)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Detected entrypoint: %s\n", entrypoint)
	// Output: Detected entrypoint: main.py
}

// Example_executeSync demonstrates synchronous execution.
// This example requires a running server.
func Example_executeSync() {
	c := client.New(getServerURL())

	tarData, err := client.TarFromMap(map[string]string{
		"main.py": `print("Hello, World!")`,
	})
	if err != nil {
		fmt.Printf("Error creating tar: %v\n", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := c.ExecuteSync(ctx, tarData, &client.Metadata{
		Entrypoint: "main.py",
	})
	if err != nil {
		fmt.Printf("Error executing: %v\n", err)
		return
	}

	fmt.Printf("Exit code: %d\n", result.ExitCode)
	fmt.Printf("Output: %s", strings.TrimSpace(result.Stdout))
	// Output:
	// Exit code: 0
	// Output: Hello, World!
}

// Example_executeWithConfig demonstrates execution with custom configuration.
// This example requires a running server.
func Example_executeWithConfig() {
	c := client.New(getServerURL())

	tarData, _ := client.TarFromMap(map[string]string{
		"main.py": `
import sys
print(f"Args: {sys.argv[1:]}")
print("Config test passed")
`,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := c.ExecuteSync(ctx, tarData, &client.Metadata{
		Entrypoint: "main.py",
		ScriptArgs: []string{"--verbose", "test.txt"},
		Config: &client.ExecutionConfig{
			TimeoutSeconds:  60,
			NetworkDisabled: true,
			MemoryMB:        512,
		},
	})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Exit code: %d\n", result.ExitCode)
	// Output:
	// Exit code: 0
}

// Example_multiFileExecution demonstrates executing multiple files.
// This example requires a running server.
func Example_multiFileExecution() {
	c := client.New(getServerURL())

	tarData, _ := client.TarFromMap(map[string]string{
		"main.py": `
from helper import calculate
result = calculate(5, 3)
print(f"Result: {result}")
`,
		"helper.py": `
def calculate(a, b):
    return a * b
`,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := c.ExecuteSync(ctx, tarData, &client.Metadata{
		Entrypoint: "main.py",
	})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Output: %s", strings.TrimSpace(result.Stdout))
	// Output: Output: Result: 15
}

// Example_asyncExecution demonstrates asynchronous execution.
// This example requires a running server.
func Example_asyncExecution() {
	c := client.New(getServerURL())

	tarData, _ := client.TarFromMap(map[string]string{
		"main.py": `
import time
time.sleep(1)
print("Async complete!")
`,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Submit asynchronously
	execID, err := c.ExecuteAsync(ctx, tarData, &client.Metadata{
		Entrypoint: "main.py",
	})
	if err != nil {
		fmt.Printf("Error submitting: %v\n", err)
		return
	}

	fmt.Printf("Submitted execution: %s\n", execID[:10]+"...")

	// Wait for completion
	result, err := c.WaitForCompletion(ctx, execID, 500*time.Millisecond)
	if err != nil {
		fmt.Printf("Error waiting: %v\n", err)
		return
	}

	fmt.Printf("Status: %s\n", result.Status)
	fmt.Printf("Output: %s", strings.TrimSpace(result.Stdout))
	// Output:
	// Submitted execution: exe_550e84...
	// Status: completed
	// Output: Async complete!
}
