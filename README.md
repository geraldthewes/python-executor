# python-executor

A high-performance, self-hosted remote Python execution service designed for secure, isolated, and reproducible execution of untrusted or ad-hoc Python code — from single scripts to multi-gigabyte ML projects.

## Features

- 🔒 **Secure Execution** - Docker-in-Docker isolation with strict resource limits
- 📦 **Multi-file Projects** - Support for entire project directories via uncompressed tar archives
- ⚡ **Sync & Async** - Both synchronous and asynchronous execution modes
- 🌐 **Multiple Interfaces** - REST API, Go client, Python client, and CLI tool
- 🔌 **Consul Integration** - Optional distributed state storage with in-memory fallback
- 📝 **API Documentation** - Complete API reference with Swagger UI at `/docs`
- 🐍 **Custom Requirements** - Install dependencies via requirements.txt
- ⚙️ **Pre-execution Commands** - Run setup commands (apt install, etc.)
- 🎯 **Resource Limits** - Configurable memory, CPU, disk, and timeout limits
- 🚫 **Network Isolation** - Network disabled by default for security
- 🔄 **Environment Variables** - Pass environment variables with `--env`
- 📋 **Script Arguments** - Pass arguments to scripts with `--` separator

## Use Cases

- AI agent tool-calling (MCP servers, LLM function calling)
- Secure online judges and education platforms
- Internal automation and data-processing microservices
- CI/CD code generation steps
- Remote script execution for distributed systems

## For AI Code Agents

Two integration approaches are available:

### Option A: Simple JSON API (Recommended for most AI agents)

Use `POST /api/v1/eval` with plain JSON — no tar archives or multipart encoding needed.

```bash
# Single script
curl -X POST http://localhost:8080/api/v1/eval \
  -H "Content-Type: application/json" \
  -d '{"code": "print(\"Hello, World!\")"}'

# With specific Python version
curl -X POST http://localhost:8080/api/v1/eval \
  -H "Content-Type: application/json" \
  -d '{"code": "import sys; print(sys.version)", "python_version": "3.11"}'

# With timeout configuration
curl -X POST http://localhost:8080/api/v1/eval \
  -H "Content-Type: application/json" \
  -d '{"code": "print(\"Hello\")", "config": {"timeout_seconds": 30}}'

# Multiple files
curl -X POST http://localhost:8080/api/v1/eval \
  -H "Content-Type: application/json" \
  -d '{
    "files": [
      {"name": "main.py", "content": "from helper import greet\ngreet()"},
      {"name": "helper.py", "content": "def greet(): print(\"Hello!\")"}
    ],
    "entrypoint": "main.py"
  }'
```

**Supported Python versions:** 3.10, 3.11, 3.12 (default), 3.13

**Best for:** Simple scripts, quick evaluations, LLM tool-calling
**Limitation:** 100KB total code size

### Option B: Client Libraries

For large projects or complex file structures, use the client libraries which handle tar archives automatically.

**Python:**
```bash
pip install git+https://github.com/geraldthewes/python-executor.git#subdirectory=python
```

**Go:**
```bash
go get github.com/geraldthewes/python-executor/pkg/client
```

**Best for:** Large projects, complex file structures, multi-MB codebases

Both approaches return the same response format:
```json
{
  "execution_id": "exe_...",
  "status": "completed",
  "stdout": "Hello, World!\n",
  "stderr": "",
  "exit_code": 0,
  "duration_ms": 150
}
```

**Error responses include structured error information:**
```json
{
  "execution_id": "exe_...",
  "status": "completed",
  "stdout": "",
  "stderr": "Traceback...\nNameError: name 'x' is not defined",
  "exit_code": 1,
  "error_type": "NameError",
  "error_line": 1,
  "duration_ms": 120
}
```

See [API Reference](docs/api.md) for complete documentation.

## Quick Start

### Using Docker Compose

```bash
cd deploy
docker-compose up
```

The server will be available at `http://localhost:8080`.

### Using the CLI

Install:
```bash
go install github.com/geraldthewes/python-executor/cmd/python-executor@latest
```

Basic usage:
```bash
# Run from stdin
echo 'print("Hello, World")' | python-executor run

# Run a single file
python-executor run script.py

# Run a directory
python-executor run ./my-project/ --entrypoint main.py

# Run with custom limits
python-executor run script.py --timeout 600 --memory 2048 --network

# Pass environment variables to the script
python-executor run script.py --env API_KEY --env DEBUG=true

# Pass arguments to the script (after --)
python-executor run script.py -- arg1 arg2 --verbose

# Combined: env vars and script arguments
python-executor run script.py --env SERVICE_HOST -e SERVICE_PORT -- http://host:port

# Async execution
id=$(python-executor submit long-script.py --async)
python-executor follow $id

# Specify server URL via flag
python-executor --server http://my-server:8080 run script.py

# Specify server URL via environment variable
export PYEXEC_SERVER=http://my-server:8080
python-executor run script.py
```

