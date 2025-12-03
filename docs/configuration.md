# Configuration

python-executor is configured via environment variables.

## Server Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `PYEXEC_HOST` | `0.0.0.0` | HTTP server bind address |
| `PYEXEC_PORT` | `8080` | HTTP server port |
| `PYEXEC_LOG_LEVEL` | `info` | Log level (debug, info, warn, error) |
| `PYEXEC_SERVER` | `http://localhost:8080` | Server base URL (used by CLI) |

## Docker Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `PYEXEC_DOCKER_SOCKET` | `/var/run/docker.sock` | Path to Docker socket |
| `PYEXEC_NETWORK_MODE` | `host` | Network mode for execution containers (`host` or `bridge`) |
| `PYEXEC_DNS_SERVERS` | `8.8.8.8,8.8.4.4` | DNS servers for execution containers (comma-separated) |

## Execution Defaults

These values are used when not specified in the request metadata:

| Variable | Default | Description |
|----------|---------|-------------|
| `PYEXEC_DEFAULT_TIMEOUT` | `300` | Default execution timeout (seconds) |
| `PYEXEC_DEFAULT_MEMORY_MB` | `1024` | Default memory limit (MB) |
| `PYEXEC_DEFAULT_DISK_MB` | `2048` | Default disk limit (MB) |
| `PYEXEC_DEFAULT_CPU_SHARES` | `1024` | Default CPU shares |
| `PYEXEC_DEFAULT_IMAGE` | `python:3.12-slim` | Default Docker image |

## Consul Configuration (Optional)

| Variable | Default | Description |
|----------|---------|-------------|
| `PYEXEC_CONSUL_ADDR` | `localhost:8500` | Consul address |
| `PYEXEC_CONSUL_TOKEN` | `` | Consul ACL token |
| `PYEXEC_CONSUL_PREFIX` | `python-executor` | Key prefix in Consul KV |

If `PYEXEC_CONSUL_ADDR` is not set, the server will use in-memory storage.

## Cleanup Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `PYEXEC_CLEANUP_TTL` | `300` | Time to keep completed executions (seconds) |

Cleanup runs every 5 minutes and removes executions older than the TTL.

## Example Configuration

```bash
# Basic setup
export PYEXEC_PORT=8080
export PYEXEC_LOG_LEVEL=debug

# With Consul
export PYEXEC_CONSUL_ADDR=consul.service.consul:8500
export PYEXEC_CONSUL_TOKEN=my-secret-token

# Custom defaults
export PYEXEC_DEFAULT_TIMEOUT=600
export PYEXEC_DEFAULT_MEMORY_MB=2048
```

## CLI Configuration

The CLI tool supports the same configuration via environment variables:

```bash
export PYEXEC_SERVER=http://localhost:8080
python-executor run script.py
```

Or via command-line flags:

```bash
python-executor run --server http://localhost:8080 script.py
```

The `PYEXEC_SERVER` environment variable and `--server` flag allow you to specify the base URL of the python-executor server when using the CLI tool. This is particularly useful when running the CLI against a remote server or a different local port.
