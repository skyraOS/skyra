# Per-turn startup behind the service host

Date: 2026-10-28 · Apple M1 Pro, 16 GB · go1.25.0

| Measurement | Code | Notes |
|---|---|---|
| Native comparison | `d6eb32a` | First measurement after the service host (49, `b283d8b`) and the named workspace (`1c0079c`). codex-cli 0.160.0, `gpt-6.1-sol` (low) |
| Phase split | `d6eb32a` | The same 12 Skyra runs, read from their records by `controller/codex/latency/phases.py` |

This replaces the startup and per-call figures in
[`OVERHEAD-VS-INFERENCE.md`](OVERHEAD-VS-INFERENCE.md), measured at
`9049bb7` on the resident design that 49 removed. Its fake-path sweep and
the replay figures in [`performance.md`](performance.md) were not rerun and
still stand.

## Question

The service host changed what a turn costs:
- Before, `skyra ask` launched a resident app-server per run, and a
  conversation held one between turns. Startup was +2.0 s once per run.
  Later turns of a conversation were never timed.
- Now every turn opens a fresh Codex session through `service-io`, and the
  service host launches it. So startup is paid on **every turn**, in chat as
  well as `ask`.

What does a turn's startup cost now, and where does it go?

## Answer

**About 1.15 s to `turn/start`, roughly 0.9 s more than native Codex.** That
is about half the old +2.0 s, but it is now paid per turn. Tool calls got
cheaper: **0.18 s each**, against 0.38 s before.

| Median of 6 (range) | Native Codex | Skyra | Skyra − native |
|---|---|---|---|
| **plain**: time to answer | 3.90 s (3.01–4.26) | 4.39 s (4.18–4.93) | **+0.5 s** |
| └ start → `turn/start` | 0.28 s (0.16–1.33) | 1.16 s (1.08–1.35) | +0.9 s |
| **tools**: time to answer | 10.42 s (9.13–12.36) | 13.99 s (12.78–14.79) | **+3.6 s** |
| └ start → `turn/start` | 0.15 s (0.14–0.31) | 1.13 s (1.04–1.44) | +1.0 s |
| └ per tool call (n=18) | not separately measured | **0.18 s** (0.13–0.36) | |
| &nbsp;&nbsp;└ first call / later calls | | 0.34 s / 0.17 s | |

Against 2026-10-13 (`9049bb7`):

| | Then | Now |
|---|---|---|
| Skyra start → `turn/start` | 2.30–2.44 s | **1.15 s** |
| Gap to native at `turn/start` | +2.0 s | **+0.9 s** |
| Paid | once per `ask`, once per conversation | **every turn** |
| Per tool call | 0.38 s | **0.18 s** |

## Where the 1.15 s goes

Medians of all 12 Skyra runs, from their records:

| Phase | What happens | Time |
|---|---|---|
| to open | Controller boot, genesis, think, narrowing, the adapter's first step, up to the `codex.open` invoke | **0.30 s** (0.27–0.35) |
| open | `codex.open`: the service host renders the session's profile and proxy and launches the app-server | **0.52 s** (0.45–0.56) |
| handshake | initialize, initialized and `thread/start`, each a recorded exchange through `service-io` | **0.34 s** (0.30–0.57) |
| | **start → `turn/start`** | **1.15 s** (1.04–1.44) |

- **The open got faster.** The old resident launch took 1.2–1.9 s. The
  service host does the same work (Seatbelt profile, isolated Codex home,
  egress proxy) in about 0.5 s.
- **The kernel's share is about half.** "To open" and the handshake are
  kernel steps: each is a record, a process and a commit. Together they are
  0.64 s of the 1.15 s. Native does its own handshake inside its 0.15–0.28 s.
- **Tool calls.** The first call costs 0.34 s and later ones about 0.17 s.
  The likely cause of the drop from 0.38 s is the durability policy (40, 41
  V5, built 2026-10-16), which cut the fake path's round trip from ~218 ms
  to ~76 ms. Its native comparison was never rerun until now.

## What taking the kernel out of the hot path should buy

Proposal [50](next/proposal/50-write-policies-proposal.md) (tabled) takes the
kernel out of the loop: it hands the apprentice a ready process and is
entered again only on a trap (`exit()`, a commit point, a policy decision).
Writes are folded per region at commit points instead of committed per step.

On these numbers, that targets the kernel's share of every turn:
- **Startup:** the 0.30 s up to the open, and most of the 0.34 s handshake.
  The handshake would then cost about what native's does.
- **Tool calls:** most of the 0.17–0.34 s each, which is mostly records and
  commits, not the `read` itself.
- **Not the open.** The service host's ~0.5 s launch is outside the kernel.
  Only reusing or pre-starting a session would remove it.

So a turn could plausibly start in about 0.6–0.7 s, and tool calls cost tens
of milliseconds. **This is an estimate from the phase split, not a
measurement.** Nothing in 50 is built.

## Method

Same as [the 10-13 comparison](OVERHEAD-VS-INFERENCE.md#against-native-codex):
- **Prompts.**
  - **plain:** "Reply with exactly: ok"
  - **tools:** read three one-line files by absolute path, one tool call
    each, then reply with their contents
- **Rounds.** 6 interleaved rounds of native `codex exec` and `skyra ask`,
  24 real model runs. Every answer was correct, and every tools run made 3
  calls.
- **Native timing.** From process start, using `--json` event times. The
  ~1 s exit is left out.
- **Skyra timing.**
  - Total: from the run's first state write to its last, so the `go build`
    is left out.
  - Start: up to the record that delivers `turn/start`.
  - Per call: from the observation's `observed_at` to the record that
    delivers the result.
- **Phases.** From record mtimes:
  - `codex.open`'s invoke
  - its return, which carries the incarnation and pid
  - the `codex.exchange` invoke carrying `turn/start`

Reproduce:

```sh
controller/codex/latency/compare.sh 6        # real model calls
python3 controller/codex/latency/phases.py ~/.skyra/runs/<run>/.skyra
```

## Caveats

- **n = 6.** Native tools runs spread over 3.2 s. The two runtimes show the
  model different prompts and tools, so the +3.6 s tools gap is mostly model
  time. The fixed costs come to about 1.5 s of it: +1.0 s startup and
  3 × 0.18 s per call.
- **Native's first plain run** reached `turn.started` at 1.33 s, against
  0.16–0.93 s for the rest. The median ignores it.
- **`ask` only.** Chat turns open the same way, but they were not timed
  separately.

## Not yet measured

- **A chat turn end to end**, including the conversation's own steps.
- **Claude through its adapter**, on the same two prompts.
- **Native per-tool cost**, so the per-call comparison is still against an
  estimate.
- **Longer tasks:** ten or more tool calls, or higher effort.