See [Examples](docs/examples.md) for more CLI usage patterns.

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

Install from repository:
```bash
pip install git+https://github.com/geraldthewes/python-executor.git#subdirectory=python
```

Usage:
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
| POST | `/api/v1/eval` | Execute code via simple JSON (AI-friendly) |
| POST | `/api/v1/exec/sync` | Execute code synchronously |
| POST | `/api/v1/exec/async` | Submit code for async execution |
| GET | `/api/v1/executions/{id}` | Get execution status and result |
| DELETE | `/api/v1/executions/{id}` | Kill a running execution |
| GET | `/health` | Health check endpoint |

See [API Documentation](docs/api.md) for detailed API specs. **Note:** Use client libraries instead of implementing the HTTP protocol directly.

## Building from Source

Requirements:
- Go 1.23+
- Docker
- Make

Build commands:
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

Binaries will be in `./bin/`.

## Deployment

### Local Development

```bash
# With in-memory storage
make run-server

# With Consul (requires Consul running)
export PYEXEC_CONSUL_ADDR=localhost:8500
make run-server
```

### Docker

```bash
docker build -t python-executor -f deploy/Dockerfile .

docker run -d \
  --name python-executor \
  --privileged \
  -p 8080:8080 \
  -v /var/run/docker.sock:/var/run/docker.sock \
  python-executor
```

### Nomad

See [deploy/nomad.hcl](deploy/nomad.hcl) for a production-ready Nomad job specification.

```bash
nomad job run deploy/nomad.hcl
```

## Configuration

Configuration is via environment variables:

```bash
# Server
export PYEXEC_PORT=8080
export PYEXEC_LOG_LEVEL=info
export PYEXEC_SERVER=http://localhost:8080

# Consul (optional)
export PYEXEC_CONSUL_ADDR=consul:8500

# Defaults
export PYEXEC_DEFAULT_TIMEOUT=300
export PYEXEC_DEFAULT_MEMORY_MB=1024

# Docker network settings
export PYEXEC_NETWORK_MODE=host          # or "bridge"
export PYEXEC_DNS_SERVERS=8.8.8.8,8.8.4.4
```

The `PYEXEC_SERVER` environment variable specifies the base URL of the python-executor server. This can be overridden with the `--server` flag when using the CLI tool.

**Note on Ports**: The python-executor server defaults to port 8080. If your scripts use frameworks like FastAPI/uvicorn (which default to port 8000), ensure your port configuration is explicit to avoid mismatches.

See [Configuration Guide](docs/configuration.md) for all options.

## Security

python-executor uses multiple security layers:

- **Docker Isolation** - Each execution in separate container
- **Non-root User** - Code runs as UID 1000
- **Resource Limits** - Memory, CPU, disk, timeout enforcement
- **Network Disabled** - No network access by default
- **Read-only Root** - Filesystem is read-only
- **Path Sanitization** - Prevents path traversal attacks

**Important:** This service requires Docker socket access and should be deployed with appropriate security measures. See [Security Guide](docs/security.md) for best practices.

## Documentation

- [API Reference](docs/api.md) - Complete API specification (use client libraries instead of raw HTTP)
- [Configuration](docs/configuration.md) - Environment variables and settings
- [Security](docs/security.md) - Security considerations and best practices
- [Examples](docs/examples.md) - Usage examples for all interfaces
- [PRD](docs/PRD.md) - Product Requirements Document

## Project Structure

```
python-executor/
├── cmd/
│   ├── server/              # API server
│   └── python-executor/     # CLI tool
├── internal/
│   ├── api/                 # HTTP handlers
│   ├── config/              # Configuration
│   ├── executor/            # Docker execution engine
│   ├── storage/             # State storage backends
│   └── tar/                 # Tar archive handling
├── pkg/
│   └── client/              # Go client library
├── python/
│   └── python_executor_client/  # Python client
├── deploy/
│   ├── Dockerfile
│   ├── docker-compose.yaml
│   └── nomad.hcl
└── docs/                    # Documentation
```

## Contributing

Contributions welcome! Please:

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Ensure all tests pass (`make test`)
5. Submit a pull request

## License

MIT License - see [LICENSE](LICENSE) for details

## Acknowledgments

Inspired by:
- [unclefomotw/code-executor](https://github.com/unclefomotw/code-executor)
- [geraldthewes/nomad-mcp-builder](https://github.com/geraldthewes/nomad-mcp-builder)
