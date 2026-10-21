#!/bin/bash
# Native Codex against Skyra on real Codex: the same model, effort and
# prompts, interleaved, TRIALS rounds. See docs/OVERHEAD-VS-INFERENCE.md.
#
#   controller/codex/latency/compare.sh 6
#
# native: start = to turn.started, answer = to the final agent message, both
#         from the process start; the ~1 s codex exec takes to exit is left out.
# skyra:  from the run's state writes (run.py); the build `skyra ask` does
#         first is left out. Tool paths are absolute (see the doc).
set -u
cd "$(dirname "$0")"
here=$PWD repo=$(cd ../../.. && pwd)
ws=$(cd "$(mktemp -d)" && pwd -P)
printf 'alpha\n' > "$ws/a.txt"; printf 'beta\n' > "$ws/b.txt"; printf 'gamma\n' > "$ws/c.txt"
model=${MODEL:-gpt-6.1-sol} effort=${EFFORT:-low}
plain="Reply with exactly: ok"
tools="Read $ws/a.txt, then $ws/b.txt, then $ws/c.txt, with a separate tool call for each file, one after another (not in parallel and not in one command). Then reply with only their three contents joined by commas."
for i in $(seq 1 "$1"); do
  for p in plain tools; do
    text=${!p}
    (cd "$ws" && codex exec -m "$model" -c model_reasoning_effort="$effort" --json --skip-git-repo-check --ephemeral -s read-only "$text" </dev/null 2>/dev/null) |
      P=$p perl -MTime::HiRes=time -ne 'BEGIN{$t=time} $d=time-$t; $s=$d if /"type":"turn.started"/; $n++ if /item.completed.*command_execution/; if(/agent_message.*"text":"([^"]*)"/){$a=$1;$m=$d} END{printf "native %s start=%.3f answer=%.3f calls=%d %s\n",$ENV{P},$s,$m,$n,$a}'
    out=$(cd "$ws" && "$repo/skyra" ask "$text" 2>&1)
    run=$(printf '%s\n' "$out" | sed -n 's/.*state: \(.*\)$/\1/p' | tail -1)
    echo "skyra $p $(python3 "$here/run.py" "$run") $(printf '%s\n' "$out" | grep '^answer:' | tail -1)"
  done
done
