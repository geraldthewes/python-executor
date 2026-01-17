# API Reference

> **IMPORTANT: Use the Client Libraries**
>
> The python-executor API uses `multipart/form-data` with **tar archives**, not JSON.
> This is complex to implement correctly. **Use the official client libraries instead:**
>
> **Python:**
> ```bash
> pip install git+https://github.com/geraldthewes/python-executor.git#subdirectory=python
> ```
>
> **Go:**
> ```bash
> go get github.com/geraldthewes/python-executor/pkg/client
> ```
>
> The client libraries handle tar archive creation, metadata encoding, multipart requests,
> and response parsing automatically. See [Quick Start with Client Libraries](#quick-start-with-client-libraries) below.

## Quick Start with Client Libraries

### Python (Recommended)

```python
from python_executor_client import PythonExecutorClient

client = PythonExecutorClient("http://localhost:8080")

# Execute code from a dict of files
result = client.execute_sync(
    files={"main.py": "print('Hello, World!')"},
    entrypoint="main.py"
)

print(result.stdout)     # Hello, World!
print(result.exit_code)  # 0
```

### Go (Recommended)

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

    // Create tar from a map of files
    tarData, _ := client.TarFromMap(map[string]string{
        "main.py": `print("Hello from Go!")`,
    })

    result, err := c.ExecuteSync(context.Background(), tarData, &client.Metadata{
        Entrypoint: "main.py",
    })
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println(result.Stdout)
}
```

---

## Raw HTTP API Reference

> **Warning:** Only use the raw HTTP API if you cannot use a client library.
> The API expects `multipart/form-data` with an **uncompressed tar archive**, not JSON.

### Base URL

All API endpoints are prefixed with `/api/v1`.

### Content Type

**All POST requests must use `multipart/form-data`**, not `application/json`.

### Request Format

POST endpoints require two form fields:

| Field | Type | Description |
|-------|------|-------------|
| `tar` | file | **Uncompressed** tar archive containing Python files |
| `metadata` | string | JSON string with execution parameters |

### Metadata Schema

The `metadata` field must be a JSON string with the following structure:

```json
{
  "entrypoint": "main.py",
  "docker_image": "python:3.11-slim",
  "requirements_txt": "numpy\npandas",
  "pre_commands": ["apt-get update", "apt-get install -y libfoo"],
  "stdin": "input data for the script",
  "env_vars": ["API_KEY=secret", "DEBUG=true"],
  "script_args": ["--verbose", "input.txt"],
  "config": {
    "timeout_seconds": 300,
    "network_disabled": true,
    "memory_mb": 1024,
    "disk_mb": 2048,
    "cpu_shares": 1024
  }
}
```

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `entrypoint` | string | Yes | - | Python file to execute (e.g., `main.py`) |
| `docker_image` | string | No | `python:3.11-slim` | Docker image to use |
| `requirements_txt` | string | No | - | Contents of requirements.txt (enables network) |
| `pre_commands` | string[] | No | - | Shell commands to run before execution |
| `stdin` | string | No | - | Data to provide on stdin |
| `env_vars` | string[] | No | - | Environment variables (`KEY=value` format) |
| `script_args` | string[] | No | - | Arguments to pass to the Python script |
| `config.timeout_seconds` | int | No | 300 | Maximum execution time |
| `config.network_disabled` | bool | No | true | Disable network access |
| `config.memory_mb` | int | No | 1024 | Memory limit in MB |
| `config.disk_mb` | int | No | 2048 | Disk limit in MB |
| `config.cpu_shares` | int | No | 1024 | CPU shares |

---

## Endpoints

### POST /api/v1/eval

Execute code using a simple JSON interface. This endpoint is designed for AI agents and simple integrations.
Supports REPL-style evaluation where the last expression's value is returned.

**Request:**
- Content-Type: `application/json`

**Request Body:**

```json
{
  "code": "print('Hello, World!')",
  "python_version": "3.12",
  "stdin": "input data",
  "eval_last_expr": true,
  "config": {
    "timeout_seconds": 30
  }
}
```

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `code` | string | No* | - | Python code to execute (creates `main.py`) |
| `files` | array | No* | - | Multiple files with `name` and `content` |
| `entrypoint` | string | No | `main.py` or first file | File to execute |
| `stdin` | string | No | - | Standard input to provide |
| `python_version` | string | No | `3.12` | Python version: `3.10`, `3.11`, `3.12`, `3.13` |
| `eval_last_expr` | bool | No | false | Enable REPL-style evaluation of last expression |
| `config.timeout_seconds` | int | No | 300 | Maximum execution time |

\* Either `code` or `files` must be provided.

**Multi-file Example:**

```json
{
  "files": [
    {"name": "main.py", "content": "from helper import greet\ngreet()"},
    {"name": "helper.py", "content": "def greet(): print('Hello!')"}
  ],
  "entrypoint": "main.py",
  "python_version": "3.11"
}
```

**Response:** `200 OK`

```json
{
  "execution_id": "exe_550e8400-e29b-41d4-a716-446655440000",
  "status": "completed",
  "stdout": "Hello, World!\n",
  "stderr": "",
  "exit_code": 0,
  "duration_ms": 150,
  "result": null
}
```

**REPL-style Response (with eval_last_expr: true):**

```json
{
  "execution_id": "exe_550e8400-e29b-41d4-a716-446655440000",
  "status": "completed",
  "stdout": "",
  "stderr": "",
  "exit_code": 0,
  "duration_ms": 120,
  "result": "4"
}
```

| Response Field | Description |
|----------------|-------------|
| `result` | The repr() of the last expression's value when `eval_last_expr` is true. `null` if the last statement was not an expression or `eval_last_expr` is false. |

**Error Response (with structured error fields):**

When Python code fails with an error, the response includes parsed error information:

```json
{
  "execution_id": "exe_550e8400-e29b-41d4-a716-446655440000",
  "status": "completed",
  "stdout": "",
  "stderr": "Traceback (most recent call last):\n  File \"main.py\", line 1, in <module>\n    print(undefined_var)\nNameError: name 'undefined_var' is not defined\n",
  "exit_code": 1,
  "error_type": "NameError",
  "error_line": 1,
  "duration_ms": 120
}
```

| Error Field | Description |
|-------------|-------------|
| `error_type` | Python exception type (e.g., `SyntaxError`, `NameError`, `TypeError`) |
| `error_line` | Line number where the error occurred |

**Errors:**
- `400 Bad Request` - Invalid request format or unsupported Python version
- `413 Request Entity Too Large` - Code exceeds 100KB limit
- `500 Internal Server Error` - Execution failed

---

### POST /api/v1/exec/sync

Execute code synchronously and wait for the result.

**Request:**
- Content-Type: `multipart/form-data`
- Fields: `tar` (file), `metadata` (JSON string)

**Response:** `200 OK`

```json
{
  "execution_id": "exe_550e8400-e29b-41d4-a716-446655440000",
  "status": "completed",
  "stdout": "Hello, World!\n",
  "stderr": "",
  "exit_code": 0,
  "started_at": "2024-01-15T10:30:00Z",
  "finished_at": "2024-01-15T10:30:01Z",
  "duration_ms": 1234
}
```

**Errors:**
- `400 Bad Request` - Invalid request format or missing fields
- `500 Internal Server Error` - Execution failed

---

### POST /api/v1/exec/async

Submit code for asynchronous execution. Returns immediately with an execution ID.

**Request:**
- Content-Type: `multipart/form-data`
- Fields: `tar` (file), `metadata` (JSON string)

**Response:** `202 Accepted`

```json
{
  "execution_id": "exe_550e8400-e29b-41d4-a716-446655440000"
}
```

**Errors:**
- `400 Bad Request` - Invalid request format
- `500 Internal Server Error` - Failed to create execution

---

### GET /api/v1/executions/{id}

Get the status and result of an execution.

**Parameters:**
- `id` (path) - Execution ID

**Response:** `200 OK`

```json
{
  "execution_id": "exe_550e8400-e29b-41d4-a716-446655440000",
  "status": "running",
  "stdout": "",
  "stderr": "",
  "exit_code": 0,
  "started_at": "2024-01-15T10:30:00Z",
  "finished_at": null,
  "duration_ms": 0
}
```

**Status Values:**
- `pending` - Waiting to start
- `running` - Currently executing
- `completed` - Finished successfully
- `failed` - Execution failed
- `killed` - Terminated by user

**Errors:**
- `404 Not Found` - Execution not found

---

### DELETE /api/v1/executions/{id}

Kill a running execution.

**Parameters:**
- `id` (path) - Execution ID

**Response:** `200 OK`

```json
{
  "status": "killed"
}
```

**Errors:**
- `404 Not Found` - Execution not found
- `500 Internal Server Error` - Failed to kill execution

---

### GET /health

Health check endpoint.

**Response:** `200 OK`

```json
{
  "status": "ok"
}
```

---

## Response Schema

### ExecutionResult

```json
{
  "execution_id": "string",
  "status": "pending|running|completed|failed|killed",
  "stdout": "string",
  "stderr": "string",
  "exit_code": 0,
  "error": "string (only if failed)",
  "error_type": "string (e.g., SyntaxError, NameError)",
  "error_line": 0,
  "started_at": "ISO 8601 timestamp",
  "finished_at": "ISO 8601 timestamp",
  "duration_ms": 0,
  "result": "string (REPL expression value, or null)"
}
```

| Field | Description |
|-------|-------------|
| `error_type` | Python exception type extracted from stderr (e.g., `SyntaxError`, `NameError`, `TypeError`). Only present when `exit_code != 0`. |
| `error_line` | Line number where the error occurred, extracted from Python traceback. Only present when `exit_code != 0`. |
| `result` | The repr() of the last expression's value when `eval_last_expr` is true. `null` if the last statement was not an expression or when using exec/sync endpoint. |

### Error Response

```json
{
  "error": "error message"
}
```

---

## curl Examples

> **Reminder:** Use the client libraries instead. These examples are for debugging only.

### Create a tar archive

```bash
# From a single file
tar -cf code.tar script.py

