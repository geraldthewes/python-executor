#!/usr/bin/env python3
"""Test Python code examples from documentation against the deployed server.

This script extracts Python code blocks from markdown files and executes them
against the python-executor service to validate they work correctly.

Usage:
    python scripts/test-examples.py [--update] [--server URL]

Options:
    --update    Update markdown files with actual output
    --server    Server URL (default: http://pyexec.cluster:9999/)
"""

import argparse
import re
import sys
from pathlib import Path

# Add the python client to the path
sys.path.insert(0, str(Path(__file__).parent.parent / "python"))

from python_executor_client import PythonExecutorClient


SERVER_URL = "http://pyexec.cluster:9999/"
DOCS_DIR = Path(__file__).parent.parent / "docs"


def extract_python_examples(content: str) -> list[dict]:
    """Extract Python code blocks from markdown content.

    Returns a list of dicts with 'code', 'start_line', 'end_line' keys.
    Only extracts blocks marked with ```python that appear to be executable
    (not just import statements or class definitions).
    """
    examples = []
    lines = content.split('\n')
    in_block = False
    block_start = 0
    block_lines = []
    block_lang = ""

    for i, line in enumerate(lines):
        if line.startswith('```python'):
            in_block = True
            block_start = i
            block_lines = []
            block_lang = "python"
        elif line.startswith('```') and in_block:
            in_block = False
            code = '\n'.join(block_lines)

            # Skip blocks that are just examples/templates or don't produce output
            if _is_executable_example(code):
                examples.append({
                    'code': code,
                    'start_line': block_start,
                    'end_line': i,
                    'language': block_lang,
                })
        elif in_block:
            block_lines.append(line)

    return examples


def _is_executable_example(code: str) -> bool:
    """Check if a code block is an executable example.

    Returns False for:
    - Import-only blocks
    - Class/function definitions without execution
    - Blocks with ellipsis (...)
    - Blocks that use PythonExecutorClient (testing the client itself)
    """
    # Skip blocks that use the client (we're testing against the server directly)
    if 'PythonExecutorClient' in code:
        return False

    # Skip blocks with just imports
    lines = [l.strip() for l in code.split('\n') if l.strip() and not l.strip().startswith('#')]
    if all(l.startswith(('import ', 'from ')) for l in lines):
        return False

    # Skip blocks with ellipsis (placeholders)
    if '...' in code:
        return False

    # Skip blocks that are just definitions without calls
    if all(l.startswith(('def ', 'class ', '@', 'import ', 'from ')) or l.startswith(' ') for l in lines):
        return False

    return True


def test_example(client: PythonExecutorClient, code: str) -> tuple[bool, str, str]:
    """Execute a code example and return (success, stdout, stderr)."""
    try:
        result = client.execute_sync(
            files={"main.py": code},
            entrypoint="main.py",
            timeout_seconds=30,
        )
        success = result.exit_code == 0
        return success, result.stdout or "", result.stderr or ""
    except Exception as e:
        return False, "", str(e)


def test_markdown_file(filepath: Path, client: PythonExecutorClient, update: bool = False) -> list[dict]:
    """Test all Python examples in a markdown file.

    Returns a list of test results.
    """
    content = filepath.read_text()
    examples = extract_python_examples(content)

    results = []
    for i, example in enumerate(examples):
        code = example['code']
        print(f"  Testing example {i+1}/{len(examples)} (lines {example['start_line']+1}-{example['end_line']+1})...")

        success, stdout, stderr = test_example(client, code)

        result = {
            'file': str(filepath),
            'example_num': i + 1,
            'start_line': example['start_line'] + 1,
            'end_line': example['end_line'] + 1,
            'success': success,
            'stdout': stdout,
            'stderr': stderr,
            'code': code[:100] + '...' if len(code) > 100 else code,
        }
        results.append(result)

        if success:
            print(f"    PASS")
        else:
            print(f"    FAIL: {stderr[:100] if stderr else 'No output'}")

    return results


def main():
    parser = argparse.ArgumentParser(description="Test Python examples from documentation")
    parser.add_argument('--update', action='store_true', help="Update markdown with actual output")
    parser.add_argument('--server', default=SERVER_URL, help=f"Server URL (default: {SERVER_URL})")
    parser.add_argument('files', nargs='*', help="Specific markdown files to test (default: all in docs/)")
    args = parser.parse_args()

    # Create client
    print(f"Connecting to {args.server}...")
    client = PythonExecutorClient(args.server, timeout=60)

    # Health check
    try:
        import requests
        resp = requests.get(f"{args.server.rstrip('/')}/health", timeout=5)
        resp.raise_for_status()
        print("Server is healthy")
    except Exception as e:
        print(f"ERROR: Server health check failed: {e}")
        sys.exit(1)

    # Find markdown files
    if args.files:
        md_files = [Path(f) for f in args.files]
    else:
        md_files = list(DOCS_DIR.glob("*.md"))

    # Test each file
    all_results = []
    for filepath in md_files:
        if not filepath.exists():
            print(f"WARNING: {filepath} not found, skipping")
            continue

        print(f"\nTesting {filepath.name}...")
        results = test_markdown_file(filepath, client, args.update)
        all_results.extend(results)

    # Summary
    passed = sum(1 for r in all_results if r['success'])
    failed = sum(1 for r in all_results if not r['success'])
    total = len(all_results)

    print(f"\n{'='*60}")
    print(f"SUMMARY: {passed}/{total} examples passed, {failed} failed")

    if failed > 0:
        print("\nFailed examples:")
        for r in all_results:
            if not r['success']:
                print(f"  - {r['file']}:{r['start_line']} - {r['stderr'][:80] if r['stderr'] else 'No error output'}")
        sys.exit(1)

    print("\nAll examples passed!")
    sys.exit(0)


if __name__ == "__main__":
    main()
