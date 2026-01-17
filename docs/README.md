# Python Executor Documentation

Remote Python code execution service with client libraries for Python, Go, and CLI.

## Quick Links

| Document | Description |
|----------|-------------|
| [HTTP API Reference](http-api.md) | REST API endpoints, schemas, and curl examples |
| [Python Client](python-client.md) | Python client library reference |
| [Go Client](go-client.md) | Go client library reference |
| [CLI Reference](cli.md) | Command-line interface documentation |
| [Configuration](configuration.md) | Server configuration options |
| [Security](security.md) | Security model and considerations |
| [Examples](examples.md) | Code examples and usage patterns |

## Getting Started

### Choose Your Interface

**For AI Agents / Simple Integrations:**
Use the JSON-based `/api/v1/eval` endpoint:
```bash
curl -X POST http://pyexec.cluster:9999/api/v1/eval \
  -H "Content-Type: application/json" \
  -d '{"code": "print(2 + 2)"}'
```

**For Python Applications:**
```python
from python_executor_client import PythonExecutorClient

client = PythonExecutorClient("http://pyexec.cluster:9999/")
result = client.execute_sync(
    files={"main.py": "print('Hello!')"},
    entrypoint="main.py"
)
print(result.stdout)
```

**For Go Applications:**
```go
c := client.New("http://pyexec.cluster:9999/")
tarData, _ := client.TarFromMap(map[string]string{
    "main.py": `print("Hello!")`,
})
result, _ := c.ExecuteSync(ctx, tarData, &client.Metadata{
    Entrypoint: "main.py",
})
fmt.Println(result.Stdout)
```

**For Shell Scripts / CLI:**
```bash
echo 'print("Hello")' | python-executor --server http://pyexec.cluster:9999/ run
```

## Installation

### Python Client
```bash
pip install git+https://github.com/geraldthewes/python-executor.git#subdirectory=python
```

### Go Client
```bash
go get github.com/geraldthewes/python-executor/pkg/client
```

### CLI
```bash
# Build from source
make build-cli
sudo cp bin/python-executor /usr/local/bin/
```

## Server

### Running the Server
```bash
# Build
make build-server

# Run locally
./bin/python-executor-server
```

### Docker
```bash
docker run -p 8080:8080 -v /var/run/docker.sock:/var/run/docker.sock \
  registry.cluster:5000/python-executor:latest
```

## API Overview

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/v1/eval` | POST | Simple JSON execution (for AI agents) |
| `/api/v1/exec/sync` | POST | Execute code synchronously |
| `/api/v1/exec/async` | POST | Submit code for async execution |
| `/api/v1/executions/{id}` | GET | Get execution status |
| `/api/v1/executions/{id}` | DELETE | Kill execution |
| `/health` | GET | Health check |

See [HTTP API Reference](http-api.md) for full details.

## Documentation Generation

This documentation is partially auto-generated from source code:

```bash
# Generate all docs
make docs-generate

# Test examples against deployed server
make docs-test

# Full pipeline (generate + test)
make docs
```
