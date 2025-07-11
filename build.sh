#!/bin/bash

# TEN VAD Go wrapper build script

set -e

echo "=== TEN VAD Go wrapper build script ==="

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "Error: Go compiler not found, please install Go first"
    exit 1
fi

echo "Go version: $(go version)"

# Check if we're in the correct directory
if [ ! -f "go.mod" ]; then
    echo "Error: Please run this script in the ten-vad directory"
    exit 1
fi

# Check if TEN VAD library files exist
echo "Checking TEN VAD library files..."
if [ ! -d "lib" ]; then
    echo "Warning: lib directory not found, please ensure TEN VAD library files are properly installed"
fi

# Clean previous builds
echo "Cleaning previous builds..."
rm -rf build/
rm -f ten_vad.test
mkdir -p build/

# Note: Run tests using ./run_test.sh
echo "Starting build..."

# Build command line tool
echo "Building command line tool..."
go build -o build/vad_demo main.go

# macOS platform needs rpath patch
if [[ "$OSTYPE" == "darwin"* ]]; then
    echo "macOS platform, patching rpath..."
    # Get absolute path of current script directory
    SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
    # Calculate absolute path of lib directory
    LIB_PATH="$SCRIPT_DIR/lib/macOS"
    install_name_tool -add_rpath "$LIB_PATH" build/vad_demo 2>/dev/null || echo "vad_demo rpath already exists"
fi

# Build library (optional)
# echo "Building library..."
# go build -buildmode=c-shared -o build/libten_vad_go.so .

echo "=== Build completed ==="
echo "Executable: build/vad_demo"

# Show usage instructions
echo ""
echo "=== Usage Instructions ==="
echo "Run command line tool:"
echo "  ./build/vad_demo -input <WAV file> -output <result file>"
echo ""
echo "View help:"
echo "  ./build/vad_demo -help"
echo ""
echo "Show version:"
echo "  ./build/vad_demo -version"
echo ""
echo "Test testset audio file:"
echo "  ./build/vad_demo -input testset/testset-audio-01.wav -output test_result.txt" 