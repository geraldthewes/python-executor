#!/usr/bin/env python3
"""Performance benchmark for python-executor service.

This script measures performance across various features of the python-executor
service and saves results for future reference.

Usage:
    python scripts/benchmark.py [options]

Options:
    --server URL          Server URL (default: http://pyexec.cluster:9999/)
    --categories CATS     Comma-separated list of categories to run
    --iterations N        Number of iterations per test (default: 3)
    --output-json FILE    Save results to JSON file
    --output-markdown FILE Save results to Markdown file
    --quick               Quick mode (1 iteration, skip slow tests)
    --list                List available tests and exit
"""

import argparse
import json
import logging
import statistics
import sys
import time
from concurrent.futures import ThreadPoolExecutor, as_completed
from dataclasses import dataclass, field, asdict
from datetime import datetime, timezone
from pathlib import Path
from typing import Callable, Optional

import requests

# Add the python client to the path
sys.path.insert(0, str(Path(__file__).parent.parent / "python"))

from python_executor_client import PythonExecutorClient
from python_executor_client.types import ExecutionResult, ExecutionStatus

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s - %(levelname)s - %(message)s"
)
logger = logging.getLogger(__name__)

SERVER_URL = "http://pyexec.cluster:9999/"


@dataclass
class TestResult:
    """Result of a single test iteration."""
    success: bool
    client_latency_ms: float
    server_duration_ms: Optional[float] = None
    overhead_ms: Optional[float] = None
    error: Optional[str] = None


@dataclass
class BenchmarkResult:
    """Aggregated result for a benchmark test."""
    name: str
    category: str
    description: str
    iterations: int
    successes: int
    failures: int
    latency_min_ms: Optional[float] = None
    latency_max_ms: Optional[float] = None
    latency_mean_ms: Optional[float] = None
    latency_stddev_ms: Optional[float] = None
    server_duration_mean_ms: Optional[float] = None
    overhead_mean_ms: Optional[float] = None
    errors: list[str] = field(default_factory=list)


@dataclass
class BenchmarkSuite:
    """Complete benchmark run results."""
    server_url: str
    started_at: str
    finished_at: str
    total_duration_seconds: float
    iterations_per_test: int
    results: list[BenchmarkResult]


class Benchmark:
    """Individual benchmark test definition."""

    def __init__(
        self,
        name: str,
        category: str,
        description: str,
        func: Callable[[], TestResult],
        slow: bool = False,
    ):
        self.name = name
        self.category = category
        self.description = description
        self.func = func
        self.slow = slow


