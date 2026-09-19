# v14 — Invoke and Return local runtime

**Status:** single-wave local runtime implemented and tested; v13 is unchanged.
The design history below records the proposal and accepted rules implemented in
the [component packages](README.md). Compiled matching and general execution policy remain open.

[protocol/protocol.go](protocol/protocol.go) incorporates the latest protocol decisions. The earlier v14 reply-route
and fresh-return-invocation proposal is superseded by this document.

## The central distinction

**Invoke establishes work. Return contributes its observation.** A Move transports either transition.
Creating a child does not fill its observation slot. Dispatch acknowledgement is not a return.

```text
Move { Root, Target, Op }
  Invoke { InvocationID, Action, Params, ReturnFd, Register, Horizon, Focus }
  Return { Observation }
```

Target contains invocation identity plus entity/address/host routing. An Invoke identifies its own
execution through Invoke.InvocationID; its Target routes to an endpoint. A Return targets the
invocation that produced it. It does not allocate another invocation or execute an endpoint.

This deliberately replaces the earlier rule that every Move was a fresh invocation. MoveID leaves
the execution wire. EngramID remains a durable-storage concern; hashing action/params is not a
prerequisite for the local echo loop. No decisions about durable storage encoding are made here.

## Execution-bound returns

```text
execute captures child Invoke.InvocationID
  → driver.Ingress(move, output)
  → output(raw) → same driver.Egress(raw)
  → Return.Target.Invocation = captured child ID
  → Scheduler.Write resolves that child
  → stored child Invoke.ReturnFd identifies its parent
  → parent's matching child observation is filled
```

The driver cannot choose another invocation to answer. execute binds the identity, root, and
source metadata; Scheduler verifies that the event source and return target match a known entry.
Raw output does not supply trusted authority, parent links, register mutations, or IDs.

Return no longer needs InReplyTo, ReturnTo, action/params, a register, or Horizon. It uses the route
established by the dispatched Invoke. Horizon authorization still applies to every dispatched
Invoke, including a parent resumption. An external caller cannot inject a Return through Submit.

For forward output, execute allocates the child Invoke.InvocationID, sets ReturnFd to the captured
producer, and inherits its root and granted Horizon. Driver-selected Focus can only select inherited
grants. Requested endpoint/action is validated, then authorized at dispatch. The child gets its own
independent continuation and is recorded in the parent's register.

## Register ownership and parent inputs

The authoritative register lives on the resident Invoke, within Scheduler's continuation table.
A dispatch Move contains a detached Invoke.Register snapshot. Drivers adapt that snapshot to their
endpoint; they cannot write through it to Scheduler. Output-supplied registers are discarded.

A parent's child invocation record contains the child Move with an empty register. The child's
independent continuation owns its actual frame. This avoids recursively copying live frames into
parent records. Register.Order preserves declaration order alongside the child map.

When all children in the local probe have returned, Scheduler resumes the **existing parent Invoke**
with its updated register and saved context. Return itself does not carry the parent inputs or
invoke the parent driver. The root fixture distinguishes initial execution from resumption by the
empty versus populated child register; it does not keep a second aggregation state.

## Compiled expectations and interpretation — latest decision

Scheduler owns compiled expectations on Continuation.CompiledRegister alongside the live state.
The live register still records invocations and returned observations; the compiled register
represents what was expected. Its internal fields remain open and are not a duplicate live frame.

Scheduler's private `interpret(live Register, compiled *CompiledRegister) Assessment` method
compares live state against those expectations and returns an assessment:
Waiting for missing observations, Satisfied for present matching observations, or Mismatch for a
contract violation. These are interpretation results, not additional dispatch lifecycle states.
A boolean contract matcher can be stubbed once required inputs are available.

Scheduler alone applies the assessment and selects what proceeds. Satisfied permits following an
already-defined compiled continuation; it does not identify the next instruction by itself.
Without compiled expectations, declaration-path selection remains open. Invoke/Return and
declaration/compiled execution remain separate distinctions.

This replaces the separate Interpreter interface; there is no interpreter dependency to wire.
The method reads without mutation, under the scheduler lock; the write path applies the assessment
under that same lock. Compiled matching and CompiledRegister fields remain explicitly stubbed.
Nil compiled expectations do not imply satisfaction. State-to-route mapping remains open.
The single-wave resumption rules below are implemented for the local proof;
they do not define general compiled or declaration behavior.

