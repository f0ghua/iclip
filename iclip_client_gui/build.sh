#!/bin/bash
# Build script for iclip_client_gui (Linux)

set -e

cd "$(dirname "$0")"

echo "=== Building iclip_client_gui (Linux) ==="

# Download dependencies
echo "[1/2] Downloading dependencies..."
go mod tidy

# Build
echo "[2/2] Compiling..."
go build -v -ldflags="-s -w" -o iclip_client_gui

echo ""
echo "=== Build complete ==="
ls -lh iclip_client_gui
echo ""
echo "To run: ./iclip_client_gui"
