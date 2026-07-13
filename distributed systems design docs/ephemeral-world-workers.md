# Ephemeral World Workers

**Status:** HIGH-LEVEL PROPOSAL · 2026-07-13

## Purpose

Apply DNA v2's durable-substrate / ephemeral-worker pattern to world execution.

The world remains durable authoritative state. `perceive`, `engage`, `move`,
and other affordances become temporary workers created for one situated
invocation. A worker receives a bounded local surface, performs one operation,
commits explicit effects where permitted, emits an inspectable outcome, and
then disappears.

```text
durable world state + situated impulse + local capability
                         |
                         v
               ephemeral world worker
  resolve -> snapshot -> validate -> execute -> commit -> return
                         |
                         v
        durable world state + reafferent observation
```

This reuses the computational pattern of DNA v2 without making the world
cognitive. World workers determine what happened. Cognitive workers determine
what that outcome means.

## Core Separation

### Durable world substrate

The world owns:

- place and space identity;
- topology and adjacency;
- occupants and presence;
- authoritative local state;
- registered affordances;
- permissions and capability boundaries;
- resource versions; and
- committed invocation outcomes.

### Ephemeral world worker

A worker owns one invocation's temporary computation:

- invocation identity;
- actor, source, and directed target;
- primitive and payload;
- bounded local traversal surface;
- authority presented for this invocation;
- snapshot or precondition version;
- deadline and cancellation state;
- proposed effects; and
- terminal outcome.

The worker is not durable memory. Anything that must survive is committed to
the world or returned as an observation for cognition to record.

### Cognition

Cognition owns interpretation, causal belief, goals, plans, prediction,
grading, action selection, and long-term experiential memory.

The world may report `file X was written`. Cognition may infer that writing the
file advanced a goal. The world must not make that inference on cognition's
behalf.

## Correspondence With DNA v2

| DNA v2 | World execution |
|---|---|
| DNA / competence registry | Place capability registry |
| Durable causal nodes | Durable spaces, resources, occupants, and state |
| Gene | Affordance implementation |
| Ephemeral root | One affordance invocation worker |
| Active causal set | Bounded local world projection |
| Membrane / orchestrator | Place dispatcher |
| Board row | Invocation record and lifecycle state |
| Worker condition | Outcome status |
| Durable causal commit | Atomic world-effect commit |

This is behavioral symmetry, not ontology collapse. Cognitive and world
workers may share lifecycle machinery while operating on different substrates
under different laws.

## Invocation And Outcome

The exact Go API remains open, but every invocation needs this semantic shape:

```text
Invocation
  ID
  Primitive
  ActorID
  Source
  TargetID
  Payload
  Authority
  SnapshotVersion / Preconditions
  Deadline
  IdempotencyKey
```

The impulse remains the transient event carried across the membrane. Deadline,
authority, retry state, and idempotency belong to the invocation work item or
Spacetime—not to durable world content.

Every worker terminates with an explicit outcome:

```text
Outcome
  InvocationID
  Status
  Result
  Effects
  ObservedVersion
  CommittedVersion
  Provenance
```

Minimum status vocabulary:

- `succeeded` — effects committed and the result is authoritative;
- `blocked` — topology or state prevented the operation;
- `denied` — authority did not permit it;
- `stale` — a snapshot or precondition no longer matched;
- `unavailable` — the target or backing device could not be reached;
- `cancelled` — execution ended before commit;
- `partial` — only explicitly listed effects committed; and
- `indeterminate` — the caller cannot prove whether a remote effect committed.

`indeterminate` is not failure. Cognition may need to perceive the target again
before retrying.

## Worker Lifecycle

### 1. Resolve

Resolve the directed target and its registered affordance. Resolution is local
to the mounted world or projected partition.

### 2. Snapshot

Capture the smallest authoritative surface required by the operation. Reads do
not hydrate or lock unrelated world state.

### 3. Validate

Validate target existence, locality, reachability, capability availability,
authority, payload shape, resource preconditions, deadline, and cancellation.

### 4. Execute

