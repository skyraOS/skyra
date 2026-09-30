# Skyra: where we are

Date: 2026-09-30  
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
byte for byte. Today it runs real Codex conversations in the browser, the
terminal and a desktop app; Codex calls `echo` and `read`, and every turn is a
replayable lineage.

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
| **P5** | **Live state is durable state** | a crash loses nothing that cannot be read back; a Horizon, the parked work, its authority are all on disk | one atomic commit per handoff; Horizons committed; resume derived from the tree | built (restore after a host dies is not) |
| **P6** | **The runtime knows no actor** | Codex, the browser and the human live in adapters, never in the trusted core | package boundaries; the audit in [next/20](next/20-boundry-crossing.md) | built |
| **P7** | **No program starts another unrecorded** | exec is allowed only for the program being run; every other program starts through an emission the runtime recorded | the exec exclusion in every rendered profile, with an OS test | built |
| **P8** | **Frame ⊆ authority** | everything a model is shown is invokable; visibility grants nothing | `tools` is derived from reach and references | built |
| **P9** | **Approval-free self-extension** | an agent registers a tool it wrote under its own Horizon and runs it with zero prompts | `system.register`, a Starlark tool run by the runtime's checker | **not built: the thesis demo** |
| **P10** | **Recurrence is the resource** | what was decided once is kept: checks learned at novelty, procedures remembered | checks and bindings as records; the log as the test suite | designed ([next/28](next/28-types-observations-checks.md)) |

P1–P3 are the README's three invariants. P4–P8 keep them true. P9 and P10 are
what the whole thing is for.

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
`Claimed`, `Parked`); effects are requests (`Commit`, `Run`, `Spawn`,
`SendHorizon`, `Settle`, `TryClaim`, `Park`…). `TestMachineIsPure` fails if
`machine` imports anything that touches the OS, the clock or the tree. That
purity is what lets replay feed recorded inputs back and get the same records.

### 3. States derived from where the emitter stands (`interpret`)

A program never says "invoke", "return" or "fault". It emits operands,
`{"to": <address>, "payload": <JSON>}`, and the runtime derives the state
(next/14, [next/24](next/24-the-room.md)):

```text
addresses the slot its frame discharges (SKYRA_FRAME), fits its output contract   → return
names a capability or reference in reach, fits its input contract                 → invoke
names anything else                                                               → fault (horizon)
addresses nothing                                                                 → fault (leaf)
not UTF-8, or not exactly {to, payload}                                           → fault (malformed)
does not fit the contract of what it addresses                                    → fault (contract)
no output                                                                         → fault (custody)
```

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
  a binary pinned by hash, an input contract, an output contract. It declares
  nothing about what it may reach.
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
```

A stamp is `sha256(record)`, a pointer, verified by reading the record back.
Each record names its predecessor and the hash of its input bytes (top-down
provenance); a resolve names every discharge it waited on (bottom-up). Only
intake, a supervisor fault and a resolve write records. **Replay** walks a
lineage from genesis and re-derives every record and every Horizon
(`replay.HorizonAt`) with the pure machine.

### 6. Custody travels as file descriptors

The release carries the Horizon's hash **plus the frame stack as open directory
fds** over `SCM_RIGHTS`. Every write in a frame goes through its fd, so a step
can only settle the frames it was handed. Descriptors are close-on-exec
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

### 8. Continuations are stateless

A continuation reads everything it knows on stdin and keeps nothing between
runs:

```json
{"frame": "<the slot it returns to>", "input": {...},
 "tools": {"<name or reference>": <input contract>, ...},
 "steps": [{"to": "...", "payload": {...}, "state": "return", "value": ...}, ...]}
