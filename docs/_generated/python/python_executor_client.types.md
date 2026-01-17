<!-- markdownlint-disable -->

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
    duration_ms: Optional[int] = None
) → None
```




### <kbd>attributes</kbd>
- ```duration_ms``` (Optional)
- ```error``` (Optional)
- ```execution_id``` (str)
- ```exit_code``` (Optional)
- ```finished_at``` (Optional)
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
