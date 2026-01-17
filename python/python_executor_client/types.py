"""Type definitions for python-executor client.

This module contains the data types used by the PythonExecutorClient:
- ExecutionStatus: Enum for execution states
- ExecutionConfig: Resource limits and settings
- Metadata: Execution parameters
- ExecutionResult: Response from the server
"""

from dataclasses import dataclass
from datetime import datetime
from enum import Enum
from typing import Optional


class ExecutionStatus(str, Enum):
    """Status of a code execution.

    Attributes:
        PENDING: Execution is queued but not yet started.
        RUNNING: Execution is currently in progress.
        COMPLETED: Execution finished successfully (exit code may be non-zero).
        FAILED: Execution failed due to an internal error (not a script error).
        KILLED: Execution was terminated by the user.

    Example:
        >>> result = client.get_execution(exec_id)
        >>> if result.status == ExecutionStatus.COMPLETED:
        ...     print(result.stdout)
    """
    PENDING = "pending"
    RUNNING = "running"
    COMPLETED = "completed"
    FAILED = "failed"
    KILLED = "killed"


@dataclass
class ExecutionConfig:
    """Resource limits and execution settings.

    Configure resource constraints and network access for code execution.

    Attributes:
        timeout_seconds: Maximum execution time in seconds. Default is 300 (5 min).
        network_disabled: If True, the container has no network access. Default is True.
        memory_mb: Memory limit in megabytes. Default is 1024 (1 GB).
        disk_mb: Disk space limit in megabytes. Default is 2048 (2 GB).
        cpu_shares: CPU shares (relative weight). Default is 1024.

    Example:
        >>> config = ExecutionConfig(
        ...     timeout_seconds=60,
        ...     network_disabled=False,  # Allow network for pip install
        ...     memory_mb=2048
        ... )
        >>> metadata = Metadata(entrypoint="main.py", config=config)
    """
    timeout_seconds: int = 300
    network_disabled: bool = True
    memory_mb: int = 1024
    disk_mb: int = 2048
    cpu_shares: int = 1024

    def to_dict(self):
        """Convert to dictionary for JSON serialization."""
        return {
            "timeout_seconds": self.timeout_seconds,
            "network_disabled": self.network_disabled,
            "memory_mb": self.memory_mb,
            "disk_mb": self.disk_mb,
            "cpu_shares": self.cpu_shares,
        }


@dataclass
class Metadata:
    """Execution metadata specifying how to run the code.

    Attributes:
        entrypoint: The Python file to execute (e.g., "main.py").
        docker_image: Docker image to use. Default is "python:3.11-slim".
        requirements_txt: Contents of requirements.txt for pip install.
            Note: Network must be enabled for package installation.
        pre_commands: Shell commands to run before executing Python.
        stdin: Data to provide on standard input to the script.
        config: Resource limits. See ExecutionConfig.
        env_vars: Environment variables as "KEY=value" strings.
        script_args: Arguments to pass to the Python script (sys.argv).

    Example:
        >>> metadata = Metadata(
        ...     entrypoint="main.py",
        ...     requirements_txt="requests\\nnumpy",
        ...     env_vars=["API_KEY=secret"],
        ...     script_args=["--verbose", "input.txt"],
        ...     config=ExecutionConfig(network_disabled=False)
        ... )
    """
    entrypoint: str
    docker_image: Optional[str] = None
    requirements_txt: Optional[str] = None
    pre_commands: Optional[list[str]] = None
    stdin: Optional[str] = None
    config: Optional[ExecutionConfig] = None
    env_vars: Optional[list[str]] = None
    script_args: Optional[list[str]] = None

    def to_dict(self):
        """Convert to dictionary for JSON serialization."""
        data = {"entrypoint": self.entrypoint}

        if self.docker_image:
            data["docker_image"] = self.docker_image
        if self.requirements_txt:
            data["requirements_txt"] = self.requirements_txt
        if self.pre_commands:
            data["pre_commands"] = self.pre_commands
        if self.stdin:
            data["stdin"] = self.stdin
        if self.config:
            data["config"] = self.config.to_dict()
        if self.env_vars:
            data["env_vars"] = self.env_vars
        if self.script_args:
            data["script_args"] = self.script_args

        return data


@dataclass
class ExecutionResult:
    """Result of a code execution.

    Contains the output and status of a completed or in-progress execution.

    Attributes:
        execution_id: Unique identifier for this execution.
        status: Current status (pending, running, completed, failed, killed).
        stdout: Standard output from the Python script.
        stderr: Standard error from the Python script.
        exit_code: Process exit code (0 = success, non-zero = error).
        error: Error message if the execution failed internally.
        started_at: When execution started (UTC).
        finished_at: When execution finished (UTC).
        duration_ms: Total execution time in milliseconds.
        result: REPL expression result when eval_last_expr is enabled.
            Contains the repr() of the last expression's value, or None
            if the last statement was not an expression.

    Example:
        >>> result = client.execute_sync(
        ...     files={"main.py": "print('hello')"}
        ... )
        >>> print(result.status)
        completed
        >>> print(result.exit_code)
        0
        >>> print(result.stdout)
        hello

        REPL-style evaluation:
        >>> result = client.eval("x = 5\\nx * 2")
        >>> print(result.result)
        10
    """
    execution_id: str
    status: ExecutionStatus
    stdout: Optional[str] = None
    stderr: Optional[str] = None
    exit_code: Optional[int] = None
    error: Optional[str] = None
    started_at: Optional[datetime] = None
    finished_at: Optional[datetime] = None
    duration_ms: Optional[int] = None
    result: Optional[str] = None

    @classmethod
    def from_dict(cls, data: dict) -> "ExecutionResult":
        """Create an ExecutionResult from an API response dictionary.

        Args:
            data: Dictionary from the JSON API response.

        Returns:
            ExecutionResult: Parsed result object.
        """
        return cls(
            execution_id=data["execution_id"],
            status=ExecutionStatus(data["status"]),
            stdout=data.get("stdout"),
            stderr=data.get("stderr"),
            exit_code=data.get("exit_code"),
            error=data.get("error"),
            started_at=datetime.fromisoformat(data["started_at"].rstrip("Z")) if data.get("started_at") else None,
            finished_at=datetime.fromisoformat(data["finished_at"].rstrip("Z")) if data.get("finished_at") else None,
            duration_ms=data.get("duration_ms"),
            result=data.get("result"),
        )
