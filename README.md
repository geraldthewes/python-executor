# python-executor

A high-performance, self-hosted remote Python execution service designed for secure, isolated, and reproducible execution of untrusted or ad-hoc Python code — from single scripts to multi-gigabyte ML projects.

## Features

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

## Use Cases

- AI agent tool-calling (MCP servers, LLM function calling)
- Secure online judges and education platforms
- Internal automation and data-processing microservices
- CI/CD code generation steps
- Remote script execution for distributed systems

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
echo 'print("Hello, World!")' | python-executor run

# Run a single file
python-executor run script.py

# Run a directory
python-executor run ./my-project/ --entrypoint main.py

# Run with custom limits
python-executor run script.py --timeout 600 --memory 2048 --network

# Async execution
id=$(python-executor submit long-script.py --async)
python-executor follow $id
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

Install:
```bash
cd python
pip install .
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
| POST | `/api/v1/exec/sync` | Execute code synchronously |
| POST | `/api/v1/exec/async` | Submit code for async execution |
| GET | `/api/v1/executions/{id}` | Get execution status and result |
| DELETE | `/api/v1/executions/{id}` | Kill a running execution |
| GET | `/health` | Health check endpoint |

See [API Documentation](docs/api.md) for detailed API specs.

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

# Consul (optional)
export PYEXEC_CONSUL_ADDR=consul:8500

# Defaults
export PYEXEC_DEFAULT_TIMEOUT=300
export PYEXEC_DEFAULT_MEMORY_MB=1024
```

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
