# Work order — DNA 15 implementation rules

**Status: W0–W4 complete within the local implementation scope.**
The user authorized implementation and approved the three formerly open failure
rules. Declarations, bodies, integrated proof and required checks are recorded in
[DECISIONS.md](DECISIONS.md); settled outcomes are in
[boundary-crossings.md](boundary-crossings.md). Permitted fixture-policy and compiled
matching stubs remain explicit.

## Objective

Build the smallest local engine that composes proposal 3's transition loop with
kernel-owned registration. Prove boot → root Invoke → two child Invokes → child
Returns → one root resumption → root Return through real component calls.

Stub explicitly deferred policy, not the mechanics required for that proof.
A passing build alone is not a working loop.

## Authority and scope

Read these local documents as the working set:

1. [contracts.go](contracts.go): the current v14 declaration baseline, consolidated
   into one file. It is not yet the complete v15 implementation surface.
2. [Loop proposal](proposals/loop/loop-proposal.md): intended event ownership,
   transition composition, placement gate, and authorization boundary.
3. [Placement algorithm](proposals/loop/placement-gated-continuation-algorithm.md):
   source association, resident handoff state, and processing order.
4. [Registration proposal](proposals/kernel-registration/proposal.md): boot,
   capability separation, endpoint installation, and public API migration.

The proposals specify intended behavior; the declared contracts specify the
currently available surface. A mismatch is a surface-change item, not permission
to conceal missing state or retain superseded behavior. Existing v14 comments in
contracts.go describe the baseline, including its old authorization placement.

Keep all implementation changes, tests, and reports inside `skyra-dna-15`.
The directory name is exactly `skyra-dna-15`, without a `v`. Do not modify v14,
older versions, module configuration, or external specification directories.
Use v14 as a read-only implementation reference when a scheduled item needs it.
Historical work orders and linked older proposals are not additional authority.
Preserve unrelated and pre-existing user changes.

## Declare the surface before implementing it

Keep structs, interfaces, supporting type declarations, and contract comments in
`contracts.go`. Implementation bodies belong in separate Go files in this package
once implementation is explicitly started. Do not scatter duplicate declarations
across component files or introduce a package decomposition in this increment.

Before dependent implementation, schedule and settle the concrete declarations
for these already identified changes:

| Surface item | Required declaration or revision |
| --- | --- |
| Resident handoff | Add `Continuation.Handoff` with sequence, placement, pending output, and consumption state. Leave the semantic Register separate. |
| Event association | Represent the two event families and bind producer InvocationID plus handoff sequence to feedback/output. Remove the authorization verdict from dispatch results. |
| Authorization | Declare the Move stamp and the bounded trusted-Horizon check used by continuation processing. Do not relocate the arbitrary v14 authority callback under the lock. |
| Kernel boot | Declare registration-only initialization, kernel fixture state, separate resolver/registrar adapters, and any public facade required for capability separation. |
| Public CTU | Remove public registration access; retain Submit and Snapshot. Specify constructor inputs for separate kernel/root setup. |
| Composition methods | Record the concrete transition, event-application, assessment, selection/reservation, and dispatch signatures required to compose the engine. |
| Test fixtures | Declare any additional fixture shapes needed for deterministic placement, authorization, and callback tests. Test code is not an exception to the surface rule. |

This list identifies the permitted design work; it is not a blanket authorization
to invent fields or signatures while filling bodies. Schedule the contract update
as its own reviewable item before dependent runtime items. Record the exact change,
its owning component, and the proposal requirement it serves.

Once a work item's surface is established, implement within it. Do not add structs,
fields, interfaces, or alter signatures to make a test pass. Propose the smallest
missing surface separately and continue independent work. Existing marker methods
and interface implementations omitted from the baseline file must be supplied
when their implementation item begins; declarations alone do not implement them.

## Engine ownership rules

