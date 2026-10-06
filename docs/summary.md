# Skyra: where we are

Date: 2026-10-06
Status: current summary of the whole project: the properties it is after, the
patterns that produce them, what runs today, and what does not yet. The
[README](../README.md) is the front door and [`startup-doc.md`](startup-doc.md)
is how to run it. This file explains *why it works the way it does*.
([`design/summary.md`](design/summary.md) is an older snapshot from 2026-09-25.)

---

## In one paragraph

Skyra is an agent runtime whose principal is the **bounded, auditable agent**,
not the process. A human grants a **Horizon** once: an explicit set of what an
agent may invoke and touch. Everything inside it runs without approval,
including tools the agent builds for itself; anything outside it is refused by
the operating system. **Authority is never held, only derived**: at every step
the runtime reads the durable state and the incoming proposal and derives what
the step may do. A single `skyra` binary runs **once per step and exits**. Each
run takes custody of untrusted bytes, derives what they mean from where the
emitter stands, commits a content-addressed record, and runs the next program
under a macOS Seatbelt profile **rendered from its Horizon**. All durable state
is a hash-linked tree, and replaying it from genesis re-derives every decision
byte for byte. An agent can now **remember a strand as a procedure**: the
interpreter compiles work that already ran into a reusable program, installs
its contract only in the Horizon that remembered it, and runs it again without
the model while its refinements hold. Procedures are the emerging application
programming model: contracts are their callable interfaces, invocations are
their executions, actors provide bounded effects, and Horizons define which
compositions are expressible. Today Skyra runs real Codex conversations in the
browser, terminal and desktop app; remembered procedures are built and tested
with the scripted agent and fake Codex. Moving the conversation application
itself onto procedures, with state reconstructed from lineage endpoints, is
proposed in [next/proposal/33](next/proposal/33-procedure-adapter-extraction-proposal.md).

---

## The problem, and the move

Coding agents ask before they act: *allow this command? this file? this tool?*
Each prompt is cheap, but together they fail. People stop reading, approve by
reflex, and then switch to full auto, which trades the prompts for **ambient
authority**: a shell, the whole workspace and the network, inside a sandbox the
agent *reports* about itself. That is **tool acceptance fatigue**. Self-extension
makes it worse: an agent that writes its own tools produces a stream of new
code, and each piece wants its own approval.

**The move is to put the approval on the right object.** A human can
meaningfully decide *what an agent may touch*, once. They cannot meaningfully
review every action inside that. So the human decides the Horizon. Anything
created under it can only rearrange authority already granted, and the OS
checks this on every invocation, so nothing inside needs a prompt. The human is
asked only when the agent reaches past the Horizon.

**The contribution is coupling, not invention.** Permissions, plans and audit
logs all exist. Skyra welds them into the execution model instead of bolting
them on beside it. The full argument, from the six observed pains to the
invariants they force, is in [`grips-and-invariants.md`](grips-and-invariants.md).

---

## The properties we are going for

Each is something the system should make **true by construction**. A semantic
claim is never made unless a mechanism enforces it, or it is listed as a gap.

