# Python Client Library

The Python client provides a simple interface for executing Python code remotely.

## Installation

```bash
pip install git+https://github.com/geraldthewes/python-executor.git#subdirectory=python
```

## Quick Start

```python
from python_executor_client import PythonExecutorClient

client = PythonExecutorClient("http://pyexec.cluster:9999/")

# Execute code from a dict of files
result = client.execute_sync(
    files={"main.py": "print('Hello, World!')"},
    entrypoint="main.py"
)

print(result.stdout)     # Hello, World!
print(result.exit_code)  # 0
```

## API Reference


<a href="https://github.com/geraldthewes/python-executor/blob/main/python/python/python_executor_client/client.py#L0"><img align="right" style="float:right;" src="https://img.shields.io/badge/-source-cccccc?style=flat-square" /></a>

# <kbd>module</kbd> `python_executor_client.client`
Python client for python-executor.

This module provides a Python client library for the python-executor service,
which allows remote execution of Python code in isolated Docker containers.


**Example:**

    Basic usage with a simple script:

    ```python
    >>> from python_executor_client import PythonExecutorClient
    >>> client = PythonExecutorClient("http://pyexec.cluster:9999/")
    >>> result = client.execute_sync(
    ...     files={"main.py": "print('Hello, World!')"},
    ...     entrypoint="main.py"
    ... )
    >>> print(result.stdout)
    Hello, World!
    >>> print(result.exit_code)
    0

    Multi-file execution:

    >>> result = client.execute_sync(
    ...     files={
    ...         "main.py": "from helper import greet; greet()",
    ...         "helper.py": "def greet(): print('Hello from helper!')"
    ...     },
    ...     entrypoint="main.py"
    ... )
    ```





---

<a href="https://github.com/geraldthewes/python-executor/blob/main/python/python/python_executor_client/client.py#L43"><img align="right" style="float:right;" src="https://img.shields.io/badge/-source-cccccc?style=flat-square" /></a>

## <kbd>class</kbd> `PythonExecutorClient`
Client for the python-executor remote code execution service.

This client handles all the complexity of the python-executor API, including:
- Creating tar archives from files, directories, or dicts
- Encoding metadata as JSON
- Managing multipart/form-data requests
- Parsing execution results


**Attributes:**

- <b>`base_url`</b>: The base URL of the python-executor server.
- <b>`timeout`</b>: HTTP request timeout in seconds.
- <b>`session`</b>: The underlying requests Session object.


**Example:**

```python
    >>> client = PythonExecutorClient("http://pyexec.cluster:9999/")
>>> result = client.execute_sync(
...     files={"main.py": "print('Hello!')"},
...     entrypoint="main.py"
... )
>>> result.exit_code
0
    ```


<a href="https://github.com/geraldthewes/python-executor/blob/main/python/python/python_executor_client/client.py#L67"><img align="right" style="float:right;" src="https://img.shields.io/badge/-source-cccccc?style=flat-square" /></a>

### <kbd>constructor</kbd> `__init__`

```python
PythonExecutorClient(base_url: str, timeout: int = 300)
```

Initialize the Python executor client.


**Args:**

- <b>`base_url`</b>: Base URL of the python-executor server (e.g., "http://pyexec.cluster:9999/").
- <b>`timeout`</b>: HTTP request timeout in seconds. Default is 300 (5 minutes).


**Example:**

```python
    >>> client = PythonExecutorClient("http://pyexec.cluster:9999/")
>>> client = PythonExecutorClient("http://localhost:8080", timeout=60)
    ```





---

<a href="https://github.com/geraldthewes/python-executor/blob/main/python/python/python_executor_client/client.py#L163"><img align="right" style="float:right;" src="https://img.shields.io/badge/-source-cccccc?style=flat-square" /></a>

### <kbd>method</kbd> `execute_async`

```python
execute_async(
    files: Optional[dict[str, str], Path, str] = None,
    tar_data: Optional[bytes] = None,
    metadata: Optional[Metadata] = None,
    **kwargs
) → str
```

