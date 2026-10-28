#!/bin/bash
# Scenario equivalence (41 + 42 build plan, V8): every controller/run.sh
# scenario commits byte-identical trees under two builds of skyra.
#
#   scripts/equivalence.sh <old-rev> [<new-rev>]
#
# <new-rev> defaults to the working tree. Targets, contracts and the registry
# are built once, from the working tree, and shared, so only the controller
# differs. Each scenario boots with a fixed genesis nonce under the same root
# path, once per build, and the two trees are compared file by file.
# A scenario that differs from itself under the old build (two runs) is
# reported as nondeterministic and compared by shape only: which files exist
# under which kinds, and how many.
#
# CHECK_SYNC=1 also checks the new build's durability policy (40; 41 V5)
# from its SKYRA_SYNC_TRACE: no full flush except a forced one, a forced
# flush in each process immediately before every target run, resident start
# and park, and one before boot prints the answer.
set -euo pipefail
ROOT=$(cd "$(dirname "$0")/.." && pwd)
old=${1:?usage: equivalence.sh <old-rev> [<new-rev>]}
new=${2:-}
work=$(mktemp -d /private/tmp/skyra-eq-XXXXXX)
trap 'kill ${host:-} 2>/dev/null || true; rm -rf "$work"' EXIT
export SKYRA_GENESIS_NONCE=00112233445566778899aabbccddeeff

build_skyra() { # rev dir
  if [ -z "$1" ]; then
    (cd "$ROOT" && go build -trimpath -o "$2/skyra" ./controller/skyra)
  else
    mkdir -p "$work/src-$1"
    git -C "$ROOT" archive "$1" | tar -x -C "$work/src-$1"
    (cd "$work/src-$1" && go build -trimpath -o "$2/skyra" ./controller/skyra)
  fi
}
mkdir -p "$work/bin" "$work/old" "$work/new"
build_skyra "$old" "$work/old"
build_skyra "$new" "$work/new"
c="$ROOT/controller"
(cd "$ROOT" && for t in echo:targets/echo agent:targets/agent codex:codex/adapter host:hosts/actor fake-codex:codex/fake read:targets/read; do
  go build -trimpath -o "$work/bin/${t%%:*}" "./controller/${t#*:}"
done)
(cd "$c/interp" && go build -trimpath -o "$work/bin/interp" .)

look="$work/look"
mkdir -p "$look/sub"
printf 'alpha\n' > "$look/a.txt"; printf 'beta\n' > "$look/b.txt"; printf 'gamma\n' > "$look/sub/c.txt"
sock=/private/tmp/skyra-eq-$$.sock
cat > "$work/bin/registry.json" <<JSON
{
  "echo":  {"path": "echo",  "contract": "$c/targets/echo/contract.json"},
  "agent": {"path": "agent", "contract": "$c/targets/agent/contract.json"},
  "interpreter": {"path": "interp", "contract": "$c/interp/examples/greet-contract.json", "state": true},
  "read":  {"path": "read",  "contract": "$c/targets/read/contract.json", "authority": {"read": [{"path": "$look", "scope": "subpath"}]}},
  "codex": {"path": "codex", "contract": "$c/codex/adapter/contract.json", "connect": ["$sock"]}
}
JSON

start_host() {
  rm -rf "$work/host-state" "$sock"; mkdir -p "$work/host-state"
  "$work/bin/host" -sock "$sock" -state "$work/host-state" -- "$work/bin/fake-codex" 2>/dev/null &
  host=$!
  for _ in $(seq 100); do [ -S "$sock" ] && return; sleep 0.05; done
  echo "actor host did not start" >&2; exit 1
}
stop_host() { kill "$host" 2>/dev/null || true; wait "$host" 2>/dev/null || true; host=; }

# dump lists every file in a tree with its hash, temp files and logs excluded.
dump() {
  (cd "$1" && find . -type f ! -name '.tmp-*' ! -name '*.log' -print0 | sort -z | xargs -0 shasum -a 256 | awk '{print $2, $1}')
}
shape() { sed -E 's/[0-9a-f]{64}/H/g; s/ H$//' | sort | uniq -c; }

