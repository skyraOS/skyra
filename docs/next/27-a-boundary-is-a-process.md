# A Boundary is a process

Proposal: **Not in the proposal.** Written after the proposal. Its order (one envelope, supervision, near and far in the record) is not yet in [next/proposal](proposal/README.md).

Date: 2026-09-30  
Status: agreed in discussion, not built. Builds on [24](24-the-room.md) and
[26](26-agents-and-facets.md). Settles the `continuation` flag left by
25.

## The rule

> **A Boundary is our process.** It is the kernel's abstraction over authority,
> a stack and durable state, the way an OS process is the OS's abstraction
> over memory and credentials. OS processes are only the workers that carry
> out its steps, and they are disposable.

Our kernel is the ephemeral controller: a wrapper around every handoff. Every
emission crosses it, is derived, and is committed before anything runs. It
keeps our definition of a process separate from the OS's.

## The two kinds of process

| | OS process | Boundary |
|---|---|---|
| Defined by | the OS kernel | our kernel (the ephemeral controller) |
| Holds | memory, credentials, descriptors | authority (reach, references, grants), its stack, its history: all durable state |
| Lives | while it runs | from its creation record until it returns |
| Executes as | itself | any number of short-lived OS processes, one per step, each under the Boundary's rendered profile |
| A crash kills | the process | one step. The Boundary goes on from durable state |
| Supervised by | its parent PID | the Boundary that created it |

A Boundary does not map to one OS process. Executables and registrations are
state; the kernel starts an OS process per step, as it does now. Hosting many
invocations in one OS process would need them loaded into one address space
(a shared library, a plugin, a WASM module), which is what Agoric does for the
objects of a vat. We do not need it: a Boundary is defined by the kernel, not by
an address space.

## Near and far invokes

An invoke either stays in the current Boundary or goes to another one. The
emitter does not say which. It emits `{"to", "payload"}`, and the derivation
decides from what `to` names:

| `to` names | The invoke | The child runs under |
|---|---|---|
| a name in this Boundary's reach (near) | stays in this Boundary | this Boundary |
| a reference from the room, such as a `system.narrow` result (far) | creates a Boundary | the narrowed Boundary |
| another agent's facet (far, 26) | enters that agent's Boundary | that agent's Boundary |
| this frame's slot | a return | — |

Both kinds cross the kernel the same way and are recorded the same way. The
record states which it was, because the derivation knows, even though the
emitter's syntax is the same.

## Every call is recorded

Calls inside a Boundary are recorded, like every call, and recording is only a
write to state. That is the difference from Agoric, which records only what
crosses the kernel and replays a vat by determinism. Recording every call:

- keeps invariant 1: the emitter is never the checker, because the checker is
  always the kernel, outside every worker;
- keeps workers free of any determinism requirement: a private method (26) is
  a recorded step like any other;
- makes supervision cheap: a supervisor restarts from what is on disk, and
  never reconstructs memory.

## The creator supervises

The Boundary that creates another supervises it:

- **The room supervises the programs it starts.** Genesis is the room's
  creation; everything the operator seeds runs in a Boundary the room made.
- **A Boundary made by `system.narrow` is supervised by whoever narrowed.** The
  caller that invoked the reference is the one waiting on it.
- **Supervision is over durable state.** A worker that dies ends one step. The
  supervisor reads the Boundary's state and runs the step again (at least once
  for effects, exactly once for records; 24, "Crash semantics"), or faults it.
- **A Boundary ends with a return** into its supervisor's history (24,
  "Creation belongs to the creator").

## One envelope for every worker

Because every invocation inside a Boundary is the same kind of thing, every
worker reads the same stdin:

```json
{"frame": "<the slot it returns to>", "input": <its payload>, "tools": {...}, "steps": [...]}
```

- The `continuation` flag goes (25, item 1). There is one format, not two to
  choose between.
- A program that invokes and is resumed is read the same way on every run.
  Today a "leaf" that invokes, which a shared Boundary allows, is resumed with
  a format it was not started with.
- `SKYRA_FRAME` stays as a convenience for simple tools; `frame` in the
  envelope says the same.
- A plain binary that knows nothing of Skyra understands neither format. Its
  output is a leaf fault either way, and it keeps what it printed.

## Prior art

- **Erlang.** An Erlang process is not an OS process: the BEAM defines it and
  runs many on a few OS threads. Supervision attaches to the Erlang process,
  never the thread; a supervisor restarts a child from a known state. A
  Boundary is defined by our kernel and runs as a stream of OS processes;
  supervision attaches to the Boundary.
- **E and Agoric.** A vat is one authority boundary holding many objects. A
  call to a near reference stays in the vat; a send to a far reference goes
  through the kernel to another vat. Invoking is one idea, and the reference
  decides which. `vatAdmin.createVat` gives the creator an `adminNode`: the
  creator supervises.
- **Waldo et al., *A Note on Distributed Computing* (1994).** Treating local
  and remote calls as the same is dangerous: remote ones can partially fail
  and are slower. E's answer, and ours, is to keep one concept of invoking but
  make the difference visible: in E in the reference's type, here in the
  derivation and the record.

## What changes

| Where | Today | Here |
|---|---|---|
| `release` (`machine.go`) | a continuation gets the envelope, anything else its bare payload, by `Surface.Continuation` | every worker gets the envelope |
| `Surface.Continuation`, `Declared.Continuation`, `"continuation": true` in four contracts | the stdin format | removed |
| `echo`, `read`, `codexio`, the resume probe | read their bare payload | read `input` from the envelope |
| `interpret` / the record | a reference invoke and a name invoke look alike in the record | the record says whether the invoke stayed in the Boundary or created or entered one |
| supervision | a supervisor fault discharges the step's own frame; nothing restarts it | the creator Boundary supervises: restart from durable state, or fault |
| docs | "an instance per step", "a surface's process" | a Boundary is the process; an instance is a worker |

## Open questions

- **What the record adds for a far invoke.** A field naming the Boundary
  entered or created, or a derived fact replay recomputes. It changes the
  encoding, so it belongs with improvements item 9.
- **Restart policy.** Erlang supervisors choose one-for-one, one-for-all, and
  limits on restarts. Which of these a Boundary's supervisor needs, and where
  the policy is written: in the creator, or in the room.
- **When a step is run again.** Recovery needs to tell a worker that died from
  one that is still running. Custody (write-once claims) settles who acts; the
  trigger (a supervisor noticing) is open.

## Order

1. One envelope for every worker; the `continuation` flag removed.
2. The docs say a Boundary is the process and an instance is a worker.
3. The record states near or far, with improvements item 9.
4. Supervision and restart, with 24's crash semantics.
