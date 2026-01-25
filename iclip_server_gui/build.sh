#!/bin/bash
# Build script for iclip_server_gui (Windows cross-compile from Linux)

set -e

cd "$(dirname "$0")"

echo "=== Building iclip_server_gui (Windows) ==="

# Check for MinGW
if ! command -v x86_64-w64-mingw32-gcc &> /dev/null; then
    echo "ERROR: x86_64-w64-mingw32-gcc not found"
    echo "Please install: sudo apt install gcc-mingw-w64-x86-64"
    exit 1
fi

# Download dependencies
echo "[1/2] Downloading dependencies..."
go mod tidy

# Build
echo "[2/2] Cross-compiling for Windows..."
CGO_ENABLED=1 \
CC=x86_64-w64-mingw32-gcc \
GOOS=windows \
GOARCH=amd64 \
go build -v -ldflags="-s -w -H=windowsgui" -o iclip_server_gui.exe

echo ""
echo "=== Build complete ==="
ls -lh iclip_server_gui.exe
echo ""
echo "Copy iclip_server_gui.exe and config.ini to Windows to run"
