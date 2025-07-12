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

if [ ! -d "include" ]; then
    echo "Warning: include directory not found, please ensure TEN VAD header files are properly installed"
fi

# Set environment variables for Linux
if [[ "$OSTYPE" == "linux-gnu"* ]] || [[ "$OSTYPE" == "linux"* ]]; then
    echo "Setting Linux environment variables..."
    
    # Get current paths
    CURRENT_LIBRARY_PATH="$LD_LIBRARY_PATH"
    
    # Set library path (append to existing if any)
    if [ -n "$CURRENT_LIBRARY_PATH" ]; then
        export LD_LIBRARY_PATH="$CURRENT_LIBRARY_PATH:$(pwd)/vad/lib/Linux/x64"
    else
        export LD_LIBRARY_PATH="$(pwd)/vad/lib/Linux/x64"
    fi
    
    echo "Environment variables set:"
    echo "LD_LIBRARY_PATH: $LD_LIBRARY_PATH"
fi

# Set environment variables for macOS
if [[ "$OSTYPE" == "darwin"* ]]; then
    echo "Setting macOS environment variables..."
    
    # Get current paths
    CURRENT_FRAMEWORK_PATH="$DYLD_FRAMEWORK_PATH"
    CURRENT_LIBRARY_PATH="$DYLD_LIBRARY_PATH"
    
    # Set framework path (usually only one path needed)
    export DYLD_FRAMEWORK_PATH="$(pwd)/vad/lib/macOS"
    
    # Set library path (append to existing if any)
    if [ -n "$CURRENT_LIBRARY_PATH" ]; then
        export DYLD_LIBRARY_PATH="$CURRENT_LIBRARY_PATH:$(pwd)/vad/lib/macOS"
    else
        export DYLD_LIBRARY_PATH="$(pwd)/vad/lib/macOS"
    fi
    
    export CGO_CFLAGS="-I$(pwd)/vad/include"
    export CGO_LDFLAGS="-F$(pwd)/vad/lib/macOS -framework ten_vad -Wl,-rpath,$(pwd)/vad/lib/macOS"
    
    echo "Environment variables set:"
    echo "DYLD_FRAMEWORK_PATH: $DYLD_FRAMEWORK_PATH"
    echo "DYLD_LIBRARY_PATH: $DYLD_LIBRARY_PATH"
    echo "CGO_CFLAGS: $CGO_CFLAGS"
    echo "CGO_LDFLAGS: $CGO_LDFLAGS"
fi

# Set environment variables for Windows
if [[ "$OSTYPE" == "msys"* ]] || [[ "$OSTYPE" == "cygwin"* ]]; then
    echo "Setting Windows environment variables..."
    
    # Set library path for Windows
    export PATH="$PATH:$(pwd)/vad/lib/Windows/x64"
    
    echo "Environment variables set:"
    echo "PATH updated with library path"
fi

# Set environment variables for Android
if [[ "$OSTYPE" == "linux-android"* ]]; then
    echo "Setting Android environment variables..."
    
    # Get current paths
    CURRENT_LIBRARY_PATH="$LD_LIBRARY_PATH"
    
    # Set library path based on architecture
    if [[ "$(uname -m)" == "armv7l" ]]; then
        LIB_PATH="$(pwd)/vad/lib/Android/armeabi-v7a"
    elif [[ "$(uname -m)" == "aarch64" ]]; then
        LIB_PATH="$(pwd)/vad/lib/Android/arm64-v8a"
    else
        LIB_PATH="$(pwd)/vad/lib/Android/arm64-v8a"
    fi
    
    # Set library path (append to existing if any)
    if [ -n "$CURRENT_LIBRARY_PATH" ]; then
        export LD_LIBRARY_PATH="$CURRENT_LIBRARY_PATH:$LIB_PATH"
    else
        export LD_LIBRARY_PATH="$LIB_PATH"
    fi
    
    echo "Environment variables set:"
    echo "LD_LIBRARY_PATH: $LD_LIBRARY_PATH"
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
go build -o build/vad_demo cmd/vad_demo.go

# Linux platform needs rpath patch
if [[ "$OSTYPE" == "linux-gnu"* ]] || [[ "$OSTYPE" == "linux"* ]]; then
    echo "Linux platform, setting up library path..."
    # Get absolute path of current script directory
    SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
    # Calculate absolute path of lib directory
    LIB_PATH="$SCRIPT_DIR/vad/lib/Linux/x64"
    echo "Library path set to: $LIB_PATH"
    
    # Try to set rpath using patchelf if available
    if command -v patchelf &> /dev/null; then
        echo "Using patchelf to set rpath..."
        patchelf --set-rpath "$LIB_PATH" build/vad_demo 2>/dev/null && echo "rpath set successfully"
    else
        echo "patchelf not available, using LD_LIBRARY_PATH for runtime"
        echo "Note: To run the program, set LD_LIBRARY_PATH:"
        echo "export LD_LIBRARY_PATH=\$LD_LIBRARY_PATH:$LIB_PATH"
    fi
fi

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

# Add runtime environment setup instructions
echo ""
echo "=== Runtime Environment Setup ==="
if [[ "$OSTYPE" == "linux-gnu"* ]] || [[ "$OSTYPE" == "linux"* ]]; then
    echo "For Linux, set LD_LIBRARY_PATH before running:"
    echo "  export LD_LIBRARY_PATH=\$LD_LIBRARY_PATH:$(pwd)/vad/lib/Linux/x64"
    echo "  ./build/vad_demo [options]"
elif [[ "$OSTYPE" == "darwin"* ]]; then
    echo "For macOS, DYLD_LIBRARY_PATH is already set in this session"
    echo "  ./build/vad_demo [options]"
elif [[ "$OSTYPE" == "msys"* ]] || [[ "$OSTYPE" == "cygwin"* ]]; then
    echo "For Windows, PATH is already updated in this session"
    echo "  ./build/vad_demo [options]"
fi 