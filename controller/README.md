# The stateless controller

Date: 2026-09-27  
Status: runnable. Standard library only; macOS (uses `sandbox-exec`). A
deterministic fake Codex app-server for the demo and tests is in
[`codex/fake/`](codex/fake/).
The design is in [`../docs/design/`](../docs/design/) and [`../docs/next/`](../docs/next/).

```sh
./run.sh                  # every scenario, each ending in a replay check
go test ./controller/...  # transitions, replay, and custody
(cd interp && go test ./...)  # the interpreter, its own module
skyra                     # from anywhere: chat with real Codex in the browser (see ../startup-doc.md)
skyra ask "TEXT"          # one question to real Codex, printed with its lineage
skyra replay DIR          # re-derive any finished conversation or run from its records
skyra metric DIR          # the research metric: procedures remembered, and calls to them with no model step
```

One `skyra` binary runs **once per step**, then exits. Every decision is made
by one pure function; all durable state is a content-addressed, hash-chained
record tree in `.skyra/`.

## What it does now

| | Built | Where |
|---|---|---|
| **Operators are derived** | A target never says "invoke" or "return". It prints operands, `{"to": address, "payload": JSON}`, and `interpret` derives the state from where it stands: its own slot is a return, a name in reach an invoke; anything else is a typed fault (`malformed`, `horizon`, `contract`, `leaf`, `custody`) that carries the observation it was derived from | `machine.interpret` |
| **Input contracts** | Every surface declares an input contract, and how it reads its input (`continuation`). It declares nothing about what it may reach, and nothing about what it returns. Fit is strict | `core/contract/`, `targets/*/contract.json` |
| **One authority per Horizon** | Reach and OS grants belong to the Horizon, and every call shares its caller's (next/24). The sandbox is rendered from the Horizon, the same for everything under it, plus the exclusions: `.skyra/` is always denied, and exec is allowed only for the program being run | `machine.Narrow`, `horizon.Render` |
| **The Horizon as an actor** (next/32) | Every Horizon holds `system`, whose contracts are the syscalls: `{"to": "system", "payload": {"narrow": …}}` returns a reference to a narrowed Horizon for one actor, which joins the caller's reach; invoking it runs the actor under the narrower authority. `extend-reach` adds a contract to the emitter's own reach entry, throughout its Horizon, with no approval: the actor is derived from the emitter, never named. A narrowing that asks for anything not held is a horizon fault (next/26) | `machine/room.go` |
| **`strand`** (next/32) | A syscall that hands the requester the stamps from an observation up to the invoke made through a contract it names, found by climbing frames, with every record of the work beneath: each resolve's discharges and the chains that led to them. A call into a narrowed child appears as one call and its return. Refused (a horizon fault) when no such invoke caused the observation, or when the walk would leave the requester's Horizon | `replay.ReadStrand`, `machine.stranded` |
| **Remember** (next/32) | A worker that holds `remember` (the interpreter's registered contract) passes an observation (a step's `observation` stamp) and the contract that requested it; the interpreter reads the strand, compiles it into a procedure under that contract, and adds `(interpreter, contract)` to its own reach with `extend-reach`, with no approval. Calling it runs the remembered calls with no model step | `interp/`, `targets/agent` (`remember:`) |
| **Running what was remembered** (next/32) | One representation per contract: inside its refinement the compiled body runs, with no model; until the model writes refinements, the refinement is the base types, at entry and at every edge where the model was resumed. Outside it, the interpreter invokes the actor the procedure was remembered from, an ordinary instruction: from the start for a different input, or at the edge that refused, through the actor's `resume` contract when it holds one, shown the command, what already ran and why, as names and presentations | `interp` (`runProcedure`, `fallback`), `targets/agent` (`look:`) |
| **The thesis demo (P9)** (next/32) | In `run.sh`: the agent does a task; a worker remembers it; the procedure runs on new input with no model step and no approval; from a Horizon that keeps it but drops `echo` it faults at intake; the same contract is not remembered twice; an actor cannot extend itself with what it holds. Codex (fake) remembers a call of its own and calls the procedure. `skyra metric` reads the research metric from the records | `targets/agent`, `codex/fake`, `runtime/metric` |
| **Codex and remember** (next/32) | With the interpreter in reach, both Codex adapters offer `remember` (one of the turn's own calls, by index: the adapter, a worker, turns it into the observation's stamp and its contract), `procedure` (input to the interpreter, which selects by fit), and one tool per procedure that exists at `thread/start`, by title. The launcher registers the interpreter | `codex/resident`, `codex/adapter`, `../skyra` |
| **Refinements** (next/32) | At remember, before the interpreter extends itself, the actor a procedure was remembered from writes its refinements in CEL when it holds a `refine` contract: shown its command, rounds, result, base types and constants, never a hash. Each is type-checked, bounded in cost and checked against its own strand; a refusal is shown, and it may try again. Where present they replace the base types: inside them the procedure runs compiled, outside them the actor runs | `interp/refine`, `interp` (`runRemember`, `refusedInput`, `refusedEdge`) |
| **Contracts by (target, schema)** (next/32) | A name in reach holds a set of contracts: its registered one, and any it added to itself. At intake the payload selects the one it fits; none, or more than one (ambiguous), is a contract fault. A continuation is shown every contract, and told which one it is executed against | `machine.Selected`, `horizon.Surface.All` |
| **Horizons committed, looked up, re-derived** | A Horizon is released by committing it; the child is handed its hash, looks it up, and refuses it unless its own commit derives exactly that Horizon | `machine.horizonWrite`, `machine.accept` |
| **Dependencies and resolve** | A continuation invokes, waits on a slot, and resumes with what came back. Slots are opened before the child runs; the resolve record names every discharge (the join rule, by hashing); a fan-in claim picks the one resumer | `machine.gather`, `machine.claim`, `tree` |
| **The call stack as capabilities** | The frame stack travels as directory fds: the frame a return discharges, then each frame waiting below it | `handshake`, `instance.Drive` |
| **Replay derives everything** | Intake records, supervisor faults and resolves are each re-derived from their stored input; nothing is skipped | `replay` |
| **Adapters** | A stateless Codex or Claude adapter invokes recorded open, exchange, inspect and stop operations on a session of the service host; the browser host holds parked conversations | `codex/resident/`, `claude/resident/`, `service/io/`, `hosts/web/` |
| **Provider sandbox** | The service host launches each Codex or Claude process under the profile its registration declares, with private state and an allowlisted egress proxy; the runtime only ever runs a process to its exit | `service/`, `codex/launch/`, `claude/launch/`, `codex/provider/` |
| **Native attempt interruption** | The adapter interrupts the exact turn and confirms a terminal event, then returns the attempt to its parent as an observation (`refused`); runtime transitions remain provider-agnostic | `codex/resident/` |
| **Real Codex** | Codex's tool calls become derived invokes of tools that run under the conversation's Horizon; a native command is interrupted and reported | `skyra ask`, `skyra` |
| **Ingress by parking** (next/16) | An async surface's invocation is parked with its host, which holds a reference to the work (its stamp) and its descriptors, never its authority. `skyra resume` derives the instance from the tree, refuses work that is not parked or was already resumed, and the successor checks the event like any output | `machine.parked`, `runtime/park/`, `skyra resume` |
| **Conversations** | The user side is a continuation whose steps are the transcript, and Codex's parent. A fault in a Codex turn is shown and the conversation goes on. Each conversation is its lineage: saved, listed with its history, replayable | `targets/user/`, `hosts/web/` |