class BenchmarkRunner:
    """Runs benchmark tests against the python-executor service."""

    def __init__(self, server_url: str, iterations: int = 3, skip_slow: bool = False):
        self.server_url = server_url.rstrip("/")
        self.iterations = iterations
        self.skip_slow = skip_slow
        self.client = PythonExecutorClient(server_url, timeout=300)
        self.benchmarks: list[Benchmark] = []
        self._register_benchmarks()

    def _register_benchmarks(self):
        """Register all benchmark tests."""
        # A. Basic Execution
        self._register("basic_print", "basic", "Simple print statement",
                       self._test_basic_print)
        self._register("basic_computation", "basic", "CPU-bound Fibonacci",
                       self._test_basic_computation)
        self._register("basic_memory", "basic", "Allocate 1M integers",
                       self._test_basic_memory)
        self._register("basic_empty", "basic", "Empty pass (overhead)",
                       self._test_basic_empty)

        # B. Sync vs Async
        self._register("sync_simple", "sync_async", "Sync execution",
                       self._test_sync_simple)
        self._register("async_simple", "sync_async", "Async execution (2s poll)",
                       self._test_async_simple)
        self._register("async_fast_poll", "sync_async", "Async (0.5s poll)",
                       self._test_async_fast_poll)
        self._register("async_slow_poll", "sync_async", "Async (5s poll)",
                       self._test_async_slow_poll, slow=True)

        # C. Multi-file Projects
        self._register("multifile_two", "multifile", "2 files with import",
                       self._test_multifile_two)
        self._register("multifile_five", "multifile", "5 interdependent files",
                       self._test_multifile_five)
        self._register("multifile_ten", "multifile", "10 files project",
                       self._test_multifile_ten)

        # D. Dependencies (requires network)
        self._register("deps_none", "dependencies", "No dependencies",
                       self._test_deps_none)
        self._register("deps_small", "dependencies", "Small package (colorama)",
                       self._test_deps_small, slow=True)
        self._register("deps_large", "dependencies", "Large package (numpy)",
                       self._test_deps_large, slow=True)

        # E. Python Versions
        self._register("version_310", "versions", "Python 3.10",
                       self._test_version_310)
        self._register("version_311", "versions", "Python 3.11 (default)",
                       self._test_version_311)
        self._register("version_312", "versions", "Python 3.12",
                       self._test_version_312)

        # F. Resource Constraints
        self._register("resource_default", "resources", "Default settings",
                       self._test_resource_default)
        self._register("resource_low_mem", "resources", "256 MB memory",
                       self._test_resource_low_mem)
        self._register("resource_high_mem", "resources", "2048 MB memory",
                       self._test_resource_high_mem)

        # G. stdin/stdout
        self._register("io_no_stdin", "io", "No input",
                       self._test_io_no_stdin)
        self._register("io_small_stdin", "io", "100 bytes input",
                       self._test_io_small_stdin, slow=True)
        self._register("io_large_stdout", "io", "100KB output",
                       self._test_io_large_stdout)

        # H. Concurrent Executions
        self._register("concurrent_2", "concurrent", "2 parallel executions",
                       self._test_concurrent_2)
        self._register("concurrent_5", "concurrent", "5 parallel executions",
                       self._test_concurrent_5)
        self._register("concurrent_10", "concurrent", "10 parallel executions",
                       self._test_concurrent_10, slow=True)

        # I. REPL Eval
        self._register("eval_simple", "eval", "Simple expression 2+2",
                       self._test_eval_simple)
        self._register("eval_multiline", "eval", "Multi-line with expression",
                       self._test_eval_multiline)
        self._register("eval_import", "eval", "Import then expression",
                       self._test_eval_import)
        self._register("eval_no_result", "eval", "Print (returns null)",
                       self._test_eval_no_result)
        self._register("eval_complex", "eval", "List comprehension",
                       self._test_eval_complex)
        self._register("eval_vs_exec", "eval", "Compare eval vs exec/sync",
                       self._test_eval_vs_exec)

    def _register(self, name: str, category: str, description: str,
                  func: Callable[["BenchmarkRunner"], TestResult], slow: bool = False):
        """Register a benchmark test."""
        self.benchmarks.append(Benchmark(name, category, description, func, slow))

    def health_check(self) -> bool:
        """Check if server is healthy."""
        try:
            resp = requests.get(f"{self.server_url}/health", timeout=5)
            resp.raise_for_status()
            return True
        except Exception as e:
            logger.error(f"Health check failed: {e}")
            return False

    def list_tests(self) -> list[dict]:
        """List all available tests."""
        tests = []
        for b in self.benchmarks:
            tests.append({
                "name": b.name,
                "category": b.category,
                "description": b.description,
                "slow": b.slow,
            })
        return tests

    def get_categories(self) -> list[str]:
        """Get list of unique categories."""
        return sorted(set(b.category for b in self.benchmarks))

    def run(self, categories: Optional[list[str]] = None) -> BenchmarkSuite:
        """Run all benchmarks and return results."""
        started_at = datetime.now(timezone.utc).isoformat() + "Z"
        start_time = time.perf_counter()

        results = []
        benchmarks_to_run = self.benchmarks

        if categories:
            benchmarks_to_run = [b for b in self.benchmarks if b.category in categories]

        if self.skip_slow:
            benchmarks_to_run = [b for b in benchmarks_to_run if not b.slow]

        total = len(benchmarks_to_run)
        for i, benchmark in enumerate(benchmarks_to_run, 1):
            logger.info(f"[{i}/{total}] Running {benchmark.name}: {benchmark.description}")
            result = self._run_benchmark(benchmark)
            results.append(result)
            self._print_result(result)

        finished_at = datetime.now(timezone.utc).isoformat() + "Z"
        total_duration = time.perf_counter() - start_time

        return BenchmarkSuite(
            server_url=self.server_url,
            started_at=started_at,
            finished_at=finished_at,
            total_duration_seconds=round(total_duration, 2),
            iterations_per_test=self.iterations,
            results=results,
        )

    def _run_benchmark(self, benchmark: Benchmark) -> BenchmarkResult:
        """Run a single benchmark multiple times and aggregate results."""
        test_results: list[TestResult] = []

        for i in range(self.iterations):
            try:
                result = benchmark.func()
                test_results.append(result)
            except Exception as e:
                logger.error(f"  Iteration {i+1} error: {e}")
                test_results.append(TestResult(
                    success=False,
                    client_latency_ms=0,
                    error=str(e)
                ))

        return self._aggregate_results(benchmark, test_results)

    def _aggregate_results(self, benchmark: Benchmark, results: list[TestResult]) -> BenchmarkResult:
        """Aggregate multiple test results into a benchmark result."""
        successes = sum(1 for r in results if r.success)
        failures = len(results) - successes
        errors = [r.error for r in results if r.error]

        latencies = [r.client_latency_ms for r in results if r.success]
        server_durations = [r.server_duration_ms for r in results if r.success and r.server_duration_ms]
        overheads = [r.overhead_ms for r in results if r.success and r.overhead_ms]

        result = BenchmarkResult(
            name=benchmark.name,
            category=benchmark.category,
            description=benchmark.description,
            iterations=len(results),
            successes=successes,
            failures=failures,
            errors=errors[:3],  # Keep first 3 errors
        )

        if latencies:
            result.latency_min_ms = round(min(latencies), 2)
            result.latency_max_ms = round(max(latencies), 2)
            result.latency_mean_ms = round(statistics.mean(latencies), 2)
            if len(latencies) >= 2:
                result.latency_stddev_ms = round(statistics.stdev(latencies), 2)

        if server_durations:
            result.server_duration_mean_ms = round(statistics.mean(server_durations), 2)

        if overheads:
            result.overhead_mean_ms = round(statistics.mean(overheads), 2)

        return result

    def _print_result(self, result: BenchmarkResult):
        """Print a benchmark result."""
        status = "PASS" if result.failures == 0 else f"FAIL ({result.failures}/{result.iterations})"
        latency = f"{result.latency_mean_ms:.0f}ms" if result.latency_mean_ms else "N/A"
        server = f"{result.server_duration_mean_ms:.0f}ms" if result.server_duration_mean_ms else "N/A"
        overhead = f"{result.overhead_mean_ms:.0f}ms" if result.overhead_mean_ms else "N/A"
        logger.info(f"  {status} - latency: {latency}, server: {server}, overhead: {overhead}")

    def _execute_and_measure(
        self,
        files: dict[str, str],
        **kwargs
    ) -> TestResult:
        """Execute code and measure timing."""
        start = time.perf_counter()
        try:
            result = self.client.execute_sync(files=files, **kwargs)
            elapsed_ms = (time.perf_counter() - start) * 1000

            # Success if completed and exit_code is 0 or None (some executions don't return exit code)
            success = result.status == ExecutionStatus.COMPLETED and (result.exit_code is None or result.exit_code == 0)
            server_duration_ms = result.duration_ms
            overhead_ms = None
            if server_duration_ms is not None:
                overhead_ms = elapsed_ms - server_duration_ms

            return TestResult(
                success=success,
                client_latency_ms=round(elapsed_ms, 2),
                server_duration_ms=server_duration_ms,
                overhead_ms=round(overhead_ms, 2) if overhead_ms else None,
                error=result.stderr if not success else None,
            )
        except Exception as e:
            elapsed_ms = (time.perf_counter() - start) * 1000
            return TestResult(
                success=False,
                client_latency_ms=round(elapsed_ms, 2),
                error=str(e),
            )

    def _execute_async_and_measure(
        self,
        files: dict[str, str],
        poll_interval: float = 2.0,
        **kwargs
    ) -> TestResult:
        """Execute code asynchronously and measure timing."""
        start = time.perf_counter()
        try:
            exec_id = self.client.execute_async(files=files, **kwargs)
            result = self.client.wait_for_completion(exec_id, poll_interval=poll_interval)
            elapsed_ms = (time.perf_counter() - start) * 1000

            success = result.status == ExecutionStatus.COMPLETED and (result.exit_code is None or result.exit_code == 0)
            server_duration_ms = result.duration_ms
            overhead_ms = None
            if server_duration_ms is not None:
                overhead_ms = elapsed_ms - server_duration_ms

            return TestResult(
                success=success,
                client_latency_ms=round(elapsed_ms, 2),
                server_duration_ms=server_duration_ms,
                overhead_ms=round(overhead_ms, 2) if overhead_ms else None,
                error=result.stderr if not success else None,
            )
        except Exception as e:
            elapsed_ms = (time.perf_counter() - start) * 1000
            return TestResult(
                success=False,
                client_latency_ms=round(elapsed_ms, 2),
                error=str(e),
            )

    def _execute_eval_and_measure(self, code: str) -> TestResult:
        """Execute code via /api/v1/eval and measure timing."""
        start = time.perf_counter()
        try:
            resp = requests.post(
                f"{self.server_url}/api/v1/eval",
                json={"code": code, "eval_last_expr": True},
                timeout=60,
            )
            elapsed_ms = (time.perf_counter() - start) * 1000
            resp.raise_for_status()

            data = resp.json()
            # Eval endpoint returns status: "completed" on success
            success = data.get("status") == "completed"
            server_duration_ms = data.get("duration_ms")
            overhead_ms = None
            if server_duration_ms is not None:
                overhead_ms = elapsed_ms - server_duration_ms

            return TestResult(
                success=success,
                client_latency_ms=round(elapsed_ms, 2),
                server_duration_ms=server_duration_ms,
                overhead_ms=round(overhead_ms, 2) if overhead_ms else None,
                error=data.get("stderr") if not success else None,
            )
        except Exception as e:
            elapsed_ms = (time.perf_counter() - start) * 1000
            return TestResult(
                success=False,
                client_latency_ms=round(elapsed_ms, 2),
                error=str(e),
            )

    # ===== A. Basic Execution Tests =====

    def _test_basic_print(self) -> TestResult:
        return self._execute_and_measure(
            files={"main.py": 'print("hello")'},
            entrypoint="main.py",
        )

    def _test_basic_computation(self) -> TestResult:
        code = """
def fib(n):
    if n <= 1:
        return n
    return fib(n-1) + fib(n-2)

result = fib(30)
print(f"fib(30) = {result}")
"""
        return self._execute_and_measure(
            files={"main.py": code},
            entrypoint="main.py",
        )

    def _test_basic_memory(self) -> TestResult:
        code = """
data = list(range(1_000_000))
print(f"Allocated {len(data)} integers")
"""
        return self._execute_and_measure(
            files={"main.py": code},
            entrypoint="main.py",
        )

    def _test_basic_empty(self) -> TestResult:
        return self._execute_and_measure(
            files={"main.py": "pass"},
            entrypoint="main.py",
        )

    # ===== B. Sync vs Async Tests =====

    def _test_sync_simple(self) -> TestResult:
        return self._execute_and_measure(
            files={"main.py": 'print("sync test")'},
            entrypoint="main.py",
        )

    def _test_async_simple(self) -> TestResult:
        return self._execute_async_and_measure(
            files={"main.py": 'print("async test")'},
            entrypoint="main.py",
            poll_interval=2.0,
        )

    def _test_async_fast_poll(self) -> TestResult:
        return self._execute_async_and_measure(
            files={"main.py": 'print("fast poll test")'},
            entrypoint="main.py",
            poll_interval=0.5,
        )

    def _test_async_slow_poll(self) -> TestResult:
        return self._execute_async_and_measure(
            files={"main.py": 'print("slow poll test")'},
            entrypoint="main.py",
            poll_interval=5.0,
        )

    # ===== C. Multi-file Tests =====

    def _test_multifile_two(self) -> TestResult:
        files = {
            "main.py": "from helper import greet\ngreet()",
            "helper.py": "def greet():\n    print('Hello from helper')",
        }
        return self._execute_and_measure(files=files, entrypoint="main.py")

    def _test_multifile_five(self) -> TestResult:
        files = {
            "main.py": """
from utils import add, multiply
from config import VALUE
from data import get_data
from output import format_output

result = multiply(add(VALUE, 5), 2)
data = get_data()
print(format_output(result, data))
""",
            "utils.py": """
def add(a, b):
    return a + b

def multiply(a, b):
    return a * b
""",
            "config.py": "VALUE = 10",
            "data.py": "def get_data():\n    return {'items': [1, 2, 3]}",
            "output.py": "def format_output(result, data):\n    return f'Result: {result}, Data: {data}'",
        }
        return self._execute_and_measure(files=files, entrypoint="main.py")

    def _test_multifile_ten(self) -> TestResult:
        files = {"main.py": "results = []\n"}
        for i in range(1, 10):
            files[f"module{i}.py"] = f"def func{i}():\n    return 'module{i}'"
            files["main.py"] += f"from module{i} import func{i}\nresults.append(func{i}())\n"
        files["main.py"] += "print(f'Loaded {len(results)} modules')"
        return self._execute_and_measure(files=files, entrypoint="main.py")

    # ===== D. Dependency Tests =====

    def _test_deps_none(self) -> TestResult:
        return self._execute_and_measure(
            files={"main.py": 'print("no deps")'},
            entrypoint="main.py",
        )

    def _test_deps_small(self) -> TestResult:
        code = """
from colorama import Fore
print(f"{Fore.GREEN}Success{Fore.RESET}")
"""
        return self._execute_and_measure(
            files={"main.py": code},
            entrypoint="main.py",
            requirements_txt="colorama",
            network_disabled=False,
        )

    def _test_deps_large(self) -> TestResult:
        code = """
import numpy as np
arr = np.array([1, 2, 3, 4, 5])
print(f"NumPy array: {arr}, mean: {arr.mean()}")
"""
        return self._execute_and_measure(
            files={"main.py": code},
            entrypoint="main.py",
            requirements_txt="numpy",
            network_disabled=False,
        )

    # ===== E. Python Version Tests =====

    def _test_version_310(self) -> TestResult:
        code = "import sys\nprint(f'Python {sys.version}')"
        return self._execute_and_measure(
            files={"main.py": code},
            entrypoint="main.py",
            docker_image="python:3.10-slim",
        )

    def _test_version_311(self) -> TestResult:
        code = "import sys\nprint(f'Python {sys.version}')"
        return self._execute_and_measure(
            files={"main.py": code},
            entrypoint="main.py",
            docker_image="python:3.11-slim",
        )

    def _test_version_312(self) -> TestResult:
        code = "import sys\nprint(f'Python {sys.version}')"
        return self._execute_and_measure(
            files={"main.py": code},
            entrypoint="main.py",
            docker_image="python:3.12-slim",
        )

    # ===== F. Resource Constraint Tests =====

    def _test_resource_default(self) -> TestResult:
        return self._execute_and_measure(
            files={"main.py": 'print("default resources")'},
            entrypoint="main.py",
        )

    def _test_resource_low_mem(self) -> TestResult:
        return self._execute_and_measure(
            files={"main.py": 'print("low memory")'},
            entrypoint="main.py",
            memory_mb=256,
        )

    def _test_resource_high_mem(self) -> TestResult:
        return self._execute_and_measure(
            files={"main.py": 'print("high memory")'},
            entrypoint="main.py",
            memory_mb=2048,
        )

    # ===== G. stdin/stdout Tests =====

    def _test_io_no_stdin(self) -> TestResult:
        return self._execute_and_measure(
            files={"main.py": 'print("no stdin")'},
            entrypoint="main.py",
        )

    def _test_io_small_stdin(self) -> TestResult:
        code = """
import sys
data = sys.stdin.read()
print(f"Received {len(data)} bytes")
"""
        return self._execute_and_measure(
            files={"main.py": code},
            entrypoint="main.py",
            stdin="x" * 100,
        )

    def _test_io_large_stdout(self) -> TestResult:
        code = """
output = "x" * 100_000
print(output)
"""
        return self._execute_and_measure(
            files={"main.py": code},
            entrypoint="main.py",
        )

    # ===== H. Concurrent Tests =====

    def _run_concurrent(self, count: int) -> TestResult:
        """Run multiple async executions concurrently."""
        start = time.perf_counter()
        try:
            code = f'print("concurrent test {count}")'
            exec_ids = []

            # Submit all executions
            for i in range(count):
                exec_id = self.client.execute_async(
                    files={"main.py": code},
                    entrypoint="main.py",
                )
                exec_ids.append(exec_id)

            # Wait for all to complete using threads
            results = []
            with ThreadPoolExecutor(max_workers=count) as executor:
                futures = {
                    executor.submit(self.client.wait_for_completion, eid, 1.0): eid
                    for eid in exec_ids
                }
                for future in as_completed(futures):
                    results.append(future.result())

            elapsed_ms = (time.perf_counter() - start) * 1000
            all_success = all(
                r.status == ExecutionStatus.COMPLETED and (r.exit_code is None or r.exit_code == 0)
                for r in results
            )
            server_durations = [r.duration_ms for r in results if r.duration_ms]
            avg_server = statistics.mean(server_durations) if server_durations else None

            return TestResult(
                success=all_success,
                client_latency_ms=round(elapsed_ms, 2),
                server_duration_ms=round(avg_server, 2) if avg_server else None,
                overhead_ms=round(elapsed_ms - avg_server, 2) if avg_server else None,
            )
        except Exception as e:
            elapsed_ms = (time.perf_counter() - start) * 1000
            return TestResult(
                success=False,
                client_latency_ms=round(elapsed_ms, 2),
                error=str(e),
            )

    def _test_concurrent_2(self) -> TestResult:
        return self._run_concurrent(2)

    def _test_concurrent_5(self) -> TestResult:
        return self._run_concurrent(5)

    def _test_concurrent_10(self) -> TestResult:
        return self._run_concurrent(10)

    # ===== I. REPL Eval Tests =====

    def _test_eval_simple(self) -> TestResult:
        return self._execute_eval_and_measure("2 + 2")

    def _test_eval_multiline(self) -> TestResult:
        return self._execute_eval_and_measure("x = 5\ny = 10\nx + y")

    def _test_eval_import(self) -> TestResult:
        return self._execute_eval_and_measure("import math\nmath.sqrt(16)")

    def _test_eval_no_result(self) -> TestResult:
        return self._execute_eval_and_measure('print("hello")')

    def _test_eval_complex(self) -> TestResult:
        return self._execute_eval_and_measure("[x**2 for x in range(5)]")

    def _test_eval_vs_exec(self) -> TestResult:
        """Compare eval endpoint vs exec/sync for same code."""
        code = "2 + 2"

        # Measure eval
        eval_result = self._execute_eval_and_measure(code)

        # Measure exec/sync with equivalent code
        exec_result = self._execute_and_measure(
            files={"main.py": f"result = {code}\nprint(result)"},
            entrypoint="main.py",
        )

        # Return combined result (using eval latency but comparing)
        return TestResult(
            success=eval_result.success and exec_result.success,
            client_latency_ms=eval_result.client_latency_ms,
            server_duration_ms=eval_result.server_duration_ms,
            overhead_ms=eval_result.overhead_ms,
            error=f"eval: {eval_result.client_latency_ms}ms, exec: {exec_result.client_latency_ms}ms" if eval_result.success and exec_result.success else eval_result.error or exec_result.error,
        )


