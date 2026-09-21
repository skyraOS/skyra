# The OS-kernel frame — what to adopt, what to re-impose

Date: 2026-09-21
Status: Load-bearing ideas. Discussion input, not a decision. Reconciles an
externally-derived "process machine for agents" summary with the Skyra
constraints established in this folder. Companion to
[mediated-actors-over-a-kernel.md](mediated-actors-over-a-kernel.md),
[ports-as-the-location-abstraction.md](ports-as-the-location-abstraction.md), and
[engrams-surfaces-and-pipe-protocol.md](../engrams-surfaces-and-pipe-protocol.md).

## The frame is right

An outside model (no knowledge of Skyra) independently derived the same shape:
build a small **process machine** for agents from OS primitives, not another
agent framework.

```
EXECUTABLE (inert behavior)
    -> spawn -> PROCESS (live instance)
        -> communicates through PORTS (address + authority)
            -> resolve through BINDINGS (indirection)
                -> other processes / resources
```

- **Executable** = inert code; no authority because it isn't running.
- **Process** = a live instance. Agents and tools are the same thing at different
  lifecycles ("tool" = short-lived process, "agent" = long-lived process,
  "subagent" = child process, "tool call" = spawn -> request -> response -> exit).
- **Port** = a stable, addressable, capability-scoped interface.
- **Binding** = `port -> current implementation`.
- **Kernel** owns executables, processes, ports, bindings, permissions, spawn,
  and termination.

Small kernel API: `spawn / port / bind / send / unbind / kill`. This is Skyra's
own thesis ("the possibility is a process; make the kernel the first real
driver"). The independent re-derivation is evidence the shape is real. Adopt it
as the ACT/OS substrate.

## Adopt outright: port space = Horizon

The frame's strongest contribution is **ports as a per-process address space**,
not one global table. Process A's `port 3` and process B's `port 3` are different
things; a process can only reach what is mapped into *its* port space. This is
task-local Mach ports / process-local fds, and it realizes object-capability
confinement directly: no ambient authority; you can only name what you were given.

It collapses two prior concepts into one:

> **A process's port space *is* its Horizon.** "If it isn't mapped in, it can't be
> reached" and "what an agent can't touch is the horizon's complement" are the
> same statement. The per-process port table is the capability set.

Keepers that are exactly right:

- **Authority disappears with the process.** Lifecycle is the authority
  lifecycle; `kill` is the revocation primitive.
- **Same executable, different port spaces, different capabilities.** This is
  profile-vs-invocation plus Horizon, made concrete.
- **Structural, not behavioral, security.** Don't trust a prompt that says it
  won't touch production — don't put a production port in its space.

## Re-impose: three constraints the outside view could not know

### 1. Topology manipulation is itself capability-checked, and attenuating

The frame's one internal contradiction: it says *"don't give the process a
production port"* (confinement) while also saying *"the home process spawns,
creates ports, binds, grants — orchestration is just process and port
manipulation."* These fight. **If a process can mint and bind its own ports,
confinement is fiction** — it grants itself a production port and walks out.

Resolution (the mediated-actor rule): `spawn / bind / create_port / kill` are
**kernel** operations, never calls a process makes directly. A process only
*requests* them, as Moves the kernel authorizes against that process's current
port space. And capability flow is **attenuating**: a parent may delegate only a
subset of what it holds; broader authority is minted by the kernel, never by a
peer. Without this, "reachability = authority" leaks the moment any process
orchestrates.

### 2. The wire is async Moves, not synchronous `send -> response`

The frame's `send(port, task); receive(port)` is RPC-shaped. DNA-17 proved that
does not hold: real sessions emit **asynchronously and unprompted** (a Codex turn
starts, calls a tool mid-turn, resumes). The unit is a **Move that re-enters the
loop**, not a call that returns a value. The port is right; the `-> response` is
too synchronous. Everything crosses as a Move; correlation and resume are
protocol, not return values. (Only a Move crosses a pipe.)

### 3. This is the ACT/OS half — keep the THINK/engram layer on top

The frame models execution and authority well but has **no compiled-cognition
layer**: profile -> **Engram** (the serialized, ordered, memoizable instruction
list), THINK-compiles-once, "recurrence is the resource." Skyra needs both halves.

> **The Engram is what a Move compiles into before the kernel dispatches it. The
> OS runs engrams; it does not replace them.**

Do not let the clean process-machine picture erase the compile layer.

## Minor

- The process-table `pid` must be a runtime-assigned monotonic id, never the OS
  pid (reuse/aliasing — the thing ports exist to escape), and never an *address*.
  Only ports address.

## Mapping to the small kernel

| Frame primitive | Skyra reading |
|---|---|
| Executable map (name -> executable) | The executable namespace; entries are inert until spawned. |
| Process table (id -> executable/state) | Backings (level 3); id is runtime-assigned, not an address. |
| Port table (port -> protocol/permissions/owner) | Per-process port space = that process's Horizon. |
| Binding table (port -> process/resource) | The rebindable `port -> backing` layer; kernel-owned. |
| Supervision (spawn/wait/restart/kill) | The supervisor that rebinds ports across backing death. |
| `send(port, message)` | Emit a **Move** (async, re-enters the loop); not RPC. |
| `spawn / bind / port / kill` | Kernel ops; reachable to a process only as authorized Moves. |

## The one line

> **GPT built the kernel; the mediation rules and the cognition layer are what
> make it Skyra and not just a process manager — adopt port-space-as-Horizon, but
> keep topology capability-checked and attenuating, the wire async Moves, and the
> engram/THINK layer on top.**
