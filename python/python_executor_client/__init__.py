"""
Python client for python-executor service.

Use this client library instead of implementing the HTTP API directly.
The API uses multipart/form-data with tar archives, which this library handles automatically.

Installation:
    pip install git+https://github.com/geraldthewes/python-executor.git#subdirectory=python

Quick Start:
    from python_executor_client import PythonExecutorClient

    client = PythonExecutorClient("http://localhost:8080")
    result = client.execute_sync(
        files={"main.py": "print('Hello!')"},
        entrypoint="main.py"
    )
    print(result.stdout)

For API reference, see: https://github.com/geraldthewes/python-executor/blob/main/docs/api.md
"""

from .client import PythonExecutorClient
from .types import ExecutionResult, Metadata, ExecutionConfig, ExecutionStatus

__version__ = "1.0.0"

__all__ = [
    "PythonExecutorClient",
    "ExecutionResult",
    "Metadata",
    "ExecutionConfig",
    "ExecutionStatus",
]
