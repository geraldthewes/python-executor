"""Type definitions for python-executor client."""

from dataclasses import dataclass
from datetime import datetime
from enum import Enum
from typing import Optional


class ExecutionStatus(str, Enum):
    """Execution status."""
    PENDING = "pending"
    RUNNING = "running"
    COMPLETED = "completed"
    FAILED = "failed"
    KILLED = "killed"


@dataclass
class ExecutionConfig:
    """Execution configuration."""
    timeout_seconds: int = 300
    network_disabled: bool = True
    memory_mb: int = 1024
    disk_mb: int = 2048
    cpu_shares: int = 1024

    def to_dict(self):
        return {
            "timeout_seconds": self.timeout_seconds,
            "network_disabled": self.network_disabled,
            "memory_mb": self.memory_mb,
            "disk_mb": self.disk_mb,
            "cpu_shares": self.cpu_shares,
        }


@dataclass
class Metadata:
    """Execution metadata."""
    entrypoint: str
    docker_image: Optional[str] = None
    requirements_txt: Optional[str] = None
    pre_commands: Optional[list[str]] = None
    stdin: Optional[str] = None
    config: Optional[ExecutionConfig] = None
    env_vars: Optional[list[str]] = None
    script_args: Optional[list[str]] = None

    def to_dict(self):
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
    """Execution result."""
    execution_id: str
    status: ExecutionStatus
    stdout: Optional[str] = None
    stderr: Optional[str] = None
    exit_code: Optional[int] = None
    error: Optional[str] = None
    started_at: Optional[datetime] = None
    finished_at: Optional[datetime] = None
    duration_ms: Optional[int] = None

    @classmethod
    def from_dict(cls, data: dict):
        """Create from API response."""
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
        )
