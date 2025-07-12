#!/bin/bash

# Prepare release package with library files

set -e

echo "=== Preparing release package ==="

# Create lib and include directories in vad package
mkdir -p vad/lib
mkdir -p vad/include

# Copy library files
echo "Copying library files..."
cp -r lib/* vad/lib/
cp -r include/* vad/include/

# Fix permissions for macOS framework
if [[ "$OSTYPE" == "darwin"* ]]; then
    echo "Fixing macOS framework permissions..."
    if [ -d "vad/lib/macOS/ten_vad.framework" ]; then
        chmod +x vad/lib/macOS/ten_vad.framework/ten_vad 2>/dev/null || true
        chmod +x vad/lib/macOS/ten_vad.framework/Versions/A/ten_vad 2>/dev/null || true
        
        # Re-sign framework if needed
        echo "Re-signing framework..."
        codesign --force --sign - vad/lib/macOS/ten_vad.framework 2>/dev/null || echo "Code signing failed (this is normal if not needed)"
    fi
fi

# Update cgo paths for release
echo "Updating cgo paths..."
sed -i.bak 's|#cgo CFLAGS: -I${SRCDIR}/../include|#cgo CFLAGS: -I${SRCDIR}/include|g' vad/vad.go
sed -i.bak 's|#cgo linux LDFLAGS: -L${SRCDIR}/../lib/Linux/x64|#cgo linux LDFLAGS: -L${SRCDIR}/lib/Linux/x64|g' vad/vad.go
sed -i.bak 's|#cgo windows,amd64 LDFLAGS: -L${SRCDIR}/../lib/Windows/x64|#cgo windows,amd64 LDFLAGS: -L${SRCDIR}/lib/Windows/x64|g' vad/vad.go
sed -i.bak 's|#cgo windows,386 LDFLAGS: -L${SRCDIR}/../lib/Windows/x86|#cgo windows,386 LDFLAGS: -L${SRCDIR}/lib/Windows/x86|g' vad/vad.go
sed -i.bak 's|#cgo darwin LDFLAGS: -F${SRCDIR}/../lib/macOS|#cgo darwin LDFLAGS: -F${SRCDIR}/lib/macOS|g' vad/vad.go
sed -i.bak 's|#cgo android,arm LDFLAGS: -L${SRCDIR}/../lib/Android/armeabi-v7a|#cgo android,arm LDFLAGS: -L${SRCDIR}/lib/Android/armeabi-v7a|g' vad/vad.go
sed -i.bak 's|#cgo android,arm64 LDFLAGS: -L${SRCDIR}/../lib/Android/arm64-v8a|#cgo android,arm64 LDFLAGS: -L${SRCDIR}/lib/Android/arm64-v8a|g' vad/vad.go
sed -i.bak 's|#cgo ios LDFLAGS: -L${SRCDIR}/../lib/iOS|#cgo ios LDFLAGS: -L${SRCDIR}/lib/iOS|g' vad/vad.go

# Clean up backup files
rm -f vad/vad.go.bak

echo "Release package prepared successfully!"
echo "Library files are now included in the vad package"
echo ""
echo "Directory structure:"
echo "vad/"
echo "├── vad.go"
echo "├── audio.go"
echo "├── process.go"
echo "├── util.go"
echo "├── lib/"
echo "│   ├── Linux/"
echo "│   ├── Windows/"
echo "│   ├── macOS/"
echo "│   │   └── ten_vad.framework/"
echo "│   ├── Android/"
echo "│   └── iOS/"
echo "└── include/"
echo "    └── ten_vad.h"
echo ""
echo "Framework permissions and code signing have been handled." 