Submit Python code for asynchronous execution.

This method submits code and returns immediately with an execution ID.
Use this for long-running scripts or when you need to run multiple
scripts concurrently.


**Args:**

- <b>`files`</b>: Python files to execute. Same options as execute_sync().
- <b>`tar_data`</b>: Pre-built tar archive bytes (alternative to files).
- <b>`metadata`</b>: Full Metadata object for advanced configuration.
- <b>`**kwargs`</b>: Shorthand for metadata fields. See execute_sync() for options.


**Returns:**

- <b>`str`</b>: Execution ID that can be used with get_execution() and wait_for_completion().


**Raises:**

- <b>`ValueError`</b>: If neither files nor tar_data is provided.
- <b>`requests.HTTPError`</b>: If the server returns an error response.


**Example:**

```python
    >>> client = PythonExecutorClient("http://pyexec.cluster:9999/")
>>> exec_id = client.execute_async(
...     files={"main.py": "import time; time.sleep(60); print('done')"}
... )
>>> print(f"Submitted: {exec_id}")
Submitted: exe_550e8400-e29b-41d4-a716-446655440000

>>> # Later, check status or wait for completion
>>> result = client.wait_for_completion(exec_id)
    ```


---

<a href="https://github.com/geraldthewes/python-executor/blob/main/python/python/python_executor_client/client.py#L82"><img align="right" style="float:right;" src="https://img.shields.io/badge/-source-cccccc?style=flat-square" /></a>

### <kbd>method</kbd> `execute_sync`

```python
execute_sync(
    files: Optional[dict[str, str], Path, str] = None,
    tar_data: Optional[bytes] = None,
    metadata: Optional[Metadata] = None,
    **kwargs
) → ExecutionResult
```

Execute Python code synchronously and wait for the result.

This method submits code for execution and blocks until completion.
Use this for short-running scripts where you need the result immediately.


**Args:**

- <b>`files`</b>: Python files to execute. Can be:
- <b>`- dict[str, str]: Mapping of filename to content (e.g., {"main.py"`</b>: "print('hi')"})
- <b>`- Path`</b>: Path to a file or directory
- <b>`- str`</b>: String path to a file or directory
- <b>`tar_data`</b>: Pre-built tar archive bytes (alternative to files).
    If provided, files parameter is ignored.
- <b>`metadata`</b>: Full Metadata object for advanced configuration.
    If provided, kwargs are ignored.
- <b>`**kwargs`</b>: Shorthand for metadata fields:
- <b>`- entrypoint (str)`</b>: Script to run (auto-detected if not specified)
- <b>`- docker_image (str): Docker image (default`</b>: python:3.11-slim)
- <b>`- requirements_txt (str)`</b>: Contents of requirements.txt
- <b>`- pre_commands (list[str])`</b>: Shell commands to run before execution
- <b>`- stdin (str)`</b>: Data to provide on stdin
- <b>`- timeout_seconds (int)`</b>: Execution timeout
- <b>`- network_disabled (bool)`</b>: Disable network access
- <b>`- memory_mb (int)`</b>: Memory limit in MB
- <b>`- disk_mb (int)`</b>: Disk limit in MB
- <b>`- cpu_shares (int)`</b>: CPU shares


**Returns:**

- <b>`ExecutionResult`</b>: Object containing stdout, stderr, exit_code, and execution metadata.


**Raises:**

- <b>`ValueError`</b>: If neither files nor tar_data is provided.
- <b>`requests.HTTPError`</b>: If the server returns an error response.


**Example:**

Simple execution:

```python
    >>> client = PythonExecutorClient("http://pyexec.cluster:9999/")
>>> result = client.execute_sync(
...     files={"main.py": "print('Hello, World!')"},
...     entrypoint="main.py"
... )
>>> result.exit_code
0
>>> result.stdout
'Hello, World!\n'

Execution with dependencies:

>>> result = client.execute_sync(
...     files={"main.py": "import numpy; print(numpy.__version__)"},
...     requirements_txt="numpy",
...     network_disabled=False
... )

Execution from a directory:

>>> result = client.execute_sync(files=Path("./myproject/"))
    ```