def results_to_json(suite: BenchmarkSuite) -> dict:
    """Convert benchmark suite to JSON-serializable dict."""
    return {
        "server_url": suite.server_url,
        "started_at": suite.started_at,
        "finished_at": suite.finished_at,
        "total_duration_seconds": suite.total_duration_seconds,
        "iterations_per_test": suite.iterations_per_test,
        "results": [asdict(r) for r in suite.results],
        "summary": {
            "total_tests": len(suite.results),
            "passed": sum(1 for r in suite.results if r.failures == 0),
            "failed": sum(1 for r in suite.results if r.failures > 0),
        },
    }


def results_to_markdown(suite: BenchmarkSuite) -> str:
    """Convert benchmark suite to Markdown report."""
    lines = [
        "# Python Executor Benchmark Results",
        "",
        f"**Server:** {suite.server_url}",
        f"**Started:** {suite.started_at}",
        f"**Duration:** {suite.total_duration_seconds}s",
        f"**Iterations:** {suite.iterations_per_test} per test",
        "",
        "## Summary",
        "",
        f"- Total tests: {len(suite.results)}",
        f"- Passed: {sum(1 for r in suite.results if r.failures == 0)}",
        f"- Failed: {sum(1 for r in suite.results if r.failures > 0)}",
        "",
        "## Results by Category",
        "",
    ]

    # Group by category
    categories = {}
    for r in suite.results:
        if r.category not in categories:
            categories[r.category] = []
        categories[r.category].append(r)

    for category, results in categories.items():
        lines.append(f"### {category.replace('_', ' ').title()}")
        lines.append("")
        lines.append("| Test | Status | Latency (ms) | Server (ms) | Overhead (ms) |")
        lines.append("|------|--------|-------------|-------------|---------------|")

        for r in results:
            status = "PASS" if r.failures == 0 else f"FAIL ({r.failures}/{r.iterations})"
            latency = f"{r.latency_mean_ms:.0f}" if r.latency_mean_ms else "N/A"
            server = f"{r.server_duration_mean_ms:.0f}" if r.server_duration_mean_ms else "N/A"
            overhead = f"{r.overhead_mean_ms:.0f}" if r.overhead_mean_ms else "N/A"
            lines.append(f"| {r.name} | {status} | {latency} | {server} | {overhead} |")

        lines.append("")

    # Latency statistics table
    lines.append("## Latency Statistics")
    lines.append("")
    lines.append("| Test | Min (ms) | Max (ms) | Mean (ms) | StdDev (ms) |")
    lines.append("|------|----------|----------|-----------|-------------|")

    for r in suite.results:
        min_lat = f"{r.latency_min_ms:.0f}" if r.latency_min_ms else "N/A"
        max_lat = f"{r.latency_max_ms:.0f}" if r.latency_max_ms else "N/A"
        mean_lat = f"{r.latency_mean_ms:.0f}" if r.latency_mean_ms else "N/A"
        stddev = f"{r.latency_stddev_ms:.0f}" if r.latency_stddev_ms else "N/A"
        lines.append(f"| {r.name} | {min_lat} | {max_lat} | {mean_lat} | {stddev} |")

    lines.append("")
    lines.append("---")
    lines.append("*Generated by python-executor benchmark script*")

    return "\n".join(lines)


