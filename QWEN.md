# python-executor Project Analysis

## Project Overview

python-executor is a high-performance, self-hosted remote Python execution service designed for secure, isolated, and reproducible execution of untrusted or ad-hoc Python code. It supports execution from single scripts to multi-gigabyte ML projects.

## Key Features

- 🔒 **Secure Execution** - Docker-in-Docker isolation with strict resource limits
- 📦 **Multi-file Projects** - Support for entire project directories via uncompressed tar archives
- ⚡ **Sync & Async** - Both synchronous and asynchronous execution modes
- 🌐 **Multiple Interfaces** - REST API, Go client, Python client, and CLI tool
- 🔌 **Consul Integration** - Optional distributed state storage with in-memory fallback
- 📝 **OpenAPI Ready** - Self-documenting API with Swagger UI
- 🐍 **Custom Requirements** - Install dependencies via requirements.txt
- ⚙️ **Pre-execution Commands** - Run setup commands (apt install, etc.)
- 🎯 **Resource Limits** - Configurable memory, CPU, disk, and timeout limits
- 🚫 **Network Isolation** - Network disabled by default for security

## Architecture

The project follows a modular architecture with the following key components:

- **Server**: API server (cmd/server) that handles HTTP requests
- **CLI**: Command-line interface (cmd/python-executor)
- **Go Client**: Library for Go applications (pkg/client)
- **Python Client**: Library for Python applications (python/python_executor_client)
- **Internal Modules**:
  - api: HTTP handlers and routing
  - config: Configuration management
  - executor: Docker execution engine
  - storage: State storage backends
  - tar: Tar archive handling

## Build and Deployment

### Build Process
The project uses Go 1.25.4 and Make for building:

```bash
# Build everything
make build

# Build server only
make build-server

# Build CLI only
make build-cli

# Run unit tests
make test

# Build Docker image
make docker-build
```

### Deployment Options
- **Local Development**: `make run-server`
- **Docker**: `docker build -t python-executor -f deploy/Dockerfile .`
- **Nomad**: See `deploy/nomad.hcl` for a production-ready Nomad job specification

## Configuration

Configuration is via environment variables:
```bash
# Server
export PYEXEC_PORT=8080
export PYEXEC_LOG_LEVEL=info

# Consul (optional)
export PYEXEC_CONSUL_ADDR=consul:8500

# Defaults
export PYEXEC_DEFAULT_TIMEOUT=300
export PYEXEC_DEFAULT_MEMORY_MB=1024
```

## Security

The service uses multiple security layers:
- Docker Isolation - Each execution in separate container
- Non-root User - Code runs as UID 1000
- Resource Limits - Memory, CPU, disk, timeout enforcement
- Network Disabled - No network access by default
- Read-only Root - Filesystem is read-only
- Path Sanitization - Prevents path traversal attacks

## Usage Examples

### Using Docker Compose
```bash
cd deploy
docker-compose up
```

### Using the CLI
```bash
# Install
go install github.com/geraldthewes/python-executor/cmd/python-executor@latest

# Basic usage
echo 'print("Hello, World!")' | python-executor run
python-executor run script.py
python-executor run ./my-project/ --entrypoint main.py
python-executor run script.py --timeout 600 --memory 2048 --network
id=$(python-executor submit long-script.py --async)
python-executor follow $id
```

### Using the Go Client
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

### Using the Python Client
```python
from python_executor_client import PythonExecutorClient

client = PythonExecutorClient("http://localhost:8080")

result = client.execute_sync(
    files={"main.py": "print('Hello from Python!')"},
    entrypoint="main.py"
)

print(result.stdout)
print(f"Exit code: {result.exit_code}")
```

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/exec/sync` | Execute code synchronously |
| POST | `/api/v1/exec/async` | Submit code for async execution |
| GET | `/api/v1/executions/{id}` | Get execution status and result |
| DELETE | `/api/v1/executions/{id}` | Kill a running execution |
| GET | `/health` | Health check endpoint |

## Development Conventions

- Uses Go 1.25.4
- Follows standard Go project structure
- Unit tests with `go test`
- Integration tests with `go test -tags=integration`
- Linting with golangci-lint
- Docker-in-Docker for execution isolation
- Comprehensive documentation in docs/ directory