#!/bin/bash

set -e

echo "=== Compiling Go test binary ==="
go test -c -o ten_vad.test

# Detect platform
if [[ "$OSTYPE" == "darwin"* ]]; then
    echo "=== Detected macOS, patching rpath to ten_vad.framework ==="
    # Get absolute path of current script directory
    SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
    # Calculate absolute path of lib directory
    LIB_PATH="$SCRIPT_DIR/lib/macOS"
    install_name_tool -add_rpath "$LIB_PATH" ./ten_vad.test 2>/dev/null || echo "rpath already exists"
else
    echo "=== Detected Linux/Windows, running tests directly ==="
fi

echo "=== Running tests ==="
./ten_vad.test -test.v