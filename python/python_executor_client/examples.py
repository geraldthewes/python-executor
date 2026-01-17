#!/usr/bin/env python3
"""Examples demonstrating the Python client for python-executor.

These examples require a running python-executor server.
Set PYEXEC_SERVER environment variable or use default http://pyexec.cluster:9999/

Run examples:
    python -m python_executor_client.examples
"""

import os

from .client import PythonExecutorClient
from .types import ExecutionStatus


def get_server_url() -> str:
    """Get server URL from environment or use default."""
    return os.environ.get("PYEXEC_SERVER", "http://pyexec.cluster:9999/")


def example_eval_simple():
    """Simple REPL-style evaluation of an expression."""
    print("=== Example: Simple Eval ===")
    client = PythonExecutorClient(get_server_url())

    result = client.eval("2 + 2")

    print(f"Code: 2 + 2")
    print(f"Exit code: {result.exit_code}")
    print(f"Result: {result.result}")
    print()


def example_eval_multiline():
    """Multi-line code with variable assignment and expression."""
    print("=== Example: Multi-line Eval ===")
    client = PythonExecutorClient(get_server_url())

    code = """x = 10
y = 5
x * y"""

    result = client.eval(code)

    print(f"Code:\n{code}")
    print(f"Exit code: {result.exit_code}")
    print(f"Result: {result.result}")
    print()


def example_eval_with_import():
    """Using imports with REPL evaluation."""
    print("=== Example: Eval with Import ===")
    client = PythonExecutorClient(get_server_url())

    code = "import math\nmath.sqrt(16)"

    result = client.eval(code)

    print(f"Code: {code}")
    print(f"Exit code: {result.exit_code}")
    print(f"Result: {result.result}")
    print()


def example_eval_with_print():
    """Demonstrating that print produces stdout, not a result."""
    print("=== Example: Eval with Print ===")
    client = PythonExecutorClient(get_server_url())

    code = 'print("Hello from eval!")'

    result = client.eval(code)

    print(f"Code: {code}")
    print(f"Exit code: {result.exit_code}")
    print(f"Stdout: {result.stdout!r}")
    print(f"Result: {result.result}")  # None because print is not an expression
    print()


def example_eval_list_comprehension():
    """Evaluating a list comprehension."""
    print("=== Example: List Comprehension ===")
    client = PythonExecutorClient(get_server_url())

    code = "[x**2 for x in range(5)]"

    result = client.eval(code)

    print(f"Code: {code}")
    print(f"Exit code: {result.exit_code}")
    print(f"Result: {result.result}")
    print()


def example_eval_with_python_version():
    """Specifying a Python version."""
    print("=== Example: Eval with Python Version ===")
    client = PythonExecutorClient(get_server_url())

    code = "import sys; sys.version_info[:2]"

    result = client.eval(code, python_version="3.11")

    print(f"Code: {code}")
    print(f"Python version requested: 3.11")
    print(f"Exit code: {result.exit_code}")
    print(f"Result: {result.result}")
    print()


def example_execute_sync():
    """Traditional synchronous execution (for comparison)."""
    print("=== Example: Execute Sync (for comparison) ===")
    client = PythonExecutorClient(get_server_url())

    result = client.execute_sync(
        files={"main.py": "print('Hello, World!')"},
        entrypoint="main.py"
    )

    print(f"Exit code: {result.exit_code}")
    print(f"Stdout: {result.stdout!r}")
    print(f"Result: {result.result}")  # None because execute_sync doesn't use eval
    print()


def main():
    """Run all examples."""
    print(f"Server URL: {get_server_url()}\n")

    examples = [
        example_eval_simple,
        example_eval_multiline,
        example_eval_with_import,
        example_eval_with_print,
        example_eval_list_comprehension,
        example_eval_with_python_version,
        example_execute_sync,
    ]

    for example in examples:
        try:
            example()
        except Exception as e:
            print(f"Error running {example.__name__}: {e}\n")


if __name__ == "__main__":
    main()