| # | Property | Means | Enforced by | Status |
|---|---|---|---|---|
| **P1** | **The emitter is never the checker** | whatever a model or tool outputs is untrusted bytes; a different process derives and commits it before anything acts on it | intake runs in the successor; `interpret` is the only path from bytes to state | built |
| **P2** | **Authority is derived, only narrows, and the OS enforces it** | every step runs under a Horizon derived from the tree and the proposal; a call shares its caller's Horizon unless it asks the room to narrow; nothing widens one | `machine.Narrow`, `machine.accept`, the room's `system.narrow`, `horizon.Render` | built |
| **P3** | **Truth is derived, not stored** | only observations are stored; everything else is re-derived by one pure function | `machine.Step` is pure; `replay` re-derives every record | built, with known gaps |
| **P4** | **Custody** | each piece of work is held by exactly one live process or one durable commit | stamp, release, exit; a parent exits only after its successor stamps | built |
| **P5** | **Live state is durable state** | a crash loses nothing that cannot be read back; a Horizon, the parked work, its authority are all on disk | one atomic commit per handoff; Horizons committed; resume derived from the tree; runtime I/O interruptions recorded as failures, never as discharges | built (restore after a host dies is not; an interrupted request is reported `delivery-unknown`, never repeated) |
| **P6** | **The runtime knows no actor** | Codex, the browser and the human live in adapters, never in the trusted core | package boundaries; the audit in [next/20](next/20-boundry-crossing.md) | built |
| **P7** | **No program starts another unrecorded** | exec is allowed only for the program being run; every other program starts through an emission the runtime recorded | the exec exclusion in every rendered profile, with an OS test | built |
| **P8** | **Frame ⊆ authority** | everything a model is shown is invokable; visibility grants nothing | `tools` is derived from reach and references | built |
| **P9** | **Approval-free self-extension** | an agent remembers a strand it ran as a procedure under its own Horizon, and runs it with zero prompts | `remember` compiles the strand; the interpreter names it with the Horizon's `extend-reach`, which needs no approval ([next/32](next/32-procedures.md)) | **built** (section H, 2026-10-03): the scripted demo and the fake Codex in `run.sh`; real Codex not yet run |
| **P10** | **Recurrence is the resource** | what was decided once is kept: a remembered procedure runs again on new input without the model, inside its refinement | procedures and refinements in the interpreter; base types computed only for what is remembered ([next/32](next/32-procedures.md)); replaces routes keyed by type ([next/30](next/30-types-by-origin-routes.md)) | **built** (section H, 2026-10-03): procedures run inside their refinements with no model step; `skyra metric` reads it from the records. Routes (2026-10-01) are removed |
| **P11** | **Hashes never reach a model** | a model sees names and presentations; hashes go only to workers the runtime starts | adapter projections; references are shown by names such as `system.ref.<n>`; Codex names a remembered call by its turn-local index | built, with the JSON-RPC correlation-id exception documented in F14 |
| **P12** | **An actor's state is its accepted lineage** | an actor can reconstruct its logical state from an immutable first invocation and its latest accepted endpoint, under its Horizon and versioned projector | explicit-anchor `system.strand`, monotonic accepted checkpoints and deterministic projection ([proposal 33](next/proposal/33-procedure-adapter-extraction-proposal.md)) | proposed; checkpoint publication and recovery are the first experiment |

P1–P3 are the README's three invariants. P4–P8 and P11 keep them true. P9 and
P10 are what the whole thing is for. P12 is the state model proposal 33 makes
explicit; it is not yet an implementation claim.

---

## The patterns, and how they produce the properties

### 1. One binary, once per step; a Horizon is the process

There is no long-running controller. One `skyra` process runs one step and
exits, and durable state lives in `.skyra/`. **A Horizon is our process**
([next/27](next/27-a-horizon-is-a-process.md)): the kernel's abstraction over
authority, a stack and durable state. OS processes are only the workers that
carry out its steps, and they are disposable: a worker that dies ends one step,
not the Horizon. The kernel is the ephemeral wrapper around every handoff.

```text
 INTAKE    receive unchecked bytes; derive the state from them and the Horizon they were emitted under;
           commit a record; reply with its stamp
 RELEASE   the parent reads the record back, checks it hashes to the stamp, derives the child's Horizon,
           commits it, and sends {its hash + the frame stack as directory fds} over SCM_RIGHTS.
           The child looks the Horizon up and refuses it unless its own commit derives exactly that Horizon
 EXECUTE   run the program under a Seatbelt profile rendered from that Horizon. Its output is not parsed here
 HANDOFF   spawn a successor with the raw output; open a slot if it is an invoke; exit once it has stamped
 SETTLE    a return or fault discharges the slot it answers, once
 GATHER    if every slot the waiting continuation opened is discharged, commit a resolve naming them all
 CLAIM     exactly one sibling wins the right to resume that continuation
```

### 2. A pure core and an effectful shell

```go
func Step(s Instance, e Event) (Instance, []Effect)   // machine: decides everything, touches nothing
func Drive(root, up, stack, s, e)                     // instance: performs effects, decides nothing
```

Events are facts (`Received`, `Released`, `Emitted`, `Stamped`, `Settled`,
`Claimed`, `StrandRead`, `Parked`); effects are requests
(`Commit`, `Run`, `Spawn`, `SendHorizon`, `Settle`, `TryClaim`, `ReadStrand`,
`Park`…). `TestMachineIsPure` fails if
`machine` imports anything that touches the OS, the clock or the tree. That
purity is what lets replay feed recorded inputs back and get the same records.

