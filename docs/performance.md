# Performance

Two costs matter for a runtime that runs one process per step and keeps all
state in a record tree:

1. **Running:** the controller overhead each live step adds on top of the model.
2. **Checking:** how long replay takes to re-derive and verify a run.

Short version: against native Codex, Skyra costs **~2 s more startup per
one-shot run and ~0.4 s per tool call**. Both are fixed per step and don't
depend on model speed. Replay re-verifies about **47,000 records per second**
from memory (about 7,800 from disk) and scales linearly with the length of a
run.

Everything here was measured on one machine: Apple M1 Pro, 16 GB, macOS,
go1.25.0, 2026-10-13.

| | Code |
|---|---|
| Controller overhead, fake app-server | `55579a1` |
| Controller overhead, against native Codex (codex-cli 0.160.0, `gpt-6.1-sol` low) | `9049bb7`, clean worktree; controller code identical to `55579a1` |
| Replay | `9049bb7` |

## Controller overhead

The full write-up, with method, per-phase breakdown, where the fsync time
goes, and caveats, is
[`OVERHEAD-VS-INFERENCE.md`](../OVERHEAD-VS-INFERENCE.md). In short:

- **The overhead is fixed per step, not per second of inference.** Against a
  fake app-server with injected latency, wall time tracks model time with a
  slope of 0.99–1.0 per model step, plus a constant of ~250 ms per model round
  trip. About two-thirds of that constant is `F_FULLFSYNC`.
- **Against native Codex (`codex exec`), median of 6 interleaved runs:**

  | Time to answer | Native | Skyra | Gap |
  |---|---|---|---|
  | "Reply with exactly: ok" | 4.39 s | 6.70 s | +2.3 s |
  | Read 3 files, one tool call each | 15.53 s | 18.60 s | +3.1 s |

  Measured directly:
  - Skyra's start to `turn/start` is +2.0 s. It launches a fresh, sandboxed
    resident app-server per `skyra ask`.
  - Each tool call costs Skyra 0.38 s.

  On the tools task those fixed costs come to about 20% (~3.5 s of 15.5 s).
  Model time varies more than that between runs, so with n = 6, treat the
  difference between totals as within noise.

To reproduce: `controller/codex/latency/sweep.sh` (fake path) and
`controller/codex/latency/compare.sh` (native comparison, real model calls).

## Replay

Replay re-derives every record from its stored input with `machine.Step`, and
checks that every stored object hashes to its name. Nothing is re-executed.

Synthetic version-2 runs (agent → echo rounds, each resolve carrying its
round's complete emission), median of 5 on the M1 Pro:

| Records | From memory | From the tree on disk |
|---|---|---|
| 103 | 2.0 ms | 10.2 ms |
| 1,003 | 23 ms | 107 ms |
| 10,003 | 0.21 s | 1.29 s |

That is about 47,000 records/s from memory and 7,800–10,000 records/s from
disk, linear in run length. The first disk iteration at 10,003 records took
4.2 s and the other four took 1.23–1.35 s, so the median leaves it out.

The first draft measured these on a 2-core Xeon under Linux:
- from memory: 4.8 ms, 51 ms and 0.50 s
- from disk: 8.2 ms, 84 ms and 0.90 s

The M1 Pro is about 2.3× faster from memory but slower from disk. The disk
gap wasn't investigated.

The runs published on the website (`website/runs/`) replay natively in under
4 ms each on the M1 Pro. The largest, `remember` (65 records), takes 3.7 ms,
and the rest take 0.12–0.33 ms. As WebAssembly under Node, the engine the
browser uses, they took under 40 ms on the Xeon (`remember`: 36 ms). The
WebAssembly runs aren't part of the benchmark command and weren't rerun here.

Profiling on the Xeon showed SHA-256 taking about a third of replay time and
JSON most of the rest, including `released()` re-marshaling a Horizon for
each record.

To reproduce:

```
go test -run '^$' -bench . -count 5 ./controller/runtime/replay/ ./website/export/
```

## Implications

- Overhead is **per step, not per second of inference**. To reduce it, cut
  steps or per-step cost, not waiting:
  - fewer process launches per round trip
  - cheaper flushes (`F_FULLFSYNC` is about two-thirds of the fake path's
    overhead; see
    [proposal 40](next/proposal/40-durability-policy-proposal.md))
  - faster Horizon lookup
- A tool-heavy turn pays for each tool round trip: ~250 ms on the fake path
  and ~380 ms with real Codex. A 10-call agentic turn adds ~4 s of controller
  time no matter how fast the model is. That is the case to watch as models
  get faster.
- For one-shot runs, the ~2 s resident startup is the largest single cost
  against native Codex.
- Replay stays cheap: the byte-identical check at the end of each run takes
  ~12 ms for 4–7 records, and a 10,000-record run re-verifies in ~0.2 s from
  memory, or ~1.3 s from disk.

## Not yet measured

- **Startup across a conversation.** A held resident should take most of the
  ~2 s startup off every turn after the first. That wasn't timed.
- **Native per-tool cost**, so the per-call comparison is Skyra's 0.38 s
  against an estimate.
- **Other models and longer tasks**: higher effort, 10+ tool calls, and
  Claude through its adapter.
- **WebAssembly replay on Apple silicon.**