## Packages

Six groups, by what a package is. Imports only run downward: `core/` imports
nothing outside itself, `runtime/` builds on `core/`, the web host builds on
`runtime/`, and the programs under `targets/`, `codex/` and `interp/` share no
code with any of them; `interp/` is its own module, so its dependencies never
reach the controller's standard-library Go. `core/` is pure,
and `TestMachineIsPure` fails if `machine` imports anything that touches the
OS, the tree or the network.

**[`core/`](core/): decides, never touches the OS**

| Package | Owns |
|---|---|
| [`contract`](core/contract/) | JSON-types schemas, strict decoding, the fit rule, canonical payloads |
| [`horizon`](core/horizon/) | Authority: `Surface` (identity and contracts), `Horizon` (reach, references, grants; the frame a return discharges and the continuation waiting on it), `Render` (the Seatbelt profile) |
| [`record`](core/record/) | The record, its states and fault kinds, canonical `Encode`, `Stamp` |
| [`machine`](core/machine/) | **The decision tree:** `Step`, `interpret`, `Narrow`, fan-out, resolve, the Horizon's syscalls. **Pure, enforced** |

**[`runtime/`](runtime/): the I/O around the core, inside one step**

| Package | Owns |
|---|---|
| [`instance`](runtime/instance/) | The shell: `Drive` performs effects and turns outcomes into events. Makes no decisions |
| [`tree`](runtime/tree/) | `.skyra/` on disk: records, inputs, frames, slots, write-once results and claims. The only package that touches `.skyra/` |
| [`replay`](runtime/replay/) | Re-derives a lineage from genesis with the machine, every sibling a resolve names included; walks a strand for the `strand` syscall. Reads the tree, writes nothing |
| [`boot`](runtime/boot/) | The operator: genesis, the seed, then waits for its frame (kqueue) |
| [`registry`](runtime/registry/) | The operator's registry: surfaces, binaries (pinned by hash), contracts |
| [`handshake`](runtime/handshake/) | The parent ↔ successor socket: messages plus the frame stack as fds |
| [`custody`](runtime/custody/) | The manifest of non-frame descriptors carried beside a lineage |
| [`park`](runtime/park/) | Parking a frame with a host, and resuming it |
| [`metric`](runtime/metric/) | The research metric (next/32), from the records alone: procedures remembered per task, and calls to them with zero model steps and zero approvals |
| [`trace`](runtime/trace/) | The process's log lines on stderr |

