#!/bin/bash
set -e

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
MODE="${1:-mock}"

cleanup() {
  echo ""
  echo "shutting down..."
  kill $MOCK_PID $NEXT_PID 2>/dev/null
  wait $MOCK_PID $NEXT_PID 2>/dev/null
  exit 0
}
trap cleanup INT TERM

cd "$ROOT/frontend"

if [ ! -d "node_modules" ]; then
  echo "installing frontend dependencies..."
  npm install
fi

case "$MODE" in
  mock)
    echo "starting mock WS server on :8080/ws..."
    npm run mock &
    MOCK_PID=$!
    sleep 1

    echo "starting Next.js dev server..."
    npm run dev &
    NEXT_PID=$!

    echo ""
    echo "  frontend:  http://localhost:3000"
    echo "  mock ws:   ws://localhost:8080/ws"
    echo ""
    echo "ctrl-c to stop both."
    wait
    ;;

  runtime)
    echo "starting v.05 runtime..."
    cd "$ROOT/skyra-v.05"
    go run . &
    MOCK_PID=$!
    sleep 2

    cd "$ROOT/frontend"
    echo "starting Next.js dev server..."
    npm run dev &
    NEXT_PID=$!

    echo ""
    echo "  frontend:  http://localhost:3000"
    echo "  runtime:   ws://localhost:8080/ws"
    echo ""
    echo "ctrl-c to stop both."
    wait
    ;;

  *)
    echo "usage: ./scripts/dev.sh [mock|runtime]"
    echo ""
    echo "  mock     — mock WS server + Next.js (default)"
    echo "  runtime  — real v.05 Go runtime + Next.js"
    exit 1
    ;;
esac
