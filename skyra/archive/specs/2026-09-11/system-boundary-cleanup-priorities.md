# System boundary cleanup priorities

**Date:** 2026-09-11  
**Status: PROPOSED REFACTOR DIRECTION.** This note identifies the next boundary
work after the DNA16 vertical slice. It does not authorize a mechanical split or
rename before the transition laws and ownership contracts are written down.

## Objective

Reduce the number of policies accumulated in the DNA16 `Scheduler`, make the
continuation interpreter a precise component, and give drivers one clear
normalization boundary without leaking provider protocol or live host machinery
into the continuation runtime.

The intended trusted path is:

```text
untrusted Move
    |
    v
MMU / Horizon check             one authority gate for this crossing
    |
    v
Continuation interpreter       continuation-language meaning only
    |
    v
Transition commit              authoritative state transaction
    |
    v
Runtime effect runner          dispatch, host commands, notifications
```

An Invoke and a Return each cross this path once. "One MMU check" means one
mandatory authority gate before interpretation of each Move; it does not mean
that only the invoke side of the system is checked. Kernel mints and installs the
Horizon. The MMU checks it. The interpreter never mints authority or consults the
kernel to decide an individual Move.

## 1. Replace the current Scheduler god object with explicit roles

DNA16's `Scheduler` currently combines continuation storage, authorization,
language interpretation, child admission, join/resumption, placement state,
dispatch selection, failure propagation, and root notification. The first task
is to name the target roles before moving methods.

Target ownership:

| Role | Owns | Must not own |
| --- | --- | --- |
| Kernel | boot, possibility policy, Horizon minting, system scheduling policy, privileged construction | continuation-language interpretation |
| MMU / Horizon gate | one fast authority verdict for an Invoke or Return crossing | driver lookup, operand meaning, continuation mutation |
| Continuation interpreter | Invoke/Return grammar, register semantics, child declaration, joins, resumption, terminal/fault language rules | host processes, transport, endpoint resolution, authority minting |
| Continuation store / transaction | authoritative records, atomic compare-and-commit, snapshots | deciding what a Move means |
| Runtime coordinator | feeds checked Moves to the interpreter and executes committed effects | inventing continuation-language laws |
| Dispatcher | resolves a canonical endpoint binding and requests execution | capability interpretation, register mutation |

The component currently named `Scheduler` should not be renamed until this split
is real. Its eventual name should describe whichever residual role remains. The
name `ContinuationInterpreter` belongs only to the component that fully
interprets the continuation language.

Acceptance criteria:

- Continuation-language behavior can be tested without a registry, driver,
  process, transport, callback, or wall clock.
- Kernel scheduling policy can change without changing continuation grammar.
- The interpreter receives an already-authorized Move and has no path for
  minting, widening, or independently rechecking authority.
- Storage applies a declared transition; it does not derive semantic effects.
- No external callback or effect executes inside the state transaction.

## 2. Specify the continuation interpreter as a transition function

Before extracting code, declare the interpreter's complete input, state, output,
and laws. The useful conceptual interface is:

```text
Interpret(current continuation state, authorized Move)
    -> next-state transaction
    + declared runtime effects
    | semantic rejection
```

The exact Go signature remains open. The important constraint is that the result
is data, not immediate I/O. A result may declare child admissions, a parent
resumption, a root result, a fault, or a dispatch request, but executing those
effects belongs outside the interpreter.

Close at least these language questions before generalizing the implementation:

- whether one continuation may declare multiple child waves;
- register completion and mismatch semantics;
- whether partial observations can resume a parent;
- cancellation and fault propagation through the invocation tree;
- duplicate and conflicting Return behavior;
- terminal-state laws; and
- the migration from `Action + Params` to ordered `Operands`.

DNA16's clone-and-rollback behavior is a useful executable specification, but
the target is an explicit write set rather than cloning the complete resident
map. Sequence/generation fencing remains part of transition admission.

Acceptance criteria:

- A table of state + Move -> transition/rejection covers every legal state.
- Property tests cover duplicates, reordering, stale generations, and terminal
  state monotonicity.
- A transition identifies the records it reads and writes.
- The same interpreter can run against an in-memory or durable store.

## 3. Decompose Driver into semantic adapter, live session, and transport

The DNA16 `CodexDriver` currently contains three different responsibilities:

