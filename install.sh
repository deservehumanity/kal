#!/usr/bin/env bash

set -e

APP_NAME="kal"
BIN_DIR="./bin"
INSTALL_PATH="/usr/local/bin/$APP_NAME"
APP_DIR="./cmd/kal/"

mkdir -p "$BIN_DIR"

go build -o "$BIN_DIR/$APP_NAME" "$APP_DIR"

if [ ! -f "$BIN_DIR/$APP_NAME" ]; then
    echo "build failed: binary not found at $BIN_DIR/$APP_NAME"
    exit 1
fi

mv "$BIN_DIR/$APP_NAME" "$INSTALL_PATH"

if ! command -v "$APP_NAME" >/dev/null 2>&1; then
    echo "installation complete, but $APP_NAME is not in PATH"
    echo "ensure /usr/local/bin is in your PATH"
fi
