# Tomorrow's agenda: semantics first

Date: 2026-09-25 (for the next session)  
Status: agenda. Produces `transitions.md`. Revised after
[`thoughts/14`](14-interpreted-not-stated.md).

## The governing question

> **What is the complete set of valid executions this machine admits?**

Not: "What sequence of handlers should we call?" Handlers, reductions and I/O
buckets come **after** the semantics, and exist only to make them true.

## What thought 14 changed

States are **derived, not stated.** An emitter supplies only operands (which
capability, what payload). The runtime derives the state with one pure function,
`interpret(state, output)`. That simplifies this agenda:

| Before | Now |
|---|---|
| Five transitions, each with its own validity | **One function, `interpret`**. Each transition's validity is the predicate under which `interpret` yields it. Defining `interpret` is most of the transition semantics |
| Invoke and Return as separate transitions | **Return is `invoke(reply, payload)`**. They share validity, the contract check and durability. Return differs only in that `reply` is single use and it discharges a dependency instead of opening one |
| "Who initiates, who claims?" per transition | **One answer for all:** the initiator is always some output, the claimer is always the runtime |
| Input contracts and return contracts defined separately | **The reply capability's input contract is the return contract.** Defining the input contract defines both |
| Invalid cases include forged control flow | **Unrepresentable.** The emitter cannot express "I return", "I fault", or a return to someone other than its parent. Invalid cases are now about **operands**: unparseable, out of reach, fits no contract, ambiguous |
| An enforcement entry per derivation | **One mechanism for every derivation:** `interpret` is pure and deterministic, and replay proves it |

**Names are notation.** A state is a predicate over system state (stack,
memory, Horizon, output). The names are labels, like bound variables in a
program. A state is real only if it passes the checklist: a derivation rule,
invariants, and enforcement.

## Keep these distinct

| | Answers | Example (Return) |
|---|---|---|
| **Semantic postcondition** | what must become true | the owner's dependency is discharged, once |
| **Completion witness** | what durable evidence proves it became true | `frames/<owner>/deps/<i>` exists and names this record |
| **Reduction / effect** | what operation helps make it true | `Claim(slot)` |
| **OS enforcement** | what prevents the runtime from violating it | `O_EXCL` on the slot file |

**Rule:** a semantic rule never claims something the machine does not enforce.
Every invariant has an enforcement entry, or is marked **not enforced yet**.

## Order

