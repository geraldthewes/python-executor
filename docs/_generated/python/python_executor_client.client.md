<!-- markdownlint-disable -->

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

_This file was automatically generated via [lazydocs](https://github.com/ml-tooling/lazydocs)._
