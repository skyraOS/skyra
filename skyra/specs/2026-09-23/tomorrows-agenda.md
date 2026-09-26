# Tomorrow's agenda: semantics first

Date: 2026-09-25 (for the next session)  
Status: agenda. Produces `transitions.md`.

## The governing question

> **What is the complete set of valid executions this machine admits?**

Not: "What sequence of handlers should we call?" Handlers, reductions and I/O
buckets come **after** the semantics, and exist only to make them true.

## Keep these distinct

| | Answers | Example (Return) |
|---|---|---|
| **Semantic postcondition** | what must become true | the owner's dependency is discharged, once |
| **Completion witness** | what durable evidence proves it became true | `frames/<owner>/deps/<i>` exists and names this record |
| **Reduction / effect** | what operation helps make it true | `Claim(slot)` |
| **OS enforcement** | what prevents the runtime from violating it | `O_EXCL` on the slot file |

**Rule:** a semantic rule never claims something the machine does not enforce.
Every invariant has an enforcement entry, or is marked **not enforced yet**.

## The transitions

1. **Invoke**
2. **Return**
3. **Interrupt**
4. **Fault**
5. **Resolve**: resolve → invoke when the last dependency is discharged. This is
   where the join completes, the completion hash forms, and an interrupt resumes.

## The checklist, for each transition, in order

### 1. Validity

- What must already be true before this transition is legal?
- What inputs does it consume? Include **which memory root** it reads, if it
  checks contracts.
- **Who initiates it, and who claims it?** These are always different parties,
  because the emitter is never the checker. State both.

### 2. Meaning

- What semantic fact becomes true when it succeeds?
- What does it mean for it to be complete?
- What authority is **created, narrowed, preserved, transferred or terminated**?

### 3. Cardinality

- How many times may it occur for the same frame, slot, dependency or Potential?
- What must be unique?
- Which transitions are mutually exclusive?

### 4. Durability

- What facts must be durable before it can be considered complete?
- What durable evidence proves the semantic postcondition occurred (the
  completion witness)?
- If the process crashes immediately afterwards, **can the machine determine the
  one legal next state?**

### 5. Nondeterminism

- Which inputs are nondeterministic (model output, time, randomness)?
- Where is each recorded before it is used, so replay reproduces the transition?

### 6. Invalid executions

- In what states is it illegal?
- What malformed or contradictory histories must be impossible?
- Does an invalid case become a Fault, an Interrupt, a rejection, or something
  else?

### 7. Enforcement

For every invariant above:

- What concrete mechanism enforces it?
- Which layer is it: **semantic, storage, capability or OS**?
- Is it **prevented** (the violation cannot happen) or only **detected** (caught
  afterwards, for example by replay)? DNA-19's lesson: detection is not
  prevention.
- If nothing enforces it, mark it **not enforced yet**.

## Vocabulary discipline

Keep these terms separate, and use each for one thing only.

| Term | Means |
|---|---|
| **Transition** | Invoke, Return, Interrupt, Fault or Resolve |
| **Transition instance** | one bounded occurrence of a transition |
| **Potential** | a declared address to an executable |
| **Frame** | *to define:* the durable node for one invocation (its directory, its slots). Decide how it relates to Potential |
| **Continuation** | a durable description of how suspended work resumes |
| **Dependency** | outstanding work required before a Potential may resolve |
| **Discharge** | closes a dependency. Replaces "settle", which the code still uses |
| **Terminate** | ends authority. Never "discharge" |
| **Complete** | a transition instance has its completion witness |
| **Authority** | what the computation may do: its Horizon |
| **Horizon** | a frame's authority: surface, reach, and the node fd |
| **Reach** | the surfaces a Horizon may invoke |
| **Release** | the parent issuing a child's Horizon |
| **Custody** | who currently owns responsibility for the Move |
| **Stamp** | the hash of a committed record; the receipt for custody |
| **Effect / reduction** | a lower-level operation used to realize a transition |
| **Handler** | a fixed executor of effects; makes no semantic decisions |

## Cross-layer check

- For each semantic rule: **what concrete runtime or OS mechanism makes it true?**
- For each runtime mechanism: **which semantic invariant does it exist to
  enforce?**

Anything with no semantic justification is a candidate for removal. Anything
with no enforcement mechanism is a gap.

## Order

1. **Invoke** first. Its validity section forces the definition of the **input
   contract**, the most load-bearing undefined term.
2. **Return** and **Fault**: most of the material is in [`faults.md`](faults.md)
   and [`thoughts/13`](stateless-controller/thoughts/13-dependencies.md).
