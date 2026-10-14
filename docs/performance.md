# Performance

Two costs matter for a runtime that runs one process per step and keeps all
state in a record tree:

1. **Running:** the controller overhead each live step adds on top of the model.
2. **Checking:** how long replay takes to re-derive and verify a run.

Short version: the controller adds about **250 ms per model round trip**, a
fixed cost that does not grow with inference time, so it is about 3% of a
10-second model turn. Replay re-verifies about **20,000 records per second**
and scales linearly with the length of a run.

| | Machine | Date | Commit |
|---|---|---|---|
| Controller overhead | Apple M1 Pro, 16 GB, macOS, go1.25.0 | 2026-10-13 | `55579a1` |
| Replay | 2-core 2.8 GHz Xeon, Linux, go1.25.0 | 2026-10-13 | `55579a1` |

The two sets were measured on different machines. Replay numbers on Apple
silicon are likely faster; they have not been rerun there yet.

## Controller overhead vs inference time

### Question

As model inference gets slower, does the controller's overhead stay fixed, or
does it grow with it? Possible causes of growth would be polling, backoff, or
keeping a process idle while the model thinks.

### Answer

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

So the overhead's share shrinks as the model gets slower:

| Inference per model step | `hello` overhead share | `tool:hello` overhead share |
|---|---|---|
| 0 ms (fake, instant) | ~100% | ~100% |
| 250 ms | ~52% | ~51% |
| 1 s | ~21% | ~20% |
| 2 s | ~12% | ~11% |
| 5 s | ~5% | ~5% |
| 10 s (typical real Codex turn) | ~3% | ~3% |

### Method

- The fake Codex app-server (`controller/codex/fake`) was put behind a small
  stdio proxy that holds back each **model output** for `N` ms before
  forwarding it. A model output is an `item/tool/call` or an `item/completed`
  line. Everything else (initialize, thread/start, turn/started, replies)
  passes through immediately. This simulates inference without touching
  controller code.
- The proxy ran under the actor host exactly as `run.sh` sets it up
  (`bin/host -sock … -- latency -ms N ./bin/fake-codex`).
- Each trial was one `skyra boot -to codex -text <scenario>` against a fresh
  `.skyra/` root, timed end to end, including the closing replay check.
- Overhead = wall time − N × (model steps).
- N swept 0, 50, 100 … 450, 500, 1000, 2000, 3000, 4000, 5000, 6000 ms: 120
  trials, 4–10 per cell.

The proxy, for reproduction:

```go
// latency: delays each model output (tool call or agent message) by -ms.
for {
	line, err := r.ReadString('\n')
	if line != "" {
		if strings.Contains(line, `"method":"item/tool/call"`) || strings.Contains(line, `"method":"item/completed"`) {
			time.Sleep(time.Duration(*ms) * time.Millisecond)
		}
		io.WriteString(os.Stdout, line)
	}
	if err != nil {
		break
	}
}
```

### Where the overhead goes

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

At 6000 ms latency the trace is identical, except that the two `execute` lines
each grow by exactly 6000 ms: the extra time is entirely inside the model
wait. The controller does nothing differently while the model thinks. Reads
from the actor host are event-driven (a condition variable), with no polling
interval to miss.

### Caveat: jitter after long waits

At 4–6 s of injected latency, some runs pick up an extra 100–300 ms
(`tool:hello` cells: 4000 ms median 851 ms in one batch, 577–814 ms in a
rerun, 490–584 ms in another). It is intermittent and does not scale with N:
5000 and 6000 ms runs were often a flat ~520 ms. It did not reproduce in the
timestamped runs, and keeping the CPU busy in parallel did not remove it. It
looks like macOS scheduling or wake-up latency after a process sits idle for
several seconds, not controller logic. It is small next to multi-second
inference (<3%).

## Replay

Replay re-derives every record from its stored input with `machine.Step`, and
checks that every stored object hashes to its name. Nothing is re-executed.

Synthetic version-2 runs (agent → echo rounds, each resolve carrying its
round's complete emission), median of 5:

| Records | From memory | From the tree on disk |
|---|---|---|
| 103 | 4.8 ms | 8.2 ms |
| 1,003 | 51 ms | 84 ms |
| 10,003 | 0.50 s | 0.90 s |

That is about 20,000 records/s from memory and 11,000 records/s from disk,
linear in run length.

The runs published on the website (`website/runs/`) replay in under 10 ms
natively and under 40 ms as WebAssembly under Node, the engine the browser
uses. The largest, `remember` (65 records), takes about 9 ms natively and
36 ms as WebAssembly.

Profiling shows SHA-256 takes about a third of replay time and JSON most of
the rest, including `released()` re-marshaling a Horizon for each record.

To reproduce:

```
go test -run '^$' -bench . ./controller/runtime/replay/ ./website/export/
```

## Implications

- Overhead is **per step, not per second of inference**. To reduce it, cut
  steps or per-step cost, not waiting:
  - fewer process launches per round trip
  - fewer fsynced commits per step (intake + gather ≈ 60 ms)
  - faster Horizon lookup
- A tool-heavy turn pays ~250 ms for each tool round trip. A 10-call agentic
  turn adds ~2.5 s of controller time no matter how fast the model is. That is
  the case to watch as models get faster.
- Replay stays cheap: the byte-identical check at the end of each run takes
  ~12 ms for 4–7 records, and a 10,000-record run re-verifies in about half a
  second.

## Not yet measured

- **Skyra against Codex running the same task natively.** The numbers above
  isolate Skyra's own overhead with a fake model; they do not compare it with
  a conventional runtime doing the same work.
- **Replay on Apple silicon**, so both sets come from one machine.
