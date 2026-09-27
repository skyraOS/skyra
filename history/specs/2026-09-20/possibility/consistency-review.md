# Possibility and execution model — consistency review

Date: 2026-09-20

Status: Open findings. Documentation review, not an implementation or security
proof. No design decisions are changed by this review.

Subsequent discussion: [Engrams, Surfaces, and the pipe protocol](engrams-surfaces-and-pipe-protocol.md)
refines the bounded-unit profiles, composition name, and external executable
boundary. Findings below describe the earlier snapshot and have not been silently
marked resolved by that refinement.

## Scope and interpretation

Reviewed the September 20 Possibility notes and proposal declarations against
the September 19 architecture, process, stdio, identity, and provider documents.
The dated provider/OS reference inventories are treated as source context, not
as freshly verified claims about external software. DNA-17 is a historical
implementation reference, not the acceptance target for the new proposal.

The latest discussion establishes these constraints: the loop owns authority;
all work leaving it for execution is authorized; preparation is a trap and a
fault is an error; state operations are Moves; StateOwner is a Surface;
execution context is serialized Moves and handlers, ideally one handler list;
the executor does not receive work with unresolved prerequisites.

Findings distinguish **conflicting/stale descriptions** from **missing contracts**.
An explicitly open decision is not a bug. It becomes important here when another
document or declaration assumes an answer or guarantee that depends on it.
Build-excluded proposal files with unresolved types are intentionally illustrative;
their failure to constitute a compilable package is not a finding.

Several conflicts come from carrying earlier assistant-authored proposals forward
and adding caveats instead of updating their main diagrams and declarations.

## C1 — High: the phase rule and executable state Moves describe different loops

**Type:** Conflicting scope of an invariant.

