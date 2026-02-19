#!/bin/bash

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
OUT_DIR="$PROJECT_DIR/out"

echo "=== Building ==="

mkdir -p "$OUT_DIR"
rm -rf "$OUT_DIR"/*

if ! command -v go &> /dev/null; then
    echo "WARNING: Go not found, skipping WASM build (using existing main.wasm if present)"
    if [ ! -f "$OUT_DIR/main.wasm" ]; then
        echo "ERROR: No WASM file found. Please install Go or ensure out/main.wasm exists."
        exit 1
    fi
else
    echo "=== Go WASM Build ==="
    GOROOT=$(go env GOROOT)
    echo "GOROOT: $GOROOT"

    WASM_EXEC_PATHS=(
        "$GOROOT/lib/wasm/wasm_exec.js"
        "$GOROOT/misc/wasm/wasm_exec.js"
    )

    WASM_EXEC_SRC=""
    for p in "${WASM_EXEC_PATHS[@]}"; do
        if [ -f "$p" ]; then
            WASM_EXEC_SRC="$p"
            break
        fi
    done

    if [ -z "$WASM_EXEC_SRC" ]; then
        echo "ERROR: wasm_exec.js not found"
        exit 1
    fi

    cp "$WASM_EXEC_SRC" "$OUT_DIR/wasm_exec.js"
    echo "Copied wasm_exec.js"

    echo "Building WebAssembly..."
    cd "$PROJECT_DIR/backend"
    GOOS=js GOARCH=wasm go build -o ../out/main.wasm .

    echo "WASM build successful!"
fi

echo "=== TypeScript Compilation ==="
cd "$PROJECT_DIR"
npm run compile

echo "=== VSIX Package ==="
npx vsce package

echo "=== Build Complete ==="
ls -la "$PROJECT_DIR"/*.vsix 2>/dev/null || echo "No VSIX generated"
