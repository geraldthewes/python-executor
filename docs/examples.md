# Examples

## CLI Examples

### Basic Usage

```bash
# Execute a simple script from stdin
echo 'print("Hello, World!")' | python-executor run

# Execute a single file
python-executor run hello.py

# Execute from multiple files
python-executor run \
  --file main.py \
  --file utils.py \
  --file config.json

# Execute an entire directory
python-executor run ./my-project --entrypoint main.py
```

### With Dependencies

```bash
# Using --file with implicit requirements.txt
python-executor run \
  --file main.py \
  --file requirements.txt

# Or specify requirements inline
cat > /tmp/script.py <<'EOF'
import numpy as np
print(np.array([1, 2, 3]).sum())
EOF

python-executor run /tmp/script.py --requirements "numpy==1.24.0"
```

### Resource Limits

```bash
# Set custom limits
python-executor run script.py \
  --timeout 600 \
  --memory 2048 \
  --cpu 2048

# Enable network access
python-executor run script.py --network
```

### Async Execution

```bash
# Submit async
id=$(python-executor submit long-running.py --async)

# Follow until complete
python-executor follow $id

# Or just check status
python-executor get $id
```

## Go Client Examples

### Basic Execution

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/geraldthewes/python-executor/pkg/client"
)

func main() {
    c := client.New("http://localhost:8080")

    // Create tar from map
    tarData, err := client.TarFromMap(map[string]string{
        "main.py": `print("Hello from Go!")`,
    })
    if err != nil {
        log.Fatal(err)
    }

    // Execute
    result, err := c.ExecuteSync(context.Background(), tarData, &client.Metadata{
        Entrypoint: "main.py",
    })
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println(result.Stdout)
}
```

### With Dependencies

```go
tarData, err := client.TarFromMap(map[string]string{
    "main.py": `
import pandas as pd
df = pd.DataFrame({"a": [1, 2, 3]})
print(df.sum())
`,
})

result, err := c.ExecuteSync(ctx, tarData, &client.Metadata{
    Entrypoint:      "main.py",
    RequirementsTxt: "pandas==2.0.0",
})
```

### Async Execution

```go
// Submit async
execID, err := c.ExecuteAsync(ctx, tarData, metadata)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Execution ID: %s\n", execID)

// Wait for completion
result, err := c.WaitForCompletion(ctx, execID, 2*time.Second)
if err != nil {
    log.Fatal(err)
}

fmt.Println(result.Stdout)
```

### From Directory

```go
// Create tar from directory
tarData, err := client.TarFromDirectory("./my-project")
if err != nil {
    log.Fatal(err)
}

// Auto-detect entrypoint
entrypoint, err := client.DetectEntrypoint(tarData)
if err != nil {
    log.Fatal(err)
}

result, err := c.ExecuteSync(ctx, tarData, &client.Metadata{
    Entrypoint: entrypoint,
    Config: &client.ExecutionConfig{
        TimeoutSeconds:  600,
        MemoryMB:        2048,
        NetworkDisabled: false,
    },
})
```

## Python Client Examples

### Basic Usage

```python
from python_executor_client import PythonExecutorClient

client = PythonExecutorClient("http://localhost:8080")

# Execute from dict
result = client.execute_sync(
    files={"main.py": "print('Hello from Python!')"},
    entrypoint="main.py"
)

print(result.stdout)
print(f"Exit code: {result.exit_code}")
```

### With Dependencies

```python
result = client.execute_sync(
    files={
        "main.py": """
import requests
response = requests.get('https://api.github.com')
print(response.status_code)
""",
    },
    entrypoint="main.py",
    requirements_txt="requests==2.31.0",
    network_disabled=False,  # Enable network
)
```

### From Directory

```python
from pathlib import Path

result = client.execute_sync(
    files=Path("./my-project"),
    entrypoint="main.py",
    timeout_seconds=600,
    memory_mb=2048,
)
```

### Async Execution

```python
# Submit async
exec_id = client.execute_async(
    files={"train.py": "# Long training script..."},
    entrypoint="train.py",
    timeout_seconds=7200,
    memory_mb=16384,
)

print(f"Execution ID: {exec_id}")

# Wait for completion
result = client.wait_for_completion(exec_id, poll_interval=5.0)
print(result.stdout)
```

### Pre-commands

```python
result = client.execute_sync(
    files={"main.py": "import cv2; print(cv2.__version__)"},
    entrypoint="main.py",
    pre_commands=[
        "apt-get update -y",
        "apt-get install -y libgl1",
    ],
    requirements_txt="opencv-python==4.8.0",
)
```

## REST API Examples

### Synchronous Execution

```bash
# Create a tar file
tar cf code.tar main.py utils.py

