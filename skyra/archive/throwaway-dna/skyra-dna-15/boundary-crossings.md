# Boundaries and settled failure rules

All implementation, tests, declarations and reports remain in skyra-dna-15.
v14 was read only. No module configuration or external specifications changed.

## Boundaries that held

- One scheduler mutex owns semantic Registers and resident Handoffs. write does
  not select, dispatch or notify; transition owns progression and reservations.
- Validators, drivers, kernel initialization, routing and host callbacks run
  outside the scheduler lock. Pure trusted-grant checks run inside it.
- Drivers receive detached Moves. Pending output stays resident until placement;
  accepted children/observations alone enter Registers. Saved contexts persist.
- Source IDs and sequence numbers are runtime bound. Stale placement cannot
  satisfy a resumption. Only Invoke is dispatched; Return directly records results.
- One private registry is shared by separate non-embedding adapters. Its resolver
  cannot assert Registrar, and public CTU exposes no registration method.
- Host callback errors propagate without execution-failure replay. Rejected
  admission cannot fault existing work just by claiming its identity.

## Settled failure seams

The user explicitly approved the proposed rules. No required policy seam remains
blocked in this increment.

| Former seam | Implemented outcome |
| --- | --- |
| Expired/denied continuation stamp | No successful effects; reject resident output and fault its producing workflow. External admission remains rejection only. |
| Consumption after rejection | Clear pending output, set OutputConsumed, retain the original Rejection diagnostic, and fault once without retries. Propagate processing and notification errors. |
| Late execution-fault precedence | Ignore older-handoff faults. Apply current-handoff faults to active workflows even after the producing child returned. Preserve accepted observations and successful placement receipts. Terminal workflows cannot reopen. |

The W0 failure-policy revision preceded dependent implementation. ErrPolicyBlocked
was removed; Handoff.Rejection now diagnoses a consumed rejection. Tests verify the
settled outcomes, including exactly one notification attempt when that attempt fails.

## Checks

Scoped build, vet and race tests passed. Deterministic tests demonstrate the real
loop, placement ordering, failed-placement discard, atomic batches, authorization
before effects, source/context/snapshot boundaries, fault suppression, callback
errors, registration concurrency and boot capability separation. No timing sleeps
or manufactured aggregate root result substitute for the integrated proof.

## Corrected event-ordering defects

The issues.md review found two gaps in the earlier proof. Matching buffered output
repeated inside a placement envelope is now applied once without losing selection
candidates; conflicting envelopes reject before partial effects. Buffered Returns
now recheck the resident context at semantic application and use the established
consumed-rejection failure path on cancellation. Permanent tests cover both root
and child cancellation and real loop completion after the repeated envelope.
These fixes required no new surface or ownership changes.
