#!/bin/sh
# Every theorem may rest only on propext, Classical.choice and Quot.sound.
set -eu
cd "$(dirname "$0")"
lake build >/dev/null
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT
{
  echo "import Invariants"
  echo "open Refine Shape Walk Dispatch"
  grep -oE '^theorem [A-Za-z_.]+' Invariants/*.lean | awk '{print "#print axioms " $2}'
} > "$tmp/Axioms.lean"
out=$(lake env lean "$tmp/Axioms.lean")
bad=$(printf '%s\n' "$out" | grep "axioms" | grep -vE "does not depend on any axioms|\[(propext|Classical\.choice|Quot\.sound)(, (propext|Classical\.choice|Quot\.sound))*\]" || true)
if [ -n "$bad" ]; then printf '%s\n' "$bad"; exit 1; fi
echo "$(printf '%s\n' "$out" | grep -c axioms) theorems: only propext, Classical.choice, Quot.sound"
