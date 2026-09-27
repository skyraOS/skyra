# Faults and interrupts

Date: 2026-09-23  
Status: agreed in outline. Builds on
[`thoughts/13-dependencies.md`](thoughts/13-dependencies.md).

When a step cannot go on normally, one of two things has happened, and they are
handled differently:

| | **Fault** (error) | **Interrupt** |
|---|---|---|
| Means | something broke, and the runtime proves it | the step cannot go on without something from outside it |
| Terminal? | yes | no: resumable |
| Discharges the dependency? | yes, with a fault | **no**: the dependency stays open while the decision is pending |
| Procedure | propagate; a continuation may recover | suspend, ask, then resume or turn into a fault |
| Kinds | custody, malformed move, horizon, authority | novelty, procedure |
| Routed to | the parent (discharges its dependency) | an LLM (produces what is missing) |

The dividing question, the same one CPUs use: **can the work resume after
something outside handles it?** If yes, it is an interrupt. If no, it is a fault.

Put another way: **a fault is something the runtime proves broke. An interrupt
is something durable memory does not have**: a contract or a procedure. Both
interrupts route to an LLM, which produces what is missing.

**Failure is not a fault.** A target that could not do its work returns, and its
return is checked against the parent's contract like any other. If the contract
covers that shape, the continuation handles it. If not, it is a novelty
interrupt, which routes to review (an LLM or the operator). There is no separate
failure track.

## Faults: errors the runtime proves

A fault is a terminal state of a Move (D9). It discharges a dependency exactly
like a return, carrying a fault instead of a value. Every fault is something the
runtime observed itself and can vouch for. A target's own claims are never
faults: they are returns.

| Fault | What broke | Examples | Observed by | Name in the minimal shape today |
|---|---|---|---|---|
| **Custody fault** | the protocol | the target died without emitting; the successor died before stamping; a stamp failed verification | the instance still holding custody | "supervisor fault" (`Supervisor: true`) |
| **Malformed move fault** | the bytes | they do not parse as exactly one Move; unknown fields | the successor, at intake | "fault (rejected at intake)", parse errors |
| **Horizon fault** | the authority, attempted | an Invoke names a target not in reach, for example a leaf that invokes | the successor, at intake | "fault (rejected at intake)", "not in reach" |
| **Authority fault** | the authority, insufficient | a child cannot fulfil its contract under its Horizon: "I cannot continue without more authority" | the successor, at intake, which confirms the request is outside the child's reach | not modelled |

**Horizon fault versus authority fault: trying versus asking.** A horizon fault
is a child that *tried* to invoke outside its reach: a violation. An authority
fault is a child that *asks* for authority it lacks: a legitimate request. Both
are terminal for the child and both go to its parent, but the parent can tell a
misbehaving child from one that needs help.

### Authority fault

**Definition.** A child **cannot fulfil its contract under the conditions it was
given**: its Horizon does not let it do what the contract requires. It signals
"I cannot continue without more authority".

**Still runtime-proven.** The child asks, for example with an Invoke to a reserved
`escalate` surface whose arguments name what it needs. Intake confirms the gap:
what it asks for really is outside its `Reach`. The runtime does not take the
child's word; it verifies the claim.

**Returned to the parent.** The child faults, which is terminal, so it does not
sit suspended. The fault discharges its slot, and the parent's continuation
decides:

```text
child: needs X, X not in its reach
  → authority fault: discharges the child's slot in its parent's frame
  → parent's continuation:
      holds X → creates a new invocation whose Horizon includes X (a plan)
      lacks X → default: propagate; the parent faults too, discharging its own parent's slot
  → the fault climbs the stack until an ancestor can act, or it reaches the operator
```

Escalation is ordinary fault propagation, with no special mechanism. **Plans**
enter where an ancestor holds the authority and creates a new Horizon with it.
Plans are otherwise out of scope here.

### A faulted dependency is the parent's view, not a fourth kind

From the parent's side, all four arrive the same way: *my dependency was
discharged with a fault*. The parent's procedure does not depend on the cause.
The record keeps the cause, so a human or a recovering continuation can tell
them apart.