### 3. States derived from where the emitter stands (`interpret`)

A program never says "invoke", "return" or "fault". It emits operands,
`{"to": <address>, "payload": <JSON>}`, and the runtime derives the state
(next/14, [next/24](next/24-the-room.md)):

```text
addresses the slot its frame discharges (SKYRA_FRAME), any JSON                  → return
names a capability or reference in reach, fits its input contract                 → invoke
names anything else                                                               → fault (horizon)
addresses nothing                                                                 → fault (leaf)
not UTF-8, or not exactly {to, payload}                                           → fault (malformed)
does not fit the contract of what it addresses                                    → fault (contract)
no output                                                                         → fault (custody)
```

The emission is the target's first line of stdout, at most 16 MiB; anything
after it is discarded.

- **No name is reserved.** The slot is the frame's id, a stamp no capability
  can be named. The old `"reply"` word, which let an emitter state its own
  operator, is gone.
- **A leaf fault keeps what was printed.** A plain binary that knows nothing of
  Skyra can be registered; its output arrives as the fault's observation, and
  the parent decides whether it was a failure.
- **Every fault carries its observation** as its payload: the bytes that did
  not fit, the address outside the Horizon, what a program printed. A parent can
  pass it on.
- **A refused output is still committed,** as a fault. A stamp means custody,
  not approval.
- **A return declares nothing.** There is no output contract any more: the
  runtime computes a return's type from the return itself (section 13).

### 4. The Horizon: one authority, shared, derived

```go
type Horizon struct {
    Stamp     string             // this frame's latest commit
    Surface   *Surface           // what this frame runs: identity and contracts only
    Reach     map[string]Surface // what emissions under it may invoke
    Authority Authority          // the OS grants everything under it runs with
    Refs      map[string]Grant   // references received from the room
    Frame     string             // the slot a return discharges
    Resume    *Horizon           // the continuation waiting on that slot
}
```

- **The Horizon is the authority boundary, not the call.** Every call shares its
  caller's reach and grants. Being a leaf or a continuation is behaviour, not
  authority. To contain a tool more tightly, its caller narrows it.
- **The room is the root Horizon;** its reach is the registry table. Genesis is
  its creation, authored by the operator and carrying a nonce, so two boots of
  one registry are two histories. Every registry grant joins the room's one
  authority.
- **Narrowing goes through the room.** `system.narrow` is a facet of the room,
  implemented by the runtime: it returns a reference to a narrowed Horizon for
  one actor, the reference joins the caller's reach, and invoking it runs the
  actor under less. A narrowing that asks for anything not held is a horizon
  fault, derived at intake.
- **A surface is identity plus contracts** ([next/25](next/25-surface-refactor.md)):
  a binary pinned by hash and an input contract. It declares nothing about what
  it may reach, and nothing about what it returns.
- **`Render(Horizon, floor)`** denies by default, adds the Horizon's grants,
  and applies the exclusions no Horizon can grant or remove: `.skyra/` is always
  denied, and exec is allowed for the program being run and nothing else.
- **Committed, looked up, re-derived.** A Horizon is released by committing it;
  the child is handed only its hash. The hash is a handle, not authority: the
  child re-derives the Horizon from its own commit and refuses one that does not
  match.

### 5. Records, stamps, and the tree

```text
.skyra/
  records/<sha256>                  every record, content-addressed
  inputs/<sha256>                   the exact bytes each record was derived from
  horizons/<sha256>                 every Horizon released
  frames/<F>/result                 write-once: what discharged F's slot
  frames/<F>/deps/<round>/<child>   a slot F opened, written before the child runs
  frames/<F>/resolved-<round>       write-once fan-in claim: the resolve that won
  frames/<F>/emitted-<round>        write-once: the hash of the complete emission that opened the round
  frames/<F>/failure                write-once: a runtime I/O interruption, not a discharge
  interpreter/procedures/<key>      procedure bodies owned by the interpreter actor
  interpreter/frames/<frame>        durable procedure positions
```