# From multiple files
tar -cf code.tar main.py utils.py requirements.txt

# From a directory
tar -cf code.tar -C ./my-project .
```

### Execute synchronously

```bash
# Create a simple script
echo 'print("Hello, World!")' > main.py
tar -cf code.tar main.py

# Execute
curl -X POST http://localhost:8080/api/v1/exec/sync \
  -F "tar=@code.tar" \
  -F 'metadata={"entrypoint":"main.py"}'
```

### Execute with requirements

```bash
curl -X POST http://localhost:8080/api/v1/exec/sync \
  -F "tar=@code.tar" \
  -F 'metadata={"entrypoint":"main.py","requirements_txt":"numpy\npandas","config":{"network_disabled":false}}'
```

### Execute asynchronously

```bash
# Submit
EXEC_ID=$(curl -s -X POST http://localhost:8080/api/v1/exec/async \
  -F "tar=@code.tar" \
  -F 'metadata={"entrypoint":"main.py"}' | jq -r .execution_id)

# Poll for result
curl http://localhost:8080/api/v1/executions/$EXEC_ID
```

### Kill an execution

```bash
curl -X DELETE http://localhost:8080/api/v1/executions/$EXEC_ID
```

---

## Size Limits

- Maximum request size: 100 MB
- Maximum tar archive size: 100 MB

---

## OpenAPI/Swagger

When the server is running, Swagger UI is available at:

```
http://localhost:8080/docs/index.html
```

The OpenAPI specification is available at:

```
http://localhost:8080/docs/doc.json
```
