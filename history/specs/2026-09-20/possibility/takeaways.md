# Possibility discussion takeaways

Date: 2026-09-20

Status: Discussion checkpoint. Established rules, proposed directions, and open
questions are distinguished below. No runtime implementation changes.

Follow-up: [authority namespace and execution context](authority-namespace-and-execution-context.md)
refines the execution handoff to serialized Moves and their handlers, ideally one
handler list, and makes state preparation use the ordinary Move execution model.

## Established: a Surface declares its pre-invocation operations

**A Surface declares what system-level operations—reads and writes—need to occur
before an invocation can occur.** These operations concern adapter state. They
do not let the Surface decide runtime authority or invocation settlement.

Interpret performs the required reads and prepares the writes. Commit applies
the prepared writes. If required state is unavailable, the invocation is
**trapped**. Only after the required operations complete and commit does the
invocation reach execution.

```text
Incoming Move
    -> interpretation and Horizon admission
    -> declared pre-invocation reads and prepared writes
    -> commit
    -> execute with complete inputs and bindings

Required state unavailable -> trap
```

Authority checks and preparation may need related state reads; the diagram does
not finalize their internal ordering or interfaces within interpretation.

**No operation/Move arrives at the executor with an unresolved dependency needed
to run.** The executor does not look up missing adapter state or wait for
prerequisite work. Waiting for an external operation's response is different
from waiting for missing execution prerequisites.

This is stronger than merely keeping state locks out of workers. The rejected
framing was that an executing worker discovers missing adapter prerequisites,
parks, and later resumes. Those prerequisites must be resolved upstream or the
invocation is trapped. What corrective work follows a trap remains open.

For a `start-thread` Surface, its declaration identifies the existing adapter
state and pending records needed to issue the operation. A newly returned
`threadID` is an execution outcome to record for subsequent work; it is not an
unresolved prerequisite of the original thread-creation execution.

## Established distinctions

- **Adapter-state ownership:** the state belongs logically to the Surface.
  StateOwner can store and serialize it without owning its application meaning.
  Surface-defined adapter operations do not authorize changes to Horizon,
  invocation parentage, or authoritative runtime settlement.
- **Preparation and execution:** state work and external execution have separate
  processing resources. Reads and writes stay together in one consistent state
  turn; there is no separate read pool and write pool.
- **Worker and invocation lifetime:** workers are disposable request/response
  shells. Continuity needed after a worker exits must live outside its private
  process memory. A logical invocation is distinct from an individual worker.
- **Registration and grants:** registering a Surface establishes a capability's
  definition and implementation binding. Registration alone does not grant a
  Possibility permission to use it.
- **Provisioning and use:** the provisioner can install restricted access once;
  routine use need not ask it to mint permission again. The provisioned boundary
  must still enforce scope and applicable liveness rules.

## Proposed direction, still being tested

1. **SurfaceRegistry holds registered definitions and implementation bindings.**
   Execution does not occur at the registry map. Provisioning reads from it to
   establish a Possibility's permitted execution contexts.
2. **Horizon combines permission and executable resolution.** A caller should
   not obtain an executable binding independently of its installed authority.
   The current direction places provisioned execution contexts on Horizon;
   the exact representation remains open.
3. **Possibility owns its invocation structure.** The map indexes a branching
   call stack of frames, dependencies, and return destinations. StateOwner may
   store and mutate those records without taking away that logical ownership.
4. **Execution belongs to a Possibility, with external machinery carrying it
   out.** This is a working ownership statement, not a settled process layout.
   A Possibility can outlive disposable workers; whether and how it spans other
   processes remains part of the authority/isolation discussion.
5. **Start the state pool with one worker.** The separate-pools sketch proposes
   one long-lived serialized state-turn processor alongside concurrent execution.
   DNA-17 currently uses a mutex-serialized StateExecutor, not this proposed pool.
6. **Keep SurfaceRegistry and StateOwner separate as services/components.**
   Registration and state provisioning may require coordinated changes. Separate
   interfaces do not yet imply separate OS processes.

## Correction to the three-role layout

The proposal folders were composed into authority owner, execution broker, and
agent/Surface worker to test those boundaries. They are not a settled architecture.

**Placing StateOwner under authority owner was too coarse.** Storing and
serializing protected state is different from deciding or provisioning authority.
StateOwner may instead be a provisioned resource service: each Surface context
receives a restricted interface to its own state. It remains responsible for
isolation and consistency without becoming the source of grants.

Likewise, Surface is an abstraction that crosses trust roles. Its definition,
adapter state, trusted preparation, and execution behavior need not reside in
one component merely because they appear in one struct.

The existing directories remain a record of the hypothesis. Read them with this
checkpoint; no physical reorganization is implied by these notes.

## Open questions

1. **State centralization:** must all adapter state use StateOwner, or is it an
   offered capability? What continuation state must survive worker disposal?
2. **Operation declarations:** what are the exact permitted reads/writes and
   their interfaces? How are Surface-defined operations executed safely without
   blocking shared state processing?
3. **State scope:** does state belong to the registered Surface, its provisioned
   context within one Possibility, or an explicitly shared resource?
4. **Move versus Request:** is Request only the internal prepared write artifact,
   or a separate caller-facing state protocol? Uniform invocation and internal
   state primitives must not be conflated.
5. **Atomicity:** which adapter updates must commit together with runtime records
   or emitted Moves, and how is that guarantee represented?
6. **Trap/fault handling:** what does TrapController receive, what may it do,
   how does its outcome correlate to the original invocation, and what happens
   if the controller itself fails or is denied?
7. **Trust boundary:** who provisions authority, who enforces an installed grant,
   and who realizes the execution binding? The word "kernel" remains unsettled.
8. **Process placement and lifetime:** where do registry, StateOwner, provider
   contexts, and execution machinery live? Long-lived storage is not itself
   crash durability.

## Related

- [Purpose of the Possibility investigation](README.md)
- [Three-role code-on-paper proposal](proposal/README.md)
- [Pre-existing Surface ingress/egress profiles](../../2026-09-19/surface.md)
- [Interpret/commit discipline](../../2026-09-19/improvements.md)
- [Separate state and execution pools](../../2026-09-19/persistent-parent-and-separate-pools.md)
