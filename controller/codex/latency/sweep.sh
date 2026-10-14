#!/bin/sh
# The overhead sweep in OVERHEAD-VS-INFERENCE.md: for each latency, TRIALS boots
# of each scenario against the fake Codex behind latency, each with a fresh
# .skyra/ root. Prints: ms scenario wall_ms records answer
#
#   codex/latency/sweep.sh 6 0 250 500 1000 2000 4000
#   SCEN=tool:hello codex/latency/sweep.sh 4 50 100 150
set -eu
cd "$(dirname "$0")/../.."
n=$1; shift
mkdir -p bin
go build -o bin/skyra ./skyra
go build -o bin/echo ./targets/echo
go build -o bin/codex ./codex/adapter
go build -o bin/host ./hosts/actor
go build -o bin/fake-codex ./codex/fake
go build -o bin/latency ./codex/latency
now() { perl -MTime::HiRes=time -e 'printf "%d", time*1000'; }
for ms in "$@"; do
  # A unix socket path must be short, and real (the sandbox matches real paths).
  sock=/private/tmp/skyra-lat-$$.sock
  ./bin/host -sock "$sock" -state "$(mktemp -d)" -- ./bin/latency -ms "$ms" ./bin/fake-codex 2>/dev/null &
  host=$!
  trap 'kill $host 2>/dev/null || true' EXIT
  while [ ! -S "$sock" ]; do sleep 0.05; done
  reg=$(mktemp)
  cat > "$reg" <<JSON
{
  "echo":  {"path": "$PWD/bin/echo",  "contract": "$PWD/targets/echo/contract.json"},
  "codex": {"path": "$PWD/bin/codex", "contract": "$PWD/codex/adapter/contract.json", "connect": ["$sock"]}
}
JSON
  for scen in ${SCEN:-hello tool:hello}; do
    i=0
    while [ "$i" -lt "$n" ]; do
      t0=$(now)
      out=$(./bin/skyra boot -root "$(mktemp -d)" -registry "$reg" -to codex -text "$scen" 2>/dev/null)
      t1=$(now)
      rec=$(printf '%s\n' "$out" | sed -n 's/^replay: \([0-9]*\) records.*/\1/p')
      ans=$(printf '%s\n' "$out" | grep '^answer:' | head -1)
      echo "$ms $scen $((t1-t0)) $rec $ans"
      i=$((i+1))
    done
  done
  kill "$host" 2>/dev/null; wait "$host" 2>/dev/null || true
  rm -f "$reg"
done
