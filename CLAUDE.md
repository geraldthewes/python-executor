# Python Executor

Remote Python code execution service with CLI, Go client, and Python client.

## Build Commands

- `make build` - Build both server and CLI
- `make build-cli` - Build CLI only
- `make build-server` - Build server only
- `make test` - Run all tests
- `make docker-build` - Build Docker image
- `make docker-push` - Build and push Docker image to registry

## Deployment

- `make nomad-restart` - Restart the Nomad service after pushing a new image
- `nomad job allocs python-executor` - List allocations for the job
- `nomad alloc logs <alloc-id>` - View stdout logs for an allocation
- `nomad alloc logs -stderr <alloc-id>` - View stderr logs for an allocation

## Usage

```bash
# Run code from stdin
echo 'print("Hello")' | bin/python-executor --server http://pyexec.cluster:9999/ run

# Run a file
bin/python-executor --server http://pyexec.cluster:9999/ run script.py

# Run a directory
bin/python-executor --server http://pyexec.cluster:9999/ run ./myproject/
```

## Project Structure

- `cmd/python-executor/` - CLI tool
- `cmd/server/` - API server
- `pkg/client/` - Go client library
- `internal/executor/` - Docker execution engine
- `internal/api/` - HTTP handlers
