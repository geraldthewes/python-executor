#!/bin/bash
# generate-docs.sh - Assembles final documentation from generated and manual content

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
DOCS_DIR="$PROJECT_ROOT/docs"
GENERATED_DIR="$DOCS_DIR/_generated"

echo "Assembling documentation..."

# Assemble Python client documentation
if [[ -f "$GENERATED_DIR/python/python_executor_client.client.md" ]]; then
    cat > "$DOCS_DIR/python-client.md" << 'HEADER'
# Python Client Library

The Python client provides a simple interface for executing Python code remotely.

## Installation

```bash
pip install git+https://github.com/geraldthewes/python-executor.git#subdirectory=python
```

## Quick Start

```python
from python_executor_client import PythonExecutorClient

client = PythonExecutorClient("http://pyexec.cluster:9999/")

# Execute code from a dict of files
result = client.execute_sync(
    files={"main.py": "print('Hello, World!')"},
    entrypoint="main.py"
)

print(result.stdout)     # Hello, World!
print(result.exit_code)  # 0
```

## API Reference

HEADER
    # Append generated content (skip the title from lazydocs)
    tail -n +2 "$GENERATED_DIR/python/python_executor_client.client.md" >> "$DOCS_DIR/python-client.md"

    # Append types documentation if it exists
    if [[ -f "$GENERATED_DIR/python/python_executor_client.types.md" ]]; then
        echo "" >> "$DOCS_DIR/python-client.md"
        echo "---" >> "$DOCS_DIR/python-client.md"
        echo "" >> "$DOCS_DIR/python-client.md"
        tail -n +2 "$GENERATED_DIR/python/python_executor_client.types.md" >> "$DOCS_DIR/python-client.md"
    fi

    echo "  ✓ Python client documentation assembled"
else
    echo "  ⚠ Python generated docs not found, skipping python-client.md assembly"
fi

# Assemble Go client documentation
if [[ -f "$GENERATED_DIR/go/client.md" ]]; then
    cat > "$DOCS_DIR/go-client.md" << 'HEADER'
# Go Client Library

The Go client provides a type-safe interface for executing Python code remotely.

## Installation

```bash
go get github.com/geraldthewes/python-executor/pkg/client
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/geraldthewes/python-executor/pkg/client"
)

func main() {
    c := client.New("http://pyexec.cluster:9999/")

    // Create tar from a map of files
    tarData, _ := client.TarFromMap(map[string]string{
        "main.py": `print("Hello from Go!")`,
    })

    result, err := c.ExecuteSync(context.Background(), tarData, &client.Metadata{
        Entrypoint: "main.py",
    })
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println(result.Stdout)
}
```

## API Reference

HEADER
    # Append generated content (skip the header from gomarkdoc)
    cat "$GENERATED_DIR/go/client.md" >> "$DOCS_DIR/go-client.md"
    echo "  ✓ Go client documentation assembled"
else
    echo "  ⚠ Go generated docs not found, skipping go-client.md assembly"
fi

# Assemble CLI documentation
if [[ -f "$GENERATED_DIR/cli/python-executor.md" ]]; then
    cat > "$DOCS_DIR/cli.md" << 'HEADER'
# CLI Reference

The `python-executor` CLI provides command-line access to the Python execution service.

## Installation

```bash
# From source
make build-cli
sudo cp bin/python-executor /usr/local/bin/

# Or download from releases
# See https://github.com/geraldthewes/python-executor/releases
```

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `PYEXEC_SERVER` | Server URL | `http://localhost:8080` |

## Quick Start

```bash
# Run code from stdin
echo 'print("Hello")' | python-executor --server http://pyexec.cluster:9999/ run

# Run a file
python-executor --server http://pyexec.cluster:9999/ run script.py

# Run a directory
python-executor --server http://pyexec.cluster:9999/ run ./myproject/
```

## Command Reference

HEADER
    # Append generated CLI docs
    cat "$GENERATED_DIR/cli/python-executor.md" >> "$DOCS_DIR/cli.md"

    # Append subcommand docs
    for subcmd in "$GENERATED_DIR/cli/python-executor_"*.md; do
        if [[ -f "$subcmd" ]]; then
            echo "" >> "$DOCS_DIR/cli.md"
            echo "---" >> "$DOCS_DIR/cli.md"
            echo "" >> "$DOCS_DIR/cli.md"
            cat "$subcmd" >> "$DOCS_DIR/cli.md"
        fi
    done

    echo "  ✓ CLI documentation assembled"
else
    echo "  ⚠ CLI generated docs not found, skipping cli.md assembly"
fi

echo "Documentation assembly complete!"