A stamp is `sha256(record)`, a pointer, verified by reading the record back.
Each record names its predecessor and the hash of its input bytes (top-down
provenance); a resolve names every discharge it waited on (bottom-up). Only
intake, a supervisor fault and a resolve write records. **Replay** walks a
lineage from genesis and re-derives every record and every Horizon
(`replay.HorizonAt`) with the pure machine. A resolve must name distinct,
finished discharges, each settling a distinct slot of its round. A genesis of
version 2 also commits each round's complete emission, so replay re-derives
every slot it opened and checks that none is missing or out of order. Older
histories get the other checks but cannot prove round coverage.

Every object is published by an exclusive link of a fully written, synced
temporary file, and an existing object is re-verified against its hash. New
trees are readable only by their owner.

### 6. Custody travels as file descriptors

The release carries the Horizon's hash **plus the frame stack as open directory
fds** over `SCM_RIGHTS`. Every write in a frame goes through its fd, so a step
can only settle the frames it was handed. Writes are resolved relative to that
descriptor, never through the process's working directory, so concurrent
claims in one process are safe. Descriptors are close-on-exec
everywhere, because a sandboxed program that inherited a frame fd could list a
directory its profile denies. A parent exits only after the successor accepts
custody (P4).

### 7. Dependencies, slots and the join rule

An invoke opens a slot in the invoker's frame before the child runs; a return
or fault discharges exactly one. A continuation resumes only when every slot of
its round is discharged (wait for all; there is no cancellation), and the
resolve record must name every discharge, so it cannot be written early. When
siblings race, a write-once claim lets exactly one resume the continuation.
Faults are discharges too: the continuation decides what to do with one.
An emission may fan out to several children; the complete emission is
committed, then every slot is opened, before any child is released, and the continuation resumes once all of them discharge.
The interpreter uses this to run independent ready procedure nodes together.

### 8. Continuations are stateless

A continuation reads everything it knows on stdin and keeps nothing between
runs:

```json
{"frame": "<the slot it returns to>", "input": {...},
 "tools": {"<name or reference>": <input contract>, ...},
 "steps": [{"to": "...", "payload": {...}, "state": "return", "value": ...,
            "observation": "<discharge stamp>"}, ...],
 "contract": { ...the contract this frame was invoked through... }}
```

The discharge stamp is available to trusted workers so `remember` can name the
strand to compile; adapters do not show it to a model. A model names one of its
own calls by a turn-local index, and the adapter resolves that index.

Its memory is its lineage, so it can resume in any process, after any crash,
and replay reproduces it. ([next/27](next/27-a-horizon-is-a-process.md) gives
every program this one envelope; today a surface still declares
`"continuation": true` to receive it.)

### 9. Adapters: stateful actors behind stateless steps

The Codex app-server holds a turn open; a browser comes and goes. Each actor's
semantics live in an adapter (P6).

| Piece | Role |
|---|---|
| `codex.start` | a resident surface: the runtime launches one app-server per conversation under its registered profile (private home and scratch, egress only through an allowlisting proxy) and puts its connection into custody |
| `codex.exchange` / `inspect` / `stop` | `codexio`: one bounded, recorded operation on that connection. Every byte Codex said is in the lineage |
| `codex` | `codexresident`: stateless. It rebuilds the whole protocol state from its steps each run and emits exactly one next operation |

The adapter offers Codex only its model-facing tools (`echo`, `read`), emits
each tool call as an invoke of the tool it names, and returns a tool's result
or fault to Codex within the same turn. Any other provider request or native
item is a **violation**: the adapter interrupts the exact turn, confirms it
ended, and **returns the attempt to its parent as an observation** (`refused`),
so the conversation goes on in the same Codex thread. The app-server's profile
renders none of the Horizon's shared grants: its native reads would be
observations with no record.

### 10. Ingress by parking; the user outranks the agent

The browser is an async surface. Invoking it **parks** the work with the
webhost, which holds **a stamp and the descriptors, never authority**. When a
message arrives, `skyra resume` derives the parked instance and its Horizon from
the tree, refuses work that is not parked or was already resumed, and a
successor checks the message like any output, addressed to the parked frame's
slot. A human who types JSON is still text.

The user side is a continuation whose steps are the transcript, and Codex's
parent: Codex can only return to it, and only the user side ends the
conversation. Each conversation is its lineage: saved, listed, rebuilt from
records, replayable. The browser UI, the TUI and the desktop app speak the same
protocol to the same webhost.

