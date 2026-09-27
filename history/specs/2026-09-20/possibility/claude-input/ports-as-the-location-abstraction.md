# Ports as the location abstraction

Date: 2026-09-20
Status: Load-bearing ideas. Discussion input, not a decision. Records why the
persistent identity of an executable must be a runtime-owned port, not a process.
Companion to [mediated-actors-over-a-kernel.md](mediated-actors-over-a-kernel.md)
and [engrams-surfaces-and-pipe-protocol.md](../engrams-surfaces-and-pipe-protocol.md).

## Why identity cannot be a PID

Once almost everything executed is an out-of-binary executable reached over a
pipe, the runtime must know where everything is at all times. That stable
identity **cannot be a process**, for the same reason it exists at all:

- A PID is **OS-owned** — you do not get to pick or reserve one when you want it.
- A PID is **ephemeral** — it dies with the process.
- A PID is **recycled** — a dead PID's number is later handed to an unrelated
  process. Routing to "the old PID" and hitting a new process is the classic
  aliasing bug.

A PID is therefore a *backing-side, ephemeral, OS-assigned* handle. Identity has
to be a name **you** own and assign — a stable index in **your** table. That is
the port. This re-derives Erlang registered-names over Pids (a supervisor
re-registers on restart), Mach port names over tasks, and DNS names over IPs:
mature systems refuse to let the OS-given instance handle be the identity.

## The three levels

Conflating these is what made "it can't be a process" hard to place. Separated,
the ephemerality problem disappears.

1. **Port** — the stable, capability-scoped endpoint: an endpoint int plus its
   entry in the runtime's port table. Moves are addressed to the port, never to a
   process. This is what you assign a backing to. **Stable.**
2. **Binding** — the current association `port -> backing`. Rebindable. Process
   dies -> the binding breaks; the port persists and is rebound. **Kernel-owned.**
3. **Backing** — the process/executable that parses Moves and does work.
   **Ephemeral. Never the identity.**

"The runtime moves actors into ports and decommissions them from ports" is the
lifecycle of level 2. It is a **kernel operation**: it mutates the port table,
which is control state, not adapter state. An actor may *request* a spawn via a
Move; only the kernel binds.

This is the same stable-descriptor / ephemeral-instance shape seen throughout:
**Port : Process :: Profile : Invocation :: Engram : dispatch-run.**

## You already built the port table

DNA-17's `Identifier` (opaque uint64) plus `fdEntry` (scope / generation /
liveness in the writer) *is* a port plus a port table. It was scoped to one
execution; the generalization is to make it the runtime's authoritative
"where is everything" map — `port -> {authority scope, liveness/generation,
current binding}` — with its lifetime decoupled from any process. That map is
substrate, in-binary, and cannot itself be reached over a pipe (there must be a
statically known root: a small set of well-known/bootstrap ports — the loop,
StateOwner — like Mach's bootstrap port).

## Trap 1: port reuse must not alias

If port 7 is decommissioned and later bound to an unrelated actor, a stale Move
or a stale grant aimed at "7" now hits the new occupant — the exact bug PIDs were
abandoned to escape. Two clean fixes:

- **(a) Never reuse an index.** Allocate monotonically (uint64; inexhaustible). A
  stale reference hits a dead entry and is rejected, never aliased. Cost: the
  table grows and dead entries need eventual GC.
- **(b) Index + generation.** The port is `(7, gen)`; references carry the gen
  they were minted against; delivery rejects a mismatch. DNA-17 already has this
  (`Identifier` + `fdEntry.generation`, "zero is never a live fd").

Pick one and **scope Horizon grants to it**, so a decommission cannot silently
transfer a capability to a new occupant. (This is also where the port meets
authority: a port in a Horizon is a stable capability; because the port survives
rebind, a grant survives a backing crash+restart without rewriting the Horizon.)

## Trap 2: a port persists a NAME, not a MAILBOX

"Persistent identity = port" tempts fusing two different kinds of persistent:

- Persisting the **name** is cheap — a table entry that outlives the backing.
- Persisting the **queue** (buffering Moves while nothing is bound) is a
  durability project.

Decide the unbound-port semantics deliberately. A Move to a decommissioned port
either:

- **faults** (live-only — start here), or
- **buffers until rebind** (durable mailbox — later; bounded, with a defined
  "never rebound" answer).

Take the persistent-name win now; do not accidentally sign up for durable queues
before you mean to.

## One fork to name

"Move an actor into a port" — is it ever the *same* actor relocating from port A
to port B (migration / re-addressing), or only *a port's backing getting
(re)bound* (same port, swapped occupant)? Almost certainly the latter: port
stable, backing swapped, senders never notice. Migration (an actor keeping its
state while changing address) is a separate, harder capability; adopt it only if
actually wanted.

## Rebinding across death: the supervisor

Ephemerality is handled not by the port but by a **supervisor**: when a backing
dies, the supervisor spawns a fresh process and rebinds the same port. The port,
the Horizon, and every Move-sender are untouched. Mach supplies the
endpoint/capability half; Erlang-style supervision supplies the rebind-on-death
half.

## How this lands on the findings

- **C10** (disposable executions vs. persistent logical work): resolved by the
  Port/Backing split. The backing is disposable; the port is the persistent
  logical identity. The old one-session `FD` / single-child `WaitingOn` fields
  were trying to be both at once.
- **C11** (execute-time liveness fence vs. the state-read boundary): if liveness
  becomes "the port's binding is live / the generation matches" checked at the
  transport/table, the fence moves off the canonical-read path and the old
  contradiction dissolves.

## The one line

> **The port is the one name you own; the process is the occupant the kernel
> binds and decommissions; guard reuse with a monotonic index or a generation;
> and remember a port persists a name, not (yet) a queue.**