Perform bounded computation or call the backing adapter. Long-running remote
work occurs without holding a global world lock.

### 5. Commit

Apply declared effects to the authoritative substrate. Local effects are
atomic at the smallest coherent resource boundary. A stale precondition
rejects the commit instead of overwriting newer state.

### 6. Return

Emit one result impulse carrying the outcome across the membrane. The result is
a reafferent observation. DNA v2 decides how to decompose, retain, grade, or
learn from it.

## Primitive Workers

### Perceive

A perceive worker reads a bounded local projection without mutation. Inputs
may include source space, radius or traversal budget, query, modality,
requested freshness, and authority.

Its result distinguishes freshly observed dynamic state from cached structural
state. Staleness and observation time are provenance.

### Engage

An engage worker invokes one directed affordance on one target. It resolves the
target locally, validates authority and preconditions, calls the capability,
commits declared effects, and returns an outcome.

### Move

A move worker validates adjacency, occupancy, boundaries, and destination
rules before atomically changing presence. Crossing a device membrane may make
the committed effect a mount change rather than a coordinate change.

### Wait

A wait worker advances or subscribes to local time without pretending no world
change occurred. Its result may include events observed during the wait.

New affordances use the same lifecycle. They do not add branches to a central
`World.Observe` switch.

## Place Capability Registry

Each place exposes only affordances genuinely available there:

```text
place / space
  -> primitive key
      -> worker factory or handler
```

Examples:

```text
grid-point:perceive
grid-point:move
entity:engage
filesystem:list
filesystem:read
filesystem:write
browser:navigate
clipboard:write
media:play
```

The registry describes situated capabilities, not arbitrary remote commands.
Different places may implement the same primitive while preserving one
invocation/outcome contract.

## Concurrency And Commit Rules

- Read-only workers may run concurrently over versioned snapshots.
- Mutating workers serialize only at resources they may change.
- No network or model call holds a global world lock.
- Each mutation declares proposed effects before commit where practical.
- Commits verify the versions or preconditions read during validation.
- Repeated delivery with the same idempotency key returns the prior committed
  outcome rather than duplicating an effect.
- Subordinate work remains bounded under one root invocation identity and one
  terminal outcome.

## Distributed Device Worlds

For a simulated world, execution is in process. For a device-backed world, the
same contract may marshal execution through an adapter or remote service:

```text
world worker
  -> device-space adapter
      -> HTTP / gRPC / WebSocket / local bridge
          -> device capability
```

The agent cannot distinguish a stub space from an API-backed space by cognitive
contract. It may perceive legitimate differences in availability, latency,
freshness, authority, and failure outcomes.

Remote execution additionally requires stable device/world/space identity,
authenticated grants, invocation deduplication, cancellation, reconnect and
lease semantics, versioned hydration, and indeterminate-outcome recovery
through fresh perception.

## Mount Interaction

The membrane remains the agent-to-world mount. Workers execute below the
currently mounted boundary.

- **Relocation:** a move or engage worker commits re-seating onto another whole
  world.
- **Projection:** a mount worker attaches a remote partition as a child of the
  current workspace.

Mount changes cannot silently redirect already-running invocations. Each
invocation is bound to the world and space identities resolved at its start.

## Daily Resource Budget

Token use is already explicit enough to support a monetary operating envelope.
The user may set a maximum price per day, and the runtime schedules cognitive
and world work within that limit.

This introduces a fourth quantity that must remain distinct:

```text
belief strength       = learned support or tension
traversal amplitude   = reach allocated within one graph walk
query-token budget    = permitted specificity of one worker emission
resource-spend budget = permitted external cost over a time window
```

The daily spend budget must not directly weaken beliefs or truncate a causal
claim. It controls whether, when, and at what service tier a worker may run.
If the remaining budget cannot support a complete valid operation, the runtime
defers it, chooses an explicitly cheaper valid strategy, or returns
`budget-exhausted`.

### Budget account

The runtime owns a durable account for each policy scope, initially one user
per local day:

