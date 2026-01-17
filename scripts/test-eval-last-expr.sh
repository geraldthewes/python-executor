#!/bin/bash
# test-eval-last-expr.sh - Test REPL-style expression evaluation feature
#
# Usage:
#   ./scripts/test-eval-last-expr.sh [--server URL]
#
# Tests the eval_last_expr feature against the python-executor server.

set -e

SERVER="${PYEXEC_SERVER:-http://localhost:8080}"

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

# Remove trailing slash
SERVER="${SERVER%/}"
API_URL="$SERVER/api/v1/eval"

echo "Testing eval_last_expr feature against $SERVER"
echo ""

# Check if jq is available
if ! command -v jq &> /dev/null; then
    echo "ERROR: jq is required but not installed"
    exit 1
fi

# Health check
echo "Checking server health..."
if ! curl -sf "${SERVER}/health" > /dev/null 2>&1; then
    echo "ERROR: Server health check failed at ${SERVER}/health"
    exit 1
fi
echo "Server is healthy"
echo ""

PASSED=0
FAILED=0

# Test function for eval_last_expr
test_eval() {
    local name="$1"
    local code="$2"
    local expected_result="$3"
    local expected_stdout="$4"

    echo -n "Testing: $name... "

    # Make the request
    response=$(curl -s -X POST "$API_URL" \
        -H "Content-Type: application/json" \
        -d "{\"code\": $(echo "$code" | jq -Rs .), \"eval_last_expr\": true}")

    # Check exit code
    exit_code=$(echo "$response" | jq -r '.exit_code')
    if [[ "$exit_code" != "0" ]]; then
        echo "FAIL (exit_code: $exit_code)"
        echo "  Response: $(echo "$response" | jq -c .)"
        ((FAILED++))
        return 1
    fi

    # Check result
    actual_result=$(echo "$response" | jq -r '.result // "null"')
    if [[ "$expected_result" != "null" ]] && [[ "$actual_result" != "$expected_result" ]]; then
        echo "FAIL (result mismatch)"
        echo "  Expected result: $expected_result"
        echo "  Got result: $actual_result"
        ((FAILED++))
        return 1
    fi

    if [[ "$expected_result" == "null" ]] && [[ "$actual_result" != "null" ]]; then
        echo "FAIL (expected null result)"
        echo "  Got result: $actual_result"
        ((FAILED++))
        return 1
    fi

    # Check stdout if specified
    if [[ -n "$expected_stdout" ]]; then
        actual_stdout=$(echo "$response" | jq -r '.stdout')
        if [[ "$actual_stdout" != *"$expected_stdout"* ]]; then
            echo "FAIL (stdout mismatch)"
            echo "  Expected stdout to contain: $expected_stdout"
            echo "  Got stdout: $actual_stdout"
            ((FAILED++))
            return 1
        fi
    fi

    echo "PASS"
    ((PASSED++))
    return 0
}

echo "=== Basic Expression Tests ==="

test_eval \
    "simple addition" \
    "2 + 2" \
    "4"

test_eval \
    "multiplication" \
    "6 * 7" \
    "42"

test_eval \
    "float division" \
    "10 / 4" \
    "2.5"

echo ""
echo "=== Multi-line with Expression ==="

test_eval \
    "variable then expression" \
    "x = 5
y = 10
x + y" \
    "15"

test_eval \
    "calculation chain" \
    "a = 100
b = a // 3
b * 2" \
    "66"

test_eval \
    "import then expression" \
    "import math
math.sqrt(16)" \
    "4.0"

echo ""
echo "=== No Expression (null result) ==="

test_eval \
    "assignment only" \
    "x = 5" \
    "null"

test_eval \
    "print statement" \
    "print('hello')" \
    "null" \
    "hello"

test_eval \
    "multiple assignments" \
    "x = 1
y = 2
z = 3" \
    "null"

echo ""
echo "=== Complex Expressions ==="

test_eval \
    "list expression" \
    "[1, 2, 3]" \
    "[1, 2, 3]"

test_eval \
    "list comprehension" \
    "[x**2 for x in range(5)]" \
    "[0, 1, 4, 9, 16]"

test_eval \
    "dict expression" \
    "{'a': 1, 'b': 2}" \
    "{'a': 1, 'b': 2}"

test_eval \
    "string expression" \
    "'hello world'" \
    "'hello world'"

test_eval \
    "tuple expression" \
    "(1, 2, 3)" \
    "(1, 2, 3)"

test_eval \
    "set expression" \
    "{1, 2, 3}" \
    "{1, 2, 3}"

echo ""
echo "=== Mixed Print and Expression ==="

test_eval \
    "print then expression" \
    "print('computing...')
2 ** 10" \
    "1024" \
    "computing..."

test_eval \
    "multiple prints then expression" \
    "print('step 1')
print('step 2')
'done'" \
    "'done'" \
    "step"

echo ""
echo "=== Built-in Functions ==="

test_eval \
    "len function" \
    "len([1, 2, 3, 4, 5])" \
    "5"

test_eval \
    "sum function" \
    "sum(range(101))" \
    "5050"

test_eval \
    "sorted function" \
    "sorted([3, 1, 4, 1, 5])" \
    "[1, 1, 3, 4, 5]"

echo ""
echo "=== Backward Compatibility (without eval_last_expr) ==="

# Test that normal execution still works
echo -n "Testing: normal execution without flag... "
response=$(curl -s -X POST "$API_URL" \
    -H "Content-Type: application/json" \
    -d '{"code": "print(42)"}')

stdout=$(echo "$response" | jq -r '.stdout')
result=$(echo "$response" | jq -r '.result // "null"')

if [[ "$stdout" == *"42"* ]] && [[ "$result" == "null" ]]; then
    echo "PASS"
    ((PASSED++))
else
    echo "FAIL"
    echo "  Expected stdout with 42 and null result"
    echo "  Got: stdout=$stdout, result=$result"
    ((FAILED++))
fi

echo ""
echo "=== Summary ==="
echo "Passed: $PASSED"
echo "Failed: $FAILED"
echo ""

if [[ "$FAILED" -gt 0 ]]; then
    echo "Some tests failed!"
    exit 1
fi

echo "All eval_last_expr tests passed!"
exit 0