---

<a href="https://github.com/geraldthewes/python-executor/blob/main/python/python/python_executor_client/client.py#L216"><img align="right" style="float:right;" src="https://img.shields.io/badge/-source-cccccc?style=flat-square" /></a>

### <kbd>method</kbd> `get_execution`

```python
get_execution(execution_id: str) → ExecutionResult
```

Get the current status and result of an execution.


**Args:**

- <b>`execution_id`</b>: The execution ID returned by execute_async().


**Returns:**

- <b>`ExecutionResult`</b>: Current status and any available output.
    The status field indicates whether the execution is still running.


**Raises:**

- <b>`requests.HTTPError`</b>: If the execution is not found (404) or server error.


**Example:**

```python
    >>> client = PythonExecutorClient("http://pyexec.cluster:9999/")
>>> result = client.get_execution("exe_550e8400-e29b-41d4-a716-446655440000")
>>> print(result.status)
running
    ```


---

<a href="https://github.com/geraldthewes/python-executor/blob/main/python/python/python_executor_client/client.py#L243"><img align="right" style="float:right;" src="https://img.shields.io/badge/-source-cccccc?style=flat-square" /></a>

### <kbd>method</kbd> `kill`

```python
kill(execution_id: str) → None
```

Terminate a running execution.

Forcefully stops the Docker container running the Python code.


**Args:**

- <b>`execution_id`</b>: The execution ID to kill.


**Raises:**

- <b>`requests.HTTPError`</b>: If the execution is not found (404) or server error.


**Example:**

```python
    >>> client = PythonExecutorClient("http://pyexec.cluster:9999/")
>>> exec_id = client.execute_async(files={"main.py": "import time; time.sleep(3600)"})
>>> client.kill(exec_id)
    ```


---

<a href="https://github.com/geraldthewes/python-executor/blob/main/python/python/python_executor_client/client.py#L265"><img align="right" style="float:right;" src="https://img.shields.io/badge/-source-cccccc?style=flat-square" /></a>

### <kbd>method</kbd> `wait_for_completion`

```python
wait_for_completion(
    execution_id: str,
    poll_interval: float = 2.0,
    max_wait: Optional[float] = None
) → ExecutionResult
```

Wait for an asynchronous execution to complete.

Polls the server periodically until the execution finishes.


**Args:**

- <b>`execution_id`</b>: The execution ID returned by execute_async().
- <b>`poll_interval`</b>: Seconds between status checks. Default is 2.0.
- <b>`max_wait`</b>: Maximum seconds to wait before raising TimeoutError.
    None means wait indefinitely. Default is None.


**Returns:**

- <b>`ExecutionResult`</b>: Final result with stdout, stderr, and exit_code.


**Raises:**

- <b>`TimeoutError`</b>: If max_wait is exceeded before completion.
- <b>`requests.HTTPError`</b>: If the server returns an error.


**Example:**

```python
    >>> client = PythonExecutorClient("http://pyexec.cluster:9999/")
>>> exec_id = client.execute_async(
...     files={"main.py": "import time; time.sleep(5); print('done')"}
... )
>>> result = client.wait_for_completion(exec_id, poll_interval=1.0)
>>> print(result.stdout)
done

With timeout:

>>> try:
...     result = client.wait_for_completion(exec_id, max_wait=10.0)
... except TimeoutError:
...     client.kill(exec_id)
    ```


---

<a href="https://github.com/geraldthewes/python-executor/blob/main/python/python/python_executor_client/client.py#L317"><img align="right" style="float:right;" src="https://img.shields.io/badge/-source-cccccc?style=flat-square" /></a>

### <kbd>method</kbd> `eval`

