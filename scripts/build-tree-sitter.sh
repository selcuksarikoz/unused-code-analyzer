#!/bin/bash

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
THIRD_PARTY="$SCRIPT_DIR/third_party"
OUT_DIR="$PROJECT_DIR/out"

export PATH="/opt/homebrew/Cellar/emscripten/5.0.1/bin:$PATH"

echo "=== Building WASM with tree-sitter ==="

# Create combined C source
COMBINED_SRC="/tmp/tree-sitter-combined.c"
echo "" > "$COMBINED_SRC"

# Add tree-sitter core
cat "$THIRD_PARTY/tree-sitter/lib/src/parser.c" >> "$COMBINED_SRC"

# Add language parsers
for lang in javascript tree-sitter-typescript/tsx tree-sitter-typescript/typescript tree-sitter-python tree-sitter-go; do
    if [ -f "$THIRD_PARTY/$lang/src/parser.c" ]; then
        echo "// ===== $lang =====" >> "$COMBINED_SRC"
        cat "$THIRD_PARTY/$lang/src/parser.c" >> "$COMBINED_SRC"
    fi
done

echo "Combined C source created: $(wc -l < $COMBINED_SRC) lines"

# Build WASM
echo "Building WASM..."
emcc "$COMBINED_SRC" \
    -Os \
    -s WASM=1 \
    -s EXPORTED_RUNCTIONS="['_parse_javascript', '_parse_typescript', '_parse_python', '_parse_go', '_malloc', '_free']" \
    -s EXPORTED_METHODS="['_malloc', '_free']" \
    -s ALLOW_MEMORY_GROWTH=1 \
    -s EXPORT_ES6=0 \
    -s MODULARIZE=0 \
    -o "$OUT_DIR/main.wasm"

echo "Build complete: $OUT_DIR/main.wasm"
