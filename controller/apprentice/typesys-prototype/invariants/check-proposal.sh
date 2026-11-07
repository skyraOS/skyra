#!/bin/sh
# Proposal 55's Lean blocks are a copy of Invariants/Core.lean, in order.
set -eu
cd "$(dirname "$0")"
p=../../../../docs/next/proposal/55-one-tool-trees-invariant-proposal.md
awk '/^```lean/{f=1;next} /^```/{if(f){f=0;print ""};next} f' "$p" | diff - Invariants/Core.lean >/dev/null \
  || { echo "proposal 55's Lean differs from Invariants/Core.lean"; exit 1; }
echo "proposal 55 matches Invariants/Core.lean"