```python
eval(
    code: str,
    *,
    files: Optional[list[dict[str, str]]] = None,
    entrypoint: Optional[str] = None,
    stdin: Optional[str] = None,
    python_version: Optional[str] = None,
    timeout_seconds: Optional[int] = None,
    eval_last_expr: bool = True
) → ExecutionResult
```

Execute code with REPL-style expression evaluation.

This method uses the simplified /api/v1/eval endpoint which accepts JSON
and automatically evaluates the last expression in the code, returning
its value in the result field.


**Args:**

- <b>`code`</b>: Python code to execute. Creates a main.py with this content.
- <b>`files`</b>: Optional list of file dicts with "name" and "content" keys.
    Takes precedence over code if provided.
- <b>`entrypoint`</b>: File to execute. Defaults to "main.py" or first file.
- <b>`stdin`</b>: Standard input to provide to the script.
- <b>`python_version`</b>: Python version to use ("3.10", "3.11", "3.12", "3.13").
    Defaults to server default (typically 3.12).
- <b>`timeout_seconds`</b>: Maximum execution time in seconds.
- <b>`eval_last_expr`</b>: If True (default), evaluate the last expression and
    return its value in result. If False, behave like normal execution.


**Returns:**

- <b>`ExecutionResult`</b>: Object containing stdout, stderr, exit_code, and result.
    The result field contains the repr() of the last expression's value,
    or None if the last statement was not an expression.


**Raises:**

- <b>`requests.HTTPError`</b>: If the server returns an error response.


**Example:**

Simple expression:

```python
>>> result = client.eval("2 + 2")
>>> print(result.result)
4
```

Multi-line code:

```python
>>> result = client.eval("x = 5\ny = 10\nx + y")
>>> print(result.result)
15
```

Using imports:

```python
>>> result = client.eval("import math\nmath.sqrt(16)")
>>> print(result.result)
4.0
```

With print (result is None):

```python
>>> result = client.eval("print('hello')")
>>> print(result.stdout)
hello
>>> print(result.result)
None
```


---

