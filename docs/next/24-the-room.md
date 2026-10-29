# The room: authority derived from state and proposal

Proposal: **In the proposal.** Tracked in [next/proposal](proposal/README.md) as B1–B11; built in part.

Revision (2026-10-28): "the room" is retired as a name. What this document
calls the room is the root Boundary, which boot writes as genesis
(`runtime/boot`). Its syscalls (`system.narrow`, `extend-reach`, `trace`, was `strand`)
are the kernel's, answered at every step, and live in
`core/machine/system.go` (was `room.go`). The rest of this document keeps
the old name.

Date: 2026-09-30  
Status: agreed in discussion; built in part on 2026-09-30 (see the [proposal status](proposal/README.md#status)): one authority per Boundary, the profile per Boundary with the exclusions, Boundaries committed and looked up by hash, resume derived from the tree, genesis authored. Follows 14
and [20](20-boundry-crossing.md), and builds on items 1, 2, 8 and 9 of
[improvements](../improvements.md). It supersedes 14's "a return is invoking
a single-use `reply` capability".

## The rule

> **Authority is never held. It is derived from where you stand and what you
> propose.** A read of the durable state plus the incoming proposal derives
> the authority a step has, at every step. A hash is only a handle to what
> that derivation already determines.

```text
authority(X) = derive( state at predecessor(X), proposal(X) )   … back to genesis
```

There is no token to steal. Anyone can compute X's authority from the log,
and the answer is always what X's position allows, never more than its
parent's. Escalating would mean changing the chain, and the chain is hashed.
This is the same move items 1 and 2 make for state: `interpret` derives the
state from position and envelope, and nothing an emitter says stands in for
it.

Two corollaries:

- **Live state is durable state.** Everything a step acts on is on disk and
  re-derivable. Any process, the root included, can die, and nothing is lost
  that cannot be read back.
- **Nothing comes from nowhere.** Every identifier resolves to a record, and
  every record traces to an authored origin (improvements, item 9).

## The Boundary is the execution context

A Boundary is the execution context of one program for as long as it runs.
It is not a message built fresh at every step.

- **Authority is re-derived at every step, and the Boundary changes only when
  one is created.** Each step, the derivation either matches the current
  Boundary or creates a new one.
- **Every invocation shares its parent's Boundary,** unless the parent
  requests a narrowing. An ordinary invoke (a frame pushed), a return (a
  frame popped), a fault, a resolve: the child is handed the **same** Boundary
  hash.
- **A narrowing is requested, never implied.** Only a parent asking for less
  authority creates a new Boundary, and the child is handed a **different**
  hash. Nothing about the callee (what it is, whether it is a leaf) narrows
  anything. The request is an invoke of the room (below).
- **Each Boundary has exactly one head: its creation.** The creation marks the
  start of the program.

A frame is where a call returns; a Boundary is the authority it runs under.
Every invoke opens a frame. Only a requested narrowing opens a Boundary
(improvements, item 3).

### One Boundary, one authority

Everything invoked under a Boundary runs with the whole of its authority:
reads, writes, sockets and reach alike. Being invoked inside a Boundary is
already within its authority, so a call is not narrowed any further.

- **The Seatbelt profile is rendered from the Boundary,** not from the
  surface. Every program under one Boundary gets the same profile, except for
  the exclusions below.
- **Reach is the capability registry:** the names a node can interact with.
  It is authority, and it is not coupled to what a program is. A leaf and a
  continuation under the same Boundary hold the same reach; whether a program
  uses it is its behaviour, not its grant.
- **The containment boundary is the Boundary, not the call.** Like a process:
  every function in it shares its authority, and dropping privileges means
  starting a narrower one. To contain a tool, run it under a narrowed Boundary.
- **A Boundary's id is its authority, its parent and its input contract.**

  ```text
  boundary id = hash( authority, the invoke that created it, its input contract )
  ```

  The invoke is a stamp, unique by its chain, so no two Boundaries share an id
  and no second id is needed. The per-step cursor (`Stamp`) is not part of
  it. The root has no parent: genesis is authored and carries its own nonce
  (item 9).
- **The stack lives with the Boundary, in durable state.** Today's
  `frames/<F>/deps`, `result` and `resolved-<round>` already are the durable
  stack; the `Resume` chain in the struct is an in-memory copy of it, and
  goes. A step reads what it needs from the stack to compute, then writes to
  it. The writes are **write-once entries** (a slot opened, a result, a claim
  won), never an overwrite, so siblings running in parallel cannot clobber
  each other and a crash never leaves a half-written stack.

### Narrowing goes through the room

A narrowing is an invoke of the room's `system.narrow` facet, on the same
protocol as every other call. It returns a reference to a new, narrowed
Boundary, which the parent then invokes. The envelope stays `{"to",
"payload"}`. See [26](26-agents-and-facets.md).

### Exclusions

Special cases live in a few exclusions the runtime applies to every rendered
profile, whatever the Boundary says. No Boundary can grant them, and nothing
can register them.

- **The floor.** `.skyra/` is always denied (as today, `Render`).
- **Exec.** A profile allows `process-exec` of the one program being run and
  nothing else. Exec can never be registered into the runtime. Otherwise a
  target could exec `codex` directly: within the authority it shares, but
  with **no record**. Only the runtime starts a program, and only through an
  emission it interpreted.

### Creation belongs to the creator

The narrowing invoke is the last entry on the parent Boundary's stack. Its
creation of the child Boundary is recorded **in the parent's history**, and
maintained there.

- **The child starts clean.** "Boundary created" is not an event on the
  child's stack, the way a process never sees its own `fork`.
- **The child's head links back.** Its creation names the invoke that caused
  it as its predecessor, so its origin is provable without being part of its
  execution.
- **Finishing is a return.** When the child program ends, its return is
  recorded in the parent's history, where the pending invoke is. A return is
  the same thing whether it pops a frame inside a Boundary or crosses back out
  of one.

### Handed as a hash, looked up by the child

The Boundary is committed durably, content-addressed, and the child is handed
its hash. The child's first act is a lookup: it loads the Boundary by that
handle, checks that it hashes to its name, and re-derives it from the state
and its own proposal. What it finds is its whole execution context. The
runtime builds the `Boundary` struct after that read, as the in-memory form of
the derivation, but the struct is never a source.

The lookup and the commit are the runtime's own operations. They are not
programs in any Boundary and hold no reach: they are what derives Boundaries.

## The room

The root Boundary is **the room**. Its Boundary is the registry table: the
whole reach. Every other Boundary is derived from it, and it is the one
position that holds all authority. It is the trusted base.

- **Genesis is the room's creation**, an act authored by the operator. It is
  the one origin, and it carries what makes that act unique (the time of the
  grant, a nonce), so every stamp after it is globally unique (item 9).
- **The room is a position, not a process.** It can die like anything else.
  On a return it is woken, reads its state, sees which child the return came
  from, derives what follows, and exits. The controller stays stateless:
  there is no long-lived trusted process, only a root position on disk.
- **The room does the handing over.** Parents never pass authority to
  children directly; every Boundary a child receives is one the derivation
  from the room produced. (Agoric's kernel is the precedent: vats never hand
  each other capabilities; the kernel translates every reference.)

The room's own reach can grow, by registration (below). Every other Boundary
only narrows from its parent. Authority, as opposed to reach, never grows
anywhere except by the human.

## Asking for more authority is a fault

A program never asks for authority. It attempts, and the level above
decides. The precedent is the **page fault**: a process touches an address
it does not have, the fault goes up to the kernel, and the kernel either maps
the page and resumes it, or delivers SIGSEGV.

- A child attempts something outside its Boundary. The runtime **derives** a
  boundary fault; the child never states one.
- The fault discharges the child's frame, so the parent sees it in its
  steps. If the parent holds the authority (it had narrowed the child), it
  can invoke again with less narrowing (the page is mapped, and the work
  resumes). If not, it faults upward.
- The fault climbs until it reaches whoever holds the authority. Above the
  room is the human, who grants a new Boundary or refuses
  ([grips](../grips-and-invariants.md): "Escalation is renegotiation, never a
  grab").

This needs no new mechanism. A fault is already a discharge, and already
travels up the stack. It keeps the contract rule:

> **One input contract, one return contract, and fault as the one outcome
> every call can end in.**

### Faults carry their observation

Every fault carries the observation of why it faulted: the bytes that did
not fit, what was addressed outside the Boundary, what a program printed. The
parent chooses what to pass on ([21](21-observation-passing.md)): the whole
fault as one observation, or the observation singled out from inside it.

**The leaf fault.** A program that emits without addressing anything did not
follow the protocol. The runtime still handles it: it derives a `leaf` fault
carrying the output. A plain binary that only prints (`ls`, a script) can
then be registered and used without knowing Skyra's protocol, which helps
extension, and the parent decides whether what it printed is a failure. A
program that follows the protocol addresses its slot, so a structured return
never depends on guessing whether unaddressed text is JSON.

### The alternatives considered, and why not

- **A return that requests a Boundary change** would give one invocation
  several return contracts, and make a return mean two things.
- **A kernel child process to invoke for more authority** is an interrupt by
  another name: a special capability whose meaning is an operator. The fault
  already says it.

## Registration: free inside your Boundary

A program may register as much as it likes under the Boundary it was given
(improvements, item 6: a tool "can hold no authority its creator does not
already hold").

The line that makes this safe is **reach versus authority**:

| | What it is | Registration |
|---|---|---|
| **Reach** | the names you may invoke | grows |
| **Authority** | the OS grants of the Boundary | unchanged: a tool carries no grants of its own; it runs with the Boundary's |

Registering names something you could already do: the tool runs under your
Boundary's profile, so it can do nothing you could not. It widens nothing, so
it needs no approval and is not a fault.

- **Registering is an invoke** of a registration capability in your reach.
  The derivation checks the tool (bytes pinned by hash, contracts that pass
  `Check`) and creates a **new Boundary**: same authority, larger reach. Its
  creation is recorded in your history, and you are handed the new hash, like
  every other Boundary creation.
- **A registration that reaches past your Boundary is a boundary fault,**
  derived and climbing like any other: for example, naming an exclusion.
  There is no separate registration fault.
- **A tool is scoped** to your Boundary and to the Boundaries derived from it.
  To share it upward, return its id; the parent registers it under its own
  Boundary, which it always can, since its authority contains yours.

So the whole system has two cases:

- **Inside your Boundary, anything:** invoke, register, narrow, delegate. All
  of it derived, none of it approved.
- **Beyond it, a fault:** one kind, climbing to whoever can grant, with the
  human at the top. The human approves the boundary once, not every tool.

## Custody: who acts as a step

Derivation settles *what authority* a step has. It does not settle *who acts
as* that step: two processes can both derive X's authority correctly. That
is custody, and write-once storage decides it, as the fan-in claim already
does. The first to commit wins; a duplicate either lands on the same stamp
(deduplicated) or is a conflicting sibling that replay rejects.

That is also why frame handles can become hashes: once replay rejects any
mark that no record backs (item 8, direction 1), uniqueness comes from
write-once storage, not from who holds a descriptor. Until then, frame fds
stay fds.

Resident connections (the Codex socket at fd 3) stay descriptors: a live
socket cannot be a hash.

## Crash semantics

Everything is on disk, so recovery is a read:

| Found | Meaning | Recovery |
|---|---|---|
| a record, and no derivation accepted it | committed, never released | an orphan: it never ran; drop it (this retires the webhost's local rule, `533d98d`) |
| an accepted step, and no successor | released; the instance died while running | run it again from its Boundary |
| an accepted step and its successor | handed off | continue from the successor |

Records are exactly once, since they deduplicate by stamp. Effects are **at
least once**: a target that wrote a file or called Codex before dying can run
again. The runtime should say so.

## Identity

From improvements, item 9, which this depends on:

- **Hash spaces are domain-separated**: a type, a record, a frame and a
  Boundary can never share an id.
- **Every step has two hashes**: a content id (what happened, no
  predecessor) to show the same work as the same across runs, and a stamp
  (content id plus predecessor) to prove the path.
- **An id is the path of handoffs**: a child's id starts with its parent's,
  segmented by frame. Back is dropping a segment, forward is a prefix query,
  a diff is the longest common prefix, and replay walks the ids.
- **Types and principals are records.** Every contract field is an
  observation type; Codex and the human are principals whose stable id is
  their first declaration.

## Returns address the slot

From items 1 and 2:

```text
addresses the slot its frame discharges (the frame id)   → return
names anything else in reach                              → invoke
names nothing in reach                                    → fault (boundary)
addresses nothing                                         → fault (leaf), carrying the output
```

The word `"reply"` stops meaning anything. Codex's attempt at a native tool is reported
as what it is, an observation returned to the parent, not a name chosen to
trigger a fault.

## Where the code differs

| Where | Today | Here |
|---|---|---|
| `machine.Narrow` (`machine.go:456`) | builds a new Boundary at every step, returns and faults included | derives at every step; creates a Boundary only for a requested narrowing or a registration; otherwise the child shares its parent's |
| `handoff` → `issueRelease` → `SendBoundary` | the parent instance computes the child's Boundary and sends it as JSON over the socket | the derivation from the room decides; the Boundary is committed and the child is handed its hash |
| `release` / `settle` → `accept` (`machine.go:330`) | checks the stamp and the frame count | loads the Boundary by hash and re-derives it |
| `boot.Run` (`boot.go`) | genesis is `{Input: Sum(reach), Reach: reach}`: no author, not unique | genesis is the room's creation, authored by the operator |
| `interpret` (`machine.go:285`) | `"reply"` decides return; unaddressed output is a string return | the slot decides return; unaddressed output is a `leaf` fault carrying it |
| `boundary.Render` (`boundary.go:157`) | renders the profile from `h.Surface.Authority`: one per surface | renders from the Boundary: one per Boundary, plus the exclusions (floor, exec of the running program only) |
| `boundary.Boundary` (`boundary.go:126`) | authority, stack position and chain cursor in one struct | authority only (reach and grants), identified by authority + creating invoke + input contract; `Stamp` moves out, and the stack (`Frame`, `Resume`) is read from durable state |
| `runtime/registry` | read once at genesis; the registry cannot grow | the room's table; registration is an invoke that creates a Boundary |
| frame fds | the capability over marks | a handle; replaced by hashes after item 8's direction 1 |
| `hosts/web` | parked frames die with the host | a parked frame resumes from its Boundary on disk |

## Open questions

- **The shape of a narrowing.** `system.narrow`'s input contract: dropping
  names from reach, dropping OS grants, or both, and in what form.
- **Is the room's derivation run by whichever instance is woken,** or by a
  root instance started for the purpose? Either way it holds no state of its
  own; the question is which binary and which sandbox.
- **Truncation.** The chain detects edits, not deletion of the newest steps.
  Something must anchor the frontier, for example the operator's own record
  of where a history ended.
- **Registration's scope in replay.** A registered tool's bytes must be in
  the tree for replay to verify its hash. Where do they live, and who may
  write them?
- **Existing lineages.** Every stamp changes (item 9). Migrate once, or
  version the rules.

## Order

1. Items 1 and 2: returns address the slot; no change to encoding.
2. The Boundary committed and handed by hash, re-derived by the child, with
   the crash rules. Together with item 8's direction 1.
3. Boundary creation only on narrowing, and the room as the derivation's root.
4. Item 9's identity: domains, two hashes, authored genesis, path ids.
5. Registration inside the Boundary, then faults climbing to the human (the
   thesis demo, item 6).
