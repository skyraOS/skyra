# The stateless controller

Date: 2026-09-27  
Status: runnable. Standard library only; macOS (uses `sandbox-exec`). A
deterministic fake Codex app-server for the demo and tests is in
[`codex/fake/`](codex/fake/).
The design is in [`../docs/design/`](../docs/design/) and [`../docs/next/`](../docs/next/).

```sh
./run.sh                  # every scenario, each ending in a replay check
go test ./controller/...  # transitions, replay, and custody
skyra                     # from anywhere: chat with real Codex in the browser (see ../startup-doc.md)
skyra ask "TEXT"          # one question to real Codex, printed with its lineage
skyra replay DIR          # re-derive any finished conversation or run from its records
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
| **Contracts by (target, schema)** (next/32) | A name in reach holds a set of contracts: its registered one, and any it added to itself. At intake the payload selects the one it fits; none, or more than one (ambiguous), is a contract fault. A continuation is shown every contract, and told which one it is executed against | `machine.Selected`, `horizon.Surface.All` |
| **Horizons committed, looked up, re-derived** | A Horizon is released by committing it; the child is handed its hash, looks it up, and refuses it unless its own commit derives exactly that Horizon | `tree.PutHorizon`, `machine.accept` |
| **Dependencies and resolve** | A continuation invokes, waits on a slot, and resumes with what came back. Slots are opened before the child runs; the resolve record names every discharge (the join rule, by hashing); a fan-in claim picks the one resumer | `machine.gather`, `machine.claim`, `tree` |
| **The call stack as capabilities** | The frame stack travels as directory fds: the frame a return discharges, then each frame waiting below it | `handshake`, `instance.Drive` |
| **Replay derives everything** | Intake records, supervisor faults and resolves are each re-derived from their stored input; nothing is skipped | `replay` |
| **Adapters** | A stateless Codex adapter invokes recorded start, exchange, inspect and stop operations on a resident process; the browser host holds parked conversations | `codex/resident/`, `codex/io/`, `hosts/web/` |
| **Provider sandbox** | The runtime launches each Codex process under a registered Seatbelt profile with private state and a per-actor allowlisted egress proxy | `runtime/instance/`, `codex/provider/` |
| **Native attempt interruption** | The adapter interrupts the exact turn and confirms a terminal event, then returns the attempt to its parent as an observation (`refused`); runtime transitions remain provider-agnostic | `codex/resident/` |
| **Real Codex** | Codex's tool calls become derived invokes of tools that run under the conversation's Horizon; a native command is interrupted and reported | `skyra ask`, `skyra` |
| **Ingress by parking** (next/16) | An async surface's invocation is parked with its host, which holds a reference to the work (its stamp) and its descriptors, never its authority. `skyra resume` derives the instance from the tree, refuses work that is not parked or was already resumed, and the successor checks the event like any output | `machine.parked`, `runtime/park/`, `skyra resume` |
| **Conversations** | The user side is a continuation whose steps are the transcript, and Codex's parent. A fault in a Codex turn is shown and the conversation goes on. Each conversation is its lineage: saved, listed with its history, replayable | `targets/user/`, `hosts/web/` |

## Packages

Five groups, by what a package is. Imports only run downward: `core/` imports
nothing outside itself, `runtime/` builds on `core/`, the web host builds on
`runtime/`, and the programs under `targets/` and `codex/` share no code with
any of them. `core/` is pure,
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
| [`replay`](runtime/replay/) | Re-derives a lineage from genesis with the machine; reads the tree, writes nothing |
| [`boot`](runtime/boot/) | The operator: genesis, the seed, then waits for its frame (kqueue) |
| [`registry`](runtime/registry/) | The operator's registry: surfaces, binaries (pinned by hash), contracts |
| [`handshake`](runtime/handshake/) | The parent ↔ successor socket: messages plus the frame stack as fds |
| [`custody`](runtime/custody/) | The manifest of non-frame descriptors carried beside a lineage |
| [`park`](runtime/park/) | Parking a frame with a host, and resuming it |
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
| [`io`](codex/io/) | One bounded operation on a granted resident connection |
| [`provider`](codex/provider/) | Launches each Codex process under its Seatbelt profile, with an allowlisted egress proxy (Python) |
| [`fake`](codex/fake/) | A deterministic fake app-server, for the demo and tests |
| [`adapter`](codex/adapter/) | Legacy Codex adapter, retained for the demo and earlier tests |

**[`hosts/`](hosts/): the only long-lived processes, outside the step**

| Package | What it is |
|---|---|
| [`web`](hosts/web/) | The actor host in front of the human: WebSocket, parked conversation frames, past conversations derived from records |
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
write grant, so the workspace stays the first read grant (where a relative
read resolves), and a narrowing drops it by leaving it out. OS tests check
that a script written there can be read back but not run, and that nothing
can be written outside it.

A profile grants the program's own binary, fixed system loader reads, and the
Horizon's grants. A target is told its slot in `SKYRA_FRAME` and its read
grants in `SKYRA_READ` (colon-separated, like `PATH`); that is information,
and the profile is what enforces it. OS tests check that a file outside
`read` is denied, that the same file is readable when granted, and that exec
of any other program is denied. The real Codex app-server uses its registered
resident profile and per-actor proxy, and none of the Horizon's shared grants.

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

## Resident Codex adapter

The runtime knows only resident connections and registered operations. `codex.start`
launches a private app-server and adds its connection to custody. `codexio`
performs one recorded exchange or inspection on that connection. `codexresident`
rebuilds protocol state from the continuation's steps and chooses the next
operation. Before a later turn it inspects the existing process; after 30
minutes without turn activity it stops the process and starts a fresh one.
`codex.stop` shuts down the write side, observes peer EOF, and removes the
connection from custody. Codex tool calls become ordinary invokes of registered
tools. An unexpected provider request or native attempt interrupts the turn,
and is returned to the parent as an observation.

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
production path. Connection handoff now runs through `instance.Drive` and
chat. Its tests launch `sandbox-exec`, so they must run outside an enclosing
sandbox that prohibits applying Seatbelt profiles.

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
  frame fds and a resident connection cannot be written to disk: if the
  webhost stops, a waiting conversation stays viewable but cannot continue
  until restore exists.
- **`system.register` is not built.** Narrowing through the room is; adding a
  tool an agent wrote to its reach needs the tool's bytes in the tree for
  replay (improvements item 6).
- **Fan-out is not for everyone yet** (next/32, H2). An emission that is a
  JSON array is that many instructions, one successor each, every slot opened
  before any is released. A fan-out cannot return, cannot carry a resident
  connection, and cannot hold an invoke that changes the waiter's Horizon
  (`narrow`, `extend-reach`): those are emitted alone, and a branch of a
  fan-out cannot extend the Horizon its siblings share. The scripted agent fans out;
  Codex and the human still emit one instruction at a time.
- **Restore after a crash**, streaming, and the host's own restart are not built.
- **The provider profile is pinned to Codex 0.160.0** (`launch.CODEX_VERSION`,
  `instance.PinnedCodex`). A new version needs a fresh validation of startup,
  tool exposure, file effects and egress: `tool_surface.py`, `probe.py` and
  `integration.py` in `experiments/appserver/`.

## Improvements to carry forward

[`docs/design/improvements/`](../docs/design/improvements/): lessons from DNA-19, the predecessor, for the next version.

- [Profiles become input contracts](../docs/design/improvements/01-profiles-become-input-contracts.md)
- [No silent refusals: a fault-tolerant async boundary](../docs/design/improvements/02-no-silent-refusals.md)
- [A conversation is an invocation: logs, history and isolation](../docs/design/improvements/03-conversation-is-an-invocation.md)
