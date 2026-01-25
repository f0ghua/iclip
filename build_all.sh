#!/bin/bash
# Build all iclip GUI components

set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

echo "=========================================="
echo "  iclip GUI Build Script"
echo "=========================================="
echo ""

# Check Go installation
if ! command -v go &> /dev/null; then
    echo "ERROR: Go is not installed"
    echo "Please install Go 1.21+: https://go.dev/dl/"
    exit 1
fi

echo "Go version: $(go version)"
echo ""

# Build client (Linux)
echo "------------------------------------------"
echo "Building iclip_client_gui (Linux)..."
echo "------------------------------------------"
cd "$SCRIPT_DIR/iclip_client_gui"
bash build.sh
echo ""

# Build server (Windows)
echo "------------------------------------------"
echo "Building iclip_server_gui (Windows)..."
echo "------------------------------------------"
cd "$SCRIPT_DIR/iclip_server_gui"
bash build.sh
echo ""

# Summary
echo "=========================================="
echo "  Build Summary"
echo "=========================================="
echo ""
echo "Linux client:"
ls -lh "$SCRIPT_DIR/iclip_client_gui/iclip_client_gui"
echo ""
echo "Windows server:"
ls -lh "$SCRIPT_DIR/iclip_server_gui/iclip_server_gui.exe"
echo ""
echo "Done!"