3. **Interrupt**: novelty and procedure.
4. **Resolve**: the join, the completion hash, resuming an interrupt.
5. **Then the cross-layer check**, against the existing mechanisms (custody,
   handlers, [`io-buckets.md`](io-buckets.md)).

## Starting drafts

What we already believe, per transition. Everything here is a draft to confirm
or correct, not a decision.

### 1. Invoke

- **Validity:** the invoker holds custody; the target is in its reach; an input
  contract exists for the target, and the arguments fit it. Initiated by the
  invoker's target (it emits the Invoke); claimed by the successor at intake.
- **Meaning:** a new Potential exists; the invoker opens one dependency for it.
  Authority is **narrowed**: the child's Horizon is a subset of the invoker's,
  issued only by the parent.
- **Cardinality:** one dependency per invoke; one child frame per dependency.
- **Durability:** the Invoke record and the opened dependency, in one commit.
  Witness: the child's stamp, verified, and its Horizon released.
- **Nondeterminism:** the arguments may come from model output; recorded as the
  intake input before use.
- **Invalid:** invoking outside reach (horizon fault); arguments that do not
  parse (malformed move fault); no input contract for a reachable target
  (procedure interrupt); a child running with no slot waiting for it; a child
  Horizon wider than its parent's.
- **Forces:** what exactly an input contract is (target plus JSON-types args
  schema?) and what "fit" admits.

### 2. Return

- **Validity:** the Move is in a state of invoke; all its own dependencies are
  discharged (join rule); the value fits a return contract for its input contract,
  under the memory root it read. Initiated by the target; claimed by the
  successor.
- **Meaning:** the owner's dependency is discharged with the value. The child's
  authority **terminates**.
- **Cardinality:** at most one of Return or Fault per slot, ever. Mutually
  exclusive with Fault.
- **Durability:** the Return record; the slot claim. Witness: the slot file names
  this record.
- **Invalid:** returning with open dependencies; a second discharge of the same
  slot; a value with no fitting return contract (novelty interrupt, not a
  return).

### 3. Interrupt

- **Validity:** durable memory lacks something the step needs (a return contract,
  or an input contract for a reachable Potential), at the memory root read.
- **Meaning:** the Potential goes back to resolve with one more open dependency
  (the LLM's answer). The original slot stays open. Authority is **preserved**.
- **Cardinality:** several per Potential are possible (for example a novelty
  interrupt, then a procedure interrupt). Never discharges the original slot.
- **Durability:** the interrupt record and the new dependency; later, the learned
  contract and a new memory root. Witness: the new dependency exists and is open.
- **Nondeterminism:** the LLM's answer, recorded as the intake input of the step
  that receives it.
- **Invalid:** an interrupt that discharges the original slot; one that widens
  authority (that is an authority fault, to the parent); resolving without
  recording what was learned.

### 4. Fault

- **Validity:** the runtime **proves** something broke: custody, malformed move,
  horizon, or authority. Claimed by whoever observed it (the custody holder, or
  the successor at intake).
- **Meaning:** the owner's dependency is discharged with the fault. Authority
  **terminates**.
- **Cardinality:** at most one of Return or Fault per slot. Mutually exclusive
  with Return.
- **Durability:** the fault record, naming its kind and its cause; the slot
  claim. Witness: the slot file names the fault record.
- **Invalid:** a fault that is only a target's claim (a target's failure is a
  return); a fault with no cause recorded; a fault that cancels siblings (wait
  for all).

### 5. Resolve

- **Validity:** every dependency of the Potential is discharged; the fan-in claim
  is won. Inputs: the discharged values or faults, and the continuation.
- **Meaning:** the Potential enters a state of invoke with its holes filled;
  custody is taken. Authority is **preserved**.
- **Cardinality:** exactly once per Potential per wait. The fan-in claim is the
  uniqueness point.
- **Durability:** the completion record, carrying every discharge hash (the
  bottom-up Merkle link). Witness: the completion record and the fan-in claim.
- **Invalid:** resolving with any dependency open (it cannot be formed: a hash is
  missing); two resolves of the same Potential.

## Done when

- All five transitions have the checklist filled in.
- Semantic postconditions and completion witnesses are separated.
- Ownership (who initiates, who claims) and multiplicity are explicit.
- Every invariant has an enforcement mechanism, its layer, and prevented or
  detected, or is marked **not enforced yet**.
- The input contract is defined concretely.
- Crash behaviour is defined at every transition boundary.
- The vocabulary table has no term used for two things, and "frame" is defined.
- There is a list of mechanisms that justify no semantic invariant.
- There is a list of semantic invariants that currently have no enforcement
  mechanism.