**Procedure when a dependency is discharged with a fault:**

1. **Wait for all siblings anyway** (the join rule). An early fault does not
   cancel the other children, because cancellation does not exist yet.
2. When the last dependency is discharged, **the parent's continuation decides**
   what the fault means: propagate it, or recover with a different plan.
3. **Default until continuations exist: propagate.** The parent faults too,
   discharging its own parent's dependency with a fault that points at its
   cause, and so on up the stack.

**Recorded:** the fault record, pointed to by `frames/<owner>/deps/<index>`.
Propagation writes one fault per level, each pointing at the fault below it, so
the chain from the root down to the original cause stays readable.

## Interrupts: resumable

### Novelty interrupt

**Definition.** Every input contract has an expected return contract in durable
memory: a contract pair, like a function type. A Return whose value does not fit
any return contract recorded for the input contract that produced it is a
**novelty interrupt**: a return the system did not expect.

**Novel is not wrong.** A fault means something failed. A novelty interrupt
means something succeeded with a shape the system has not seen. It is a surprise
to review, not an error to propagate.

**Detected, in two places.** The interrupt is the same; only the contract
differs.

1. **At the child's Return** (the successor's intake): does the value fit the
   return contract for the Invoke it answers?
2. **At the join** (resolving the parent): the children all returned, but do the
   filled values fit what the parent's continuation needs as input? This is "I
   invoked all my children and they did not give me what I needed".

**Procedure: resolve with a review dependency.** No new state is needed. A
novelty interrupt means "I cannot continue until someone decides about this
contract", which is a dependency on a decision:

```text
a child returns a novel shape
  → intake commits the Return, marked novel, and does NOT discharge the parent's slot
  → the parent's Potential goes back to a state of resolve,
    with one more open dependency: review(contract), produced by a contract-review surface
  → review returns:
      accept → the contract pair is recorded in durable memory;
               the slot is discharged with the value; resolve continues
      refuse → the review's refusal is itself a return; the slot is discharged
               with it, and the parent's continuation decides what to do
```

This reuses what already exists: dependencies, resolve and the join rule.
Restore comes free: a Potential in resolve survives a crash untouched, so an
interrupt waiting on an LLM (or a person) can wait indefinitely. "Interrupt" names the event
that puts it there; the state stays resolve.

**Why it matters.** Known shapes take the fast path. Only novel shapes take the
slow path, and each accepted one becomes known. That is "recurrence is the
resource" from the thesis, and it is the structural check Invariant II asks for:
deviation from a typed contract is caught by shape, not only by identity.

**Requirements.**

- **Seeded contracts.** Registration declares each surface's contract pairs,
  as DNA-19 profiles already declare an `OutputContract`. Otherwise every first
  return is novel. Novelty then fires only outside what was declared or learned.
- **Structural fit.** "Fits" means shape (fields and types), not values, and it
  must be decidable purely and deterministically, so replay reproduces the same
  verdict. Echo's return contract is `{value: string}`, whatever the text says.
- **Contracts reach the step through the Horizon.** `step` is pure and cannot
  read memory. Each surface in `Reach` carries its contract pairs, like its path
  and hash. The check stays in intake, and the checker stays the successor.

### Procedure interrupt

**Definition.** A **procedure** is the series of invocations mapped to an input
contract: how to turn an input of that contract into calls on the surfaces the
Horizon reaches. Procedures are keyed by input contract alone and stored in
durable memory.

A **procedure interrupt** fires when **the Horizon exists and a child Potential is
in reach, but the parent has no input contract for that Potential yet.** The
parent may touch the Potential but does not know how to call it. With no input
contract there is no key, so no procedure can be found or built.

**Not an authority fault.** The authority is there: the Potential is in reach.
What is missing is knowledge (how to call it), not permission. That is a memory
miss, so it is an interrupt, not a fault.

**A procedure is not a plan.** A plan is what happens when a Horizon is created:
deciding what an invocation may touch. A procedure works inside a Horizon that
already exists. Plans are out of scope here.

