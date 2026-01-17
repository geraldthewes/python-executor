"""Python client for python-executor.

This module provides a Python client library for the python-executor service,
which allows remote execution of Python code in isolated Docker containers.

Example:
    Basic usage with a simple script:

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
"""

import io
import json
import tarfile
import time
from pathlib import Path
from typing import Optional, Union

import requests

from .types import ExecutionResult, Metadata, ExecutionStatus


class PythonExecutorClient:
    """Client for the python-executor remote code execution service.

    This client handles all the complexity of the python-executor API, including:
    - Creating tar archives from files, directories, or dicts
    - Encoding metadata as JSON
    - Managing multipart/form-data requests
    - Parsing execution results

    Attributes:
        base_url: The base URL of the python-executor server.
        timeout: HTTP request timeout in seconds.
        session: The underlying requests Session object.

    Example:
        >>> client = PythonExecutorClient("http://pyexec.cluster:9999/")
        >>> result = client.execute_sync(
        ...     files={"main.py": "print('Hello!')"},
        ...     entrypoint="main.py"
        ... )
        >>> result.exit_code
        0
    """

    def __init__(self, base_url: str, timeout: int = 300):
        """Initialize the Python executor client.

        Args:
            base_url: Base URL of the python-executor server (e.g., "http://pyexec.cluster:9999/").
            timeout: HTTP request timeout in seconds. Default is 300 (5 minutes).

        Example:
            >>> client = PythonExecutorClient("http://pyexec.cluster:9999/")
            >>> client = PythonExecutorClient("http://localhost:8080", timeout=60)
        """
        self.base_url = base_url.rstrip("/")
        self.timeout = timeout
        self.session = requests.Session()

    def execute_sync(
        self,
        files: Optional[Union[dict[str, str], Path, str]] = None,
        tar_data: Optional[bytes] = None,
        metadata: Optional[Metadata] = None,
        **kwargs,
    ) -> ExecutionResult:
        """Execute Python code synchronously and wait for the result.

        This method submits code for execution and blocks until completion.
        Use this for short-running scripts where you need the result immediately.

        Args:
            files: Python files to execute. Can be:
                - dict[str, str]: Mapping of filename to content (e.g., {"main.py": "print('hi')"})
                - Path: Path to a file or directory
                - str: String path to a file or directory
            tar_data: Pre-built tar archive bytes (alternative to files).
                If provided, files parameter is ignored.
            metadata: Full Metadata object for advanced configuration.
                If provided, kwargs are ignored.
            **kwargs: Shorthand for metadata fields:
                - entrypoint (str): Script to run (auto-detected if not specified)
                - docker_image (str): Docker image (default: python:3.11-slim)
                - requirements_txt (str): Contents of requirements.txt
                - pre_commands (list[str]): Shell commands to run before execution
                - stdin (str): Data to provide on stdin
                - timeout_seconds (int): Execution timeout
                - network_disabled (bool): Disable network access
                - memory_mb (int): Memory limit in MB
                - disk_mb (int): Disk limit in MB
                - cpu_shares (int): CPU shares

        Returns:
            ExecutionResult: Object containing stdout, stderr, exit_code, and execution metadata.

        Raises:
            ValueError: If neither files nor tar_data is provided.
            requests.HTTPError: If the server returns an error response.

        Example:
            Simple execution:

            >>> client = PythonExecutorClient("http://pyexec.cluster:9999/")
            >>> result = client.execute_sync(
            ...     files={"main.py": "print('Hello, World!')"},
            ...     entrypoint="main.py"
            ... )
            >>> result.exit_code
            0
            >>> result.stdout
            'Hello, World!\\n'

            Execution with dependencies:

            >>> result = client.execute_sync(
            ...     files={"main.py": "import numpy; print(numpy.__version__)"},
            ...     requirements_txt="numpy",
            ...     network_disabled=False
            ... )

            Execution from a directory:

            >>> result = client.execute_sync(files=Path("./myproject/"))
        """
        tar_bytes, meta = self._prepare_request(files, tar_data, metadata, **kwargs)

        files_data = {
            "tar": ("code.tar", tar_bytes, "application/octet-stream"),
            "metadata": (None, json.dumps(meta.to_dict()), "application/json"),
        }

        response = self.session.post(
            f"{self.base_url}/api/v1/exec/sync",
            files=files_data,
            timeout=self.timeout,
        )
        response.raise_for_status()

        return ExecutionResult.from_dict(response.json())

    def execute_async(
        self,
        files: Optional[Union[dict[str, str], Path, str]] = None,
        tar_data: Optional[bytes] = None,
        metadata: Optional[Metadata] = None,
        **kwargs,
    ) -> str:
        """Submit Python code for asynchronous execution.

        This method submits code and returns immediately with an execution ID.
        Use this for long-running scripts or when you need to run multiple
        scripts concurrently.

        Args:
            files: Python files to execute. Same options as execute_sync().
            tar_data: Pre-built tar archive bytes (alternative to files).
            metadata: Full Metadata object for advanced configuration.
            **kwargs: Shorthand for metadata fields. See execute_sync() for options.

        Returns:
            str: Execution ID that can be used with get_execution() and wait_for_completion().

        Raises:
            ValueError: If neither files nor tar_data is provided.
            requests.HTTPError: If the server returns an error response.

        Example:
            >>> client = PythonExecutorClient("http://pyexec.cluster:9999/")
            >>> exec_id = client.execute_async(
            ...     files={"main.py": "import time; time.sleep(60); print('done')"}
            ... )
            >>> print(f"Submitted: {exec_id}")
            Submitted: exe_550e8400-e29b-41d4-a716-446655440000

            >>> # Later, check status or wait for completion
            >>> result = client.wait_for_completion(exec_id)
        """
        tar_bytes, meta = self._prepare_request(files, tar_data, metadata, **kwargs)

        files_data = {
            "tar": ("code.tar", tar_bytes, "application/octet-stream"),
            "metadata": (None, json.dumps(meta.to_dict()), "application/json"),
        }

        response = self.session.post(
            f"{self.base_url}/api/v1/exec/async",
            files=files_data,
            timeout=self.timeout,
        )
        response.raise_for_status()

        return response.json()["execution_id"]

    def get_execution(self, execution_id: str) -> ExecutionResult:
        """Get the current status and result of an execution.

        Args:
            execution_id: The execution ID returned by execute_async().

        Returns:
            ExecutionResult: Current status and any available output.
                The status field indicates whether the execution is still running.

        Raises:
            requests.HTTPError: If the execution is not found (404) or server error.

        Example:
            >>> client = PythonExecutorClient("http://pyexec.cluster:9999/")
            >>> result = client.get_execution("exe_550e8400-e29b-41d4-a716-446655440000")
            >>> print(result.status)
            running
        """
        response = self.session.get(
            f"{self.base_url}/api/v1/executions/{execution_id}",
            timeout=self.timeout,
        )
        response.raise_for_status()

        return ExecutionResult.from_dict(response.json())

    def kill(self, execution_id: str) -> None:
        """Terminate a running execution.

        Forcefully stops the Docker container running the Python code.

        Args:
            execution_id: The execution ID to kill.

        Raises:
            requests.HTTPError: If the execution is not found (404) or server error.

        Example:
            >>> client = PythonExecutorClient("http://pyexec.cluster:9999/")
            >>> exec_id = client.execute_async(files={"main.py": "import time; time.sleep(3600)"})
            >>> client.kill(exec_id)
        """
        response = self.session.delete(
            f"{self.base_url}/api/v1/executions/{execution_id}",
            timeout=self.timeout,
        )
        response.raise_for_status()

    def wait_for_completion(
        self,
        execution_id: str,
        poll_interval: float = 2.0,
        max_wait: Optional[float] = None,
    ) -> ExecutionResult:
        """Wait for an asynchronous execution to complete.

        Polls the server periodically until the execution finishes.

        Args:
            execution_id: The execution ID returned by execute_async().
            poll_interval: Seconds between status checks. Default is 2.0.
            max_wait: Maximum seconds to wait before raising TimeoutError.
                None means wait indefinitely. Default is None.

        Returns:
            ExecutionResult: Final result with stdout, stderr, and exit_code.

        Raises:
            TimeoutError: If max_wait is exceeded before completion.
            requests.HTTPError: If the server returns an error.

        Example:
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
        """
        start_time = time.time()

        while True:
            result = self.get_execution(execution_id)

            if result.status in (ExecutionStatus.COMPLETED, ExecutionStatus.FAILED, ExecutionStatus.KILLED):
                return result

            if max_wait and (time.time() - start_time) > max_wait:
                raise TimeoutError(f"Execution did not complete within {max_wait}s")

            time.sleep(poll_interval)

    def _prepare_request(
        self,
        files: Optional[Union[dict[str, str], Path, str]],
        tar_data: Optional[bytes],
        metadata: Optional[Metadata],
        **kwargs,
    ) -> tuple[bytes, Metadata]:
        """Prepare tar archive and metadata for an API request.

        Internal method that handles the various input formats and constructs
        the tar archive and Metadata object needed for the API.
        """
        # Create tar if not provided
        if tar_data is None:
            if files is None:
                raise ValueError("Either files or tar_data must be provided")
            tar_data = self._create_tar(files)

        # Create metadata if not provided
        if metadata is None:
            # Detect entrypoint
            entrypoint = kwargs.pop("entrypoint", None)
            if entrypoint is None:
                entrypoint = self._detect_entrypoint(tar_data)

            from .types import ExecutionConfig
            metadata = Metadata(
                entrypoint=entrypoint,
                docker_image=kwargs.pop("docker_image", None),
                requirements_txt=kwargs.pop("requirements_txt", None),
                pre_commands=kwargs.pop("pre_commands", None),
                stdin=kwargs.pop("stdin", None),
                config=ExecutionConfig(**kwargs) if kwargs else None,
            )

        return tar_data, metadata

    def _create_tar(self, files: Union[dict[str, str], Path, str]) -> bytes:
        """Create tar archive from files."""
        buf = io.BytesIO()

        with tarfile.open(fileobj=buf, mode="w") as tar:
            if isinstance(files, dict):
                # Dict of filename -> content
                for filename, content in files.items():
                    info = tarfile.TarInfo(name=filename)
                    content_bytes = content.encode() if isinstance(content, str) else content
                    info.size = len(content_bytes)
                    tar.addfile(info, io.BytesIO(content_bytes))

            elif isinstance(files, (Path, str)):
                path = Path(files)

                if path.is_file():
                    # Single file
                    tar.add(path, arcname=path.name)
                elif path.is_dir():
                    # Directory
                    for file_path in path.rglob("*"):
                        if file_path.is_file():
                            arcname = file_path.relative_to(path)
                            tar.add(file_path, arcname=str(arcname))
                else:
                    raise ValueError(f"Path does not exist: {path}")

        return buf.getvalue()

    def _detect_entrypoint(self, tar_data: bytes) -> str:
        """Detect entrypoint from tar archive."""
        with tarfile.open(fileobj=io.BytesIO(tar_data), mode="r") as tar:
            names = [m.name for m in tar.getmembers() if m.name.endswith(".py")]

            # Priority order
            if "main.py" in names:
                return "main.py"
            if "__main__.py" in names:
                return "__main__.py"
            if names:
                return names[0]

        raise ValueError("No Python files found in archive")