[improvements.md](../../2026-09-19/improvements.md) requires all canonical reads
in interpret, all writes in commit, and one lock held across both before execution.
[takeaways.md](takeaways.md) repeats that model. In contrast,
[StateOwner's Surface contract](proposal/authority-owner/state-owner/README.md)
and [loop serialization](loop-owned-authority-and-move-serialization.md) put
state operations in the work emitted for execution.

Both can exist only if the phase rule has a defined scope: for example, a state
Surface may have an internal atomic turn even though the outer loop emits its
invocation. The documents currently continue to describe the rule globally.

**Consequence:** an implementer could either prohibit the newly required state
execution or hold the outer state lock while awaiting StateOwner, preventing the
very preparation operation from running.

**Needed:** name the outer loop and the state service's internal turn separately;
state exactly which reads/writes and locks belong to each. Do not silently restore
the old rule as a prohibition on state-operation Moves.

## C2 — High: loop authority is not reconciled with AuthorizedMove's Horizon

**Type:** Missing contract across a changed authority model.

[The execution-context note](authority-namespace-and-execution-context.md) says
state-operation Moves use the original Possibility's authority.
[The later loop note](loop-owned-authority-and-move-serialization.md) allows the
loop to authorize inspection using its own authority before admitting that
Possibility's requested work.
[AuthorityChecker](proposal/authority-owner/authority-checker/proposal.go) still
describes its output as the original Move plus that Move's Possibility Horizon.

**Example:** a Move names an unknown Possibility. The loop may legitimately read
the namespace to reject it, but that read cannot derive its authority from the
unknown Possibility. Conversely, attaching the loop's broad Horizon to the
original target invocation would authorize too much.

**Needed:** specify the authority context and original-work correlation of each
generated inspection/preparation Move. This need not enlarge AuthorizedMove; it
must establish what its Horizon means for each kind of work. Also define how
trusted source identity reaches Check: its current Move-only signature leaves
that dependency implicit, as its own README acknowledges.

## C3 — High: serialized operations do not yet preserve the old atomicity guarantee

**Type:** Unresolved guarantee assumed by older descriptions.

[The separate-pools sketch](../../2026-09-19/persistent-parent-and-separate-pools.md)
preserves reads through commit in one turn.
[The new loop](loop-owned-authority-and-move-serialization.md) emits ordered
operations and explicitly leaves their transaction boundaries open.

**Example:** two invocations each run `read counter -> write counter + 1` as
separate Moves. Each list can preserve its own order while both read the same
value and one update is lost. A single StateOwner worker serializing individual
operations does not automatically serialize the whole read/modify/write group.

**Needed:** specify which groups are atomic and how conditional writes validate
their inputs. Also decide what happens between authority evaluation, the authority
stamp write, and target execution if Horizon or a binding changes. Ordered work
must not be described as having the earlier transaction guarantee until defined.

## C4 — High: StateOwner is both an executable Surface and the old whole-loop host

**Type:** Stale composition retained beside a newer contract.

[StateOwner README](proposal/authority-owner/state-owner/README.md) now correctly
says it receives/emits Moves like any Surface and that the coordinator retains
dependency bookkeeping. Later in the same file its state pool still resolves
Horizon, manages dependencies, and dispatches external work.
[Its struct](proposal/authority-owner/state-owner/proposal.go) still contains
the whole state-turn pool, Possibilities, execution records, and registry.
[AuthorityChecker](proposal/authority-owner/authority-checker/README.md) meanwhile
says namespace ownership must not be inferred from that map.

These fields are labeled provisional; this is not evidence that the new design
requires this ownership. They are nevertheless the only concrete storage sketch.

**Consequence:** following the struct literally makes the state service its own
coordinator and can reintroduce same-queue or same-lock waits when preparation
invokes StateOwner. It also suggests the namespace-owner question is already answered.

**Needed:** separate the service's storage/serialization role from loop coordination
in the current sketch, or explicitly freeze the old struct as a historical baseline
and supply a minimal current composition. Do not decide process placement by this edit.

## C5 — High: registry lookup and provisioned Horizon bindings are competing models

**Type:** Stale binding model plus an undeclared replacement.

[SurfaceRegistry](proposal/authority-owner/surface-registry/README.md) resolves
the permitted SurfaceID against its table during each interpretation turn, with
registry access serialized by StateOwner. The newer
[takeaways](takeaways.md) and [execution-context note](authority-namespace-and-execution-context.md)
instead describe provisioning registry entries into Horizon and avoiding handler
rediscovery during execution.

[Horizon's declaration](proposal/authority-owner/horizon/proposal.go) still contains
only generation, SurfaceID grants, and TrapControllerID. It has no represented
provisioned executable binding. The registry retains `*Surface` including Driver.

**Consequence:** a handler list cannot yet be derived solely from the described
Horizon. An implementer must choose an unstated registry lookup or capture a live
handler whose replacement/sharing semantics are not defined.

**Needed:** decide when binding occurs, what provisioning retains, and what
replacement/retirement does to existing Horizons and already-emitted work.
Separate service interfaces do not, by themselves, prohibit StateOwner-backed
registry storage; the actual inconsistency is the lookup and binding lifetime.

## C6 — High: a dependency-free executor and a result-dependent handler list need a handoff contract

**Type:** Missing contract between two explicit invariants.

[The execution context](authority-namespace-and-execution-context.md) contains
serialized Moves/handlers and no stack or dependency scheduler.
[The loop note](loop-owned-authority-and-move-serialization.md) allows entries
that consume earlier results and conditional authority-stamp writes, while
retaining the rule that unresolved prerequisites never reach execution.

**Example:** `read Horizon -> evaluate authority -> write success stamp -> run
target`. Before the read/check completes, the success stamp and target are not
yet unconditionally authorized ready work. A returned result is also supposed
to re-enter authority; the executor cannot simply advance past that boundary
without a defined rule.

**Needed:** specify whether output is successive ready lists or one sequence with
explicit permitted result binding/branch semantics. Define who stops advancement
on failure and admits the next entry. Preserve the user's minimal context shape;
this finding is not a request to add the whole stack or resource bags to it.

## C7 — Medium: services have no declared-operation contract for the new coordinator to consume

**Type:** Prose/declaration gap.

[Loop serialization](loop-owned-authority-and-move-serialization.md) depends on
each service declaring ordered required reads/writes. The
[Surface struct](proposal/agent-surface-worker/surface/proposal.go) has
`Dependencies []*Surface` and unresolved ingress/egress Profiles, but no declared
mapping from an invocation to those state operations, their result relationships,
or order. A list of Surface addresses does not supply that information.

[AuthorityChecker.Check](proposal/authority-owner/authority-checker/proposal.go)
is synchronous and returns a pair or error; there is no declaration/result
contract connecting its required reads and authority-stamp write to the coordinator.
The coordinator itself remains prose-only.

**Needed:** identify whether the existing Profiles become these declarations or
are separate concepts. Define one small example end to end before adding fields
or introducing a second interpretation engine implicitly.

## C8 — Medium: private adapter state and runtime control state still share an unspecified access model

**Type:** Missing ownership/scoping contract concealed by a common service.

[Takeaways](takeaways.md) explicitly limit Surface-declared operations to adapter
state. [Loop serialization](loop-owned-authority-and-move-serialization.md) adds
trap records and authority stamps as service-declared operations. This extension
is reasonable for appropriately privileged services, but
[StateOwner's declaration](proposal/authority-owner/state-owner/proposal.go)
contains runtime maps and no adapter-state namespace or restricted operation profile.

**Example:** a target Surface and AuthorityChecker both invoke StateOwner. One may
update its own adapter continuation; the other may record authorization. A grant
to the StateOwner SurfaceID alone does not describe that difference.

**Needed:** specify the provisioned state scope and operations for each service
context, including whether adapter state is per Possibility or intentionally
shared. The common Move protocol must not turn all StateOwner callers into
general writers of runtime records.

## C9 — Medium: the proposed Surface/Driver contract mixes older and newer output mechanisms

**Type:** Explicitly inherited interface with unresolved compatibility.

[Surface proposal](proposal/agent-surface-worker/surface/proposal.go) uses the
September 19 struct, the September 18 `DriverEmission` callback/one-Move Egress,
and an undefined Profile. [The DNA-17 snapshot](proposal/authority-owner/reference/dna17.go)
uses `DriverInput`, `Emission`, error-returning Egress, and potentially multiple
Moves. These differences are labeled; the snapshot is not an implementation target.

However, the new requirement is that StateOwner implement this common shape and
that preparation may emit multiple ordered operations. No current contract
establishes normalization errors, fan-out/order, or how profile output becomes
the new handler list.

**Needed:** select or refine the proposal's driver/profile contract explicitly.
Do not infer the interface from whichever historical file happens to compile.

## C10 — Medium: disposable executions are not reconciled with dynamic provider work and frame handles

**Type:** Known lifecycle gap spanning execution examples.

[Disposable workers](../../2026-09-19/disposable-workers-state-turns-and-execution-policy.md)
have fresh handles while logical invocations persist.
[InvocationFrame](proposal/authority-owner/possibility/proposal.go) still carries
the explicitly provisional one-session `FD` and single-child `WaitingOn` fields.
[Codex integration](../../2026-09-19/codex/codex-integration.md) describes a turn
that can introduce tool work after it starts; [stdio](../../2026-09-19/stdio/stdio-surface.md)
describes a persistent receive continuation.

This is not necessarily a contradiction: an initially ready operation can later
produce a new child request. But the latest rejection of workers discovering
missing prerequisites does not yet say whether a whole Codex turn or each
protocol step is the executable unit.

**Needed:** distinguish pre-invocation requirements from new work produced during
execution, and define who owns the live provider session, later notifications,
and correlation after a worker exits. The old one-child/session fields cannot be
treated as the finished cactus/disposable-attempt model.

## C11 — Medium: the older execute-time liveness fence contradicts the state-read boundary

**Type:** Direct conflict within September 19 references.

[Handle-liveness fence](../../2026-09-19/claude-syscall/handle-liveness-fence.md)
places liveness/generation in canonical state and checks that table at execution
time. [improvements.md](../../2026-09-19/improvements.md) explicitly disallows
canonical reads after interpret, including hidden liveness interpretation in commit.

**Needed:** classify this as an admitted state operation, a new state turn, or
private resource coordination with a defined stable binding. A detached liveness
snapshot is not equivalent to current resource liveness. The older claim of a
structural stale-resource safety proof should remain unproven until that binding
and use boundary are defined.

## C12 — Medium: no single current contract supersedes the accumulated checkpoints

**Type:** Navigation and status drift that makes other findings harder to resolve.

[The entry README](README.md) starts with the three-role proposal, whose
[overview](proposal/README.md) still puts Horizon first before any execution and
retains old registry/StateOwner placement. Newer notes authorize inspection work
before that check completes and explicitly reopen those placements.

The proposal overview still asks what distinguishes trap from fault after the
newer note defines trap as preparation and fault as error. It also says no new
gate signatures have been introduced, although AuthorityChecker now exists.
Takeaways still asks whether Request is a separate public protocol, while the
new StateOwner section explicitly rejects that public path.

**Needed:** distinguish current contract from historical checkpoint and unresolved
decision. Keep history, but provide a small precedence/index table so a reader
can tell which statements to implement. This review itself does not make those
design choices.

## Secondary carryover issues

- **Spawn contracts:** [the cactus disposition](../../2026-09-19/invocation-cactus-and-possibility-isolation.md)
  defers lifecycle Surfaces and makes spawn a mechanism, while
  [the syscall table](../../2026-09-19/syscall/skyra-syscall-surface.md) and
  [start-process proposal](../../2026-09-19/syscall/start-process/proposal.md)
  still present create/ensure/attach plus fork-and-configure as active proposals.
  Their status needs a supersession pointer rather than silent coexistence.
- **Identity vocabulary:** the cactus text says ReturnTo points to a caller fd;
  [the identity note](../../2026-09-19/surfaceID-moveID-fd.md) explicitly leaves
  parentFD unresolved and distinguishes descriptors from Move identity. The new
  frame uses a UUID return destination. Also, scoped invocation maps address
  cross-Possibility collisions, not same-Possibility duplicate IDs; the earlier
  claim that issue #1 becomes structurally impossible is broader than the described fix.
- **Root representation:** cactus prose treats `rootOnce` as a frame/floor;
  DNA-17 uses it as result-deduplication storage in CU. The analogy does not yet
  supply a concrete root continuation or sink for the new frame service.
- **Surface count:** [Codex integration](../../2026-09-19/codex/codex-integration.md)
  says four top-level Surfaces but lists five, including `codex.fork`.
- **UUID sentinel:** [InvocationFrame's comment](proposal/authority-owner/possibility/proposal.go)
  says `nil` denotes the root despite a value-typed `uuid.UUID`. If intended,
  name the zero UUID / `uuid.Nil`, or specify a different optional representation.

## Suggested resolution order

1. C1 and C4: separate outer coordination from the executable state service.
2. C2 and C8: bind loop inspection and Surface state access to their proper authority.
3. C6 and C7: walk one declared operation sequence through the minimal handler list.
4. C3 and C5: define atomic groups, binding lifetime, and authority changes.
5. C9–C12 and carryovers: align interfaces, lifecycle examples, and document status.

Validation performed: local Markdown file targets across both dated spec sets
resolve. No compilation, runtime tests, live provider calls, or OS enforcement
checks were performed for this documentation review.
