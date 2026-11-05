# Human root

A Mission's human position sits above its worker trees. It is neither an act
node nor a delegate node. This package imports no worker-tree code and exposes
no new worker instruction. The existing apprentice tree, its `Kind`, `Step`,
and Core.lean remain unchanged.

The root may author human observations repeatedly, receive verified results
from its pending children, and invoke another child with selected observations
and a bounded grant. Its observation identities bind the Mission, source
record and output position. Passing Think's returned spec to Act preserves
that observation's identity, value and source. Editing the value creates a
new human observation instead.

## Pure protocol

`Step(state, ingress, event) → (state, effect)` does no I/O and never changes
the input state. `Authored` and `Invoke` require human ingress; `Returned`
requires runtime ingress; `Committed` requires storage ingress. The trusted
adapter assigns ingress from the actual channel, never from a JSON field.
The worker's local node `0` does not acquire any human privilege.

A valid input yields one `Commit` effect and a pending state. Only storage's
acknowledgement of that exact record introduces observations into scope or
releases a `Dispatch`. A duplicate request yields `Duplicate`, including the
original call reference for an invocation, and never another dispatch. Replay
re-derives the records and restores scope and pending calls without dispatching
unknown outcomes again.

The configuration pins the human principal, callable entrypoints and grantable
capability references. The references must come from the operator's catalogue
of exact contracts and resource scopes; they are not model-selected tool names.
`MayPass` requires held observations, and `MayGrant` requires a subset of held
capabilities. Empty means empty. Resolving those references, checking payloads,
narrowing actual OS grants, and running children remain runtime obligations.

## Trusted adapter

`controller/runtime/humanroot` journals the pure protocol. It uses separate
human `Author`/`Invoke` methods and a runtime `Receive` method. Receive requires
a trusted callback that verifies the runtime receipt against the known child
invocation and reads the actual outputs. A supplied receipt string alone does
not establish a result's authenticity. No generic model-facing root endpoint
is registered.

The journal pins its configuration and stores sequence- and content-addressed
records with file and directory synchronization before acknowledging commits.
A process-shared lock serializes adapters. Put it in runtime-owned storage
that worker sandboxes cannot read or write.

This is the root protocol and adapter foundation. It does not yet replace the
webhost's editable spec storage, generate Think specs, or reroute the existing
Mission Run button through the human root. Those integrations must use this
adapter and the existing worker dispatcher rather than giving a worker root
instructions or importing I/O into its pure core.

## Proof and checks

`controller/apprentice/invariants/Invariants/HumanRoot.lean` models authoring,
scoped passing, preserved origins and bounded grants. Its ingress-authentication
and runtime-receipt checks are explicit trusted-adapter assumptions; it does
not purport to prove the OS, JSON parser, filesystem durability, or full Go
state machine. Fifteen lemmas establish the corresponding model properties.
Go checks all 15,628 decision vectors written by `HumanRootVectors.lean`.
Journal tests additionally cover the human → Think → human → Act handoff,
restarts, concurrency, forged receipts, corrupt records and duplicate dispatch.
