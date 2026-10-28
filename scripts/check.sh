#!/bin/bash
# All deterministic checks. Requires macOS, Go 1.25, Node and full Xcode.
set -euo pipefail
ROOT=$(cd "$(dirname "$0")/.." && pwd)
cd "$ROOT"
go test -race -count=1 ./controller/... ./tui/... ./website/export/...
go vet ./controller/... ./tui/... ./website/export/...
(cd controller/interp && go test -race -count=1 ./... && go vet ./...)
if [ -z "${DEVELOPER_DIR:-}" ] && [ -d /Applications/Xcode.app/Contents/Developer ]; then
 export DEVELOPER_DIR=/Applications/Xcode.app/Contents/Developer
fi
(cd frontend && npm run lint && npm run typecheck && npm run build && npm run check && npm run check:real -- --desktop)
wasm=$(mktemp /tmp/skyra-replay-XXXXXX)
trap 'rm -f "$wasm"' EXIT
GOOS=js GOARCH=wasm go build -trimpath -o "$wasm" ./website/replay
./controller/run.sh
printf 'All deterministic checks passed. Real-provider tests require a login and are separate.\n'
