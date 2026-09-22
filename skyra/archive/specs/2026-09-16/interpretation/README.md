# Interpreter construction and boot

Date: 2026-09-17.

Status: accepted proposal for construction and processing boundaries. Supporting
state read operations, Result fields, and package wiring remain provisional.

Interpreter is a concrete struct constructed with `interpretation.New(state,
custody)`. Boot supplies InterpretationState and CustodyState through distinct
private adapters bound to Interpretation's identity. The custody adapter permits
transfer to Supervisor. The constructor fixes `self = Interpretation` and
`next = Supervisor`; these identities are inherent to the Interpreter's role.
StateOwner still validates identity and authority at its boundary.

Boot creates StateOwner and the restricted adapters, constructs Interpreter,
finishes wiring the components, then starts `Run(ctx)` once and supervises its
return. Construction only stores dependencies: state access and processing begin
in Run. The full broker must not be passed behind a narrower interface because
type assertions could recover other authority it exposes.

Run takes one Move and finishes its interpretation and handoff before taking
another. The internal `interpretAndTransfer` helper follows the existing
[minimal loop](../ctu/minimal-loop.md):

```text
read detached state
    -> interpret Invoke / Return or normalize a Move-level fault
    -> prepare Result with related state changes
    -> Transfer(result, self, next)
        -> stale revision: re-read and reinterpret the same Move
        -> committed: take the next Move
        -> other error: propagate
```

InterpretationState supplies Take and the typed reads needed by interpretation.
Those read signatures await the continuation record contracts. CustodyState
submits Result; StateOwner atomically commits related changes, destination
mailbox placement, and custody before acknowledging. Result's exact fields remain
open. Cancellation and ambiguous acknowledgment follow the minimal loop's
reconciliation rules.

Interpreter owns its processing machinery and temporary working data. Canonical
records and both mailboxes belong to StateOwner. Move, Invoke, and Return remain
data-only types; their semantic handlers are internal to Interpreter.

See [proposal.go](proposal.go) and the updated
[custody signature](../tu/proposal.go). These declarations are build-excluded
proposals, not an implemented runtime.