_This file was automatically generated via [lazydocs](https://github.com/ml-tooling/lazydocs)._

---


<a href="https://github.com/geraldthewes/python-executor/blob/main/python/python/python_executor_client/types.py#L0"><img align="right" style="float:right;" src="https://img.shields.io/badge/-source-cccccc?style=flat-square" /></a>

# <kbd>module</kbd> `python_executor_client.types`
Type definitions for python-executor client.

This module contains the data types used by the PythonExecutorClient:
- ExecutionStatus: Enum for execution states
- ExecutionConfig: Resource limits and settings
- Metadata: Execution parameters
- ExecutionResult: Response from the server





---

<a href="https://github.com/geraldthewes/python-executor/blob/main/python/python/python_executor_client/types.py#L16"><img align="right" style="float:right;" src="https://img.shields.io/badge/-source-cccccc?style=flat-square" /></a>

## <kbd>enum[str]</kbd> `ExecutionStatus`
Status of a code execution.


**Attributes:**

- <b>`PENDING`</b>: Execution is queued but not yet started.
- <b>`RUNNING`</b>: Execution is currently in progress.
- <b>`COMPLETED`</b>: Execution finished successfully (exit code may be non-zero).
- <b>`FAILED`</b>: Execution failed due to an internal error (not a script error).
- <b>`KILLED`</b>: Execution was terminated by the user.


**Example:**

```python
    >>> result = client.get_execution(exec_id)
>>> if result.status == ExecutionStatus.COMPLETED:
...     print(result.stdout)
    ```


### <kbd>symbols</kbd>
- **COMPLETED** = completed
- **FAILED** = failed
- **KILLED** = killed
- **PENDING** = pending
- **RUNNING** = running




---

<a href="https://github.com/geraldthewes/python-executor/blob/main/python/python/python_executor_client/types.py#L38"><img align="right" style="float:right;" src="https://img.shields.io/badge/-source-cccccc?style=flat-square" /></a>

## <kbd>dataclass</kbd> `ExecutionConfig`
Resource limits and execution settings.

Configure resource constraints and network access for code execution.


**Attributes:**

- <b>`timeout_seconds`</b>: Maximum execution time in seconds. Default is 300 (5 min).
- <b>`network_disabled`</b>: If True, the container has no network access. Default is True.
- <b>`memory_mb`</b>: Memory limit in megabytes. Default is 1024 (1 GB).
- <b>`disk_mb`</b>: Disk space limit in megabytes. Default is 2048 (2 GB).
- <b>`cpu_shares`</b>: CPU shares (relative weight). Default is 1024.


**Example:**

```python
    >>> config = ExecutionConfig(
...     timeout_seconds=60,
...     network_disabled=False,  # Allow network for pip install
...     memory_mb=2048
... )
>>> metadata = Metadata(entrypoint="main.py", config=config)
    ```


<a href="https://github.com/geraldthewes/python-executor/blob/main/python/%3Cstring%3E"><img align="right" style="float:right;" src="https://img.shields.io/badge/-source-cccccc?style=flat-square" /></a>

### <kbd>constructor</kbd> `__init__`

```python
ExecutionConfig(
    timeout_seconds: int = 300,
    network_disabled: bool = True,
    memory_mb: int = 1024,
    disk_mb: int = 2048,
    cpu_shares: int = 1024
) → None
```




### <kbd>attributes</kbd>
- ```cpu_shares``` (int)
- ```disk_mb``` (int)
- ```memory_mb``` (int)
- ```network_disabled``` (bool)
- ```timeout_seconds``` (int)



---

<a href="https://github.com/geraldthewes/python-executor/blob/main/python/python/python_executor_client/types.py#L65"><img align="right" style="float:right;" src="https://img.shields.io/badge/-source-cccccc?style=flat-square" /></a>

### <kbd>method</kbd> `to_dict`

```python
to_dict()
```

Convert to dictionary for JSON serialization.



---

<a href="https://github.com/geraldthewes/python-executor/blob/main/python/python/python_executor_client/types.py#L76"><img align="right" style="float:right;" src="https://img.shields.io/badge/-source-cccccc?style=flat-square" /></a>

## <kbd>dataclass</kbd> `Metadata`
Execution metadata specifying how to run the code.


**Attributes:**

- <b>`entrypoint`</b>: The Python file to execute (e.g., "main.py").
- <b>`docker_image`</b>: Docker image to use. Default is "python:3.11-slim".
- <b>`requirements_txt`</b>: Contents of requirements.txt for pip install.
    > [!NOTE] Network must be enabled for package installation.
- <b>`pre_commands`</b>: Shell commands to run before executing Python.
- <b>`stdin`</b>: Data to provide on standard input to the script.
- <b>`config`</b>: Resource limits. See ExecutionConfig.
- <b>`env_vars`</b>: Environment variables as "KEY=value" strings.
- <b>`script_args`</b>: Arguments to pass to the Python script (sys.argv).


**Example:**

```python
    >>> metadata = Metadata(
...     entrypoint="main.py",
...     requirements_txt="requests\nnumpy",
...     env_vars=["API_KEY=secret"],
...     script_args=["--verbose", "input.txt"],
...     config=ExecutionConfig(network_disabled=False)
... )
    ```


<a href="https://github.com/geraldthewes/python-executor/blob/main/python/%3Cstring%3E"><img align="right" style="float:right;" src="https://img.shields.io/badge/-source-cccccc?style=flat-square" /></a>

### <kbd>constructor</kbd> `__init__`

```python
Metadata(
    entrypoint: str,
    docker_image: Optional[str] = None,
    requirements_txt: Optional[str] = None,
    pre_commands: Optional[list[str]] = None,
    stdin: Optional[str] = None,
    config: Optional[ExecutionConfig] = None,
    env_vars: Optional[list[str]] = None,
    script_args: Optional[list[str]] = None
) → None
```




### <kbd>attributes</kbd>
- ```config``` (Optional)
- ```docker_image``` (Optional)
- ```entrypoint``` (str)
- ```env_vars``` (Optional)
- ```pre_commands``` (Optional)
- ```requirements_txt``` (Optional)
- ```script_args``` (Optional)
- ```stdin``` (Optional)



---

<a href="https://github.com/geraldthewes/python-executor/blob/main/python/python/python_executor_client/types.py#L109"><img align="right" style="float:right;" src="https://img.shields.io/badge/-source-cccccc?style=flat-square" /></a>

### <kbd>method</kbd> `to_dict`

```python
to_dict()
```

Convert to dictionary for JSON serialization.



---

<a href="https://github.com/geraldthewes/python-executor/blob/main/python/python/python_executor_client/types.py#L131"><img align="right" style="float:right;" src="https://img.shields.io/badge/-source-cccccc?style=flat-square" /></a>

## <kbd>dataclass</kbd> `ExecutionResult`
Result of a code execution.

Contains the output and status of a completed or in-progress execution.


**Attributes:**

- <b>`execution_id`</b>: Unique identifier for this execution.
- <b>`status`</b>: Current status (pending, running, completed, failed, killed).
- <b>`stdout`</b>: Standard output from the Python script.
- <b>`stderr`</b>: Standard error from the Python script.
- <b>`exit_code`</b>: Process exit code (0 = success, non-zero = error).
- <b>`error`</b>: Error message if the execution failed internally.
- <b>`started_at`</b>: When execution started (UTC).
- <b>`finished_at`</b>: When execution finished (UTC).
- <b>`duration_ms`</b>: Total execution time in milliseconds.
- <b>`result`</b>: REPL expression result when eval_last_expr is enabled.
    Contains the repr() of the last expression's value, or None if the
    last statement was not an expression.


**Example:**

```python
    >>> result = client.execute_sync(
...     files={"main.py": "print('hello')"}
... )
>>> print(result.status)
completed
>>> print(result.exit_code)
0
>>> print(result.stdout)
hello

# REPL-style evaluation:
>>> result = client.eval("x = 5\nx * 2")
>>> print(result.result)
10
    ```


<a href="https://github.com/geraldthewes/python-executor/blob/main/python/%3Cstring%3E"><img align="right" style="float:right;" src="https://img.shields.io/badge/-source-cccccc?style=flat-square" /></a>

### <kbd>constructor</kbd> `__init__`

```python
ExecutionResult(
    execution_id: str,
    status: ExecutionStatus,
    stdout: Optional[str] = None,
    stderr: Optional[str] = None,
    exit_code: Optional[int] = None,
    error: Optional[str] = None,
    started_at: Optional[datetime] = None,
    finished_at: Optional[datetime] = None,
    duration_ms: Optional[int] = None,
    result: Optional[str] = None
) → None
```




### <kbd>attributes</kbd>
- ```duration_ms``` (Optional)
- ```error``` (Optional)
- ```execution_id``` (str)
- ```exit_code``` (Optional)
- ```finished_at``` (Optional)
- ```result``` (Optional)
- ```started_at``` (Optional)
- ```status``` (ExecutionStatus)
- ```stderr``` (Optional)
- ```stdout``` (Optional)



---

<a href="https://github.com/geraldthewes/python-executor/blob/main/python/python/python_executor_client/types.py#L169"><img align="right" style="float:right;" src="https://img.shields.io/badge/-source-cccccc?style=flat-square" /></a>

### <kbd>classmethod</kbd> `from_dict`

```python
from_dict(data: dict) → ExecutionResult
```

Create an ExecutionResult from an API response dictionary.


**Args:**

- <b>`data`</b>: Dictionary from the JSON API response.


**Returns:**

- <b>`ExecutionResult`</b>: Parsed result object.





---

_This file was automatically generated via [lazydocs](https://github.com/ml-tooling/lazydocs)._
