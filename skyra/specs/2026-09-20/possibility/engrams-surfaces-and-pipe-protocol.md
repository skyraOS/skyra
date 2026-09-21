# Engrams, Surfaces, and the pipe protocol

Date: 2026-09-20

Status: Discussion refinement. Records the new composition and executable
boundary. Existing proposal structs have not yet been rewritten to match it.

## Read/write profiles belong to bounded units

Every independently bounded unit declares a read and a write:

- **read:** what state needs to be read.
- **write:** what state needs to be written.

**Reads occur before writes, and required writes complete before the unit
returns its outcome.** This is the profile concept being established. A profile
declares the operations; their concrete values can depend on the invocation,
read results, and the unit's decision.

```text
bounded unit
    declared reads
    -> decision / result preparation
    -> declared writes
    -> return
```

For example, an authority check reads installed authority and caller context,
decides whether the Move is admitted, writes the corresponding decision, and
then returns its outcome. A successful authorization stamp must not be written
on the denial branch.

The profile is a protocol/composition requirement. It must not be interpreted
as requiring externally authored executables to implement Go `read()` and
`write()` methods inside the runtime binary. The representation and owner of
the declaration remain to be specified.

The phase order is relative to the bounded unit's turn. Ordering alone does not
establish transaction isolation; consistency across related reads/writes remains
a separate contract. Whether a unit declares empty read/write sets, and the exact
boundaries of a unit and turn, remain to be formalized.

## Engrams are the serialized lists

**An Engram is the serialized list.** It composes addressable executable units.
The previous discussion called this serialized work an execution context; Engram
is the name now assigned to the composition.

**Codex `start-thread` is an Engram, not a Surface.** It describes composed work
performed through executable endpoints. Its exact entries and ordering still
need to be written down; this note does not invent that list.

```text
Engram
    serialized composition
        -> addressable executable
        -> addressable executable
        -> ...
```

Do not read the arrows as settling wire encoding, branching, result binding, or
whether execution receives the entire list versus successive ready portions.
Those remain open contracts from the serialization discussion.

## Surface means an addressable executable endpoint

Call each addressable executable a **Surface**. Its minimal addressing shape is
an endpoint integer: "some endpoint int."

```text
Surface
    endpoint: int
```

This is a conceptual address shape, not a finalized Go declaration or a claim
that the integer is an OS file descriptor. Scope, allocation, lifetime, reuse,
and its relationship to the older UUID SurfaceID remain open. The endpoint
number names executable behavior; the number alone is not proof of authority.

Composition belongs to the Engram. Do not retain `start-thread` as a compound
Surface merely because the previous Surface struct could contain Dependencies
and Contracts. Nor should that old struct be treated as the required object
layout of an external executable.

## Anything executable goes over pipes

**Anything executable goes over pipes.** A model can author executable behavior
that is not part of the runtime binary. The runtime does not control its native
methods or require it to implement an in-process language interface.

The executable must know how to parse a Move and speak the Skyra protocol.
That protocol is the interoperability boundary.

```text
runtime
    -> pipe: protocol Move
    -> executable endpoint
    -> pipe: protocol output / emitted Moves
    -> runtime
```

Language-specific methods and private implementation details remain inside the
executable. Runtime-side interfaces may still help implement transport and
dispatch, but they do not define an interface that external code must implement
inside the runtime process.

This replaces the assumption that the executable entries in an Engram must be
Go handler objects or function pointers. The earlier preference for a single
handler list needs to be expressed as addressable protocol execution across
pipes, with concrete wire representation still to be defined.

The rule establishes an execution boundary, not a process-count policy. It does
not yet specify whether one process hosts one endpoint or several, or how long
each process lives. Framing, protocol validation, and process/endpoint binding
also require contracts; speaking the protocol does not itself establish trust.

## What this revises

| Earlier description | Current refinement |
|---|---|
| `start-thread` is a higher-level Surface | `start-thread` is an Engram. |
| Surface carries composition, native Driver methods, and profiles in one struct | Surface is the addressable executable endpoint; Engram carries composition. Profile declaration placement remains open. |
| Execution context is a serialized list of Moves and their handlers | Engram names that serialized composition; executable entries are addressed through the pipe protocol. |
| Runtime can require external behavior to implement its Go interface | Runtime requires Move parsing and Skyra protocol conformance across pipes. |

The prior structs and provider proposals remain historical starting points.
This note records the refinement without silently rewriting all of them.

## Carry-forward constraints

- The loop owns authority; everything it emits for execution is authorized.
- The Possibility supplies the run's authority namespace. Invocation-stack
  bookkeeping need not travel with the Engram's execution handoff.
- Preparation is a trap; a fault is an error.
- No unresolved prerequisite reaches execution. Engram ordering and readiness
  still need a precise operational contract.
- StateOwner is an executable Surface speaking the same protocol; it does not
  acquire a separate public invocation path for state requests.

## Next contracts

1. Engram entry representation, encoding, ordering, and result correlation.
2. Endpoint integer scope/lifetime and its binding to a pipe and process.
3. Where bounded-unit read/write profiles are declared, validated, and stored.
4. What constitutes one bounded unit and which operations form an atomic turn.
5. Pipe framing, Move schema, response/error behavior, and protocol versioning.
6. How newly authored executable code is registered, provisioned, and bound into
   a Possibility's Horizon without becoming part of the runtime binary.

## Related

- [Possibility investigation](README.md)
- [Loop-owned authority and serialization](loop-owned-authority-and-move-serialization.md)
- [Earlier execution-context refinement](authority-namespace-and-execution-context.md)
- [Consistency review before this refinement](consistency-review.md)
