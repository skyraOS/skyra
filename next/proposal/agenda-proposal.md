# Agenda: proposal

Proposes changes to [agenda](../agenda.md). Ids refer to the
[list](README.md).

## What holds

The governing question, "semantics first", the distinction between
postcondition, witness and reduction, and the per-transition checklist all
hold. The new design is an answer to several of the agenda's questions.

## Superseded

| Agenda says | Now |
|---|---|
| "Return is `invoke(reply, payload)`" | a return **addresses the slot** (A1). It is still derived, never stated |
| "The reply capability's input contract is the return contract" | the emitter's own **output contract** is the return contract; the slot has no contract of its own |
| §2 `interpret`: "names a capability in reach, or none (→ reply)" | none → **leaf fault** carrying the output (A2) |
| §3 "The reply capability": issued at release, single use, default target, transferable, invoked by the runtime for faults | single use is the write-once `result` of the frame; there is no default target; transfer is the tail-call question; a fault is its own state carrying its observation (A6) |
| Invoke draft: "Authority is **narrowed**: the child's Horizon is a subset of the invoker's, issued only by the parent, and includes a fresh `reply`" | authority is **shared**: the child runs under the invoker's Horizon, unless the invoker narrows through the room or calls another agent (B6, D2, D5). The room issues every Horizon, never the parent |
| Fault draft: "the runtime invokes the instance's `reply` on its behalf" | the fault discharges the frame directly, carrying the observation of why |
| Fault draft: kinds "custody, malformed move, horizon, or authority" | add `leaf` (A2). Authority faults climb (24) |

## Answered questions

- **The addressing format:** `{"to": <slot | name | reference>, "payload":
  <JSON>}`.
- **What "state" contains:** the Horizon (authority: reach and grants), its
  durable stack (open slots, results, claims), and the proposal. Memory is
  still to come.
- **Where the authority fault fits:** it is a horizon fault that climbs until
  someone holds the authority; above the room, the human (24).
- **"What an input contract is attached to: a Potential, by name or by
  content hash?"**: to a surface by name now; by type id once types are
  records (E5).

## Edits to the document

- Mark §3 and the reply cases in §2 and the drafts as superseded, linking 24
  and improvements items 1 and 2.
- Replace the Invoke draft's "Authority is narrowed" sentence.
- Add `leaf` to the fault kinds.

## Code

None of its own.