**Ingress is authenticated.** Every launcher generates a session token unless
one is supplied, and the webhost refuses to start without one. Native clients
send it as a Bearer header; the browser carries it in a WebSocket subprotocol,
accepted only from the launcher's exact UI origin. Host headers must be
loopback, and frames are bounded and validated before allocation.

**Requests are admitted durably before work starts.** The webhost's request
ledger records an identity hash of each request, so reusing an ID for
different work is a `conflict`. A client keeps its draft until admission is
established. If the host restarts or the runtime is interrupted, the request is
reported `delivery-unknown`: external effects may already have happened, so
nothing is retried automatically. The ledger keeps every request in flight and
the newest 256 finished outcomes. Snapshots are kept within 16 MiB by leaving
older history out of the view; the history itself stays on disk.

### 11. Faults

| Kind | Means |
|---|---|
| `custody` | the protocol broke: no output, a dead successor, a failed stamp |
| `malformed` | the bytes are not operands |
| `horizon` | addressed something outside the Horizon, or asked for more than it holds |
| `contract` | the payload does not fit what it addressed |
| `leaf` | addressed nothing; the output is kept |

**Failure is not a fault.** A tool that could not do its job returns, for
example `read` returning `{"error":"denied"}`. **Asking for more authority is a
fault that climbs** (designed): like a page fault, it goes up until someone
holds the authority, with the human above the room. The application-level
handling for that escalation remains designed rather than built.

### 12. Observations

An observation is a value: the bytes a step received, kept verbatim and named by
hash. Everything the runtime acts on is derived from them. `read` is the first
program that brings observations in from the world, with provenance (path,
whole-file sha256 and size, line range). **Observation passing**
([next/21](next/21-observation-passing.md), proposed) would make provenance
structural: a continuation may pass its children's observations up, but not
author a replacement.

### 13. The procedure interface (built 2026-10-03)

[next/32](next/32-procedures.md) supersedes the type-and-route design in 30.
Types, routes, the Starlark runner and novelty interrupts have left the kernel.
Their replacement is a general programming interface implemented by the
interpreter actor:

```text
caller → contract → procedure → actors → recorded outcome
```

- **A contract is the callable interface.** A procedure is exposed as an input
  contract of the interpreter. The kernel sees an ordinary invoke and selects
  a contract by `(target, schema)` and structural fit; it knows nothing about
  procedures.
- **A procedure is the program.** It is a durable composition of calls,
  bindings, fan-out and returns. Its body and each running frame's position are
  held in the interpreter's actor state. Procedures may call other procedures
  exactly as they call any contract.
- **An invocation is the execution.** It gives one run an identity, frame,
  Horizon and lineage. Proposal 33 makes a conversation an enclosing,
  long-lived invocation whose turns remain descendants of the same anchor.
- **A Horizon is the programming environment.** It is the complete set of
  contracts a procedure can compose at that point. Every emitted instruction
  is checked at intake, and Horizons only narrow.
- **Actors are effectful primitives.** They own protocols, storage access,
  provider operations and live resource custody. Procedures own workflow and
  application policy.
- **A strand is reconstructable history.** `system.strand` walks from an
  observation back to the invocation made through a named contract, preserving
  Horizon opacity. `remember` compiles that bounded trace into a procedure.

`remember` is itself an interpreter contract. The caller supplies the trusted
worker with the observation and original contract; the kernel returns the
authorized strand; the interpreter compiles its calls and bindings; then
`system.extend-reach` installs the resulting contract in the caller's Horizon.
An actor can extend only itself, and extension adds a contract, never
authority. A procedure therefore exists only in the Horizon that remembered
it, while every later emission is still checked under the Horizon in which it
runs.

Remembered procedures have base refinements derived from the original calls
and values. When the originating actor exposes `refine`, it may generalize
those edges with bounded CEL predicates. If a refinement refuses, execution
falls back to the original actor through `resume` when available. While the
refinements hold, recurrence costs no model call.

Proposal 33 extends the same interface from remembered work to authored
application workflows such as `conversation.list`, `load`, `create`, `send`,
`close` and `delete`. Its state rule is:

```text
ActorState = Project(Strand(Horizon, first_invocation, accepted_endpoint))
```

