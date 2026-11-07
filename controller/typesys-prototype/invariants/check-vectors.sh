#!/bin/sh
# The committed vectors are exactly what the Lean writes today.
set -eu
cd "$(dirname "$0")"
lake build vectors >/dev/null
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT
lake exe vectors "$tmp" >/dev/null
diff -r vectors "$tmp" >/dev/null || { echo "vectors are stale: run 'lake exe vectors'"; exit 1; }
echo "vectors match the Lean"

lake build rootvectors >/dev/null
lake exe rootvectors "$tmp/human-root.json" >/dev/null
cmp ../../core/humanroot/testdata/lean.json "$tmp/human-root.json" >/dev/null || { echo "human root vectors are stale: run lake exe rootvectors ../../core/humanroot/testdata/lean.json"; exit 1; }
echo "human root vectors match the Lean"
