# Controller overhead vs inference time

Date: 2026-10-13 · Apple M1 Pro, 16 GB · go1.25.0

> **A snapshot at `55579a1` / `9049bb7`.** It measures the resident design,
> which 49 removed (`b283d8b`). The startup and per-call figures against
> native Codex are superseded by
> [`STARTUP-PER-TURN.md`](STARTUP-PER-TURN.md), measured at `d6eb32a`. The
> fake-path sweep and its fsync analysis were not rerun. Bug #41 below was
> fixed in `1c0079c`. Read the code these numbers describe with
> `git show 9049bb7:<path>`.

| Measurement | Code | Notes |
|---|---|---|
| Fake-path sweep | `55579a1` | It builds only `controller/` binaries, so the uncommitted `models` target and launcher change were never part of it |
| Native comparison | `9049bb7`, clean worktree | Controller code identical to `55579a1`. Only the latency tools and the benchmarks were added. codex-cli 0.160.0, `gpt-6.1-sol` (low) |
| Replay | `9049bb7` | `go test -run '^$' -bench . -count 5 ./controller/runtime/replay/ ./website/export/`; results in [`docs/performance.md`](performance.md#replay) |

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
Each model round trip costs roughly **250 ms** of overhead. On this fake path
only, each record costs roughly **70 ms**: a record is one `skyra` process
run once, and most of that is `F_FULLFSYNC` (see
[Where the 250 ms goes](#where-the-250-ms-goes)). With real Codex through the
resident, a tool call is about 13 records in ~0.38 s, roughly 30 ms per
record. Don't carry the 70 ms figure over to that path.

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
however long the model takes. This share table describes the fake path. Against real
Codex the costs are larger: see [Against native Codex](#against-native-codex).

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
[`controller/codex/latency/`](../controller/codex/latency/) and the sweep script
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
| `release` (Boundary lookup) | 10–40 ms |
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
The rounds ran from a clean worktree at `9049bb7`, with no uncommitted
changes. Every answer was correct, and both made exactly 3 tool calls in
every tools run: native through its shell (`cat`), Skyra through its `read`
target. The file paths in the prompt are absolute (see the note at the end of
this section).

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

The two clocks don't start and stop at quite the same points:
- **Start.** Native starts at process launch. Skyra starts at its first state
  write, which skips its own process start. That is ~25 ms, per
  [Process start](#process-start-30-ms-to-genesis).
- **End.** Native leaves out its ~1 s exit after answering. Skyra counts its
  0.2–0.5 s wind-down.

These roughly cancel, and are small next to the gaps below.

| Median of 6 (range) | Native Codex | Skyra | Skyra − native |
|---|---|---|---|
| **plain**: time to answer | 4.39 s (3.95–5.37) | 6.70 s (4.86–11.97) | **+2.3 s** |
| └ start → `turn/start` | 0.45 s | 2.44 s | +2.0 s |
| **tools**: time to answer | 15.53 s (14.71–19.32) | 18.60 s (16.89–20.94) | **+3.1 s** |
| └ start → `turn/start` | 0.26 s | 2.30 s | +2.0 s |
| └ per tool call (n=18) | not separately measured | **0.38 s** (0.34–0.55) | |

**Reading it.**
- **Startup is the largest fixed cost: about +2 s per `skyra ask`.**
  - Native `codex exec` reaches `turn.started` in 0.3–0.5 s.
  - Skyra launches a fresh resident app-server under its Seatbelt profile,
    with an isolated Codex home and an egress proxy. This took ~1.2–1.9 s in
    an earlier batch (one run took 3.3 s).
  - Skyra then spends ~0.4–1.1 s on initialize and `thread/start`, each a
    recorded exchange.

  That split comes from the earlier batch's records. The clean rerun only
  measured the total to `turn/start`.
- **Each tool call costs ~0.38 s of Skyra time.** This is more than the
  ~0.25 s per round trip on the fake path. With the resident, a tool call is
  about 13 records (73 records for the tools run against 34 for plain): the
  `read` step plus the recorded exchanges with the app-server. Native's
  matching cost, running `cat` in its own sandbox, didn't get a separate
  timing because `codex exec --json` doesn't timestamp the start of a
  command. It is likely tens of milliseconds.
- **Model time dominates, and n = 6 is small.** Native tools runs spread over
  4.6 s here; an earlier batch spread over 7 s. The two runtimes don't show
  the model the same thing:
  - Skyra replaces Codex's base prompt and offers its own tools.
  - Native carries its default prompt (~14k input tokens on the plain run)
    and makes the model write shell commands.

  So model time can differ between them for reasons unrelated to the
  runtime, and these runs can't separate that from noise. Treat the
  difference between totals as within noise. The reliable numbers are the
  fixed costs, which are measured directly:
  - startup: +2.0 s
  - tool calls: 3 × 0.38 s
  - wind-down: ~0.3 s

  For this task those come to ~3.5 s. That is close to the measured gap of
  3.1 s.

  An earlier batch, run on a tree with uncommitted launcher changes, measured
  a total gap of only +1.5 s for the same task, with the same startup and
  per-call costs. That shows how far model time alone moves the totals.

So against a conventional runtime, Skyra costs about **2 s more to start** a
one-shot run and about **0.4 s per tool call**. Both are fixed per step and
don't depend on model speed. On a 4 s answer that is about +50% (+2.3 s
measured). On the 15.5 s tools task it is about 20% in fixed costs (~3.5 s).
The measured gap was smaller (3.1 s) because Skyra's model time was shorter in
these runs, which is noise, not a result.

Reproduce with:

```sh
controller/codex/latency/compare.sh 6   # interleaved rounds; real model calls
```

`compare.sh` prints one line per run. `run.py` beside it times a finished
`skyra ask` run from its `.skyra/` state.

**Note on absolute paths.** The prompts name files by absolute path because
of a bug the first batch hit, filed as
[#41](https://github.com/skyraOS/tony-stark-was-able-tobuild-this-in-a-cave/issues/41).
`read` and `find` resolve a relative path against the first read grant, and
the registry sorts read grants by path. Any second read grant that sorts
before the workspace takes over relative paths, and `a.txt` comes back
`not_found`. In this case an uncommitted `models` target granting
`~/.claude/cache/model-catalog` did that. On the clean tree the workspace is
the only read grant, so the bug can't trigger. The paths stay absolute so
both batches use the same prompt.

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

Closed since the first draft:

- **Replay on Apple silicon.** The replay benchmarks from PR #40 now run on
  this machine at `9049bb7`:
  - about 47,000 records/s from memory, linear to 10,003 records (0.21 s)
  - about 7,800 records/s from the tree on disk (1.29 s for 10,003)
  - every published website run in under 4 ms natively

  Details are in [`docs/performance.md`](performance.md#replay). The
  WebAssembly figures there were not rerun here.

## Implications

- Overhead is **per step, not per second of inference**. To reduce it, cut
  steps or per-step cost, not waiting. Ways to do that:
  - fewer process launches per round trip
  - fewer fsynced commits per step (intake + gather ≈ 60 ms)
  - faster Boundary lookup
- A tool-heavy turn pays for each tool round trip: ~250 ms on the fake path
  and ~380 ms with real Codex through the resident. A 10-call agentic turn
  adds ~4 s of controller time no matter how fast the model is. That is the
  case to watch as models get faster.
- For one-shot runs, the ~2 s resident startup is the largest single cost
  against native Codex. Reusing a resident, or starting one ahead of time,
  would close most of that gap.
- Replay stays cheap: the byte-identical re-derivation at the end of each run
  is ~12 ms for 4–7 records. A 10,000-record run re-verifies in ~0.2 s from
  memory, or ~1.3 s from disk.

## Where the 250 ms goes

Date: 2026-10-13 · branch `sync-measure` (from `52e478a`) · same machine.
Every number below is at 0 ms injected latency against the fake app-server.

**Short answer: most of it is `F_FULLFSYNC`.** A `tool:hello` run issues
**70** flushes (`hello`: 36), about 10 per record. On this machine each one
costs **~4.6 ms**, file or directory alike, and they run strictly one after
another. That is **334 ms of a 500 ms run** (`hello`: 210 of 338). With
flushes ordered rather than forced to stable storage (`F_BARRIERFSYNC`), one
more model round trip costs ~80 ms instead of ~225 ms, and a one-step run
takes ~110 ms instead of ~250. With no flushing at all it costs ~65 ms. What remains is process launches, `sandbox-exec`, and the
adapter's round trip to the host.

### How it was measured

The instrumentation is on branch `sync-measure`, not on main. There, every
flush in `runtime/tree` (and `apprentice/proc`) goes through one helper, `flush`
(`controller/runtime/tree/sync.go`). Two environment variables, both off by
default, control it:

- `SKYRA_SYNC=full|barrier|plain|none` picks how to flush. Unset or unknown
  means `full`, the old behavior: `File.Sync()`, which Go implements on darwin
  as `fcntl(F_FULLFSYNC)` (`internal/poll/fd_fsync_darwin.go`).
  - `barrier` is `fcntl(fd, F_BARRIERFSYNC=85)`. It falls back to
    `F_FULLFSYNC` if the file system refuses it.
  - `plain` is `syscall.Fsync`.
  - `none` skips the flush and is for measurement only.
- `SKYRA_SYNC_TRACE=1` logs a line for each flush: mode, what was flushed,
  and how long it took.

Phases were attributed by timestamping every stderr line on arrival. Each
phase is the time from the previous phase line up to its own line. The
release a child logs therefore includes the parent's handoff writes, which
happen while the child waits for its Boundary.

All tests (`go test ./...` in `controller/` and `controller/apprentice/`) pass in
every mode.

In isolation (`micro` benchmark, 40 runs, 400-byte file on APFS, 99% full
internal SSD):

| Mode | file flush (median) | directory flush (median) |
|---|---|---|
| `full` (F_FULLFSYNC) | 4.7 ms | 3.9 ms |
| `barrier` (F_BARRIERFSYNC) | 0.65 ms | ~2 µs (p90 4 ms) |
| `plain` (fsync) | 26 µs | 0.5 µs |
| `none` | — | — |

### Per phase, by sync mode

`tool:hello` (2 model steps, 7 records). Each cell is wall ms, with ms
inside flushes in parentheses. A row sums that phase over all steps, using
per-step medians of 3 traced trials:

| Phase (summed over the run) | Flushes | full | barrier | plain | none |
|---|---|---|---|---|---|
| process start → `genesis` | 2 | 23 (10) | 14 (2) | 15 (0) | 17 (0) |
| `intake`: store input + commit record, ×5 | 20 | 117 (97) | 34 (13) | 24 (0) | 25 (0) |
| `release`: parent's handoff writes + Boundary lookup, ×5 | 20 | 91 (83) | 18 (10) | 7 (0) | 15 (0) |
| acceptance receipt before `execute`, ×2 | 6 | 31 (28) | 6 (4) | 2 (0) | 7 (0) |
| `execute` → emitted, ×3 | 0 | 95 (0) | 83 (0) | 95 (0) | 92 (0) |
| acceptance receipt before `settle`, ×2 | 6 | 31 (29) | 6 (3) | 2 (0) | 2 (0) |
| `gather`: settle's result + resolve commit, ×2 | 12 | 62 (57) | 13 (7) | 4 (0) | 11 (0) |
| `claim`, ×2 | 4 | 21 (19) | 4 (2) | 1 (0) | 1 (0) |
| exit + replay check | 0 | 9 (0) | 9 (0) | 9 (0) | 9 (0) |
| **total (median of run totals)** | **70** | **500 (334)** | **196 (43)** | **161 (1)** | **215 (0)** |

`hello` (1 model step, 4 records):

| Phase | Flushes | full | barrier | plain | none |
|---|---|---|---|---|---|
| process start → `genesis` | 2 | 37 (16) | 18 (2) | 14 (0) | 15 (0) |
| `intake` ×2 | 10 | 79 (62) | 20 (7) | 15 (0) | 12 (0) |
| `release` ×2 | 10 | 50 (46) | 9 (5) | 3 (0) | 4 (0) |
| acceptance before `execute` | 3 | 16 (14) | 3 (2) | 1 (0) | 1 (0) |
| `execute` → emitted | 0 | 73 (0) | 60 (0) | 66 (0) | 66 (0) |
| acceptance before `settle` | 3 | 15 (14) | 4 (2) | 1 (0) | 1 (0) |
| `gather` | 6 | 32 (29) | 6 (4) | 2 (0) | 2 (0) |
| `claim` | 2 | 10 (9) | 2 (1) | 1 (0) | 1 (0) |
| exit + replay check | 0 | 8 (0) | 8 (0) | 9 (0) | 9 (0) |
| **total** | **36** | **338 (210)** | **125 (22)** | **107 (1)** | **111 (0)** |

The flushes for one record, in order:

| Write | Flushes |
|---|---|
| intake: `inputs/` (file, dir), `records/` (file, dir) | 4 |
| release of an invoke that opens a slot: `emitted-` (file, frame dir), slot (file, deps/round dir), `deps/` dir, frame dir, `boundaries/` (file, dir) | 8 |
| release of a return: `boundaries/` (file, dir) | 2 |
| acceptance receipt: `acceptances/` (file, dir), lookup link (dir) | 3 |
| settle: `result` (file, frame dir) | 2 |
| fan-in claim: `resolved-` (file, frame dir) | 2 |

The frame-dir flush right after the `deps/` flush costs ~5 µs: nothing is
dirty by then. Every other flush pays the full ~4.6 ms.

`none` is not faster than `plain` in the traced `tool:hello`. That gap is
noise from 3 trials (one 10.8 ms `release` outlier). The untraced sweep
below has the two equal.

### End to end, by sync mode (`sweep.sh`-style, 10 trials, medians)

One actor host and one discarded warm-up boot. Modes rotate within each
trial, so drift hits all four equally. Run twice (a, b); b ran on a noisier
machine:

| Mode | `hello` a | `hello` b | `tool:hello` a | `tool:hello` b | One more round trip (`tool:hello` − `hello`, a) |
|---|---|---|---|---|---|
| `full` (default) | **250** | 338 | **476** | 517 | 226 ms |
| `barrier` | **109** | 180 | **189** | 210 | 80 ms |
| `plain` | **86** | 112 | **152** | 159 | 66 ms |
| `none` | **83** | 118 | **150** | 152 | 67 ms |

Every run answered correctly with the expected record count (4 / 7).

Two earlier batches were discarded as confounded:

- One ran `SKYRA_SYNC=$m sweep.sh 10 0` per mode. Its first trial per host was
  always an outlier (298, 176, 185 ms), and `full hello` read 404 ms for 7
  trials, then 250.
- One interleaved modes by calling `sweep.sh 1 0` per trial. **`sweep.sh`
  runs `go build` on every call. That rewrites the binaries even when nothing
  changed, and the next exec of a rewritten binary is slow** (see process
  start below). Every trial in that batch was a first exec.

Use the single-host rotation above (or warm up once) when comparing.

### Process start (~30 ms to `genesis`)

Exec cost, 30 runs each, warm:

| Binary | Size | Median exec → exit |
|---|---|---|
| `/usr/bin/true` | 0.1 MB | 1.4 ms |
| Go hello world | 2.4 MB | 2.4 ms |
| `skyra` (no args: `SealInherited`, usage, exit 2) | 5.2 MB | 2.9 ms |
| `echo` target | 3.0 MB | 2.9 ms |

**Go runtime start is not the cost:** `skyra` starts in ~3 ms, 0.5 ms more
than a hello world. `skyra boot` takes 26 ms median from exec to its
`genesis` line (`full`, range 19–40), or 11.5 ms with `SKYRA_SYNC=none`.
What runs before genesis:

| Step | Time |
|---|---|
| exec + Go runtime + `custody.SealInherited` (lists `/dev/fd`, sets CLOEXEC) | ~3 ms |
| flags, `realDir` (MkdirAll, Abs, EvalSymlinks) | <0.5 ms |
| `tree.Init` (5 × MkdirAll) | 0.3 ms |
| `registry.Load`: parse the registry and contracts, then **read and SHA-256 every registered binary** (here 2 binaries, 6.9 MB) | 4.5–6.5 ms |
| per-actor state dirs. With a `bundle`: re-read and hash the installer, run `install-bundle` as a subprocess | 0 here |
| `tree.Commit(genesis)`: 2 F_FULLFSYNC | ~10 ms (0.7 with `none`) |
| `tree.OpenFrame` | 0.1–0.3 ms |

`run()` also re-reads and hashes the target binary on every `execute`
(codex 3.9 MB, echo 3.0 MB: a few ms each). A registry with many large
binaries pays the hashing at genesis in proportion to their size.

**Code signature and quarantine.** The binaries are ad-hoc,
linker-signed (`flags=0x20002(adhoc,linker-signed)`, no TeamIdentifier). They
carry `com.apple.provenance` but **no `com.apple.quarantine`**, so the
Gatekeeper assessment doesn't run on exec (`spctl --assess` rejects them,
which doesn't matter for exec). The kernel still validates code-signature
pages on every exec; that is inside the ~3 ms.

**What does apply is a one-time first-exec cost per written file:**

| First exec after… | Cost |
|---|---|
| a new file (a byte-identical copy included) | 130–340 ms |
| a no-op `go build -o` that touched the file | 70 ms |
| a plain `touch` | 47 ms |

Every later exec takes ~3 ms. That pattern fits a syspolicyd/XProtect scan of
newly modified executables; the cause was inferred, not traced. In steady
state it doesn't touch Skyra, which never rewrites its binaries. It does
touch every benchmark that rebuilds first, and a user's first run after
install or update.

### `sandbox-exec -p` (hypothesis: the profile is compiled on every call)

20 runs each, interleaved, after one warm-up round:

| Command | Median | Min |
|---|---|---|
| `/usr/bin/true` | 2.0 ms | 1.9 ms |
| `sandbox-exec -p <real echo profile, 737 B> /usr/bin/true` | 9.5 ms | 8.8 ms |
| `sandbox-exec -p '(version 1)(allow default)' /usr/bin/true` | 9.6 ms | 8.8 ms |
| `sandbox-exec -f <same profile from a file> /usr/bin/true` | 9.5 ms | 8.8 ms |
| `echo` target, bare | 3.4 ms | 3.0 ms |
| `sandbox-exec -p <echo profile> echo` | 11.2 ms | 10.3 ms |

The profile was rendered by `boundary.Render` from a real run's accepted
Boundary, with the binary path changed to `/usr/bin/true` for the `true` rows.

**Rejected as stated.** `sandbox-exec` adds ~7.5 ms per target, but a
trivial profile costs the same as the real one, and `-f` costs the same as
`-p`. Compiling this profile isn't measurable. The cost is fixed: a second
exec (sandbox-exec, then the target) plus the sandbox library's own setup.
That is ~7.5 ms of the ~12 ms `echo` execute step. A one-off profile cache
wouldn't remove it. Applying the sandbox without the extra exec
(`sandbox_init` in a pre-exec helper) might; that is unmeasured.

### `release` (the 10–40 ms Boundary lookup)

The child's `release` phase runs from its `intake` line to `Boundary …
looked up`. Almost none of it is the lookup:

- **Child:** sends its stamp, then waits. On receipt it calls `LoadBoundary`,
  which reads `boundaries/<id>`, re-hashes it, and `json.Unmarshal`s it. That
  takes <1 ms.
- **Parent, in that window:**
  1. `tree.Load` of the child's record: read and re-hash.
  2. `issueRelease`: `checkSuccessor` plus `Narrow`, which re-derives the
     child's Boundary from its own. This is pure, under 1 ms.
  3. Then the handoff writes:
     - **Invoke that opens a slot:** `StoreInput` (deduplicated: stat, read,
       re-hash, no write), `PublishEmission` (2 flushes), `OpenDep` (4),
       `OpenFrame` (mkdir), `PutBoundary` (2). That is **8 F_FULLFSYNC,
       ~32 ms**.
     - **Return:** only `PutBoundary`: **2 F_FULLFSYNC, ~9 ms**.

**The variance is the step kind, not the lookup.** An invoke's release costs
~35 ms and a return's ~10 ms. Without flushes both take 1–3 ms.

After `release`, the child does more before it acts:

- `machine.accept` re-derives the Boundary again (`Narrow`, JSON-encode both,
  compare).
- `tree.Accept` re-reads and re-hashes the Boundary a second time and writes
  the receipt (3 flushes, ~15 ms).

Only then does `execute` or `settle` begin.

### What this changes in the earlier table

| Earlier line | What it actually is |
|---|---|
| `intake` 25–35 ms | 4 flushes (~20 ms), plus 6 for the first instance, which includes boot's `PutBoundary` of the root |
| `release` 10–40 ms | the parent's 2 or 8 flushes |
| `settle` ~15 ms | the acceptance receipt's 3 flushes |
| `gather` ~30 ms | settle's 2 flushes plus intake's 4 |
| `claim` ~10 ms | 2 flushes |

Without flushing, the controller-only phases (intake, release, acceptance,
gather, claim) total 2–7 ms per record. What remains per round trip:

- `execute`: sandbox-exec ~7.5 ms, binary re-hash, and for the Codex adapter
  its own process plus the host round trip (12–48 ms)
- ~3 ms of exec per `skyra` process
- registry hashing at genesis

A durability policy that keeps crash consistency at a fraction of this cost
is proposed in
`docs/next/proposal/40-durability-policy-proposal.md`.
