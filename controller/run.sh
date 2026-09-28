#!/bin/sh
# The stateless controller, end to end. Each scenario gets a fresh .skyra/
# state directory and ends with a replay check.
set -eu
cd "$(dirname "$0")"
mkdir -p bin
go build -o bin/skyra ./skyra
go build -o bin/echo ./echo
go build -o bin/agent ./agent
go build -o bin/codex ./codex
go build -o bin/host ./host
go build -o bin/fake-codex ./fakecodex

# The actor host in front of a deterministic fake Codex app-server (fakecodex/).
# A unix socket path must be short, and real (the sandbox matches real paths).
sock=/private/tmp/skyra-fake-$$.sock
./bin/host -sock "$sock" -state "$(mktemp -d)" -- ./bin/fake-codex 2>/dev/null &
host=$!
trap 'kill $host 2>/dev/null' EXIT
while [ ! -S "$sock" ]; do sleep 0.05; done

# The operator's registry: every surface, its binary, and its contract.
cat > bin/registry.json <<JSON
{
  "echo":  {"path": "echo",  "contract": "../echo/contract.json"},
  "agent": {"path": "agent", "contract": "../agent/contract.json"},
  "codex": {"path": "codex", "contract": "../codex/contract.json", "connect": ["$sock"]}
}
JSON

scenario() {
  name=$1; shift
  root=$(cd "$(mktemp -d)" && pwd -P)/.skyra
  echo
  echo "══ $name ══"
  ./bin/skyra boot -root "$root" -registry bin/registry.json "$@" || true
}

scenario "agent: a continuation calls echo, waits, resumes" -to agent -payload '{"text":"hello"}'
scenario "codex (fake) asks for echo, through the actor host" -to codex -payload '{"text":"tool:hello"}'
scenario "codex (fake) runs a native command: fenced" -to codex -payload '{"text":"taint-child"}'
scenario "echo hello"                      -to echo -payload '{"text":"hello"}'
scenario "echo probes .skyra/ from inside" -to echo -payload "{\"text\":\"hello\",\"probe\":\"$(cd "$(mktemp -d)" && pwd -P)/.skyra\"}"
scenario "echo (a leaf) addresses echo"    -to echo -payload '{"text":"hello","misbehave":"yes"}'
scenario "payload breaks echo's contract"  -to echo -payload '{"txt":"hello"}'