# run_one build scenario-args...: a fresh root at the same path, then the boot.
root="$work/root/.skyra"
run_one() {
  local build=$1; shift
  rm -rf "$work/root"; mkdir -p "$work/root"
  start_host
  if [ "$1" = interp ]; then shift; "$work/bin/interp" install "$root/interpreter" "$c/interp/examples/greet.json" >/dev/null; fi
  SKYRA_SYNC_TRACE=${CHECK_SYNC:-} "$work/$build/skyra" boot -root "$root" -registry "$work/bin/registry.json" "$@" >"$work/out" 2>&1 || true
  stop_host
  grep -q '^replay: .*byte-identical$' "$work/out" || { echo "  replay did not pass under $build:"; sed 's/^/    /' "$work/out"; return 1; }
  if [ -n "${CHECK_SYNC:-}" ] && [ "$build" = new ]; then
    python3 -I - "$work/out" <<'PY' >&2 || return 1
import re, sys
last, outward, forced, bad = {}, 0, 0, []
for line in open(sys.argv[1]):
    m = re.match(r"\[([^\]]+)\]\s+(\S+)\s+(.*)", line)
    if not m:
        if line.startswith("lineage") and last.get("boot") != "force":
            bad.append("boot answered before a forced flush")
        continue
    who, what, rest = m.groups()
    if what == "fsync":
        kind = rest.split()[0]
        if kind == "full":
            bad.append(f"{who}: a full flush that is not forced: {rest}")
        forced += kind == "force"
        last[who] = kind
        continue
    if (what == "execute" and "under rendered sandbox" in rest) or what == "park":
        outward += 1
        if last.get(who) != "force":
            bad.append(f"{who}: {what} {rest.strip()} without a forced flush right before it")
    last[who] = what
if forced != outward + 1:
    bad.append(f"{forced} forced flushes for {outward} outward acts and one answer")
for b in bad:
    print("  sync:", b)
sys.exit(1 if bad else 0)
PY
  fi
  dump "$root"
}

failed=0
scenario() {
  local name=$1; shift
  local a b a2
  a=$(run_one old "$@") || { failed=1; return; }
  a2=$(run_one old "$@") || { failed=1; return; }
  b=$(run_one new "$@") || { failed=1; return; }
  if [ "$a" = "$a2" ]; then
    if [ "$a" = "$b" ]; then printf 'same   %s (%d files)\n' "$name" "$(printf '%s\n' "$a" | wc -l)"; return; fi
    printf 'DIFF   %s\n' "$name"; diff <(printf '%s\n' "$a") <(printf '%s\n' "$b") | sed 's/^/    /' | head -20; failed=1
  else
    if [ "$(printf '%s\n' "$a" | shape)" = "$(printf '%s\n' "$b" | shape)" ]; then
      printf 'shape  %s (nondeterministic under the old build; same shape)\n' "$name"; return
    fi
    printf 'DIFF   %s (shape)\n' "$name"; diff <(printf '%s\n' "$a" | shape) <(printf '%s\n' "$b" | shape) | sed 's/^/    /' | head -20; failed=1
  fi
}

scenario "agent calls echo"            -to agent -payload '{"text":"hello"}'
scenario "agent narrows echo"          -to agent -payload '{"text":"narrow:hello"}'
scenario "fan-out"                     -to agent -payload '{"text":"fan:hello,world"}'
scenario "fan-out, one out of reach"   -to agent -payload '{"text":"fan:hello,!shell"}'
scenario "extend-reach"                -to agent -payload '{"text":"extend:hello"}'
scenario "extend-reach refused"        -to agent -payload '{"text":"redundant:"}'
scenario "interpreter procedure"       interp -to interpreter -payload '{"names":["ada","bo"],"greeting":"hi"}'
scenario "remember"                    -to agent -payload '{"text":"remember:hello"}'
scenario "refinement"                  -to agent -payload "{\"text\":\"look:$look\"}"
scenario "codex (fake) tool call"      -to codex -payload '{"text":"tool:hello"}'
scenario "codex (fake) native command" -to codex -payload '{"text":"taint-child"}'
scenario "codex (fake) remembers"      -to codex -payload '{"text":"remember:hello"}'
scenario "echo hello"                  -to echo -payload '{"text":"hello"}'
scenario "echo outside its Horizon"    -to echo -payload '{"text":"hello","misbehave":"yes"}'
scenario "echo contract fault"         -to echo -payload '{"txt":"hello"}'
exit $failed
