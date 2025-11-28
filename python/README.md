# python-executor-client

Python client library for the python-executor service.

## Installation

```bash
pip install python-executor-client
```

Or install from source:

```bash
cd python
pip install .
```

## Usage

```python
from python_executor_client import PythonExecutorClient

# Initialize client
client = PythonExecutorClient("http://localhost:8080")

# Execute from dict of files
result = client.execute_sync(
    files={"main.py": "print('Hello, World!')"},
    entrypoint="main.py"
)

print(result.stdout)  # Hello, World!
print(result.exit_code)  # 0

# Execute from directory
result = client.execute_sync(
    files="./my-project",
    entrypoint="main.py",
    requirements_txt="numpy\npandas",
)

# Async execution
exec_id = client.execute_async(
    files={"script.py": "import time; time.sleep(10); print('Done')"},
    entrypoint="script.py"
)

# Wait for completion
result = client.wait_for_completion(exec_id)
print(result.stdout)
```

## API Reference

### PythonExecutorClient

#### `__init__(base_url, timeout=300)`

Initialize the client.

- `base_url`: Server URL
- `timeout`: Request timeout in seconds

#### `execute_sync(files=None, tar_data=None, metadata=None, **kwargs)`

Execute code synchronously and wait for result.

- `files`: Dict of filename->content, Path to directory, or path string
- `tar_data`: Raw tar bytes (alternative to files)
- `metadata`: Metadata object
- `**kwargs`: Metadata fields (entrypoint, docker_image, requirements_txt, etc.)

Returns `ExecutionResult`.

#### `execute_async(files=None, tar_data=None, metadata=None, **kwargs)`

Submit code for async execution.

Returns execution ID string.

#### `get_execution(execution_id)`

Get execution status and result.

Returns `ExecutionResult`.

#### `kill(execution_id)`

Kill a running execution.

#### `wait_for_completion(execution_id, poll_interval=2.0, max_wait=None)`

Poll until execution completes.

Returns `ExecutionResult`.

## License

MIT