**No approval.** The Horizon already bounds what any procedure can touch, so
authority is settled before the procedure exists. A procedure is only an
arrangement of calls within that authority. Intake still checks it: every
invocation in it must be in reach, or it is an admission fault.

**Procedure: resolve with a procedure dependency.** The same shape as the novelty
interrupt:

```text
a child Potential is in reach, but the parent has no input contract for it
  → the parent's Potential goes to a state of resolve, with one open dependency:
    procedure(child Potential), produced by an LLM
  → the LLM returns the missing input contract, and the procedure keyed by it:
    a series of Invokes, possibly with holes
  → intake checks it: every invoke in reach → otherwise an admission fault
  → the input contract and its procedure are recorded in durable memory
  → its invokes open dependencies, and execution continues as normal
```

The next time the parent reaches that Potential, the input contract and its
procedure are found, and it never interrupts. A recorded procedure is the fast path. A miss is the slow path,
taken once per input contract. This is the thesis's engram, compiled and
memoized.

**Which invocations it applies to.** Only procedure-driven ones, where the
runtime looks up the series. An executable-driven invocation (a target like
Codex that decides its calls as it runs) is its own procedure and has nothing to
miss. The two coexist. The surface that composes a missing procedure is often an
executable-driven agent whose output becomes the stored procedure.

### The two interrupts side by side

Both route to an LLM: an agent produces what durable memory lacks.

| Interrupt | Durable memory is missing | Keyed by | Fires when | The LLM produces |
|---|---|---|---|---|
| **Novelty** | a return contract | the input contract that produced the return | a return has an unfamiliar shape | a verdict: accept records the contract; a refusal is a return the continuation handles |
| **Procedure** | an input contract (and so a procedure) for a reachable Potential | the child Potential | a Potential is in reach but the parent has no input contract for it | the input contract plus a procedure; intake checks it is in reach; both are recorded |

## Mapping to the minimal shape

| Here | Minimal shape today | Change |
|---|---|---|
| Custody fault | `Record.Supervisor = true` | rename the flag to a fault kind: `Kind: custody` |
| Malformed move fault | `State: fault`, `Move == nil`, parse error in `Reason` | `Kind: malformed` |
| Horizon fault | `State: fault`, `Move == nil`, "not in reach" in `Reason` | `Kind: horizon` |
| Target failure | `State: fault`, `Move.Fault` set ("emitted") | not a fault: a return checked against the contract; novel shapes interrupt |
| Authority fault | not modelled | needs a way to ask (e.g. an `escalate` surface) and `Kind: authority` |
| Faulted dependency | not modelled (no dependencies yet) | arrives with `frames/<owner>/deps/<index>` |
| Novelty interrupt | not modelled (no contracts yet) | needs contract pairs in `Surface`, and resolve |
| Procedure interrupt | not modelled (the seed's input contract is hand-written) | needs input contracts per reachable Potential, a procedure store keyed by them, and a non-leaf Horizon |

## Open

1. **Input-contract fault or interrupt.** The counterpart on the way in: an
   Invoke whose arguments fit no input contract for its target. Is it a
   malformed move fault (reject), or a novelty interrupt (ask)? Next to discuss.
2. **The `Fault` Move variant.** If target failures are returns, targets never
   emit `Fault`. It becomes how runtime-proven faults travel up the stack. Should
   intake treat a target-emitted `Fault` as a malformed move?
3. **How an authority request is expressed:** an `escalate` surface every Horizon
   reaches, or something else. Its arguments must name what is needed precisely
   enough for intake to confirm it is outside the child's reach.
4. **Who accepts a novel contract**, and is the review itself a Move (an Invoke
   to a contract-review surface), so it is recorded like everything else? The
   procedure above assumes yes.
5. **Where contract memory lives:** in genesis and registrations only, or in a
   separate content-addressed store that learned contracts are added to.
6. **Plans.** What happens when a Horizon is created. Deliberately out of scope
   here; procedures assume the Horizon already exists.
7. **What an input contract is, concretely,** since both interrupts key on it:
   the target plus the shape of its arguments?
8. **The record for a novel Return:** a field on the Return record (`novel:
   true`), or a separate interrupt record that points at it.