## Supporting additions retained from the first v14 proposal

| Gap | Implemented surface | Purpose |
|---|---|---|
| Context | Continuation.Context; Write(ctx, event) | Retain execution lifetime; do not replace parent context with a callback's context. |
| Terminal result bookkeeping | Continuation.Result | Distinguish an answered invocation from handoff state and reject conflicting duplicates. |
| Parent resumption reservation | Continuation.Resumed | Reserve one parent resumption under the mutex for this single-wave probe. |
| Fan-out admission | MoveBatch with bound Source | Record all siblings before an immediate child callback can resume the parent. |
| Failures | AdmissionFailure and ExecutionFailure | Distinguish rejected external input from a failure attributed to a known producer. |
| Endpoint contract | InputValidator on DriverRegistration | Supply a narrow executable validator without inventing generic JSON Schema support. |
| Root outcome | RootResult callback | Expose root success/failure without claiming global quiescence. |
| Test observation | Detached Snapshot methods | Inspect resident state without manufacturing execution progress. |
| Concrete fixtures | EchoDriver, RootDriver, ProbeInput, ProbeAuthority, ProbeOutput | Declare what the constrained implementer previously could not add. |

The bookkeeping and signatures in this table are implemented supporting surfaces. General
repeated fan-out/resumption and retry semantics are not solved by a single Resumed flag.

## Accepted minimal selection rules

A newly admitted Invoke may dispatch once. A Return records its child's observation in the parent.
Missing child observations keep the parent waiting; when every child in the declared batch has
returned, Scheduler selects the parent's existing Invoke once with its updated register. The
entire batch is admitted before any child is dispatched. An empty child set is not a resumption
trigger. Compiled matching remains stubbed; this local completeness rule does not assert that a
compiled contract passed. These selection rules are accepted for the echo proof.

The supporting transitions below are implemented for the local loop; general retries and
repeated fan-out remain outside the proof.

## Implemented local transitions

The local proof supports one child-declaration batch and one parent resumption per invocation.
One driver output is either a batch of Invokes or exactly one Return; mixed or streaming batches
require a later declaration-completeness rule.

1. Admit a root Invoke, or atomically admit a source-bound batch of child Invokes with their parent
   register entries. Zero State is unselected. Do not overwrite existing invocation IDs.
2. Under the scheduler mutex, select eligible Invokes and reserve handoff as Dispatched. Unlock
   before CC or driver calls. Only Invoke goes through Dispatch/execute.
3. Record authorization and delivery outcomes through Write. Accepted handoff may become InFlight;
   failed authorization/delivery becomes Faulted. Neither contributes a child observation.
4. Apply a source-bound Return to the targeted invocation's Result. Follow its stored ReturnFd to
   fill the parent register. Zero ReturnFd denotes the host/root boundary. A Return creates no new
   continuation. Identical duplicate returns are no-ops; conflicting results fail without overwriting.
5. The local completeness rule requires all declared children to have observations. Reserve
   Resumed under the same lock and select the parent's existing Invoke once with its updated register.
   A late acknowledgement must not clear Result, reopen selection, or trigger a second resumption.
6. A root result or fatal workflow fault makes one host-notification attempt outside the lock.
   Callback errors propagate; no automatic notification replay or work retry is implied.

Dispatched, InFlight, and Faulted remain the only named lifecycle states. Result and Resumed carry
separate facts. There is no fresh Return invocation or return-dispatch lifecycle to track anymore.
Records remain resident for the proof; retirement is separate work.

A dispatch callback can arrive after a return and even after parent resumption has started. The
probe must preserve its reservations and terminal facts; it must not treat an old acknowledgement
as evidence about a newer handoff. Precise multi-attempt status needs an explicit later protocol;
no hidden attempt counter or completion side table is authorized by this proposal.

## Validation, callbacks, and context

CC.Validate checks Invoke structure and endpoint input contract. Return needs an observation and a
bound invocation target, not an endpoint action/schema. Scheduler verifies resident source linkage.
Validate returns a pure verdict: Submit/execute turn it into one scheduler event. This lets
execute validate all candidate children before atomic batch admission. Horizon is checked at dispatch.

```go
Ingress(ctx context.Context, move Move, output func(string, error) error) error
Egress(raw string) ([]Move, error)
```

