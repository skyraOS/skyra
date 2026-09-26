# Tomorrow's agenda: semantics first

Date: 2026-09-25 (for the next session)  
Status: agenda. Produces `transitions.md`.

## The question

> What is the complete set of valid executions this machine admits?

Not: "What sequence of handlers should we call?" Handlers, reductions and I/O
buckets come **after** the semantics, and exist only to make them true.

## The four layers, kept separate

| Layer | Answers | Example |
|---|---|---|
| **Semantic transition** | what must become true | "the parent's dependency is discharged, exactly once" |
| **Transition instance** | one bounded execution of it | "R2 discharged frame 1's slot" |
| **Reduction / effects** | the bookkeeping and system operations that make it true | `Claim(slot)`, `Store(record)` |
| **OS enforcement** | the mechanism that guarantees the invariant holds | `O_EXCL`, content hash, `SCM_RIGHTS` fd, Seatbelt |

**Rule:** a semantic rule never claims something the machine does not enforce.
Every invariant gets an enforcement entry, or is listed as **not enforced yet**.

## The transitions

1. **Invoke**
2. **Return**
3. **Interrupt**
4. **Fault**
5. **Resolve** (added: resolve → invoke, when the last dependency is discharged;
   where the join completes, the completion hash forms, and an interrupt resumes)

## The template, for each transition

**What must be true** (the valid transition):

1. **Preconditions:** what facts must already be true.
2. **Inputs:** what data, authority, dependencies or continuation it consumes.
3. **State change:** what becomes true because of it.
4. **Durable facts:** what must be committed so it survives a crash.
5. **Authority change:** preserved, narrowed, transferred, or discharged. Includes
   **who may perform it**.
6. **Completion condition:** what proves it is finished.

**What must not be true** (the complement):

7. **Invalid cases:** states in which it must not be representable or accepted.

**Plus, for every invariant above:**

8. **Enforcement:** the mechanism that makes it hold, or "not enforced yet".

## Order

1. **Invoke** first. Its preconditions and inputs force the definition of the
   **input contract**, the most load-bearing undefined term.
2. **Return** and **Fault**: most of the material is in [`faults.md`](faults.md)
   and [`thoughts/13`](stateless-controller/thoughts/13-dependencies.md).
3. **Interrupt**: novelty and procedure.
4. **Resolve**: the join, the completion hash, resuming an interrupt.
5. **Only then**, map the existing mechanisms underneath (custody, handlers,
   [`io-buckets.md`](io-buckets.md)). Anything with no transition to justify it
   is cut. Any invariant with no mechanism is a listed gap.

## Starting drafts

To speed things up, what we already believe, per transition. Everything here is
a draft to confirm or correct, not a decision.

### 1. Invoke

**Must be true**

- Preconditions: the invoker holds custody; the target is in its Horizon's reach;
  an input contract exists for the target, and the arguments fit it.
- Inputs: the arguments; the invoker's Horizon; the input contract; memory root.
- State change: a new Potential exists; the invoker opens one dependency for it.
- Durable facts: the Invoke record, and the opened dependency, in one commit.
- Authority change: **narrowed**. The child's Horizon is a subset of the
  invoker's (`narrow`). Only the parent issues it.
- Completion: the child's stamp is verified and its Horizon released.

**Must not be true**

- Invoking outside reach (→ horizon fault).
- Arguments that do not parse (→ malformed move fault).
- No input contract for a reachable target (→ procedure interrupt).
- A child running with no dependency slot waiting for it.
- A child Horizon wider than its parent's.

**Questions it forces:** what exactly is an input contract (target plus
JSON-types args schema?); what does "fit" admit.

### 2. Return

**Must be true**

- Preconditions: the Move is in a state of invoke; all its own dependencies are
  discharged (join rule); the value fits a return contract for its input
  contract.
- State change: the owner's dependency is discharged with the value.
- Durable facts: the Return record; the slot claim.
- Authority change: **discharged**. The child's Horizon ends.
- Completion: the slot claim exists (`O_EXCL`) and names this record.

**Must not be true**

- Returning with open dependencies.
- A second return for the same slot.
- A value with no fitting return contract (→ novelty interrupt, not a return).

### 3. Interrupt

**Must be true**

- Preconditions: durable memory lacks something the step needs (a return
  contract, or an input contract for a reachable Potential).
- State change: the Potential goes back to resolve, with one more open
  dependency (the LLM's answer). The original slot stays open.
- Durable facts: the interrupt record; the new dependency; later, the learned
  contract and a new memory root.
- Authority change: **preserved**. No authority is added or removed.
- Completion: the answer discharges the new dependency, and memory has grown.

**Must not be true**

- An interrupt that discharges the original slot.
- An interrupt that widens authority (that is an authority fault, to the parent).
- Resolving without recording what was learned.

### 4. Fault

**Must be true**

- Preconditions: the runtime **proves** something broke (custody, malformed
  move, horizon, authority).
- State change: the owner's dependency is discharged with the fault.
- Durable facts: the fault record, naming its kind and its cause; the slot claim.
- Authority change: **discharged**.
- Completion: the slot claim names the fault record.

**Must not be true**

- A fault that is only a target's claim (a target's failure is a return).
- A fault with no cause recorded.
- A fault that cancels siblings (wait for all).

### 5. Resolve

**Must be true**

- Preconditions: every dependency of the Potential is discharged; the fan-in
  claim is won.
- Inputs: the discharged values or faults; the continuation.
- State change: the Potential enters a state of invoke with its holes filled.
- Durable facts: the completion record, carrying every discharge hash (the
  bottom-up Merkle link).
- Authority change: **preserved**; custody is taken.
- Completion: the completion record is committed, and custody is taken.

**Must not be true**

- Resolving with any dependency open (cannot be formed: a hash is missing).
- Two resolves of the same Potential (the fan-in claim).

## Done when

- `transitions.md` holds all five transitions, each with the template filled in.
- Every invariant has an enforcement entry or a "not enforced yet" gap.
- The input contract is defined concretely.
- There is a list of existing mechanisms with no transition to justify them.

## Before starting

- Today's docs are committed:
  - [`summary.md`](summary.md)
  - [`faults.md`](faults.md)
  - [`durable-storage.md`](durable-storage.md)
  - [`thoughts/13`](stateless-controller/thoughts/13-dependencies.md)
  - the [`io-buckets.md`](io-buckets.md) changes