**[`skyra/`](skyra/)**: the command: argv and flags only.

**[`targets/`](targets/): programs run once under a rendered sandbox, each with its `contract.json`**

| Package | What it is |
|---|---|
| [`echo`](targets/echo/) | An untrusted leaf. Prints its text |
| [`read`](targets/read/) | An untrusted leaf. Reads a range of a UTF-8 file, or lists a directory, under a workspace read grant; replies with a typed observation (path, sha256, lines, text) |
| [`write`](targets/write/) | An untrusted leaf. Writes UTF-8 text to a file in the Horizon's scratch space (`SKYRA_SCRATCH`), never through a link and never out of it; read reads it back as `scratch:PATH` |
| [`roots`](targets/roots/) | An untrusted leaf. Replies with the paths its Horizon lets it see: the workspace (where a relative read resolves), every read grant, and the scratch space. Opens nothing |
| [`agent`](targets/agent/) | A fake continuation: calls echo, then replies with what it said |
| [`user`](targets/user/) | The user's side of a conversation: a continuation whose steps are the transcript |
| [`browser`](targets/browser/) | The browser's contract: an async surface, served by the web host |

**[`codex/`](codex/): every Codex semantic**

| Package | What it is |
|---|---|
| [`resident`](codex/resident/) | Stateless Codex protocol and lifecycle decisions |
| [`provider`](codex/provider/) | Launches each Codex process under its Seatbelt profile, with an allowlisted egress proxy (Python) |
| [`fake`](codex/fake/) | A deterministic fake app-server, for the demo and tests |
| [`latency`](codex/latency/) | Holds back each model output of an app-server by a fixed delay; with `sweep.sh`, the overhead sweep, and `compare.sh`, native Codex against Skyra, in [`OVERHEAD-VS-INFERENCE.md`](../docs/OVERHEAD-VS-INFERENCE.md) |
| [`adapter`](codex/adapter/) | Legacy Codex adapter, retained for the demo and earlier tests |
| [`proceduretools`](codex/proceduretools/) | Pure procedure discovery, model-call argument translation and schema projection shared by both Codex transports |

**[`interp/`](interp/): the interpreter, an actor, its own module (next/32)**

