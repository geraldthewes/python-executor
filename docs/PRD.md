# Product Requirements Document (PRD) – python-executor v1.0  
**Final Locked Version – Ready for Implementation**  
**Date:** 28 November 2025  
**Status:** Approved  

### 1. Project Overview

python-executor is a high-performance, self-hosted, Nomad-native remote Python execution service designed for secure, isolated, and reproducible execution of untrusted or ad-hoc Python code — from single scripts to multi-gigabyte ML projects.

Primary use cases:
- AI agent tool-calling (MCP servers, LLM function calling)
- Secure online judges and education platforms
- Internal automation and data-processing microservices
- CI/CD code generation steps

### 2. Core Design Decisions (Locked)

| Decision                          | Final Choice                                                                 |
|-----------------------------------|-------------------------------------------------------------------------------|
| File transfer method              | Single uncompressed `.tar` archive only (no .tar.gz, no .zip)                |
| API transport                     | `multipart/form-data` with two parts: `tar` (file) + `metadata` (JSON)      |
| Compression                       | None – server extracts streaming tar at line speed with near-zero CPU       |
| Execution orchestrator            | HashiCorp Nomad (docker driver) + Consul KV for async state                 |
| Inner sandbox                     | Docker-in-Docker (privileged) with strict limits and non-root user          |
| Default network access            | Disabled (`--network none`)                                                 |
| Max concurrent jobs               | Configurable via Nomad group count                                          |

### 3. API Specification (v1)

Base URL: `https://python-executor.yourdomain.internal`

| Method | Endpoint                    | Mode        | Request (multipart/form-data)                                  | Response                                      |
|--------|-----------------------------|-------------|----------------------------------------------------------------|-----------------------------------------------|
| POST   | `/api/v1/exec/sync`         | Sync        | `tar` (file) + `metadata` (JSON text/file)                    | Execution result (stdout, stderr, exit_code) |
| POST   | `/api/v1/exec/async`        | Async       | Same as sync                                                   | `{ "execution_id": "exe_..." }`               |
| GET    | `/api/v1/executions/{id}`   | Poll        | —                                                              | Status + result when finished                 |
| DELETE | `/api/v1/executions/{id}`   | Kill        | —                                                              | `{ "status": "killed" }`                      |

#### Metadata JSON (exact schema)

```json
{
  "entrypoint": "main.py",
  "docker_image": "python:3.12-slim",
  "requirements_txt": "numpy\npandas==2.2",
  "pre_commands": [
    "apt update -y",
    "apt install -y gcc"
  ],
  "stdin": "optional input string",
  "config": {
    "timeout_seconds": 300,
    "network_disabled": true,
    "memory_mb": 1024,
    "disk_mb": 2048,
    "cpu_shares": 1024
  }
}
```

### 4. Execution Environment & Security (v1)

| Layer               | Implementation                                   |
|---------------------|--------------------------------------------------|
| Host                | Nomad agent with Docker socket bind-mounted     |
| Inner container     | Non-root user (UID 1000), read-only root fs      |
| Network             | `--network none` by default                      |
| Resource limits     | Enforced by Docker + Nomad                       |
| Filesystem          | All files extracted to `/work` (tmpfs overlay)   |
| Tar extraction      | Streaming, no temp files, strict path sanitization (no `../`) |
| Cleanup             | Job deregistered + alloc purged after 5 min      |

### 5. CLI Tool – `python-executor` (Full Specification)

Single static Go binary. All commands support the same rich input methods.

#### Global flags (available everywhere)
```
--server URL          Server base URL (env: PYEXEC_SERVER)
--timeout N           Override timeout (seconds)
--memory N            Memory limit (MB)
--disk N              Disk limit (MB)
--cpu N               CPU shares
--network             Allow internet access
--image NAME          Custom Docker image
--async               Submit async instead of sync
-q, --quiet           Minimal output
-v, --verbose         Debug output
```

#### Commands

```bash
run        Execute synchronously (most common)
submit     Submit asynchronously
follow     Poll or tail an execution
logs       Stream raw container logs
kill       Terminate a running job
version    Show version info
```

#### Input methods (any command that accepts code)

All five are fully supported and mutually exclusive in precedence order:

| Priority | Method                                | Example                                                                                 |
|----------|---------------------------------------|-----------------------------------------------------------------------------------------|
| 1        | `--file path` (multiple)              | `--file main.py --file utils/helpers.py --file data/model.pkl`                         |
| 2        | Explicit `.tar` file argument         | `python-executor run project.tar`                                                       |
| 3        | Directory argument                    | `python-executor run ./my-project/`                                                     |
| 4        | Single file argument                  | `python-executor run script.py`                                                         |
| 5        | Pipe (stdin)                          | `cat hello.py | python-executor run`                                                    |

The CLI always creates an uncompressed `.tar` stream in memory (spills to disk only > 500 MB).

#### Real-world examples

```bash
# Quick script
cat hello.py | python-executor run

# Full ML project from folder
python-executor run ./stable-diffusion-train/ \
  --entrypoint train.py \
  --timeout 7200 \
  --memory 16384 \
  --network

# Individual files (great for scripts)
python-executor run \
  --file main.py \
  --file requirements.txt \
  --file data/config.yaml \
  --timeout 600

# Async large job
id=$(python-executor submit training-project.tar --async --memory 32768)
python-executor follow --tail $id
```

#### Entrypoint auto-detection order
1. `--entrypoint` flag  
2. `main.py`  
3. `__main__.py`  
4. First `.py` file in archive  
5. Error (user must specify)

### 6. Deliverables

| Deliverable                                    | Description |
|------------------------------------------------|-----------|
| `python-executor` server Docker image          | < 50 MB, multi-stage Go build |
| Example Nomad job spec + Consul integration    | Production-ready |
| OpenAPI 3.1 + Swagger UI at `/docs`            | Self-documenting |
| Python client (`python-executor-client`)       | `pip install python-executor-client` |
| Go client module                               | `github.com/yourorg/python-executor-client-go` |
| CLI binary `python-executor`                   | Full featured, static, supports all 5 input methods |
| Documentation                                  | README, security guide, large-project best practices |

### 7. Quality (v1)

The code will be modular. Each function will have it's unit test

All code will be reviewed with a linter

All code will be reviewed for security flaws

A comprehensive integration test will be created for the Web API and the command line tool.

### 8. Out of Scope (v1)

- Authentication / multi-tenancy
- GPU support
- Interactive REPL / Jupyter
- Streaming output
- Webhooks
- Rootless Docker (planned v2)
- Compressed archive support


This PRD is now final and locked.  
All architectural decisions are made. Begin implementation.
