"""Python client for python-executor."""

import io
import json
import tarfile
import time
from pathlib import Path
from typing import Optional, Union

import requests

from .types import ExecutionResult, Metadata, ExecutionStatus


class PythonExecutorClient:
    """Client for python-executor service."""

    def __init__(self, base_url: str, timeout: int = 300):
        """Initialize client.

        Args:
            base_url: Base URL of the server
            timeout: HTTP request timeout in seconds
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
        """Execute code synchronously.

        Args:
            files: Dict of filename->content, Path to directory, or path string
            tar_data: Raw tar archive bytes (alternative to files)
            metadata: Execution metadata
            **kwargs: Additional metadata fields (entrypoint, docker_image, etc.)

        Returns:
            ExecutionResult with stdout, stderr, exit code
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
        """Execute code asynchronously.

        Args:
            files: Dict of filename->content, Path to directory, or path string
            tar_data: Raw tar archive bytes (alternative to files)
            metadata: Execution metadata
            **kwargs: Additional metadata fields

        Returns:
            Execution ID
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
        """Get execution status and result.

        Args:
            execution_id: Execution ID

        Returns:
            ExecutionResult
        """
        response = self.session.get(
            f"{self.base_url}/api/v1/executions/{execution_id}",
            timeout=self.timeout,
        )
        response.raise_for_status()

        return ExecutionResult.from_dict(response.json())

    def kill(self, execution_id: str) -> None:
        """Kill a running execution.

        Args:
            execution_id: Execution ID
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
        """Wait for async execution to complete.

        Args:
            execution_id: Execution ID
            poll_interval: Seconds between polls
            max_wait: Maximum seconds to wait (None = no limit)

        Returns:
            ExecutionResult when complete
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
        """Prepare tar and metadata for request."""
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