# Submit for execution
curl -X POST http://localhost:8080/api/v1/exec/sync \
  -F "tar=@code.tar" \
  -F 'metadata={
    "entrypoint": "main.py",
    "docker_image": "python:3.12-slim",
    "config": {
      "timeout_seconds": 300,
      "memory_mb": 1024,
      "network_disabled": true
    }
  }'
```

Response:
```json
{
  "execution_id": "exe_123",
  "status": "completed",
  "stdout": "Hello, World!\n",
  "stderr": "",
  "exit_code": 0,
  "started_at": "2025-11-28T10:00:00Z",
  "finished_at": "2025-11-28T10:00:01Z",
  "duration_ms": 1234
}
```

### Asynchronous Execution

```bash
# Submit async
response=$(curl -X POST http://localhost:8080/api/v1/exec/async \
  -F "tar=@code.tar" \
  -F 'metadata={"entrypoint": "main.py"}')

exec_id=$(echo $response | jq -r .execution_id)

# Poll for status
curl http://localhost:8080/api/v1/executions/$exec_id
```

### Kill Execution

```bash
curl -X DELETE http://localhost:8080/api/v1/executions/$exec_id
```

## REPL-style Expression Evaluation

The `/api/v1/eval` endpoint supports REPL-style expression evaluation. When `eval_last_expr` is `true`, the value of the last expression is captured and returned in the `result` field.

### curl Examples

```bash
# Simple calculation
curl -s -X POST http://localhost:8080/api/v1/eval \
  -H "Content-Type: application/json" \
  -d '{"code": "2 + 2", "eval_last_expr": true}' | jq .result
# Output: "4"

# Multi-line with expression at end
curl -s -X POST http://localhost:8080/api/v1/eval \
  -H "Content-Type: application/json" \
  -d '{"code": "x = 10\ny = 20\nx * y", "eval_last_expr": true}' | jq .result
# Output: "200"

# Print statement (no result, returns null)
curl -s -X POST http://localhost:8080/api/v1/eval \
  -H "Content-Type: application/json" \
  -d '{"code": "print(\"hello\")", "eval_last_expr": true}' | jq '{stdout, result}'
# Output: {"stdout": "hello\n", "result": null}

# List comprehension
curl -s -X POST http://localhost:8080/api/v1/eval \
  -H "Content-Type: application/json" \
  -d '{"code": "[x**2 for x in range(5)]", "eval_last_expr": true}' | jq .result
# Output: "[0, 1, 4, 9, 16]"
```

### Python Client

```python
from python_executor_client import PythonExecutorClient

client = PythonExecutorClient("http://localhost:8080")

# Use eval_last_expr for calculator-style interactions
result = client.execute_eval(
    code="import math\nmath.pi * 2",
    eval_last_expr=True
)

print(result.result)  # "6.283185307179586"
print(result.stdout)  # "" (empty, no print statements)
```

### Go Client

```go
// Simple eval with expression result
result, err := c.ExecuteEval(ctx, &client.SimpleExecRequest{
    Code:         "sum([1, 2, 3, 4, 5])",
    EvalLastExpr: true,
})
if err != nil {
    log.Fatal(err)
}

if result.Result != nil {
    fmt.Println("Result:", *result.Result) // "15"
}
```

---

## MCP Server Integration

If you're building an MCP (Model Context Protocol) server for AI agents:

```python
from python_executor_client import PythonExecutorClient

class PythonExecutorTool:
    def __init__(self):
        self.client = PythonExecutorClient("http://python-executor:8080")

    def run_python(self, code: str) -> dict:
        """Tool for AI agents to execute Python code."""
        result = self.client.execute_sync(
            files={"main.py": code},
            entrypoint="main.py",
            timeout_seconds=30,
            memory_mb=512,
            network_disabled=True,
        )

        return {
            "stdout": result.stdout,
            "stderr": result.stderr,
            "exit_code": result.exit_code,
            "success": result.exit_code == 0,
        }
```

## Large Projects

For ML projects with large datasets:

```bash
# Prepare project with data
tar cf ml-project.tar \
  train.py \
  model.py \
  requirements.txt \
  data/

# Submit with higher limits
python-executor run ml-project.tar \
  --entrypoint train.py \
  --timeout 7200 \
  --memory 16384 \
  --disk 10240 \
  --network
```

## Error Handling

### Go

```go
result, err := c.ExecuteSync(ctx, tarData, metadata)
if err != nil {
    log.Printf("Execution failed: %v", err)
    return
}

if result.Status == client.StatusFailed {
    log.Printf("Code failed with error: %s", result.Error)
    log.Printf("stderr: %s", result.Stderr)
}

if result.ExitCode != 0 {
    log.Printf("Non-zero exit code: %d", result.ExitCode)
}
```

### Python

```python
try:
    result = client.execute_sync(files=files, entrypoint="main.py")

    if result.exit_code != 0:
        print(f"Failed with code {result.exit_code}")
        print(f"stderr: {result.stderr}")

except Exception as e:
    print(f"Request failed: {e}")
```