- `transition()` is the sole progression entry: apply events, assess affected
  continuations, select, and reserve under the scheduler mutex. Immediate and
  delayed feedback reenter this same path.
- `write()` recognizes events and delegates application. It returns affected
  continuations, an optional notification, and a processing error. It does not
  select, dispatch, notify, or own a second execution loop.
- Helpers called under the scheduler lock do not reacquire it. Document lock
  preconditions. Do not call drivers, endpoint validators, kernel initialization,
  dispatch, host callbacks, or arbitrary policy callbacks under that lock.
- Release the mutex before external work. Before each reserved handoff, check
  resident root completion/fault under the mutex and unlock before dispatch.
  Suppression is not cancellation of a handoff already proceeding after unlock.
- Scheduler owns resident continuation state. Component interactions use their
  declared methods; do not bypass ownership through direct field access, globals,
  side tables, driver aggregation state, or unrelated payload fields.
- Received output belongs to `Continuation.Handoff`. Only accepted declarations
  and observations enter semantic registers. Both remain under the same scheduler
  mutex; no additional handoff state store or manager is needed.
- Drivers receive detached Moves and semantic Registers without resident handoff
  bookkeeping. Preserve saved execution contexts and deep snapshot isolation.

## Protocol and authorization rules

- There are two event families: runtimeEvent (also called dispatchEvent) and
  continuationEvent carrying Invoke/Return. A combined envelope is not a third
  family. Process its runtime portion first.
- Check authorization for both Invoke and Return in continuation-opcode processing
  before opcode effects. The runtime generates the check result and timestamp.
  Resolve Return's trusted Horizon from its bound resident source.
- Dispatch and runtime-event callbacks do not perform Horizon authorization or
  report an authorization verdict. Output released by placement feedback still
  goes through continuation authorization. Add no separate resumption auth hook.
- Reserve the handoff before dispatch. Match feedback by producing InvocationID
  and handoff sequence. Initial and resumed parent executions reuse the invocation
  identity but have distinct sequences.
- Buffer early output and return callback acceptance without waiting for placement.
  Placement success releases it; failure discards it and applies the runtime fault.
  Unconfirmed placement permits no continuation effects. External root admission
  has no producing handoff and retains its validated admission path.
- Validate complete child batches before atomic installation; install every child
  and parent entry before selecting any child. Preserve distinct child IDs, bound
  parent links, inherited grants, and declaration order.
- Only Invoke is dispatched. Return applies an observation directly. Preserve one
  parent resumption, identical-return idempotence, conflicting-return rejection,
  terminal results/faults, and one root-notification attempt.
- Keep execution faults, admission rejection, and processing/notification errors
  distinct. Do not replay a callback error as another execution failure or fault
  existing work solely because rejected external input claimed its identity.

## Registration and boot rules

One private registry serves the runtime. Separate non-embedding adapters expose
resolution to the controller and registration to the explicitly designated kernel.
The controller's resolver and public CTU must not reveal registration by type
assertion. This is an ownership boundary, not native-code sandboxing.

Trusted boot installs the designated kernel endpoint, wires the engine, and calls
registration-only initialization. The fixture kernel installs a fixed local list,
including the separately supplied ordinary root and echo endpoint. Initialization
starts no workflow. Publish Submit/Snapshot only after initialization succeeds and
the required root is available. Failed boot returns no public runtime.

Keep the registrar private to the kernel after boot; the fixture installs all its
endpoints during initialization. Do not deliver placeholder Admit/Fault powers.
Preserve atomic driver/validator installation, append-only bindings, duplicate
rejection, canonical-address checks, and nil/typed-nil dependency checks. Endpoint
implementations remain responsible for their own concurrency safety.

## Work sequence

