# Controller overhead vs inference time

Date: 2026-10-13  
Commit: `55579a1` (main) · Apple M1 Pro, 16 GB · go1.25.0

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

## Not yet measured

These results show that Skyra's overhead is fixed per step. They **do not**
show how that overhead compares with a conventional runtime, and nothing in
this doc should be read as that comparison.

- **Native Codex comparison.** Nobody has timed the same task on Codex run
  directly (`codex` CLI or app-server, no Skyra), so there is no baseline.
  Every number here is measured against the fake app-server. Real Codex adds
  its own costs: process and thread start, network, and tool execution under
  its own sandbox. Without a side-by-side run there is no way to say whether
  ~250 ms per round trip is more or less than what a conventional runtime
  pays for the same calls. To measure it, run identical prompts with tool
  calls through `skyra ask` and through native Codex, with the same model and
  machine, and compare wall time per tool round trip.
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
- A tool-heavy turn pays ~250 ms for each tool round trip. A 10-call agentic
  turn adds ~2.5 s of controller time no matter how fast the model is. That
  is the case to watch as models get faster.
- Replay stays cheap: the byte-identical re-derivation at the end of each run
  is ~12 ms for 4–7 records.