def main():
    parser = argparse.ArgumentParser(description="Performance benchmark for python-executor")
    parser.add_argument("--server", default=SERVER_URL,
                        help=f"Server URL (default: {SERVER_URL})")
    parser.add_argument("--categories",
                        help="Comma-separated list of categories to run")
    parser.add_argument("--iterations", type=int, default=3,
                        help="Number of iterations per test (default: 3)")
    parser.add_argument("--output-json",
                        help="Save results to JSON file")
    parser.add_argument("--output-markdown",
                        help="Save results to Markdown file")
    parser.add_argument("--quick", action="store_true",
                        help="Quick mode (1 iteration, skip slow tests)")
    parser.add_argument("--list", action="store_true", dest="list_tests",
                        help="List available tests and exit")
    args = parser.parse_args()

    iterations = 1 if args.quick else args.iterations
    skip_slow = args.quick

    runner = BenchmarkRunner(args.server, iterations=iterations, skip_slow=skip_slow)

    # List tests if requested
    if args.list_tests:
        print("Available benchmark tests:\n")
        categories = {}
        for test in runner.list_tests():
            if test["category"] not in categories:
                categories[test["category"]] = []
            categories[test["category"]].append(test)

        for category, tests in sorted(categories.items()):
            print(f"Category: {category}")
            for test in tests:
                slow_marker = " [slow]" if test["slow"] else ""
                print(f"  - {test['name']}: {test['description']}{slow_marker}")
            print()

        print(f"Categories: {', '.join(runner.get_categories())}")
        return

    # Health check
    logger.info(f"Connecting to {args.server}...")
    if not runner.health_check():
        logger.error("Server health check failed. Exiting.")
        sys.exit(1)
    logger.info("Server is healthy")

    # Parse categories
    categories = None
    if args.categories:
        categories = [c.strip() for c in args.categories.split(",")]
        logger.info(f"Running categories: {categories}")

    # Run benchmarks
    logger.info(f"Starting benchmark (iterations={iterations}, skip_slow={skip_slow})")
    suite = runner.run(categories=categories)

    # Print summary
    print("\n" + "=" * 60)
    print("BENCHMARK COMPLETE")
    print("=" * 60)
    print(f"Total tests: {len(suite.results)}")
    print(f"Passed: {sum(1 for r in suite.results if r.failures == 0)}")
    print(f"Failed: {sum(1 for r in suite.results if r.failures > 0)}")
    print(f"Duration: {suite.total_duration_seconds}s")

    # Save outputs
    if args.output_json:
        with open(args.output_json, "w") as f:
            json.dump(results_to_json(suite), f, indent=2)
        logger.info(f"Results saved to {args.output_json}")

    if args.output_markdown:
        with open(args.output_markdown, "w") as f:
            f.write(results_to_markdown(suite))
        logger.info(f"Results saved to {args.output_markdown}")

    # Exit with error if any failures
    if any(r.failures > 0 for r in suite.results):
        sys.exit(1)


if __name__ == "__main__":
    main()