1. **Define the input contract**: JSON types, and the fit rule. This also
   defines the return contract (the reply capability's input contract).
2. **Define `interpret`** as a total, pure function over (state, output). This
   is the validity section for every transition at once.
3. **Define the reply capability's lifecycle**: issued at release, single use,
   transferable for tail calls, invoked by the runtime to deliver a fault.
4. **Fill in the rest of the checklist per transition**: meaning, cardinality,
   durability, nondeterminism, invalid operands, enforcement. In order: **Invoke**
   (including its reply case, which is Return), then **Fault**, **Interrupt**,
   **Resolve**.
5. **Cross-layer check** against the existing mechanisms (custody, handlers,
   [`io-buckets.md`](../design/io-buckets.md)).

## 1. The input contract

Settle, with examples:

- A contract is a JSON-types schema (see
  [`durable-storage.md`](../design/durable-storage.md)): `object` with `fields` and
  `required`, `array` with `items`, `string`, `number`, `bool`, `null`.
- **Fit** is structural, pure and deterministic. Unknown fields do not fit.
- What an input contract is attached to: a Potential (by name, or by content
  hash?).
- The reply capability's input contract is the return contract of the input
  contract that produced it.

## 2. `interpret`

Draft, to confirm or correct:

```text
interpret(state, output):
  no output                                        → fault (custody)
  not parseable                                    → fault (malformed move)
  names a capability not in reach                  → fault (horizon)
  names a capability in reach, or none (→ reply):
      payload fits that capability's input contract   → invoke, or return if it is reply
      fits nothing known                              → interrupt (novelty)
      fits more than one contract                     → interrupt (novelty): ask, never guess
  a reachable Potential with no input contract     → interrupt (procedure)
```

Questions to settle:

- **The addressing format**: how a continuation's output names a capability,
  and how `reply` is written explicitly. Unaddressed output defaults to `reply`.
- **What "state" contains** exactly: the Horizon (reach, reply), the open
  dependencies, the memory root and its entries.
- Where the authority fault fits: a continuation that cannot fulfil its contract
  under its Horizon.

## 3. The reply capability

- **Issued** at release, as part of the child's Horizon.
- **Single use**: one `O_EXCL` claim on the parent's slot.
- **Default target** for unaddressed output: the least authority available.
- **Transferable** for tail calls. Question: does the invoker lose its own reply
  at that moment?
- **Invoked by the runtime** to deliver a fault on the instance's behalf.

## 4. The checklist, per transition

### Meaning

- What semantic fact becomes true when it succeeds?
- What does it mean for it to be complete?
- What authority is **created, narrowed, preserved, transferred or terminated**?

### Cardinality

- How many times may it occur for the same frame, slot, dependency or Potential?
- What must be unique?
- Which transitions are mutually exclusive?

### Durability

- What facts must be durable before it can be considered complete?
- What durable evidence proves the semantic postcondition occurred (the
  completion witness)?
- If the process crashes immediately afterwards, **can the machine determine the
  one legal next state?**

### Nondeterminism

- Which inputs are nondeterministic (model output, time, randomness)?
- Where is each recorded before it is used, so replay reproduces the transition?

### Invalid operands

- Which operands must `interpret` refuse, and as which fault or interrupt?
- What malformed or contradictory histories must be impossible?

### Enforcement

For every invariant:

- What concrete mechanism enforces it?
- Which layer: **semantic, storage, capability or OS**?
- **Prevented** or only **detected** (for example by replay)? Detection is not
  prevention.
- If nothing enforces it: **not enforced yet**.

## Starting drafts

What we already believe. Drafts to confirm or correct, not decisions.

### Invoke (including Return, its reply case)

- **Derived when:** the output names a capability in reach (or none, meaning
  `reply`), and the payload fits that capability's input contract, under the
  memory root read.
- **Meaning:**
  - *Ordinary capability:* a new Potential exists and the invoker opens one
    dependency for it. Authority is **narrowed**: the child's Horizon is a subset
    of the invoker's, issued only by the parent, and includes a fresh `reply`.
  - *Reply:* the parent's dependency is discharged with the payload. The
    instance's authority **terminates**.
- **Cardinality:** one dependency per ordinary invoke. At most one reply per
  instance, ever (single use). A reply and a fault are mutually exclusive for the
  same slot.
- **Durability:** the record (with its derived state) and, for an ordinary invoke,
  the opened dependency, in one commit. For a reply, the slot claim. Witnesses:
  the child's verified stamp and released Horizon; or the slot file naming this
  record.
- **Join rule:** a reply is derived only when the instance has no open
  dependencies. A reply with open dependencies cannot be formed (a discharge hash
  is missing).
- **Nondeterminism:** payloads may come from model output; recorded as the intake
  input before use.
- **Invalid operands:** out of reach (horizon fault); unparseable (malformed move
  fault); no fitting contract (novelty interrupt); ambiguous fit (novelty
  interrupt).

### Fault

- **Derived when:** the runtime **proves** something broke: custody, malformed
  move, horizon, or authority.
- **Meaning:** the runtime invokes the instance's `reply` on its behalf, with a
  fault payload naming the kind and the cause. The owner's dependency is
  discharged with the fault. Authority **terminates**.
- **Cardinality:** at most one of reply or fault per slot.
- **Durability:** the fault record and the slot claim. Witness: the slot file
  names the fault record.
- **Invalid:** a fault with no cause recorded; a fault that cancels siblings
  (wait for all). A target cannot emit a fault at all.

### Interrupt

- **Derived when:** memory lacks what `interpret` needed: a return contract
  (novelty), or an input contract for a reachable Potential (procedure).
- **Meaning:** the Potential goes back to resolve, with one more open dependency
  (the LLM's answer). The original slot stays open. Authority is **preserved**.
- **Cardinality:** several per Potential are possible. Never discharges the
  original slot.
- **Durability:** the interrupt record and the new dependency; later, the learned
  contract and a new memory root. Witness: the new dependency exists and is open.
- **Nondeterminism:** the LLM's answer, recorded as the intake input of the step
  that receives it.
- **Invalid:** an interrupt that discharges the original slot; one that widens
  authority (that is an authority fault, to the parent).

### Resolve

- **Derived when:** the stack shows every dependency of the Potential discharged,
  and the fan-in claim is won.
- **Meaning:** the Potential enters a state of invoke with its holes filled;
  custody is taken. Authority is **preserved**.
- **Cardinality:** exactly once per wait. The fan-in claim is the uniqueness
  point.
- **Durability:** the completion record, carrying every discharge hash (the
  bottom-up Merkle link). Witness: the completion record and the fan-in claim.
- **Invalid:** resolving with any dependency open (cannot be formed); two
  resolves of the same Potential.

## Vocabulary

Names are notation; the semantics are what matter. Use each term for one thing.

| Term | Means |
|---|---|
| **Transition** | a derived state change: invoke (with its reply case), fault, interrupt, resolve |
| **Transition instance** | one bounded occurrence of a transition |
| **`interpret`** | the pure function that derives a transition from (state, output) |
| **Operands** | what an emitter supplies: which capability (optional), and a payload |
| **Reply capability** | the single-use handle to the parent's slot; invoking it is a return |
| **Potential** | a declared address to an executable |
| **Frame** | *to define:* the durable node for one invocation (its directory, its slots) |
| **Continuation** | a durable description of how suspended work resumes |
| **Dependency** | outstanding work required before a Potential may resolve |
| **Discharge** | closes a dependency (replaces "settle", which the code still uses) |
| **Terminate** | ends authority. Never "discharge" |
| **Complete** | a transition instance has its completion witness |
| **Authority / Horizon** | what a frame may do: surface, reach (including `reply`), and the node fd |
| **Reach** | the capabilities a Horizon may invoke |
| **Release** | the parent issuing a child's Horizon |
| **Custody** | who currently owns responsibility for the Move |
| **Stamp** | the hash of a committed record; the receipt for custody |
| **Effect / reduction** | a lower-level operation used to realize a transition |
| **Handler** | a fixed executor of effects; makes no semantic decisions |

## Done when

- The input contract, and so the return contract, is defined concretely.
- `interpret` is written as a total, pure function, with the addressing format
  and the contents of "state" settled.
- The reply capability's lifecycle is defined.
- Every transition has the checklist filled in, with postconditions and
  completion witnesses separated.
- Every invariant has an enforcement mechanism, its layer, and prevented or
  detected, or is marked **not enforced yet**.
- Crash behaviour is defined at every transition boundary.
- "Frame" is defined, and no term is used for two things.
- There is a list of mechanisms that justify no semantic invariant, and a list of
  invariants that currently have no enforcement.