| Package | What it is |
|---|---|
| [`interp`](interp/) | The interpreter: runs procedures from its own state (`.skyra/interpreter/`, `SKYRA_STATE`), one resume per exec, against the contract it was invoked through. Emits every call now ready as one emission; records where its frame is before each. `interp install DIR SOURCE` installs a procedure by hand |
| [`proc`](interp/proc/) | Procedures: a header, a return and a body of content-addressed nodes (`{call, in, each, after}`); bindings by provenance; the planner |
| [`remember`](interp/remember/) | Compiles a strand into procedures: each continuation's control into nodes, round after round; bindings by provenance (the input, then what earlier calls returned; the rest constants); a frame whose return the model authored stays a call to it. `remember`'s body, embedded in the binary: `strand`, compile, then `extend-reach` per contract, refusing a second representation before anything is stored |
| [`refine`](interp/refine/) | Refinements in CEL (`cel.dev/cel-go`, this module's only dependency): a fixed environment (`input`, `input_type`, `returns`, `types`), pure functions, no `map` or `filter`, a cost bounded by CEL's estimator, and no error on its own strand, even one `&&` or `||` absorbed |
| [`types`](interp/types/) | Base types, moved from the kernel: shapes, contents, presentations. Nothing is typed unless it is remembered |

The staged [conversation reconstruction bundle](interp/workflows/conversations/)
adds authored procedures, accepted-frontier checkpoints and bounded multi-leaf
history reads for proposal 33. Registry entries may specify a pinned `bundle`
for a stateful continuation; boot installs it before genesis and publishes
its contracts together. Production conversation and client routing remain
on the existing path; the bundle README records the remaining migration.

**[`hosts/`](hosts/): the only long-lived processes, outside the step**

| Package | What it is |
|---|---|
| [`web`](hosts/web/) | The actor host in front of the human: WebSocket, parked conversation frames, past conversations derived from records |
| [`service`](service/) | The service host (proposal 49): runs registered long-lived programs as sessions, each opened by a handle; [`host`](service/host/) is its binary, [`io`](service/io/) the adapter that performs one bounded operation on it |
| [`actor`](hosts/actor/) | Legacy universal actor host, retained for the demo and earlier tests; current chat does not launch it |

### Binaries are registered by hash

A conversation pins each surface's binary by sha256 at genesis, and a binary
that no longer matches is a custody fault. So the launcher never registers a
file a rebuild overwrites: `./skyra` copies each built binary to
`build/controller/pinned/<sha256>/<name>` and registers that copy. A rebuild
adds copies and leaves running and saved conversations their own.

### Authority in the registry

A surface carries no authority (next/25). A registry entry may bring grants
into the room: `authority.read` and `write` as `{ "path": "...", "scope":
"literal" | "subpath" }` entries, `authority.mach` service names and
`authority.connect` Unix socket paths. Every grant in the file becomes the
room's one authority, which genesis records and everything under the root
Horizon shares until a narrowing. Exec is not a grant: a profile allows exec
of the program being run and nothing else. Paths are resolved at registry
load. Legacy top-level `connect` entries are folded into `authority.connect`.
No entry may be named `system` or `system.*`: those are the Horizon's syscalls and references.

A Horizon may also carry a **scratch** directory (`authority.scratch`): read
and write for everything under it, and, like everywhere, no exec but the
program being run, so nothing written there can be run. It is not a registry
grant: `skyra boot -scratch DIR` puts it in the room at genesis, and the web
host gives each conversation its own, `conversations/<id>/scratch`, beside its
`.skyra/` and deleted with it. It is a field of its own, not a read and a
write grant, so a narrowing drops it by leaving it out. OS tests check that a
script written there can be read back but not run, and that nothing can be
written outside it.

The **workspace** (`authority.workspace`) is where a relative read resolves,
and the directory the models are told they work in. An entry names it, and a
read grant must cover it; a second, different workspace is refused. It is
never inferred from the order of the read grants, which the registry sorts.
Targets are told it in `SKYRA_WORKSPACE`. A narrowing that drops the grant
under it drops the workspace too.

A profile grants the program's own binary, fixed system loader reads, and the
Horizon's grants. An entry marked `"state": true` keeps state of its own: its
sandbox may read and write `.skyra/<name>/`, created at genesis and named in
`SKYRA_STATE`, and nothing else of the floor (next/32). An actor can only see
its own state. A target is told its slot in `SKYRA_FRAME` and its read
grants in `SKYRA_READ` (colon-separated, like `PATH`); that is information,
and the profile is what enforces it. OS tests check that a file outside
`read` is denied, that the same file is readable when granted, and that exec
of any other program is denied. A model's process is the service host's: it
runs under the profile its program registration declares, with its own egress
proxy, and none of the Horizon's shared grants.

## The step

```go
func Step(s Instance, e Event) (Instance, []Effect)
```

`Step` makes every decision; `instance.Drive` performs effects and reports what
happened as the next event. The state, derived at intake, picks the program:

```text
intake ─┬─ invoke  → release → execute → handoff              run the target; hand off its output
        ├─ return  → release → settle → gather → claim ─┐     discharge the slot, then maybe resume the waiter
        └─ fault   → release → settle → gather → claim ─┤
                                                        └─ resolve → execute → handoff   the waiter, resumed
```

| Phase | Event in | Effects out |
|---|---|---|
| Intake | `Received`: the bytes, and the Horizon they were emitted under | `StoreInput`, `Commit`, `SendStamp` |
| Release | `Released`: the Horizon, and the number of frame fds that came with it | invoke: `Run`. Return or fault: `Settle`, then `GatherWaiter` (or `Exit` if only the operator waits) |
| Execute | `Emitted`: the target's one line | `Spawn`, for its first instruction (a JSON array is several: fan-out); a silent or failed target: a supervisor fault, which discharges this frame's slot |
| Handoff | `Stamped`: a successor's stamp and record | the next instruction: `Spawn`. Every one stamped: `OpenDep` for every new frame, then `SendHorizon` for each (invoke: new frame; return or fault: same stack; in a fan-out a fault holds a frame of its own). Then `Exit` |
| Gather | `Settled`: the waiting frame's record and every round it opened | a slot still open: `Exit`. All discharged: `StoreInput`, `Commit` (resolve), `TryClaim` |
| Claim | `Claimed` | lost: `Exit`. Won: `Pop`, `Run` (the waiting continuation, resumed in this instance) |

### Operands, and what `interpret` derives

```text
{"to": "<address>", "payload": <JSON>}

addresses the slot its frame discharges (SKYRA_FRAME), any JSON payload                  → return
names a capability or reference in reach, payload fits its input contract                → invoke
names anything else                                                                      → fault (horizon)
addresses nothing                                                                        → fault (leaf)
not readable as operands                                                                 → fault (malformed)
payload does not fit the contract of what it addresses                                   → fault (contract)
```

No name is reserved. Every fault carries, as its payload, the observation it
was derived from, so a parent can pass it on.

### A continuation's input

A continuation (a surface whose contract says `"continuation": true`) reads
everything it knows on stdin, and keeps nothing between runs:

```json
{"frame": "<its frame>", "input": {"text": "hello"},
 "tools": {"echo": {"type": "object", "required": ["text"], "fields": {"text": {"type": "string"}}}},
 "steps": [{"to": "echo", "payload": {"text": "hello"}, "state": "return", "value": "hello"}]}
```

`tools` is exactly its reach, references included: what it is shown is what
it may invoke. It returns by addressing `frame`.

## What happens: a continuation calls a tool

```text
boot                      agent (frame R1)              echo (frame R2)            echo's return
commit genesis G0
emit {"to":"agent",…}  ─▶ INTAKE invoke agent R1
open slot G0/deps/G0/R1
release [R1, G0]       ─▶ EXECUTE agent: {"to":"echo",…}
                          INTAKE (successor) invoke echo R2
                          open slot R1/deps/R1/R2
                          release [R2, R1, G0]     ─▶ EXECUTE echo: {"to":R2,…}
                                                      handoff ────────────────▶ INTAKE return "hello" R3
                                                                                SETTLE R2/result = R3
                                                                                GATHER R1's round: all discharged
                                                                                commit resolve R4 (deps: R3)
                                                                                CLAIM R1/resolved-R1: won
                                                                                EXECUTE agent, resumed: echo said "hello"
                                                                                ... return R5 settles R1, resolve R6 of G0
boot wakes (kqueue on G0), prints the lineage, replays it: 7 records, byte-identical
```

## State on disk

```text
.skyra/
  records/<sha256>                  every record, content-addressed
  inputs/<sha256>                   the exact bytes each record was derived from
  horizons/<sha256>                 every Horizon released, looked up by the child it was released to
  frames/<F>/result                 write-once: what discharged F's slot
  frames/<F>/deps/<round>/<child>   a slot F opened in a round, before the child ran
  frames/<F>/resolved-<round>       write-once fan-in claim: the resolve that won
```

## Models behind the service host

The kernel runs every target to its exit
([proposal 49](../docs/next/proposal/49-services-behind-adapters-proposal.md)).
Anything long-lived sits behind an adapter: `service-io`, in front of the
service host. The launcher starts one host per launch, and registers the
programs it may run in a file beside the registry: Codex's app-server, and
`claude-start` when the Claude CLI is installed, each with the profile, login
and egress hosts it needs. `codex.open` (or `claude.open`) asks the host to
start a session and returns its handle, an unguessable id that only the
adapter's recorded steps carry. Exchange, inspect and stop name that handle,
and the host keeps output that arrives between steps for the next one.
`codexresident` rebuilds protocol state from the continuation's steps and
chooses the next operation; each thread's whole prompt is `You are inside
Skyra, a cognitive architecture, fulfilling: <header>`. Each turn opens its
own session and stops it when done. The host reclaims a session left idle by
an exponential backoff that every request resets (30 s doubling to 16 min,
about half an hour idle), and a later operation on it sees the peer gone.
Codex tool calls become ordinary invokes of registered tools. An unexpected
provider request or native attempt interrupts the turn, and is returned to the
parent as an observation. Pasted images are not given to either model for now
(decision 2).

## Decisions, and where they live

| Decision | Code |
|---|---|
| D1 the emitter is never the checker | `intake` interprets before any `Run`; `execute` passes output on uninterpreted |
| D2 custody; the parent exits after release | `handoff` releases only after `issueRelease` passes; `drive` releases the child and exits |
| D3 the stamp is a pointer | `record.Stamp`; `tree.Load` refuses a record that does not hash to its stamp; `checkSuccessor` |
| D5 one commit per handoff | only intake, a supervisor fault, and a resolve commit |
| D6 the release is the Horizon plus fds | `SendHorizon`: the committed Horizon's id plus the frame stack over `SCM_RIGHTS`; the child looks it up and re-derives it; every write in a frame goes through its fd |
| D7 the sandbox is rendered from the Horizon; `.skyra/` always denied | `horizon.Render(self, floor)`; the floor is a parameter, not a Horizon field; exec only of the running program |
| D8 leaves | a behaviour, not an authority: a leaf shares its caller's Horizon, and output that addresses nothing is a leaf fault |
| D9 states, derived (thought 14) | `interpret`; the record stores the derived state, and replay derives it again |
| Join rule (thought 13) | `gather`: no resolve while a slot is open; the resolve names every discharge |

## Known gaps

The [resident custody experiment](experiments/resident/README.md) preceded the
service host (proposal 49), which replaced in-kernel resident connections.
Tests launch `sandbox-exec`, so they must run outside an enclosing sandbox
that prohibits applying Seatbelt profiles.

- **Records store derived fields** (`state`, `to`, `payload`, `fault`). They
  are kept as a commitment that replay checks, not as a source of truth.
- **`drive` still checks the surface hash** before running; it should report
  the observed hash for `step` to decide.
- **A fault ends the Codex turn it happened in.** The conversation goes on
  (the user side shows it), but Codex itself is not told within the turn;
  routing a horizon fault back into the same turn as an observation is still
  [improvement 03](../docs/design/improvements/03-conversation-is-an-invocation.md).
- **Parked descriptors live in the webhost's memory.** The parked work and
  its Horizon are on disk, and resume derives them from the tree, but the
  frame fds cannot be written to disk: if the webhost stops, a waiting
  conversation stays viewable but cannot continue until restore exists. A
  model's session is not among them: it lives in the service host, which
  reclaims it when idle.
- **`system.register` is not built.** Narrowing through the room is; adding a
  tool an agent wrote to its reach needs the tool's bytes in the tree for
  replay (improvements item 6).
- **Fan-out is not for everyone yet** (next/32, H2). An emission that is a
  JSON array is that many instructions, one successor each, every slot opened
  before any is released. A fan-out cannot return, and cannot hold an invoke
  that changes the waiter's Horizon (`narrow`, `extend-reach`): those are
  emitted alone, and a branch of a
  fan-out cannot extend the Horizon its siblings share. The scripted agent fans out;
  Codex and the human still emit one instruction at a time.
- **Restore after a crash**, streaming, and the host's own restart are not built.
- **The provider profile is pinned to Codex 0.160.0** (`launch.CODEX_VERSION`,
  `codex/launch.Pinned`). A new version needs a fresh validation of startup,
  tool exposure, file effects and egress: `tool_surface.py`, `probe.py` and
  `integration.py` in `experiments/appserver/`.

## Improvements to carry forward

[`docs/design/improvements/`](../docs/design/improvements/): lessons from DNA-19, the predecessor, for the next version.

- [Profiles become input contracts](../docs/design/improvements/01-profiles-become-input-contracts.md)
- [No silent refusals: a fault-tolerant async boundary](../docs/design/improvements/02-no-silent-refusals.md)
- [A conversation is an invocation: logs, history and isolation](../docs/design/improvements/03-conversation-is-an-invocation.md)