| Item | Deliverable | Dependency |
| --- | --- | --- |
| W0 — contract preparation | Exact v15 surface changes and method signatures recorded in the single contract file; unresolved semantic choices listed separately. | Before dependent bodies. |
| W1 — registration and boot | Registry adapters, dependency checks, registration-only kernel setup, public CTU boundary. | Relevant W0 declarations. |
| W2 — engine composition | Transition and application helpers, resident Handoff ownership, routing/driver wiring, saved contexts. | Relevant W0 declarations. |
| W3 — gating and authorization | Placement association and buffering, Invoke/Return stamps, atomic effects and reservation. | W2 and relevant W0 declarations. |
| W4 — integrated proof | Real boot-to-root-result echo flow plus ordering, failure, and boundary tests. | W1–W3. |

Keep each item bounded and independently reviewable; state dependencies rather
than pretending dependent items can be dropped independently. Do not widen an item
into adjacent cleanup. Record actual completion and blocked portions separately.
This schedule does not authorize starting implementation during document preparation.

## Stubs and unresolved behavior

Permitted scope stubs are the fixed local kernel registration policy, trusted test
grants/test authorization policy, and compiled matching returning Waiting with the
separate completeness rule for one parent resumption. They must be explicit and
must not be reported as production policy or compiled contract satisfaction.

The previously open expired-stamp response, late execution-fault precedence and
rejected-output consumption rules were settled by user approval. Reject resident
output without successful effects, consume it and fault its workflow once without
retries. Ignore older-handoff and terminal-workflow faults; apply current-handoff
faults to active workflows, including after the producer returned. External
admission rejection never faults existing resident work. See the local reports
for the implemented behavior and verification.

Persistence, retirement, repeated execution waves, retries, production authority
verification, network transport, dynamic loading, and registration replacement or
removal are outside this order.

## Proof and checks

Use real Submit, driver ingress/egress, bound callbacks, transition application,
and root notification for the integrated proof. Do not manufacture parent results
or call scheduler Return application directly in place of that proof. Focused unit
tests may inject events to verify rejection and ordering boundaries.

Verify the following with deterministic coordination, not timing sleeps:

- Immediate, delayed/reversed, and concurrent child completion produces the ordered
  root result, with distinct child identities and one parent resumption.
- Output-first, placement-first, and combined-envelope delivery obey the placement
  gate; immediate output callbacks return without deadlocking acknowledgement.
- Placement failure after buffered output has no continuation effects and attempts
  one root failure notification. Duplicate/stale feedback cannot reopen terminal
  work, satisfy the wrong handoff, or repeat effects.
- Invoke and Return pass authorization before effects; runtime-event callbacks do
  not perform it. A denied opcode cannot apply its successful effects. Report any
  unresolved terminal rejection behavior as blocked rather than claim full coverage.
- Whole-batch rejection installs nothing. Source binding, context preservation,
  snapshot isolation, fault suppression, and callback-error propagation hold.
- Registry duplicates, concurrent lookup/installation, typed nils, boot failure,
  and capability separation meet the registration proposal's acceptance criteria.

Run scoped checks from the `skyra` module directory once implementation exists:

```sh
go build ./src/skyra/skyra-dna-15/...
go vet ./src/skyra/skyra-dna-15/...
go test -race ./src/skyra/skyra-dna-15/...
```

Establish the starting scoped result and rerun relevant checks after each completed
item. Do not repair unrelated module failures. Existing behavior tests may change
only to reflect the documented v15 contract changes; do not weaken assertions to
hide regressions. Build and execution results are recorded in DECISIONS.md.

## Implementation report

When implementation is started, maintain concise local reports:

- `DECISIONS.md`: item IDs, declared surface changes, implementation choices within
  those boundaries, and reproduction commands/results.
- `boundary-crossings.md`: boundaries that held, blocked seams, and the smallest
  proposed changes needed to proceed. Distinguish design gaps from ordinary defects.

Report separately: declarations completed, bodies implemented, checks passed,
execution demonstrated, and behavior still stubbed or blocked. Full completion
requires the integrated proof and required checks, with unresolved required paths
identified as unfinished. Stop at the scope of the scheduled work.
