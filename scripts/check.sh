#!/bin/bash
# All deterministic checks. Requires macOS, Go 1.25, Node and full Xcode.
set -euo pipefail
ROOT=$(cd "$(dirname "$0")/.." && pwd)
cd "$ROOT"
# Frozen for now: website/export, website/replay, check:real (which runs ./skyra)
# and controller/run.sh need the controller this branch removed, so they are skipped.
go test -race -count=1 ./tui/...
go vet ./tui/...
(cd controller/apprentice/typesys-prototype && go test -race -count=1 ./... && go vet ./...)
if [ -z "${DEVELOPER_DIR:-}" ] && [ -d /Applications/Xcode.app/Contents/Developer ]; then
 export DEVELOPER_DIR=/Applications/Xcode.app/Contents/Developer
fi
(cd frontend && npm run lint && npm run typecheck && npm run build && npm run check)
printf 'All deterministic checks passed. Real-provider tests require a login and are separate.\n'
