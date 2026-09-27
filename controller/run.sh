#!/bin/sh
# The stateless controller, end to end. Each scenario gets a fresh .skyra/
# state directory and ends with a replay check.
set -eu
cd "$(dirname "$0")"
mkdir -p bin
go build -o bin/skyra ./skyra
go build -o bin/echo ./echo

# The operator's registry: every surface, its binary, and its contract.
cat > bin/registry.json <<JSON
{
  "echo": {"path": "echo", "contract": "../echo/contract.json"}
}
JSON

scenario() {
  name=$1; shift
  root=$(cd "$(mktemp -d)" && pwd -P)/.skyra
  echo
  echo "══ $name ══"
  ./bin/skyra boot -root "$root" -registry bin/registry.json "$@" || true
}

scenario "echo hello"                      -to echo -payload '{"text":"hello"}'
scenario "echo probes .skyra/ from inside" -to echo -payload "{\"text\":\"hello\",\"probe\":\"$(cd "$(mktemp -d)" && pwd -P)/.skyra\"}"
scenario "echo (a leaf) addresses echo"    -to echo -payload '{"text":"hello","misbehave":"yes"}'
scenario "payload breaks echo's contract"  -to echo -payload '{"txt":"hello"}'
