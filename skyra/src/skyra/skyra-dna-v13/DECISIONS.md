# Minimal loop probe decisions

Status: independent component implementation complete; the echo-loop proof is blocked.
No declared types, fields, or signatures were changed. No runtime fixture types,
globals, or invocation side tables were added. Existing proposal comments remain
as the design record; implemented method bodies and this report describe current behavior.

Implemented:
- CTU constructs registry → controller → scheduler, rejects missing root/authority
  dependencies, and delegates Submit and registration through their owning methods.
- Registry serializes registration/lookup, rejects duplicate addresses and nil
  drivers (including typed nils), and detaches schema bytes. Replacement is unsupported.
- Validate checks basic envelope invariants, cancellation, and registration;
  emits Move or ValidationFailure through the supplied callback. An absent schema
  imposes no schema check in this probe. Nonempty schemas fail explicitly until
  dialect and ordered-Params mapping are defined; registration checks JSON syntax only.
- Dispatch resolves a registration, invokes Authority.Check, and reports denial or
  authorized execution failure through DispatchResult. It returns the callback's
  error, not the outcome error. Execute currently rejects before driver ingress.
- Move snapshots detach parent pointers, observations, capability signatures,
  and pointer/value Ret payloads. Authority receives its own snapshot.

Explicitly unfinished: Scheduler.Write rejects without mutating state; run rejects
because it cannot propagate the submitted context; selectMoves retains an explicit
panic until eligibility is defined. Execute rejects before starting a driver because
binding and parse-failure reporting remain open. Parent dependency/contract checking
is stubbed by leaving scheduling unavailable, not by treating contracts as satisfied.
No successful Submit or driver execution is claimed.

Proposed minimal lifecycle semantics (not implemented or adopted): Call admits an
independent continuation plus its parent child entry, using zero as unselected.
Under the scheduler mutex, selection reserves zero-state Calls as Dispatched;
successful acknowledgement marks InFlight, failure marks Faulted. Ret creates its
own continuation, fills only the existing parent's InReplyTo entry, and never enters
Call selection. An existing correlated observation prevents late acknowledgement
from changing child state. InFlight would not mean complete; retirement remains open.
Duplicate calls must not overwrite entries. Invalid producer input must not mutate
an existing invocation merely by claiming its ID. No retries or hidden completion set.

Reproduce from the skyra module directory:

```sh
go build ./src/skyra/skyra-dna-v13/...
go vet ./src/skyra/skyra-dna-v13/...
go test -race ./src/skyra/skyra-dna-v13/...
```

All three passed. Six component tests cover dependency rejection, validation and
callback error semantics, dispatch lookup-failure reporting, concurrent blocked
admission, snapshot isolation, and distinct Ret identity. They do not exercise
successful registry binding/authorization, driver ingress/egress, selection, parent
returns, or completion. Direct component checks are not substitutes for the required
immediate/delayed echo proof. No production authorization implementation is supplied.
