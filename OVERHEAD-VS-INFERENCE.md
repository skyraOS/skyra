# Controller overhead vs inference time

Date: 2026-10-13  
Commit: `55579a1` (main) · Apple M1 Pro, 16 GB · go1.25.0 · native comparison:
codex-cli 0.160.0, `gpt-6.1-sol` (low)

## Question

As model inference gets slower, does the stateless controller's overhead stay
fixed, or does it grow with it? Possible causes of growth would be polling,
backoff, or keeping a process idle while the model thinks.

## Answer

**The overhead is a fixed cost per step and does not grow with inference
time.** Wall time is inference time plus a constant:

| Scenario | Model steps | Records | Fit: wall time vs injected latency | Median overhead |
|---|---|---|---|---|
| `hello` (model answers) | 1 | 4 | slope **0.99**, intercept 337 ms | **~270 ms** |
| `tool:hello` (model calls `echo`, then answers) | 2 | 7 | slope **2.03**, intercept 552 ms | **~515 ms** |

A slope of 1.0 per model step means each millisecond of inference costs one
millisecond of wall time, with no multiplier. The intercept is the controller.
Each model round trip costs roughly **250 ms** of overhead. Each record costs
roughly **70 ms**, because a record is one `skyra` process run once.

So the overhead fraction shrinks as the model gets slower:

| Inference per model step | `hello` overhead share | `tool:hello` overhead share |
|---|---|---|
| 0 ms (fake, instant) | ~100% | ~100% |
| 250 ms | ~52% | ~51% |
| 1 s | ~21% | ~20% |
| 2 s | ~12% | ~11% |
| 5 s | ~5% | ~5% |
| 10 s | ~3% | ~3% |

Against the fake app-server, the controller adds about 0.25 s per model round trip,
however long the model takes.

## Method

- The fake Codex app-server (`controller/codex/fake`) was put behind a small
  stdio proxy that holds back each **model output** for `N` ms before
  forwarding it. A model output is an `item/tool/call` or an `item/completed`
  line. Everything else (initialize, thread/start, turn/started, replies)
  passes through immediately. This simulates inference without touching
  controller code.
- The proxy ran under the actor host exactly as `run.sh` sets it up
  (`bin/host -sock … -- latency -ms N ./bin/fake-codex`).
- Each trial was one `skyra boot -to codex -text <scenario>` against a fresh
  `.skyra/` root. It was timed end to end, including the closing replay
  check.
- Overhead = wall time − N × (model steps). These are the steps in which
  the model produces output.
- N swept 0, 50, 100 … 450, 500, 1000, 2000, 3000, 4000, 5000, 6000 ms.
  That came to 120 trials, with 4–10 per cell.

To reproduce, use the proxy in
[`controller/codex/latency/`](controller/codex/latency/) and the sweep script
beside it. The script builds the binaries, runs one actor host per latency,
and prints `ms scenario wall_ms records answer` for each trial:

```sh
controller/codex/latency/sweep.sh 6 0 250 500 1000 2000 4000
SCEN=tool:hello controller/codex/latency/sweep.sh 4 50 100 150 200 250 300 350 400 450 3000 5000 6000
```

Overhead for each line is `wall_ms − ms × k`, where `k` is 1 for `hello`
and 2 for `tool:hello`.

## Where the overhead goes

Each line of the controller log was timestamped for a `tool:hello` run with
0 ms latency (501 ms total). Every millisecond is accounted for:

| Phase (per step) | Typical |
|---|---|
| process start → `genesis` | ~30 ms |
| `intake` (commit record) | 25–35 ms |
| `release` (Horizon lookup) | 10–40 ms |
| `execute` adapter → emitted (spawn Codex adapter, host round trip) | 30–70 ms |
| `execute` echo → emitted | ~17 ms |
| `settle` | ~15 ms |
| `gather` (commit resolve) | ~30 ms |
| `claim` | ~10 ms |
| final replay check (7 records, byte-identical) | ~12 ms |

At 6000 ms latency the same trace is identical. The only change is that the
two `execute` lines grow by exactly 6000 ms each: the extra time is entirely
inside the model wait. The controller does nothing differently while the
model thinks. Reads from the actor host are event-driven (a condition
variable), with no polling interval to miss.

## One caveat: jitter after long waits

At 4–6 s of injected latency, some runs pick up an extra 100–300 ms
(`tool:hello` cells: 4000 ms median 851 ms in one batch, 577–814 in a rerun,
490–584 in another). It is intermittent and doesn't scale with N: 5000 and
6000 ms runs were often a flat ~520 ms. It also didn't reproduce in the
timestamped runs. Keeping the CPU busy in parallel didn't remove it. This
looks like macOS scheduling or wake-up latency after a process sits idle for
several seconds, not controller logic. It is small next to multi-second
inference (<3%).

## Against native Codex

Everything above was measured against the fake app-server. This section
times the same tasks on real Codex two ways: through Skyra (`skyra ask`), and
run directly (`codex exec`), which is the conventional runtime.

**Setup.** codex-cli 0.160.0, `gpt-6.1-sol` at `low` effort for both. `low`
is the default Skyra's isolated Codex picks; native was pinned to match with
`-m`/`-c`. Same machine. Two prompts:
- **plain:** "Reply with exactly: ok"
- **tools:** read three one-line files with a separate tool call each, then
  reply with their contents