The first invocation is immutable identity; the accepted endpoint is the
actor's causal frontier. Together they are sufficient to reconstruct recorded
logical state when every state-relevant fact is in lineage and the projector
is versioned. They do not recreate live descriptors, processes, unjoined child
results or unobserved external effects. Explicit-anchor strand reads,
checkpoint advancement and crash-gap reconciliation are proposed, not built.

---

## One real turn, end to end

*"Read `README.md` and tell me what this project is"*, in `skyra tui`:

1. **The human types.** The webhost resumes the parked browser invocation by
   its stamp; `skyra resume` derives it from the tree. The message is intaken as
   a return to that slot. The user continuation resumes and emits
   `{"to":"codex","payload":{"text":…,"thread":…,"incarnation":…}}`.
2. **The runtime derives `invoke codex`.** Codex runs under the conversation's
   Horizon; the parent commits it and hands over its hash; the child re-derives
   it before anything runs.
3. **The adapter** rebuilds state from its steps and emits `codex.inspect`, then
   `codex.exchange` with `turn/start`, each an invoke of `codexio` on the
   connection in custody, each observation recorded.
4. **Codex asks for `read`.** The adapter checks it is offered and emits
   `{"to":"read","payload":{"path":"README.md"}}`.
5. **Intake derives `invoke read`.** It runs in a fresh process under the
   Horizon's profile (the workspace readable, `.skyra` denied, exec of nothing
   else) and emits `{"to":<its slot>,"payload":{"path":…,"sha256":…,"text":…}}`.
   Intake commits the derived return.
6. **Settle, gather, claim.** The return discharges the adapter's slot; once
   its round is complete, exactly one claimant resumes it. The adapter pairs
   the result with Codex's pending call and delivers it.
7. **Codex finishes.** The adapter returns the final text to its slot; the user
   continuation shows it and parks on the browser again.
8. **Every step is a record.** `skyra replay` re-derives all of them, byte for
   byte.

Asking for `~/.ssh/id_rsa` returns `"error":"denied"`, from the kernel, not the
tool. Asking Codex to run a shell interrupts the turn, and the attempt comes
back to the user side as an observation.

---

## What runs today

| | Status |
|---|---|
| The step protocol, custody handoff, content-addressed hash-chained records | **built**, tested |
| `interpret`: states derived from the emitter's position; the slot is the return; typed faults that carry their observation | **built** (2026-09-30) |
| Input contracts, strict structural fit at intake; output contracts removed | **built** (outputs removed 2026-10-01) |
| Fan-out: one emission may supervise several successors; all slots open before release and results resume in emission order | **built** (2026-10-03) |
| The Horizon as `system`; contracts selected by `(target, schema)`; `system.extend-reach` lets an actor add only its own contract | **built** (2026-10-03) |
| Bounded `system.strand` reads, preserving requester-Horizon visibility and narrowed-child opacity | **built** (2026-10-03) |
| Interpreter actor: installed procedures, durable frame positions, bindings, `each`, procedure-to-procedure calls | **built** (2026-10-03) |
| `remember`: compile a completed strand and expose it only in the remembering Horizon | **built** (2026-10-03) |
| Remembered procedure execution with base refinements, model fallback and bounded CEL refinements | **built** (2026-10-03) |
| The hash audit: no hash in what a model reads; references named `system.ref.<n>` | **built** (2026-10-01), one accepted exception |
| One authority per Horizon, shared by every call; profiles rendered per Horizon; `.skyra/` and exec excluded | **built** (2026-09-30), with OS tests |
| The room: genesis authored with a nonce; `system.narrow` and references that join the caller's reach | **built** (2026-09-30) |
| Horizons committed and looked up by hash; the child re-derives its Horizon | **built** (2026-09-30) |
| Resume derived from the tree; the host holds a stamp, never authority | **built** (2026-09-30) |
| Dependencies, slots, resolve, fan-in claim | **built** |
| Replay re-deriving every record and every Horizon | **built** |
| Real Codex: one resident app-server per conversation, sandboxed, egress-proxied, connection in custody, idle restart | **built** |
| Codex native attempts: interrupted and returned as an observation; the conversation continues on the same thread | **built** (2026-09-30) |
| `echo` and `read` as model-facing tools; a tool's result or fault back to Codex in the same turn | **built** |
| Conversations as lineages in the browser, the TUI and the desktop app | **built** |
| `run.sh`: scenarios for authority, fan-out, procedures and Codex behavior, each ending in a byte-identical replay and checked against its expected outcome | **built** |
| Replay rejects unfinished, repeated, unrelated, omitted or reordered discharges; version 2 histories commit each round's complete emission | **built** (2026-10-06) |
| Authenticated ingress in every launch mode; bounded, validated WebSocket framing | **built** (2026-10-06) |
| Durable request admission, conflict detection, `delivery-unknown` after a restart or runtime interruption; client drafts kept until admission | **built** (2026-10-06) |
| `skyra test`: every module, frontend, Swift, website replay and demos in one command; the same script in macOS CI | **built** (2026-10-06) |

