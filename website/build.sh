#!/bin/sh
# Builds what the static site needs from the controller: replay compiled to
# WebAssembly, and Go's loader for it. The pages themselves are plain files.
set -eu
here=$(cd "$(dirname "$0")" && pwd)
cd "$here/.."
GOOS=js GOARCH=wasm go build -trimpath -ldflags="-s -w" -o website/assets/replay.wasm ./website/replay
# The toolchain's copy is read-only; replace ours rather than write through it.
rm -f website/assets/wasm_exec.js
cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" website/assets/wasm_exec.js
chmod 644 website/assets/wasm_exec.js
echo "built website/assets/replay.wasm ($(wc -c < website/assets/replay.wasm | tr -d ' ') bytes)"