The two runtimes were interleaved for 6 rounds, 24 real model runs in all.
Every answer was correct, and both made exactly 3 tool calls in every tools
run: native through its shell (`cat`), Skyra through its `read` target. The
file paths in the prompt are absolute (see the note at the end of this
section).

**What is timed.** Both measure time to the answer.
- **Native:** time from process start to the final agent message, from
  `--json` event times. Native's ~1 s to exit after answering is left out.
- **Skyra:** time from the run's first state write to its last, from file
  times. The `go build` that `skyra ask` does first is left out. The ~0.2–0.5
  s Skyra spends after `turn/completed` (resolve, claim, replay check) is
  counted.
- **Skyra per tool call:** from when the holder observed Codex's
  `item/tool/call` (`observed_at`) to the record that delivers the result
  back to the app-server.

| Median of 6 (range) | Native Codex | Skyra | Skyra − native |
|---|---|---|---|
| **plain**: time to answer | 4.00 s (3.66–4.97) | 6.71 s (5.80–8.80) | **+2.7 s** |
| └ start → `turn/start` | 0.54 s | 2.46 s | +1.9 s |
| **tools**: time to answer | 15.87 s (13.99–21.18) | 17.40 s (15.79–18.56) | **+1.5 s** |
| └ start → `turn/start` | 0.30 s | 2.38 s | +2.1 s |
| └ per tool call (n=18) | not separately measured | **0.38 s** (0.34–0.64) | |

**Reading it.**
- **Startup is the largest fixed cost: about +2 s per `skyra ask`.**
  - Native `codex exec` reaches `turn.started` in 0.3–0.5 s.
  - Skyra launches a fresh resident app-server under its Seatbelt profile,
    with an isolated Codex home and an egress proxy. This takes ~1.2–1.9 s
    (one run took 3.3 s).
  - Skyra then spends ~0.4–1.1 s on initialize and `thread/start`, each a
    recorded exchange.
- **Each tool call costs ~0.38 s of Skyra time.** This is more than the
  ~0.25 s per round trip on the fake path. With the resident, a tool call is
  about 13 records (73 records for the tools run against 34 for plain): the
  `read` step plus the recorded exchanges with the app-server. Native's
  matching cost, running `cat` in its own sandbox, didn't get a separate
  timing because `codex exec --json` doesn't timestamp the start of a
  command. It is likely tens of milliseconds.
- **Model time dominates, and its spread is wider than the gap.** Native
  tools runs ranged over 7 s. For tools, Skyra's model-plus-wind-down portion
  was actually shorter (median 13.5 s against 15.6 s). The two runtimes don't
  show the model the same thing:
  - Skyra replaces Codex's base prompt and offers its own tools.
  - Native carries its default prompt (~14k input tokens on the plain run)
    and makes the model write shell commands.

  So the total "tools" gap of +1.5 s is within noise. The fixed costs above
  come to ~3.5 s for this task: +2.1 s startup, 3 × 0.38 s for the tool
  calls, and ~0.3 s wind-down. Those are the reliable numbers, not the gap
  between totals.

So against a conventional runtime, Skyra costs about **2 s more to start** a
one-shot run and about **0.4 s per tool call**. On a 4 s answer that is
+70%. On a 15–20 s tool-using task it is 10–20%, and the model's own
variance is larger.

Reproduce with:

```sh
controller/codex/latency/compare.sh 6   # interleaved rounds; real model calls
```

`compare.sh` prints one line per run. `run.py` beside it times a finished
`skyra ask` run from its `.skyra/` state.

**Note on absolute paths.** With the current uncommitted `models` target in
the launcher's registry, the Horizon's first read grant is
`~/.claude/cache/model-catalog` rather than the workspace. `read` resolves
relative paths against the first grant, so `a.txt` came back `not_found`.
Absolute paths sidestep this and are equally fair to both runtimes.

## Not yet measured

The native comparison above covers one-shot runs of two small tasks on one
model. These are still open:

- **Startup across a conversation.** `skyra ask` launches a resident for
  every run. In a conversation the resident is held between turns, which
  should take most of the ~2 s startup off every turn after the first.
  That wasn't timed.
- **Native per-tool cost.** Native's cost per tool call wasn't isolated, so
  the per-call comparison is Skyra's 0.38 s against an estimate, not a
  measurement.
- **Other models and longer tasks.** One model, at `low` effort, with 0 or 3
  tool calls. Ten or more tool calls, higher effort, or Claude through its
  adapter weren't run.
- **Replay on Apple silicon.** The only replay number here is incidental:
  the ~12 ms byte-identical check at the end of each run, over 4–7 records.
  No dedicated replay benchmark was run. In particular, this sweep did not
  measure how `skyra replay` scales with lineage length (hundreds or
  thousands of records, long conversations) or where its time goes.

## Implications

- Overhead is **per step, not per second of inference**. To reduce it, cut
  steps or per-step cost, not waiting. Ways to do that:
  - fewer process launches per round trip
  - fewer fsynced commits per step (intake + gather ≈ 60 ms)
  - faster Horizon lookup
- A tool-heavy turn pays for each tool round trip: ~250 ms on the fake path
  and ~380 ms with real Codex through the resident. A 10-call agentic turn
  adds ~4 s of controller time no matter how fast the model is. That is the
  case to watch as models get faster.
- For one-shot runs, the ~2 s resident startup is the largest single cost
  against native Codex. Reusing a resident, or starting one ahead of time,
  would close most of that gap.
- Replay stays cheap: the byte-identical re-derivation at the end of each run
  is ~12 ms for 4–7 records.
