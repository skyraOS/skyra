# Engram loading and fault returns

> **Resolve ownership decided:** filesystem driver exposes `Resolve(id MoveID) (*Fd, error)`.
> Scheduler checks Memory first; on a miss, kernel calls the driver and installs the returned Move/Register
> entry. This supersedes kernel-owned read/decode/verify in the older text below. Absence/error conventions
> and binding the returned fd to the incoming Move's live context remain open. See the
> [current proposal](../../src/skyra/skyra-ctu-engine/surface-proposal.md#decided-one-identity-resident-lookup-then-durable-resolution).

> **Latest decision:** [the resolution proposal](../../src/skyra/skyra-ctu-engine/surface-proposal.md#decided-one-identity-resident-lookup-then-durable-resolution)
> unifies MoveID and EngramID. Call ingress uses that ID for resident lookup, then kernel filesystem
> resolution on absence. If no remembered executable exists, declaration supplies behavior while
> kernel provisions the register. Register creation alone does not make work executable. This supersedes
> the separate invocation/backing identities and dispatch-time lookup in the older trace below.
> Filesystem errors and invalid content remain distinct from absent remembered backing.

**Date:** 2026-09-06  
**Status: DECISION / interfaces to design.** Records the identity, durable-unit, and loading decisions
from the fault-path walkthrough. No runtime implementation or final Go struct is introduced here.

This revises the kernel/CTU ownership in [ingress-and-fire.md](ingress-and-fire.md) and the
[engine proposal](../../src/skyra/skyra-ctu-engine/engine.go): the scheduler belongs to the kernel,
and the kernel may install executable backing through shared Memory's defined installation interface.
Kernel coordinates and provisions, Memory maintains resident state, and CTU executes. Scheduler remains
the sole writer of scheduling lifecycle state.

## Identity and the durable unit

| Identity | Meaning |
|---|---|
| EngramID | The durable executable unit, with the same identity whether resident or persisted; replaces BodyID terminology for that unit |
| PossibilityID | The execution chain / graph as a whole |
| MoveID | Canonical action + ordered observation params; directly addresses its register (scope/sharing open) |

No separate Attempt identity is adopted. Retry correlation remains a question to test with concrete
traces; the original MoveID survives materialization and reissue.

An engram is the immutable, durable form of a remembered executable: **one body plus its immediate,
version-specific child references**, now with explicit parentContract and childContracts fields for
its durable return obligations. See [the contract refinement](body-registers-and-return.md#parent-and-child-contracts-on-the-engram).
Each child engram has its own body and immediate children. An engram
does not physically contain the entire descendant graph. The exact body representation, child-key type,
encoding, and Engram struct remain to be designed. The current v12 Body already contains Children;
the eventual struct must avoid representing those edges twice.

Child references pin versions. Changing a descendant creates new ancestor versions wherever the child
references change. Unchanged branches remain shareable and old versions remain addressable. With
content-addressed IDs, the durable executable contents determine the engram's identity. The new contract
fields must be reconciled with the canonical encoding/versioning design; the older body-and-children-only
hash description is incomplete. Exact contract encoding and digest preimage remain open.

Structure and loading policy are separate. Loading one engram makes its body and child references
resident. Eager descendant loading versus demand loading is a separate decision, not part of the
engram's definition.

## Ownership: kernel policy, memory management, execution

- **Kernel:** coordinates and provisions, contains the scheduler, and owns `Load(EngramID)`. It decides
  what to load, retain, or evict; calls the filesystem driver directly; decodes/verifies the engram;
  and requests installation through Memory's defined interface. Eviction policy remains to be designed.
- **Scheduler:** owns MoveID lifecycle, dispatch, parking, and reissue. It routes fault recovery to the
  kernel's loader; it does not execute the faulted operation.
- **Memory:** owns resident engrams, execution frames, and their internal storage structures. It provides
  lookup, installation, register access/publication, and reclamation operations, maintaining their
  consistency and lifetime rules.
- **CTU:** executes moves by resolving through Memory, reading operands, performing operations, and
  writing results through Memory. It reports execution outcomes and faults; it does not own Memory's
  internal layout or independently manipulate its maps.
- **Filesystem driver:** supplies file I/O for kernel loading and persistence as well as ordinary
  user-file operations. No separate EngramStore interface is introduced between kernel and driver.

Kernel and CTU access the **same Memory instance**. Memory is the runtime's shared memory subsystem;
it is not a separate kernel copy and CTU copy. Both use Memory's operations instead of independently
mutating its maps. Concrete pointer/interface wiring and method signatures remain open.

The responsibility split is who decides an operation is needed versus who maintains storage correctly.
The kernel chooses an eviction candidate; Memory prevents reclamation while live execution still needs
the backing. The CTU determines an execution result; Memory maintains the register slot's consistency.
Exact pinning/refcount/locking rules remain open. Kernel provisioning does not take over ordinary
opcode execution or result production.

Illustrative operations, not committed Go signatures:

```text
Kernel.Load(id) → filesystem read + decode + verify → Memory.Install(id, engram)
CTU executes   → Memory.Resolve(moveID) → read operands → execute → Memory.Land(fd, value)
```

The conceptual installation operation is `Memory.Install(EngramID, Engram)`, not a committed Go signature.
The kernel need not return a frame for the CTU to install later: it can provision backing through this
interface itself. Installation must finish before the scheduler releases a waiting move. Multiple moves
waiting on one engram can share the load; the deduplication/waiter mechanism remains open.

This replaces the earlier three-peer topology and the absolute ban on kernel mutation of CTU memory.
Scheduling now occurs inside the kernel on the ordinary path. Loading and authority provisioning remain
exceptional work; the old phrase "hot path is kernel-free" no longer describes this topology.

## Durable storage location and filesystem-driver reuse

**Decision:** use a configurable directory in the user's application-data location, with kernel loading
using the same filesystem driver that provides ordinary user-file operations. No dedicated disk
partition or separate SSD driver is needed for this implementation; the host operating system handles
the underlying device.

On the current Mac, the proposed default storage root is:

```text
~/Library/Application Support/Skyra/
```

Engrams and internal entry-point references are application-managed durable data. Files created for the
user and explicitly exported engrams go to visible locations the user chooses. Hidden-directory status
is presentation, not authorization. Engrams are not disposable cache; CTU-resident copies are the cache.

The initial storage layout is one immutable file per engram, addressed by its content digest, with
entry-point references stored separately. Illustrative layout; exact filenames and digest sharding
remain open:

```text
Skyra/
  engrams/       durable executable objects by EngramID
  references/    remembered entry points referring to EngramIDs
```

Kernel.Load answers "load this EngramID." The reference registry answers which remembered engram
an entry point names. Updating a reference does not mutate an existing engram. Registry schema and
publication rules remain to be designed; no database or garbage-collection mechanism is selected here.

### Responsibilities and loading path

**Decision: exact-ID loading directly on the kernel.** `Kernel.Load(EngramID)` derives a file path
from the configured storage root and ID, then calls the filesystem driver. Loading a known ID requires
no directory scan, database query, or semantic search. For example, an illustrative ID `abc123` maps to
`<storage-root>/engrams/abc123`; the final filename/sharding convention remains open.

Entry-point lookup remains separate: a name resolves through an index to an EngramID, then uses the same
exact-ID load. Intent-based discovery belongs to planning and is not part of the page-fault loader.

| Component | Responsibility |
|---|---|
| Filesystem driver | File operations such as read, write, rename, and sync; knows files and bytes, not engram semantics |
| Kernel | Own Load(EngramID), map IDs to paths, decode/verify engrams, arrange persistence using the driver, and install loaded backing through Memory's interface |
| Scheduler within kernel | Account for the parked move and release it only after installation and readiness are established |

```text
Kernel.Load(EngramID)
  → derive the exact file path
  → Filesystem driver reads bytes from application storage
  → kernel decodes the engram and verifies its ID
  → Memory.Install(EngramID, Engram)
  → Scheduler releases eligible waiting moves
```

This supersedes the earlier proposed `EngramStore.Load` layer. The kernel owns the Load method and calls
the filesystem driver directly; an additional EngramStore interface is unnecessary at this stage.
Exact Go signatures and error representations remain open. A missing file, I/O failure, or invalid
contents is a load failure, not successful installation. The move stays out of execution until the
kernel accounts for that outcome; distinguish these failure causes when designing the result type.

Durable publication must provide deterministic encoding, content identity, complete-object visibility,
and a defined crash-survival guarantee on successful persistence. The kernel arranges these guarantees
through driver operations; the exact write/sync/publication protocol remains open.

### Authority and boot requirements

**Sharing the driver does not share authority.** Kernel access to Skyra's internal storage does not
automatically authorize workflow moves to modify it. User-facing filesystem operations still run under
their granted authority. The exact path-scoping/access interface remains to be designed.

**Recovery must not depend on the missing engram.** The filesystem driver and enough kernel loading code
to retrieve, decode, verify, and install engrams must be available at boot, independently of pageable
engram execution. Reusing the driver does not require routing its recovery I/O through an executable
that may itself be absent. The direct invocation interface remains open.

## Involuntary faults use the execution-return path

**The CTU reports what happened; the kernel decides what happens next.** The kernel coordinates and
provisions in response to these returns, applying lifecycle changes through its scheduler. Memory
maintains the consistency of the corresponding installation, value-publication, and retention operations.

The CTU reports an involuntary fault as a **Faulted event to the scheduler over the same return transport
used for execution outcomes**. This chooses one logical return path; it does not claim the proposal
already implements a queue, channel, or wire codec.

| Returned event | Meaning | Kernel / scheduler consequence |
|---|---|---|
| Issued | Dispatch happened; completion remains outstanding | Keep the move outstanding; do not regress a later lifecycle state |
| Landed | A result was produced | Record completion after publication is established; schedule dependents whose inputs are available |
| Faulted | Execution could not proceed; the move has not completed | Park the move, arrange recovery, and reschedule only when ready |

New invocations supply the initial work to schedule. After dispatch, CTU returns drive the execution
and recovery loop. They are not the only wake source: asynchronous loading and human confirmation can
also notify the kernel when recovery is ready. Notification interfaces remain open. None permits
dispatch before required backing/values are available or bypasses scheduler claims.

A fault returns control, not a semantic result to the parent. It neither fills the consumer's result
register nor marks the move Completed. Scheduler ingress applies the lifecycle consequence.

The CTU reports the fault once. It does **not** also independently invoke a kernel handler. Scheduler
fault handling routes the recovery to the kernel loader. For this involuntary Materialize path there is
no additional CTU-fired Trap required to initiate the same recovery. The remaining role of explicit Trap
opcodes is open.

## Materialize walkthrough

```text
Kernel scheduler → CTU: execute MoveID M
CTU resolves M → required EngramID E is not resident
CTU → Scheduler: Faulted{MoveID: M, missing EngramID: E}

Scheduler parks M and routes Materialize recovery
Kernel.Load(E) reads through the filesystem driver, decodes, and verifies E
Kernel installs E through shared Memory's interface
Scheduler makes M eligible after installation and readiness are established

Kernel scheduler → CTU: execute the same MoveID M
CTU resolves successfully, rechecks access, and executes
CTU → Scheduler: Landed{MoveID: M, result} once result publication is established
```

The fault payload above is conceptual. It must identify both the blocked MoveID and missing EngramID;
the concrete event shape and Memory.Resolve miss result are the next interface decisions.

Materialization installs missing executable backing. It does not create a replacement execution move.
The original descriptor, operands, return destination, and authority remain recoverable while parked.
Successful loading alone does not grant authority or consume inputs.

An unknown MoveID is distinct from a known live move whose engram is absent. Resolution must retain a
way to discover the required EngramID even when its executable backing is missing. The home of that
reference is still open.

## Remaining work

[One register per fd](register-frame.md) now supplies the agreed live Memory register shape and
schema-only Contract. Its ordered child observations revise the earlier scalar-slot framing; reconcile
the durable Body/Engram fields with it before implementation. The engine now uses canonical MoveID
keys and actual observation params; v12 retains its older Context/RegKey addressing.

Declaration scope is now recorded in [focus-and-move.md](focus-and-move.md): Focus exposes the live move's
available operations as `[]c.Capability` selected from Horizon, separately from compiled Dependencies.
No OperationDecl wrapper is needed. Its Move/backing-field sketch
is a proposal. [body-registers-and-return.md](body-registers-and-return.md) now records Body as the register
and return specification, with structural contracts discovered by execution and recorded at REMEMBER.
Concrete Body types and durable focus encoding still need design.

1. Design Engram and the concrete Kernel.Load signature, including canonical content identity,
   filesystem-driver operations, publication guarantees, and entry-point references. Exact-ID loading
   directly through the filesystem driver is decided; no separate EngramStore interface is required.
2. Define MoveID-to-executable/frame lookup and its missing-EngramID result.
3. Define the Faulted payload and kernel-to-Memory installation interface.
4. Specify pending recovery, load success/failure, shared loads, wakeup, and duplicate-event handling.
   Do not add Attempt IDs before the trace demonstrates a need.
5. Specify return delivery and result publication, keeping faults distinct from completion and semantic errors.
6. Reconcile the engine proposal and v12 types with these decisions before implementing the local loop.

The concrete scheduler placement, installation permission, and involuntary-fault return route are settled
here. Event encoding, method signatures, synchronization, storage implementation, and eager/lazy loading
remain open. [closed-loop-outstanding.md](closed-loop-outstanding.md) tracks the local-loop proof obligations.