## What does not run yet

| | Status |
|---|---|
| **Procedure-native conversation application**: list/load/create/send/close/delete enter through interpreter contracts rather than webhost/user/Codex lifecycle code | proposed ([proposal 33](next/proposal/33-procedure-adapter-extraction-proposal.md)) |
| **Actor reconstruction from lineage endpoints**: explicit immutable invocation anchor plus monotonic accepted endpoint; checkpoint publication and crash reconciliation | proposed; first experiment in proposal 33 |
| **Authored workflow control**: versioned bundles, conditional nodes, explicit fault handlers and an event-driven repeat/resume construct | proposed for the conversation application |
| **Provider/session extraction**: bounded Codex operations, durable provider-home identity and safe recovery of the same enclosing invocation | proposed; real restart/resume must be proven |
| **Faults that climb to the human**; the room's admin facet | designed ([next/24](next/24-the-room.md), [next/26](next/26-agents-and-facets.md)) |
| **One stdin envelope** for every program; the `continuation` flag removed | designed ([next/27](next/27-a-horizon-is-a-process.md)) |
| **Facets** declared at registration; calls into another agent's Horizon | designed ([next/26](next/26-agents-and-facets.md)) |
| **Supervision and restart** from durable state | designed ([next/27](next/27-a-horizon-is-a-process.md)) |
| **Identity**: canonical encoding, content ids and path ids | designed ([improvements item 9](improvements.md)) |
| **Real Codex authoring refinements** | not built; real Codex is offered procedures and `remember`, but H8 was exercised with fake Codex |
| **Observation passing** by id with enforced provenance | proposed ([next/21](next/21-observation-passing.md)) |
| **Restore** after a webhost dies; tail calls; a write tool | not built; fan-out is built. A restart reports unfinished requests as `delivery-unknown` instead of reviving them |

The complete list of changes, with what each became, is in
[`next/proposal/`](next/proposal/README.md).

## Known gaps, stated honestly

- **A shared Horizon contains a tool only as tightly as its caller.** Under one
  Horizon, `read` addressing `codex` derives an invoke, not a fault. That is the
  deliberate cost of making the Horizon the unit of containment; the remedy is to
  narrow, which is built but not yet used by the conversation path.
- **The stack still travels as a copy.** A Horizon is committed with its
  `Resume` chain, so the stack is durable, but the in-memory copy has not been
  replaced by reading the stack from the tree.
- **The webhost's orphan rule** (a record never released is dropped) is a local
  rule; one definition of an accepted record, checked by replay, is
  [improvements item 8](improvements.md).
- **The floor covers only this run's `.skyra`.** The `skyra` command refuses a
  workspace that overlaps other state; a state-root floor would remove the need.
- **Records still store derived fields** (state, to, payload, fault). Replay
  checks them, but they are not yet pure observations.
- **`Args` is still on the surface.** The launcher passes the real app-server's
  `--disable` flags through it; it becomes slots the owner fills (next/28).
- **Some emissions must remain singular.** A fan-out cannot return, carry a
  resident connection or contain `narrow`/`extend-reach`, because those cases
  require one owner or would make the waiting Horizon nondeterministic.
- **The provider profile is pinned to Codex 0.160.0.**
- **Procedure versioning is incomplete.** There is one remembered
  representation per contract; authored bundles, migrations and versioned
  source/node/position formats are proposal 33 work.
