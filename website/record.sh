#!/bin/sh
# Records the runs the Verify page replays, and exports each as a bundle in
# runs/. Every run is a `skyra boot` under the real sandbox, against the
# deterministic fake Codex app-server (controller/codex/fake), so recording
# needs no login and costs nothing. Real Codex runs can be added the same way.
#
# The binaries are staged under /private/tmp/skyra-site, so the paths stored
# in each run's genesis record name that directory and not this machine's
# home. Usage: website/record.sh [date]   (default: today)
set -eu
here=$(cd "$(dirname "$0")" && pwd)
repo=$(cd "$here/.." && pwd)
ctl=$repo/controller
day=${1:-$(date +%Y-%m-%d)}
stage=/private/tmp/skyra-site
rm -rf "$stage"
mkdir -p "$stage/bin" "$stage/runs"

cd "$ctl"
go build -o "$stage/bin/skyra" ./skyra
go build -o "$stage/bin/echo" ./targets/echo
go build -o "$stage/bin/agent" ./targets/agent
go build -o "$stage/bin/codex" ./codex/adapter
go build -o "$stage/bin/host" ./hosts/actor
go build -o "$stage/bin/fake-codex" ./codex/fake
(cd interp && go build -o "$stage/bin/interp" .)

sock=$stage/fake.sock
"$stage/bin/host" -sock "$sock" -state "$(mktemp -d)" -- "$stage/bin/fake-codex" 2>/dev/null &
host=$!
trap 'kill $host 2>/dev/null || true' EXIT
while [ ! -S "$sock" ]; do sleep 0.05; done

cat > "$stage/bin/registry.json" <<JSON
{
  "echo":  {"path": "echo",  "contract": "$ctl/targets/echo/contract.json"},
  "agent": {"path": "agent", "contract": "$ctl/targets/agent/contract.json"},
  "interpreter": {"path": "interp", "contract": "$ctl/interp/examples/greet-contract.json", "state": true},
  "codex": {"path": "codex", "contract": "$ctl/codex/adapter/contract.json", "connect": ["$sock"]}
}
JSON

# run NAME TITLE NOTE -- boot arguments
run() {
  name=$1 title=$2 note=$3; shift 3
  root=$stage/runs/$name/.skyra
  mkdir -p "$stage/runs/$name"
  echo "══ $name"
  # A run that faults still commits its records; boot's exit status is not the test, replay is.
  perl -e 'alarm 60; exec @ARGV' "$stage/bin/skyra" boot -root "$root" -registry "$stage/bin/registry.json" "$@" \
    | sed -n '/^lineage/,$p' || true
  (cd "$repo" && go run ./website/export -root "$root" -name "$name" -title "$title" -note "$note" \
    -recorded "$day" -o "website/runs/$name.json")
}

run answered "A question, answered with a tool" \
  "Codex (the deterministic fake app-server) is asked a question, calls echo as a tool inside its Boundary, and answers with what echo returned. Every hop is a record." \
  -to codex -payload '{"text":"tool:hello"}'

run outside "A call outside the Boundary, faulted at intake" \
  "The target emits a call to something its Boundary does not reach: the shape of what a prompt injection asks for. Intake derives a boundary fault from the bytes; nothing outside runs." \
  -to echo -payload '{"text":"hello","misbehave":"yes"}'

run probe "A probe of .skyra/ from inside, denied by the OS" \
  "The target tries to read the runtime's own state directory. The sandbox profile generated from its Boundary denies it: an actor sees only its own state, and echo has none. The probe's failure is in the record." \
  -to echo -payload "{\"text\":\"hello\",\"probe\":\"$stage/runs/probe/.skyra\"}"

run contract "A payload that breaks its contract" \
  "The call names echo, which is in reach, but its payload does not fit echo's input contract. Intake derives a contract fault; echo never starts." \
  -to echo -payload '{"txt":"hello"}'

run continuation "A continuation that waits and resumes" \
  "An agent calls echo and waits. When echo's slot is discharged, a resolve record resumes the agent, which returns its own answer." \
  -to agent -payload '{"text":"hello"}'

run remember "Work remembered as a procedure, run with no agent" \
  "An agent does a task: it calls echo twice in one emission, then once more. The work is remembered: the interpreter reads the trace, compiles it, and adds the contract to itself with no approval. Called on new input, the procedure runs with no agent in its frame. From a narrowed Boundary without echo it faults at intake; the same contract is not remembered twice; input outside its refinement goes back to the agent. The agent is scripted, not a model, and says so." \
  -to agent -payload '{"text":"remember:hello"}'

# The index the Verify page lists, in this order.
cd "$here/runs"
printf '[' > index.json
sep=
for n in remember answered outside probe contract continuation; do
  printf '%s"%s"' "$sep" "$n" >> index.json
  sep=,
done
printf ']\n' >> index.json
echo "wrote website/runs/index.json"
