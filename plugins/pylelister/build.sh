#!/bin/bash

#
# Build script for Python plugin
#
# This creates a self-contained executable that includes:
# - Python interpreter
# - Plugin code
# - All dependencies (grpcio, protobuf, etc.)
# - Generated protobuf stubs
#
# The resulting binary can run on any system without Python installed.
#

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

echo "==> Building Python plugin: pylelister"

# Check if we need to generate protobuf code
PROTOGEN_DIR="../../shared/protogen_py"
if [ ! -d "$PROTOGEN_DIR" ]; then
    echo "==> Protobuf stubs not found, generating..."
    cd ../..
    buf generate
    cd "$SCRIPT_DIR"
fi

# Create virtual environment for build if it doesn't exist
if [ ! -d "build_env" ]; then
    echo "==> Creating build virtual environment..."
    python3 -m venv build_env
fi

# Activate virtual environment
echo "==> Activating build environment..."
source build_env/bin/activate

# Install/upgrade dependencies
echo "==> Installing dependencies..."
pip install --upgrade pip
pip install -r requirements.txt

# Build with PyInstaller
echo "==> Building executable with PyInstaller..."
pyinstaller --clean pylelister.spec

# Check if build succeeded
if [ -f "dist/pylelister" ]; then
    echo "==> Build successful!"
    echo "==> Executable: $SCRIPT_DIR/dist/pylelister"
    ls -lh dist/pylelister
else
    echo "==> Build failed - executable not found"
    exit 1
fi

echo "==> Done!"