```text
SpendAccount
  Currency
  DailyLimit
  WindowStart / WindowEnd
  CommittedSpend
  ReservedSpend
  RemainingSpend
  Policy
```

Provider prices and model token accounting are inputs to this controller, not
properties of causal memory. The ledger records the pricing version used for
each estimate and reconciliation.

### Reserve, execute, reconcile

Before starting a billable worker:

1. estimate a conservative maximum cost from its provider, model, input,
   query-token limit, retries, and tool plan;
2. atomically reserve that amount against the daily account;
3. reject, defer, or downgrade before execution if reservation would exceed the
   limit;
4. execute the worker;
5. reconcile the reservation against authoritative usage returned by the
   provider; and
6. release unused reservation or record any bounded estimation variance.

Reservations are required because Efficient and Final workers, perceptions,
and device operations may run concurrently. Checking only the current spend
would allow parallel work to oversubscribe the daily limit.

### Policy

The scheduler may use remaining spend to choose among valid strategies:

- reserve capacity for expected future turns;
- select a less expensive model or local inference provider;
- reduce optional fan-out or defer background inspection;
- use already verified graph structure instead of issuing a fresh query; or
- stop cleanly when only an incomplete or malformed result could fit.

Critical safety and consistency work—such as determining whether a mutation
committed—needs a separately declared reserve or must refuse the original
mutation when recovery cannot be afforded. The runtime must never spend past
the cap merely because an operation was already started unless the user's
policy explicitly permits a small emergency reserve.

Every invocation trace records estimated, reserved, and reconciled cost. This
makes cost attributable to the worker and the durable state change it produced,
allowing evaluation of useful memory or world progress per dollar.

## Invariants

1. **The world is authoritative.** A worker cannot commit state it did not
   validate against the owning world.
2. **Workers are ephemeral.** Durable knowledge and state never live only in a
   worker instance.
3. **Outcomes are explicit.** Success, denial, staleness, partial completion,
   and uncertainty are not collapsed into plain text.
4. **Effects are attributable.** Every committed effect points to its
   invocation, actor, target, and authority.
5. **Locality is preserved.** A worker sees only the traversal and capability
   surface exposed by its place.
6. **World and cognition do not collapse.** The world records events; cognition
   interprets and learns from them.
7. **Transport is hidden, failure is not.** Protocol details stay below the
   place boundary, while their observable consequences remain honest.
8. **Retries are safe or explicit.** Mutations are idempotent, rejected, or
   reported indeterminate—never blindly repeated.
9. **Spend is reserved before work.** Parallel workers cannot collectively
   exceed the configured operating envelope through stale budget reads.
10. **Cost controls scheduling, not truth.** Monetary pressure may defer or
    choose among valid computations; it cannot silently corrupt, truncate, or
    regrade causal structure.

## Minimal Implementation Slice

Prove the pattern locally before introducing transport:

1. Add a world affordance registry and worker factory.
2. Turn `perceive` into a read-only ephemeral worker over the existing grid.
3. Turn `engage` into a directed worker with an explicit outcome.
4. Preserve current maze behavior and result impulses.
5. Add invocation IDs, statuses, and trace output.
6. Prove concurrent perceptions do not block one another.
7. Prove conflicting mutations use resource-local versioned commit.
8. Place a stub MacBook membrane in the base world.
9. Cross into a stub MacBook world through the same worker lifecycle.
10. Replace one stub capability with an adapter-backed capability without
    changing cognition, the kernel, or the invocation contract.

## Deferred Questions

- Whether world and cognitive workers share one generic worker runtime or only
  parallel interfaces.
- The exact relationship among Invocation, Impulse, Spacetime, and Board.
- Whether every world retains outcomes or reconstructs them from an append-only
  invocation log.
- The granularity of resource versions and transactional commits.
- How subscriptions and streaming perception fit one terminal outcome.
- How projected spaces advertise lease expiry and disconnection.
- Whether mount changes are `move`, `engage`, or a universal `mount` affordance
  beneath those user-facing primitives.

These questions do not block the first local worker conversion.
