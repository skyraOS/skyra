# SurfaceRegistry

Date: 2026-09-20

Status: Code-on-paper proposal. Component ownership is proposed; method contracts
and process placement remain open. No runtime changes.

## Purpose

Grants name SurfaceIDs. SurfaceRegistry holds the registered Surfaces and their
handler bindings so those identities can resolve to executable behavior.

The registry is long-lived relative to disposable workers. Its name describes
its responsibility, not an OS process. Whether it and StateOwner share a process
is separate from their being distinct components; the current starting point is
that they do share the long-lived runtime host.

"Durable" needs a further contract: stable Surface identity can survive handler
replacement, but an in-memory registry does not survive a runtime crash. Durable
registration metadata and reconstruction of live handlers remain open. A handler
pointer or live process resource is not a persistable registration record.

## Registry, Horizon, and StateOwner

| Component | Responsibility |
|---|---|
| SurfaceRegistry | Hold registered Surface definitions and their handler bindings. |
| Horizon | Determine which registered capabilities a Possibility may invoke or admit results from. |
| StateOwner | Serialize canonical reads and writes, including registry changes and execution admission records. |
| Executor | Execute the binding supplied by committed work. |

Registration does not grant permission. A grant does not establish that a live
handler exists. Interpretation resolves both conditions as one admission
decision, under the state turn's locking discipline.

```text
Incoming Move + trusted caller context
    -> Possibility.Horizon checks the crossing
    -> resolve the permitted SurfaceID through SurfaceRegistry
    -> prepare the exact binding and state changes
    -> commit
    -> execute that binding outside the lock
```

The checks and resolution above are parts of one interpretation turn, not an
external permission check followed by a freely available handler lookup.
The executor must not perform a fresh canonical registry lookup after commit.

## Proposed storage boundary

StateOwner holds a SurfaceRegistry component. The registry owns the table's
meaning; StateOwner owns serialization of canonical access. This avoids two
independent writers for Surface registration and authorization state.

The first declaration uses `SurfaceID -> Surface`, matching DNA-17's table and
the revised Surface struct's `Driver` field. No separate Handler abstraction is
introduced yet. The registered Driver supplies behavior; a disposable worker
process is an execution resource, not the registered Surface itself.

Registry maps and raw driver references are private runtime implementation
details. Surfaces do not receive a general `Resolve(id) -> handler` interface
that bypasses Horizon. The interpreter-facing restricted read profile and the
executor's opaque authorized binding still need concrete contracts.

TrapController registers as a Surface here; Horizon's TrapControllerID names it.
Its provisioning and authority are still open in the parent proposal. Registration
alone does not solve recursive traps or authorize controller execution.

## Next questions

These are the immediate questions for the next discussion. Earlier placement
suggestions in this draft are starting assumptions, not answers to these questions:

1. **Is SurfaceRegistry a persistent process or not?** Is it a process at all,
   or a component hosted in another process? What must outlive worker execution?
2. **What is its involvement in provisioning a Possibility?** What does it
   supply, validate, or bind, and who installs the Possibility's Horizon?
3. **What is the difference between provisioning a Possibility and registering
   a Surface?** Which records and authority does each operation establish, and
   can either happen independently of the other?
4. **Does SurfaceRegistry belong to what we call the kernel?** Clarify whether
   kernel means the trusted runtime responsibilities, a component, or a process.

Further registration and lifecycle questions:

1. Who may register, replace, or retire a Surface, and how do those Moves enter?
2. Is the registered Driver a shared adapter, a factory, or a resolver for
   Possibility-specific handlers? Do not infer shared provider sessions from this map.
3. What pins a binding across commit and execution if registration can change?
   Generation checks, retirement, and in-flight ownership remain to be defined.
4. What registration metadata is persisted and how are handlers rebuilt at boot?
5. What happens when a granted Surface has no available implementation?

## Sources

- [Declaration](proposal.go)
- [StateOwner proposal](../state-owner/README.md)
- [Current DNA-17 table and surfaceOf](../../../../../../src/skyra/skyra-dna-17/state.go)
- [Earlier StateOwner userState.surfaces](../../../../../2026-09-18/MessageBroker/proposal.go)
- [Revised Surface structure](../../../../../2026-09-19/surface.md)
- [Interpret/commit rule](../../../../../2026-09-19/improvements.md)