1. provider protocol interpretation and Skyra mapping;
2. live request/thread routing and receive-loop concurrency; and
3. host command delivery through `Transport`.

Separate them conceptually:

```text
Provider adapter
    provider message <-> normalized Skyra proposal
    owns provider schema and private protocol meaning

Driver session / host binding
    owns live request correlation, lifecycle, recovery status, and event delivery

Transport
    owns complete-frame movement and closure only
```

The provider adapter may be stateful in the continuation sense, but it must not
own process handles, pipes, goroutines, mutexes, contexts, or retained runtime
callbacks in its persisted state. The live session may index provider IDs, but
those indexes must be classified as derivable cache, reconciled session state, or
explicitly non-recoverable state. There must not be two silent authorities for
the same provider phase.

Replace invocation-specific retained output callbacks with a driver event sink
whose events carry explicit invocation identity and handoff generation. The sink
has one declared lifetime and stale events are rejected at its receiving boundary.

Acceptance criteria:

- The provider adapter can be tested from recorded frames without starting a
  process or goroutine.
- The host/session can be tested with a trivial adapter.
- Transport tests contain no provider concepts.
- Rebinding, cancellation, host loss, and late provider messages have explicit
  outcomes.
- Persisted driver state is the declared authority; every live index has a stated
  reconstruction or reconciliation rule.

## 4. Make normalization a single, typed trust boundary

DNA16 sends `DriverEmission.Raw` back through `Driver.Egress`, after which the
controller allocates identity, binds parentage/root, constrains inherited
authority, and validates the resulting Moves. This splits normalization across a
raw string, a second driver call, and controller rewriting.

Use two explicit stages instead:

```text
provider frame
    |
    v
driver-owned normalization
    -> typed, provider-neutral proposed Invoke / Return / state / command
    |
    v
runtime-owned adjudication
    -> bind source, root, child identity, ReturnFd, generation
    -> enforce non-escalation and structural validity
    -> produce a committable transition
```

The driver owns normalization because only it understands provider meaning. The
runtime owns adjudication because driver output is untrusted and cannot assign
authoritative identity, parentage, authority, or lifecycle state. Provider JSON,
thread IDs, request IDs, and native-tool vocabulary must not cross the normalized
proposal boundary except inside opaque driver state or opaque host commands.

Prefer a closed typed proposal family over `Raw string`. Do not expose a
half-authoritative `protocol.Move` that looks committed while still requiring
identity and authority rewriting.

Acceptance criteria:

- One driver method turns provider input plus detached state into a complete
  normalized proposal.
- No shared runtime component reparses provider bytes.
- The adjudicator is the only place that binds runtime-owned fields.
- A normalized proposal cannot represent an authority-bearing value supplied by
  the provider.
- Invalid normalization faults or rejects deterministically without partial
  state mutation.

## 5. Separate semantic commit from effect delivery and live execution state

The current commit-before-send rule is correct but incomplete: a crash can occur
after a transition commits and before its host command is delivered. At the same
time, resident continuations mix semantic state with `context.Context` and other
process-lifetime concerns.

Define two separations together:

```text
Continuation record             Execution lease
---------------------------     ---------------------------
serializable language state     context and cancellation
driver-state payload            live driver/session binding
generation and lifecycle        local goroutines/handles
pending durable effects         process-lifetime bookkeeping
```

Committed external commands need an explicit lifecycle such as pending, sent,
acknowledged, failed, or delivery-unknown. Exact retry behavior must be based on
the provider's idempotency and reconciliation guarantees, never on an assumed
successful send.

Acceptance criteria:

- An authoritative continuation record contains no context, callback, channel,
  mutex, process, or transport.
- State replacement, child effects, and durable outbound-command creation commit
  atomically.
- Restart behavior is defined for every command delivery state.
- A provider without safe reconciliation faults explicitly rather than silently
  replaying or resetting live work.
- Effect execution can fail without undoing an already committed semantic fact.

## Recommended order

Do the work in dependency order:

1. Name the roles and specify the checked-Move -> interpreted-transition path.
2. Write the continuation transition laws and tests.
3. Extract store/transaction and effect execution from the current Scheduler.
4. Introduce typed driver normalization and runtime adjudication.
5. Split live driver sessions from provider adapters, then add durable effect
   delivery and recovery semantics.

Avoid beginning with package moves or type renames. The refactor is complete when
the ownership and test seams change, not when the same methods reside in more
directories.