- **The procedure language is still a finite DAG.** It has bindings and
  `each`, but no conditional branch, explicit fault handler or durable
  event-driven repetition yet.
- **An interpreter frame position is durable actor state, but actor-state
  recovery is not yet reduced to lineage endpoints.** Proposal 33 must define
  who advances an accepted checkpoint, how stale writers are refused and how a
  crash after acceptance but before publication is reconciled.
- **Round coverage is unprovable for version 1 histories.** Replay of a
  history someone else supplies is checked only against the genesis it
  declares, so a supplied history can declare the older version.
- **Large conversation views are truncated.** A snapshot over 16 MiB leaves
  out the oldest history, and the clients do not yet say so; the host logs it.
- **Durability is ordered, not certified.** Tree and interpreter publications
  are synced before they are relied on, but no power-loss test has been run.
- **A Horizon bounds the blast radius; it does not judge intent.** A prompt
  injection that misuses a granted tool is not stopped; the evaluation must
  separate the two cases.

---

## The research question

> **Can human approval move from every action to one grant of a Horizon, so
> that an agent extends itself without asking, while matching a conventional
> runtime on cost and task success and containing strictly more?**

The baseline is the same Codex model in its own approval modes: *ask on
request* and *full auto*. Skyra should beat the first on **human interruptions
per task** and the second on **containment**, at once. Under a shared Horizon
the unit of containment is the Horizon: the claim is that effects outside it
succeed zero times, and the measure of narrowness is the narrowings a task needs.
Details: [`next/research-question.md`](next/research-question.md).

Procedures add structural measures without pretending to judge whether the
model's answer is good:

- **Model calls displaced by recurrence.** Count procedure runs, successful
  refined edges and fallbacks to the originating model. Repeated accepted
  structure should cost procedure execution rather than another inference.
- **Remembered coverage.** Measure how much later work runs inside remembered
  procedures, and where their refinements refuse. Work should increasingly be
  paid for by novelty rather than total task length.
- **Authority preservation.** A remembered procedure invoked under a narrower
  Horizon must either complete inside it or fault at intake; it must never
  recover the authority of the run from which it was compiled.

`skyra metric` reads the procedure measures from records. Proposal 33 adds a
second research direction: how much long-horizon application state can be
reconstructed from an invocation anchor and accepted endpoint without a
parallel mutable database.

| When | Work |
|---|---|
| September | semantics: contracts, `interpret`, dependencies and resolve (**done**) |
| October | Codex calling tools under generated sandboxes (**done**); the room and derived authority (**done**); fan-out and remembered procedures (**done**); procedure-native conversations and endpoint reconstruction (**proposed**) |
| November | audit and replay; the baseline harness and task set |
| December | evaluation and writeup |

---

## Where to read next

| To understand | Read |
|---|---|
| why any of this exists | [`grips-and-invariants.md`](grips-and-invariants.md) |
| the code, package by package | [`controller/README.md`](../controller/README.md), starting at [`controller/core/machine/machine.go`](../controller/core/machine/machine.go) |
| authority derived from state, the room, faults that climb | [`next/24-the-room.md`](next/24-the-room.md) |
| agents, facets, narrowing | [`next/26-agents-and-facets.md`](next/26-agents-and-facets.md) |
| a Horizon as the process | [`next/27-a-horizon-is-a-process.md`](next/27-a-horizon-is-a-process.md) |
| procedures, remembering and refinements | [`next/32-procedures.md`](next/32-procedures.md), [`proposal/32`](next/proposal/32-procedures-proposal.md) |
| procedure-native conversations and lineage endpoint reconstruction | [`proposal/33`](next/proposal/33-procedure-adapter-extraction-proposal.md) |
| types that compose, slots, blame; actor state | [`next/28`](next/28-types-observations-checks.md), [`next/29`](next/29-actors-own-their-state.md) |
| what is built and what is not, change by change | [`next/proposal/`](next/proposal/README.md) |
| the boundary rules every new crossing must answer | [`next/20-boundry-crossing.md`](next/20-boundry-crossing.md) |
| the reasoning, in the order it happened | [`design/thoughts/01–13`](design/thoughts/), then [`next/14–32`](next/) |
| the predecessor, and what it taught | [How it got here](../README.md#how-it-got-here); the code is in git history (last at `b243b03`) |