Ingress acknowledges acceptance; output can happen before that acknowledgement or later. Its error
argument reports asynchronous execution/transport failure. Egress returns candidate transitions or
a parse failure. execute attributes those failures to its captured source via ExecutionFailure.
An error returned by the output callback reports processing/delivery failure and must be surfaced.
It must not cause blind replay of an already-admitted emission.

Submit's context is the workflow execution lifetime. Children inherit it; resumed parents keep their
own saved context. Selection returns detached Continuation snapshots carrying that context and Move.
No driver/host callback executes under Scheduler's mutex. Cancellation cannot prove remote work stopped.

ProbeInput implements the narrow contract for Invoke with one observation. Drivers must register
an input validator; there is no missing-schema bypass. Full parent return-contract checking remains
stubbed separately. Authority fixtures are test policy, not a production grant/signature implementation.

## The echo proof

RootDriver's initial Invoke emits two identical child Invokes in one response. They have the same
action/params but distinct InvocationIDs. EchoDriver returns each input observation. When the parent
resumes, its Invoke.Register contains both results; RootDriver returns JSON encoding of their ordered
[]Observation to the host. It has no private result aggregation map or completion counter.

EchoDriver.Release is nil for immediate output. Delayed mode queues bound output operations through
an explicitly declared hook so a test can release them in reverse order and inspect callback errors.
ProbeOutput is only the fixtures' JSON format; its kinds are "invoke" and "return". It supplies no
trusted source ID or grants. Other drivers may use other representations.

Execution evidence covered by `ctu/ctu_test.go` and `scheduler/scheduler_test.go`:

- CTU.Submit produces one root result through actual ingress/egress and Scheduler.Write.
- Identical child inputs, distinct IDs, reverse completion, and correct parent entries.
- Immediate callbacks, late dispatch acknowledgement, and no duplicate parent resumption.
- Context cancellation, denial, malformed output, and callback errors remain observable.
- Return cannot target another invocation through raw output or public Submit.
- Snapshot mutation cannot mutate resident state; constructors reject missing dependencies.
- Scoped build, vet, and race tests pass. Compilation alone is not execution evidence.

The local runtime and fixtures are implemented. No production kernel, durable loader, network
protocol, general retry policy, multi-wave scheduler, or retirement behavior is included.

## Scheduling consolidation and recorded surface changes

The [builder change order](change.md) is implemented. `Write` mutates state,
selects affected work, and reserves handoffs under the scheduler mutex, then calls
`run(selected []Continuation)` after unlocking. `run` dispatches those snapshots
without selecting again. It checks root terminal state before each handoff,
releases the lock before CC calls, retains saved contexts, and propagates callback
errors. Remaining selected work is suppressed after an observed root fault or result.

The implementation added private `rootLocked` and `faultLocked` methods, cloning
helpers, dependency checks, empty/completed-register checks, and fixture parsing.
These additions are retained; both root helpers require the caller to hold the
scheduler mutex. The test-only `testDriver` retains its `ingress` and `egress`
function fields. The former unused `run(move Move)` has been replaced by the
snapshot-dispatch signature above. This consolidation adds no runtime structs,
fields, or public interfaces.

`Dispatch` reports failures before ingress; after authorization, `execute` owns
the ingress outcome report. Output callbacks independently write bound moves or
execution failures. A callback write error is surfaced without reclassifying it
as handoff failure or replaying it from Dispatch. A successful handoff means
authorized and accepted, not completion. Each handoff outcome is reported at
most once, and late acknowledgements preserve returned/faulted state. No retry
or compiled interpretation policy is added.

## Component package decomposition

The former `runtime.go` and `proposal.go` declarations now live with their owning
components: `protocol`, `driver`, `controller`, `scheduler`, `ctu`, and `probe`.
The root `dna` package forwards constructors and aliases public types for existing
callers. No mutable state or alternate scheduling path lives in that facade.

Shared cloning, dependency, and register helpers live in `internal/support`.
Their names are exported for use across the component packages, while Go's
`internal` boundary restricts their import to v14. `rootLocked`, `faultLocked`,
`selectMoves`, `run`, `interpret`, and fixture parsing retain private ownership.
Integration tests moved alongside CTU's private wiring; the private `run` test
moved to scheduler. No runtime fields or behavioral policies were added.
