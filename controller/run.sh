#!/bin/sh
# The stateless controller, end to end. Each scenario gets a fresh .skyra/
# state directory and ends with a replay check.
set -eu
cd "$(dirname "$0")"
mkdir -p bin
go build -o bin/skyra ./skyra
go build -o bin/echo ./targets/echo
go build -o bin/agent ./targets/agent
go build -o bin/codex ./codex/adapter
go build -o bin/host ./hosts/actor
go build -o bin/fake-codex ./codex/fake
go build -o bin/read ./targets/read
# The interpreter is its own module (next/32): the controller stays standard-library Go.
(cd interp && go build -o ../bin/interp .)

# The actor host in front of a deterministic fake Codex app-server (codex/fake/).
# A unix socket path must be short, and real (the sandbox matches real paths).
sock=/private/tmp/skyra-fake-$$.sock
./bin/host -sock "$sock" -state "$(mktemp -d)" -- ./bin/fake-codex 2>/dev/null &
host=$!
trap 'kill $host 2>/dev/null' EXIT
while [ ! -S "$sock" ]; do sleep 0.05; done

# A directory read may look at: two small files and a subdirectory (next/32, H7).
look=$(cd "$(mktemp -d)" && pwd -P)
printf 'alpha\n' > "$look/a.txt"
printf 'beta\n' > "$look/b.txt"
mkdir "$look/sub" && printf 'gamma\n' > "$look/sub/c.txt"

# The operator's registry: every surface, its binary, and its contract.
cat > bin/registry.json <<JSON
{
  "echo":  {"path": "echo",  "contract": "../targets/echo/contract.json"},
  "agent": {"path": "agent", "contract": "../targets/agent/contract.json"},
  "interpreter": {"path": "interp", "contract": "../interp/examples/greet-contract.json", "state": true},
  "read":  {"path": "read",  "contract": "../targets/read/contract.json", "authority": {"read": [{"path": "$look", "scope": "subpath"}]}},
  "codex": {"path": "codex", "contract": "../codex/adapter/contract.json", "connect": ["$sock"]}
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
scenario "agent narrows echo through its Horizon, then calls it" -to agent -payload '{"text":"narrow:hello"}'
# Fan-out (next/32, H2): one emission, two instructions, one round of two slots.
scenario "fan-out: the agent asks echo twice in one emission" -to agent -payload '{"text":"fan:hello,world"}'
scenario "fan-out: one instruction out of reach faults, its sibling runs" -to agent -payload '{"text":"fan:hello,!shell"}'
# The Horizon's syscalls (next/32, H3): the agent adds a contract to itself, then calls itself through it.
scenario "extend-reach: the agent adds a contract to itself and is invoked through it" -to agent -payload '{"text":"extend:hello"}'
scenario "extend-reach: a contract the actor already holds is refused" -to agent -payload '{"text":"redundant:"}'
# The interpreter (next/32, H5): a procedure installed by hand in its own state runs through the
# kernel, its fork and its each fanned out in one emission. The kernel sees only instructions.
root=$(cd "$(mktemp -d)" && pwd -P)/.skyra
echo
echo "══ interpreter: a hand-installed procedure fans out through the kernel ══"
./bin/interp install "$root/interpreter" interp/examples/greet.json
./bin/skyra boot -root "$root" -registry bin/registry.json -to interpreter -payload '{"names":["ada","bo"],"greeting":"hi"}' || true
# The thesis demo (P9; next/32, H6, H8): a worker remembers the agent's strand, then calls the
# procedure it became: no model step, no approval. From a Horizon without echo it faults at intake;
# the same contract is not remembered twice; an actor cannot extend itself with what it holds.
scenario "remember (P9): the agent's strand becomes a procedure, run with no agent" -to agent -payload '{"text":"remember:hello"}'
./bin/skyra metric "$root"
# Running what was remembered (next/32, H7): inside its base types the procedure runs alone;
# a listing where a file was read is outside them, and the agent resumes at that edge.
scenario "refinement: a procedure runs inside its base types, and falls back at the edge that refused" -to agent -payload "{\"text\":\"look:$look\"}"
./bin/skyra metric "$root"
scenario "codex (fake) asks for echo, through the actor host" -to codex -payload '{"text":"tool:hello"}'
scenario "codex (fake) runs a native command: fenced" -to codex -payload '{"text":"taint-child"}'
# Codex remembers (next/32, H8): remember is in its Horizon's reach, so it is offered. Codex has the
# agent relay, remembers that call, and calls the procedure it became: no agent runs for it.
scenario "codex (fake) remembers a call of its own and calls the procedure" -to codex -payload '{"text":"remember:hello"}'
./bin/skyra metric "$root"
scenario "echo hello"                      -to echo -payload '{"text":"hello"}'
scenario "echo probes .skyra/ from inside" -to echo -payload "{\"text\":\"hello\",\"probe\":\"$(cd "$(mktemp -d)" && pwd -P)/.skyra\"}"
scenario "echo addresses what is outside the Horizon" -to echo -payload '{"text":"hello","misbehave":"yes"}'
scenario "payload breaks echo's contract"  -to echo -payload '{"txt":"hello"}'
