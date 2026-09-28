# The stateless controller

Date: 2026-09-27  
Status: runnable. Standard library only; macOS (uses `sandbox-exec`). It does
not import DNA-19 (it only borrows DNA-19's fake Codex binary for a test).
The design is in [`../design/`](../design/) and [`../next/`](../next/).

```sh
./run.sh                  # every scenario, each ending in a replay check
go test ./controller/... ./history/dna19/... # transitions, replay, and custody
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
| **Operators are derived** | A target never says "invoke" or "return". It prints operands, `{"to": capability, "payload": JSON}` or plain text, and `interpret` derives invoke, return or a typed fault (`malformed`, `horizon`, `contract`, `custody`) | `machine.interpret` |
| **Input contracts** | Every surface declares an input contract and a return contract (the `reply` capability's input contract), and what it may be granted. Fit is strict | `contract/`, `*/contract.json` |
| **Authority only narrows** | A child is granted only surfaces its parent holds. A registry entry can attenuate a child's file, process, Mach and socket grants; `Narrow` checks each grant against the parent's reachable ceiling. The tool sandbox is rendered from those grants; `.skyra/` is always denied | `machine.Narrow`, `horizon.Render` |
| **Dependencies and resolve** | A continuation invokes, waits on a slot, and resumes with what came back. Slots are opened before the child runs; the resolve record names every discharge (the join rule, by hashing); a fan-in claim picks the one resumer | `machine.gather`, `machine.claim`, `tree` |
| **The call stack as capabilities** | The frame stack travels as directory fds: the frame a reply discharges, then each frame waiting below it | `handshake`, `instance.Drive` |
| **Replay derives everything** | Intake records, supervisor faults and resolves are each re-derived from their stored input; nothing is skipped | `replay` |
| **Adapters** | A stateless Codex adapter invokes recorded start, exchange, inspect and stop operations on a resident process; the browser host holds parked conversations | `codexresident/`, `codexio/`, `webhost/` |
| **Provider sandbox** | The runtime launches each Codex process under a registered Seatbelt profile with private state and a per-actor allowlisted egress proxy | `instance/`, `provider/` |
| **Native attempt interruption** | The adapter interrupts the exact turn and confirms a terminal event before emitting its horizon fault; runtime transitions remain provider-agnostic | `codexresident/` |
| **Real Codex** | Codex's tool calls become derived invokes of leaves that run under their own sandbox; a native command becomes a horizon fault | `skyra ask`, `skyra` |
| **Ingress by parking** (next/16) | An async surface's invocation is parked with its host, which holds custody (the frame stack) until an outside event arrives, then resumes it; the successor checks the event like any output | `machine.parked`, `park/`, `skyra resume` |
| **Conversations** | The user side is a continuation whose steps are the transcript, and Codex's parent. A fault in a Codex turn is shown and the conversation goes on. Each conversation is its lineage: saved, listed with its history, replayable | `user/`, `webhost/` |

## Packages

One directory per owner. Imports only run downward, and `machine` is pure:
`TestMachineIsPure` fails if it imports anything that touches the OS, the tree
or the network.

| Package | Owns | Pure? |
|---|---|---|
| [`contract`](contract/) | JSON-types schemas, strict decoding, the fit rule, canonical payloads | yes |
| [`horizon`](horizon/) | Authority: `Surface` (contracts, grants, sockets), `Horizon` (reach, the frame its reply discharges, the continuation waiting on it), `Render` (the Seatbelt profile) | yes |
| [`record`](record/) | The record, its states and fault kinds, canonical `Encode`, `Stamp` | yes |
| [`machine`](machine/) | **The decision tree:** `Step`, `interpret`, `Narrow`, resolve. **Pure, enforced** | yes |
| [`registry`](registry/) | The operator's registry: surfaces, binaries (pinned by hash), contracts | no |
| [`tree`](tree/) | `.skyra/` on disk: records, inputs, frames, slots, write-once results and claims. The only package that touches `.skyra/` | no |
| [`handshake`](handshake/) | The parent ↔ successor socket: messages plus the frame stack as fds | no |
| [`instance`](instance/) | The shell: `Drive` performs effects and turns outcomes into events. Makes no decisions | no |
| [`replay`](replay/) | Re-derives a lineage from genesis with the machine | reads only |
| [`boot`](boot/) | The operator: genesis, the seed, then waits for its frame (kqueue) | no |
| [`skyra`](skyra/) | The command: argv and flags only | — |
| [`echo`](echo/) | An untrusted leaf. Prints its text | — |
| [`agent`](agent/) | A fake continuation: calls echo, then replies with what it said | — |
| [`host`](host/) | Legacy actor host, retained for earlier tests; current chat does not launch it | — |
| [`codex`](codex/) | Legacy Codex adapter, retained for earlier tests | — |
| [`codexresident`](codexresident/) | Stateless Codex protocol and lifecycle decisions | — |
| [`codexio`](codexio/) | One bounded operation on a granted resident connection | — |
| [`park`](park/) | Parking a frame with a host, and resuming it | no |
| [`user`](user/) | The user's side of a conversation: a continuation whose steps are the transcript | — |
| [`browser`](browser/) | The browser's contract: an async surface, served by the webhost | — |
| [`webhost`](webhost/) | The actor host in front of the human: WebSocket, parked conversation frames, past conversations derived from records | — |

The targets (`echo`, `agent`, `codexresident`, `codexio`) share no code with skyra.

### Tool authority in the registry

An executable registry entry may declare `authority.read`, `write` and `exec`
as `{ "path": "...", "scope": "literal" | "subpath" }` entries, plus
`authority.mach` service names and `authority.connect` Unix socket paths.
Paths are resolved at registry load, before genesis commits the surfaces.
An optional `delegate` map limits the direct authority passed to a named tool
in the surface's `reach`. If omitted, the tool's registered ceiling is passed.
The caller does not gain its tool's direct OS permissions. Legacy top-level
`connect` entries are folded into `authority.connect` on load.

Tool profiles grant only their own binary, fixed system loader reads, and
declared resources. An OS test checks that a file outside `read` is denied and
the same file is readable when granted. The real Codex app-server uses the
registered resident profile and per-actor proxy.

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
| Execute | `Emitted`: the target's one line | `Spawn`; a silent or failed target: a supervisor fault, which discharges this frame's slot |
| Handoff | `Stamped`: the successor's stamp and record | invoke: `OpenDep`, `SendHorizon` (new frame). Return or fault: `SendHorizon` (same stack). Then `Exit` |
| Gather | `Settled`: the waiting frame's record and every round it opened | a slot still open: `Exit`. All discharged: `StoreInput`, `Commit` (resolve), `TryClaim` |
| Claim | `Claimed` | lost: `Exit`. Won: `Pop`, `Run` (the waiting continuation, resumed in this instance) |

### Operands, and what `interpret` derives

```text
{"to": "<capability>", "payload": <JSON>}   addressed
anything else                               unaddressed: its text is the payload for reply

names a capability in reach, payload fits its input contract   → invoke
names reply, or nothing, payload fits the return contract      → return
not readable as operands                                       → fault (malformed)
names a capability outside the Horizon                         → fault (horizon)
payload does not fit the contract of what it names             → fault (contract)
```

A leaf's output is always a reply: tools just print their data.

### A continuation's input

A continuation (a surface with something in its reach) reads everything it
knows on stdin, and keeps nothing between runs:

```json
{"frame": "<its frame>", "input": {"text": "hello"},
 "tools": {"echo": {"type": "object", "required": ["text"], "fields": {"text": {"type": "string"}}}},
 "steps": [{"to": "echo", "payload": {"text": "hello"}, "state": "return", "value": "hello"}]}
```

`tools` is exactly its reach: what it is shown is what it may invoke.

## What happens: a continuation calls a tool

```text
boot                      agent (frame R1)              echo (frame R2)            echo's reply
commit genesis G0
emit {"to":"agent",…}  ─▶ INTAKE invoke agent R1
open slot G0/deps/G0/R1
release [R1, G0]       ─▶ EXECUTE agent: {"to":"echo",…}
                          INTAKE (successor) invoke echo R2
                          open slot R1/deps/R1/R2
                          release [R2, R1, G0]     ─▶ EXECUTE echo: hello
                                                      handoff ────────────────▶ INTAKE return "hello" R3
                                                                                SETTLE R2/result = R3
                                                                                GATHER R1's round: all discharged
                                                                                commit resolve R4 (deps: R3)
                                                                                CLAIM R1/resolved-R1: won
                                                                                EXECUTE agent, resumed: echo said "hello"
                                                                                ... reply R5 settles R1, resolve R6 of G0
boot wakes (kqueue on G0), prints the lineage, replays it: 7 records, byte-identical
```

## State on disk

```text
.skyra/
  records/<sha256>                  every record, content-addressed
  inputs/<sha256>                   the exact bytes each record was derived from
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
leaves. Unexpected provider requests become horizon faults.

## Decisions, and where they live

| Decision | Code |
|---|---|
| D1 the emitter is never the checker | `intake` interprets before any `Run`; `execute` passes output on uninterpreted |
| D2 custody; the parent exits after release | `handoff` releases only after `issueRelease` passes; `drive` releases the child and exits |
| D3 the stamp is a pointer | `record.Stamp`; `tree.Load` refuses a record that does not hash to its stamp; `checkSuccessor` |
| D5 one commit per handoff | only intake, a supervisor fault, and a resolve commit |
| D6 the release is the Horizon plus fds | `SendHorizon`: Horizon JSON plus the frame stack over `SCM_RIGHTS`; every write in a frame goes through its fd |
| D7 the sandbox is rendered from the Horizon; `.skyra/` always denied | `horizon.Render(self, floor)`; the floor is a parameter, not a Horizon field |
| D8 leaves | a surface with an empty reach: its output is always a reply |
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
  [improvement 03](improvements/03-conversation-is-an-invocation.md).
- **Parked frames live in the webhost's memory.** fds cannot be written to
  disk: if the webhost stops, a waiting conversation stays viewable but cannot
  continue until restore exists.
- **One output per step**, so one child per round: fan-out is supported by the
  slots and the claim, but nothing emits two children yet.
- **Restore after a crash**, streaming, and the host's own restart are not built.
- **The provider profile is pinned to Codex 0.157.1.** A new version needs a
  fresh validation of startup, tool exposure, file effects and egress.

## Improvements to carry forward

[`improvements/`](improvements/): lessons from DNA-19 for the next version.

- [Profiles become input contracts](improvements/01-profiles-become-input-contracts.md)
- [No silent refusals: a fault-tolerant async boundary](improvements/02-no-silent-refusals.md)
- [A conversation is an invocation: logs, history and isolation](improvements/03-conversation-is-an-invocation.md)
