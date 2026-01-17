# HTTP API Reference

> **IMPORTANT: Use the Client Libraries**
>
> The python-executor API uses `multipart/form-data` with **tar archives**, not JSON.
> This is complex to implement correctly. **Use the official client libraries instead:**
>
> - [Python Client](python-client.md) - `pip install git+https://github.com/geraldthewes/python-executor.git#subdirectory=python`
> - [Go Client](go-client.md) - `go get github.com/geraldthewes/python-executor/pkg/client`
> - [CLI](cli.md) - Command-line interface
>
> The client libraries handle tar archive creation, metadata encoding, multipart requests,
> and response parsing automatically.

---

## Base URL

All API endpoints are prefixed with `/api/v1`.

## Content Type

**POST requests use either `multipart/form-data` or `application/json`** depending on the endpoint:

- `/api/v1/eval` - Uses `application/json` (simple endpoint for AI agents)
- `/api/v1/exec/sync` and `/api/v1/exec/async` - Use `multipart/form-data` with tar archives

---

## Endpoints

### POST /api/v1/eval

Execute code using a simple JSON interface. This endpoint is designed for AI agents and simple integrations.

**Request:**
- Content-Type: `application/json`

**Request Body:**

```json
{
  "code": "print('Hello, World!')",
  "python_version": "3.12",
  "stdin": "input data",
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
  "duration_ms": 150
}
```

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

Execute code synchronously and wait for the result. Uses multipart/form-data with tar archives.

**Request:**
- Content-Type: `multipart/form-data`
- Fields: `tar` (file), `metadata` (JSON string)

**Metadata Schema:**

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
  "duration_ms": 0
}
```

| Field | Description |
|-------|-------------|
| `error_type` | Python exception type extracted from stderr. Only present when `exit_code != 0`. |
| `error_line` | Line number where the error occurred. Only present when `exit_code != 0`. |

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

### Simple JSON execution (for AI agents)

```bash
curl -X POST http://localhost:8080/api/v1/eval \
  -H "Content-Type: application/json" \
  -d '{"code": "print(2 + 2)"}'
```

---

## Size Limits

- Maximum request size: 100 MB
- Maximum tar archive size: 100 MB
- Maximum code size for /api/v1/eval: 100 KB

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