```

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
holds the authority, with the human above the room. Interrupts are designed
next (section 13).

### 12. Observations

An observation is a value: the bytes a step received, kept verbatim and named by
hash. Everything the runtime acts on is derived from them. `read` is the first
program that brings observations in from the world, with provenance (path,
whole-file sha256 and size, line range). **Observation passing**
([next/21](next/21-observation-passing.md), proposed) would make provenance
structural: a continuation may pass its children's observations up, but not
author a replacement.

### 13. Designed next: types, checks and learned structure

[next/28](next/28-types-observations-checks.md) settles how structure is
learned:

- **Types are contracts; observations are values.** Types compose: a field can
  hold a value or a type, and filling slots (the args of the execve) makes a new
  type. `coder` is Codex's `continue-conversation` with its input and its 5–7
  tools filled. A continuation is a type, not a flag.
- **The parent owns the check** on what its children return. There are no
  preconditions: the runtime checks structure, and the callee defends itself in
  its own code. Different parents can check the same contract differently.
- **Novelty is when a check is written.** The first time a boundary is crossed
  there is no check: a novelty interrupt locks the parent (typically an LLM) in
  to write one, in **Starlark**, pure computation with no ambient authority, run
  by a runtime-owned checker. It is kept; the next crossing uses it.
- **Failures carry a trace and blame**, and the parent's handler picks a
  restart: fault, substitute (checked again), or retry within the same turn.
- **The log is the test suite.** A rewritten check runs against every recorded
  verdict at its boundary; flipped verdicts are declared, never silent.
- **Self-extension falls out of it**: an agent-written tool is Starlark source
  in state, run by the same checker, so code is state, the floor holds, and the
  exec exclusion holds.

Identity is designed in [improvements item 9](improvements.md): domain-separated
hashes, a content id beside every stamp, ids as the path of handoffs, types and
principals as records.

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
   Intake checks the object against read's output contract and commits a return.
6. **Settle, gather, claim.** The adapter resumes, pairs the result with
   Codex's pending call and delivers it.
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
| Input and output contracts, strict structural fit at intake | **built** |
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
| `skyra demo`: eight scenarios, each ending in a byte-identical replay | **built** |
| About 6,300 lines of Go in the controller (standard library only), 72 test functions | — |

## What does not run yet

| | Status |
|---|---|
| **Self-extension** (`system.register`): an agent registers a tool it wrote and runs it with zero approvals | **not built: the thesis demo.** Designed as Starlark run by the runtime's checker ([next/28](next/28-types-observations-checks.md)) |
| **Checks, novelty interrupts, handlers** | designed ([next/28](next/28-types-observations-checks.md)); `core/machine` has no interrupt state yet |
| **Faults that climb to the human**; the room's admin facet | designed ([next/24](next/24-the-room.md), [next/26](next/26-agents-and-facets.md)) |
| **One stdin envelope** for every program; the `continuation` flag removed | designed ([next/27](next/27-a-horizon-is-a-process.md)) |
| **Facets** declared at registration; calls into another agent's Horizon | designed ([next/26](next/26-agents-and-facets.md)) |
| **Supervision and restart** from durable state | designed ([next/27](next/27-a-horizon-is-a-process.md)) |
| **Identity**: canonical encoding, content ids, path ids, types as records | designed ([improvements item 9](improvements.md)) |
| **Stored procedures and memory** (P10) | designed, deferred |
| **Observation passing** by id with enforced provenance | proposed ([next/21](next/21-observation-passing.md)) |
| **Restore** after a webhost dies; tail calls and fan-out; a write tool | not built |

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
- **One output per step**, so one child per round.
- **The provider profile is pinned to Codex 0.157.1.**
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

| When | Work |
|---|---|
| September | semantics: contracts, `interpret`, dependencies and resolve (**done**) |
| October | Codex calling tools under generated sandboxes (**done**); the room and derived authority (**done**); checks and **self-extension** |
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
| types, checks, learned structure | [`next/28-types-observations-checks.md`](next/28-types-observations-checks.md) |
| what is built and what is not, change by change | [`next/proposal/`](next/proposal/README.md) |
| the boundary rules every new crossing must answer | [`next/20-boundry-crossing.md`](next/20-boundry-crossing.md) |
| the reasoning, in the order it happened | [`design/thoughts/01–13`](design/thoughts/), then [`next/14–28`](next/) |
| the predecessor, and what it taught | [How it got here](../README.md#how-it-got-here); the code is in git history (last at `b243b03`) |
