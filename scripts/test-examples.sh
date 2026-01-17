#!/bin/bash
# test-examples.sh - Test CLI examples from documentation against deployed server
#
# Usage:
#   ./scripts/test-examples.sh [--server URL]
#
# Tests basic CLI functionality against the python-executor server.

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
CLI="$PROJECT_ROOT/bin/python-executor"
SERVER="${PYEXEC_SERVER:-http://pyexec.cluster:9999/}"

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --server)
            SERVER="$2"
            shift 2
            ;;
        *)
            echo "Unknown option: $1"
            exit 1
            ;;
    esac
done

echo "Testing CLI examples against $SERVER"
echo "Using CLI: $CLI"
echo ""

# Check if CLI exists
if [[ ! -x "$CLI" ]]; then
    echo "ERROR: CLI not found at $CLI"
    echo "Run 'make build-cli' first"
    exit 1
fi

# Health check
echo "Checking server health..."
if ! curl -sf "${SERVER}health" > /dev/null 2>&1; then
    echo "ERROR: Server health check failed at ${SERVER}health"
    exit 1
fi
echo "Server is healthy"
echo ""

PASSED=0
FAILED=0

# Test function
test_example() {
    local name="$1"
    local cmd="$2"
    local expected_exit="${3:-0}"
    local expected_output="$4"

    echo -n "Testing: $name... "

    # Run the command
    set +e
    output=$(eval "$cmd" 2>&1)
    exit_code=$?
    set -e

    # Check exit code
    if [[ "$exit_code" -ne "$expected_exit" ]]; then
        echo "FAIL (exit code: $exit_code, expected: $expected_exit)"
        echo "  Output: ${output:0:200}"
        ((FAILED++))
        return 1
    fi

    # Check output if specified
    if [[ -n "$expected_output" ]] && [[ "$output" != *"$expected_output"* ]]; then
        echo "FAIL (output mismatch)"
        echo "  Expected: $expected_output"
        echo "  Got: ${output:0:200}"
        ((FAILED++))
        return 1
    fi

    echo "PASS"
    ((PASSED++))
    return 0
}

echo "=== Basic Execution Tests ==="

# Test 1: Simple stdin execution
test_example \
    "stdin execution" \
    "echo 'print(\"hello world\")' | $CLI --server $SERVER run" \
    0 \
    "hello world"

# Test 2: Math operation
test_example \
    "math operation" \
    "echo 'print(2 + 2)' | $CLI --server $SERVER run" \
    0 \
    "4"

# Test 3: Multi-line script
test_example \
    "multi-line script" \
    "echo -e 'x = 5\ny = 3\nprint(x * y)' | $CLI --server $SERVER run" \
    0 \
    "15"

# Test 4: Exit code propagation (error case)
test_example \
    "exit code propagation" \
    "echo 'import sys; sys.exit(42)' | $CLI --server $SERVER run" \
    42

# Test 5: Quiet mode
test_example \
    "quiet mode" \
    "echo 'print(\"quiet test\")' | $CLI --server $SERVER run -q" \
    0 \
    "quiet test"

# Test 6: Version command
test_example \
    "version command" \
    "$CLI version" \
    0 \
    "python-executor"

echo ""
echo "=== File Execution Tests ==="

# Create temp directory for file tests
TMPDIR=$(mktemp -d)
trap "rm -rf $TMPDIR" EXIT

# Test 7: Single file execution
echo 'print("file test")' > "$TMPDIR/test.py"
test_example \
    "single file execution" \
    "$CLI --server $SERVER run $TMPDIR/test.py" \
    0 \
    "file test"

# Test 8: Directory execution with main.py
mkdir -p "$TMPDIR/project"
echo 'print("main entry")' > "$TMPDIR/project/main.py"
echo 'helper_var = 42' > "$TMPDIR/project/helper.py"
test_example \
    "directory execution" \
    "$CLI --server $SERVER run $TMPDIR/project/" \
    0 \
    "main entry"

# Test 9: Tar file execution
tar -cf "$TMPDIR/code.tar" -C "$TMPDIR/project" .
test_example \
    "tar file execution" \
    "$CLI --server $SERVER run $TMPDIR/code.tar" \
    0 \
    "main entry"

echo ""
echo "=== Summary ==="
echo "Passed: $PASSED"
echo "Failed: $FAILED"
echo ""

if [[ "$FAILED" -gt 0 ]]; then
    echo "Some tests failed!"
    exit 1
fi

echo "All CLI tests passed!"
exit 0
