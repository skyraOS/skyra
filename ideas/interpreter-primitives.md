# Interpreter primitives: try, remember, dream, undo (draft)

Date: 2026-10-22. Status: thinking, not agreed. Builds on `novelty-by-type.md`
(novelty is a miss on the type), next/32 (procedures, refinements, remember),
`DELEGATION.md` (delegate), and the three invariants in the README.

## The claim

> The kernel holds a few guarantees: `narrow()` and `read()` (plus the atomic
> commit every step already does). On top of them, the interpreter (the
> procedure actor) holds four primitives, and every one of them returns an
> observation, never asserted state:
>
> | Primitive  | What it does                                                  | Who may hold it        |
> |------------|---------------------------------------------------------------|------------------------|
> | `try()`    | start a new strand that may fail and may or may not be kept   | agent                  |
> | `remember()` | keep a strand as a procedure: decide what matters           | human (see below)      |
> | `dream()`  | compress what is remembered into fewer, more general procedures | agent proposes, kernel checks |
> | `undo()`   | run a procedure that returns state to something equal to before | agent, within evidence |

`try()` is today's `delegate`, renamed globally: it says what delegation is in
this system, the start of a procedure that may or may not be remembered.

The layering is a microkernel's: the kernel keeps mechanisms, the interpreter
keeps policy. `narrow()` is close to seL4's capability mint, a copy with fewer
rights, never more. A useful test of minimality: every interpreter primitive
must be expressible as kernel guarantees plus commit. If one is not, the kernel
is missing something. `undo()` is the first place to check (below).

## One idea underneath: types with their own equality

Contracts are learned, not placed. They live in durable state, not in the
runtime: a contract is the args of an execve, an actor is the binary. A human
can add a contract to state and the system knows it, but keeping it first-class
in state keeps adaptation open. "Learned" here means no more than that.

A contract plus a procedure that says what counts as equal is a **setoid**: a
shape with an equivalence relation. Properties of a contract are then laws over
that equality:

```text
novelty        x is novel        iff  ∄ seen y. x ≈ y          (equality across inputs)
reversibility  f is reversible   iff  ∃ g. g(f(s)) ≈ s          (equality across time)
idempotency    f is idempotent   iff  f(f(s)) ≈ f(s)            (safe retry after unknown outcome)
```

Both novelty and reversibility are definitions that live on the contract, and
both are binary: equal or not. The procedure carries two things: the steps, and
the definition of equal. That second part is where every assumption lives,
explicit and in durable state ("this counts as reversed if you ignore
timestamps and the transaction log").

These laws are not proven. Like typeclass laws checked by QuickCheck, they are
**held as evidence**: claims that survived every trial so far, under a granted
definition of equal, revoked by one failed trial. That is invariant 3 applied
to effects: truth is derived, not stored. It is also the honest version of
"learning from lived experience": the agent can always say why it believes
something (the trials, the projection, who granted the projection).

### Shared default: not equal

When equality cannot be shown, the answer is "not equal":

- unrecognized input is **novel** → fault routes to the model (cheap);
- an action with no passing inverse is **irreversible** → fault routes to
  whoever holds the authority, usually the human (expensive).

The safe outcome falls out of the procedure failing, not of anyone remembering
to be cautious. The procedure decides equality; the fault kind decides who
handles inequality. This lines up with faults that route by kind.

## undo(): reversibility as a derived fact

### Why it matters

Transactions in databases record the state change. Agent transactions must
also hold what led to it (the provenance chain, already built) and must deal
with effects that leave the machine. Plain reads and writes are the wrong cut;
**reversible vs irreversible** is the right one.

The README's crash semantics already name the gap: after the atomic write,
intent is recorded but an external effect may have an unknown outcome. Internal
records are always recoverable; external effects are not.

### One notion, not two

An action is reversible iff there is a procedure that brings state back to
something that hashes equal to where it started, under that procedure's
definition of equal. No special cases in the runtime; special cases belong in
procedures:

- snapshot restore is just a very general procedure;
- a refund is a procedure whose equality ignores the statement line and
  compares balances. There is no separate "compensable" tier.

### How it is derived

At contract registration (or later), the agent may propose an inverse, through
the same actor or another. The runtime, never the agent (invariant 1), runs
the trial and checks the hashes:

1. Fix the projection (the definition of equal) **before** the trial.
2. Fork the state; never trial against live state. The content-addressed tree
   and per-call sandboxes make this possible for anything local.
3. Run `f`, then the proposed `g`, on the fork.
4. Hash the projection of the result and of the start. Equal → evidence for
   reversibility on this contract and input class. Not equal → no evidence.

Things to plan for:

- **The hash only covers what can be observed.** Effects that leave the
  sandbox (network, a sent message) cannot be hashed back, so the trial never
  passes. Default irreversible; only evidence promotes.
- **Byte equality is too strict.** Mtimes, inode numbers, logs and caches
  differ even when the inverse worked. The projection decides what is hashed.
- **One trial is not proof.** An inverse can pass for the tried arguments and
  fail for others. Reversibility is per contract and input class, evidence
  accumulates, and a failure in practice revokes it.
- **Registration-time reversibility is not run-time reversibility.** A write
  that passed becomes irreversible once another actor reads it (escape; undo
  would cascade). Reversibility can also be time-bounded (undo send,
  authorization windows). The runtime still tracks escape through the slot and
  record chain.

### The projection is the attack surface

A projection that hashes everything to a constant makes every action
reversible. Two guards:

- **Weakening equality is authority.** A projection that ignores more is a
  broader claim. The agent may propose one; accepting it comes from the
  Horizon or the human, like reaching past the boundary.
- **The projection is fixed before the trial,** so no search for whichever
  definition of equal lets the inverse pass.

### How it fits the rest

- **Horizon.** A Horizon says what an agent may invoke; it can also say what
  class of effect it may cause. Reversible effects run freely under the grant;
  irreversible ones count as reaching past the Horizon, the one case that
  already asks the human.
- **Procedures compose.** A remembered strand is irreversible if any step is,
  so self-built tools stay approval-free only inside the reversible zone.
  "Authority only narrows" holds.
- **Replay is not re-execution.** Replay re-derives state from records, which
  is always safe; re-running an irreversible step is not. Irreversible invokes
  likely need an idempotency key in the record, so a retry after an unknown
  outcome is detected rather than doubled.
- **Prepare / commit as slots.** Where an irreversible action can be split
  (draft vs send, card authorization vs capture), the prepare step opens a slot
  and only a separate commit discharges it. The pivot point (sagas, Garcia-Molina
  and Salem 1987: reversible steps before it, the irreversible one at it, only
  steps that cannot fail after it) becomes explicit in the record chain.

### Open: does undo() need the kernel?

Restoring is a write with real effects. If `undo()` is an ordinary invoke under
the current Horizon, it is expressible as kernel guarantees plus commit. If
restoring needs to touch something the original action's Horizon could reach
but the current one cannot, either `undo()` needs kernel support or the inverse
must be granted together with the action.

## remember(): what matters

`remember()` stores what matters, and by default the agent does not hold it:
if the agent decides what matters, every strand looks worth keeping and
procedures sprawl past anyone's ability to audit them.

Roles, as a start and not as hard lines:

- the **human** is responsible for authority and for what matters;
- the **agent** is responsible for novelty (as far as it can go without more
  authority) and for reversibility.

The roles can blur without the authority moving, using invariant 1: the agent
may *propose* "this strand seems worth remembering" (untrusted bytes), and
nothing becomes a procedure until someone with the authority commits it.

Remember is itself an action, and forgetting is its inverse. That turns the
reversibility model back on memory: an agent could remember **provisionally**,
with forgetting as the guaranteed undo, while the human decides what becomes
permanent (see `forgetting.md`).

## dream(): compression as the incentive

`try`, `remember` and `undo` are building blocks for learning. The missing
piece is an incentive to compress.

### The incentive

Minimum description length, as in DreamCoder (Ellis et al. 2021): alternate
waking (solving with procedures) with a sleep phase that refactors solved
programs into a library of abstractions, keeping the library that makes
*library + all past solutions* shortest. Stitch (Bowers et al. 2023) makes the
abstraction step fast. A new procedure earns its place if it shortens how
everything already remembered is expressed. A smaller library also lowers the
cost of remembering more.

### Generics: merging and differentiating, not widening

Several strands that serve the same base condition merge into one generic,
like deduplicating code. The shared shape is found by anti-unification
(Plotkin 1970): the least general pattern covering the examples, differences
pulled out as parameters.

- **Closed generic:** the parameter ranges only over instances already
  observed. Pure refactoring: same behavior, less code. No new claim, no grant.
- **Open generic:** the parameter ranges over anything satisfying the base
  condition. This *is* widening: it covers instances never tried, like a weak
  projection covers states never compared. Opening needs evidence over time or
  a grant through the Horizon.
- **The bound** is the base condition: the shared contract with its equality
  laws.
- **Differentiation** runs the other way: a new strand that fits the generic's
  shape but fails its equality check splits back out as a specialization.

### The kernel is the oracle

The kernel holds all the state, so every committed strand is already a test
case: inputs, outputs, exact sequence, hash-checked. `dream()` needs no
hand-written tests: propose an abstraction, replay the strands it claims to
cover through it in a sandbox, compare under the contract's equality. Pass →
the merge is safe; fail → split.

Two limits:

- **External effects replay as recordings.** Strands that touched the network
  are checked against recorded responses, like fixtures. That proves
  equivalence to what happened, not to what would happen now.
- **History cannot catch overfitting.** An abstraction tuned to every recorded
  strand always passes. To open a generic, hold strands out: build on part of
  the history, test on the rest. Passing held-out strands is the evidence an
  opening grant should rest on.

So the kernel gives both halves: a regression suite for closed generics and a
held-out set for open ones.

## Testing the primitives: ablation through narrow()

Each primitive can be removed with `narrow()` and the agent measured with and
without it. Unlike prompt or config toggles, a narrowed-out primitive is gone,
removed by the same mechanism that enforces security. This plugs into the
research question (human interruptions per task, task success, cost,
agent-built tools run with zero approvals).

- **Test combinations, not just single removals.** The primitives interact:
  provisional `remember()` depends on `undo()`. A small factorial run (each on
  or off) shows interactions that one-at-a-time removal hides.
- **Separate absent from faulting.** An agent that never sees a primitive
  behaves differently from one that sees it and is denied. Measure both; the
  second shows how gracefully it handles reaching past its Horizon.

## Framing note: the banking parallel

Database transactions were formalized using the bank transfer as the defining
example (Gray and others in the 1970s; airline reservations were the other
early driver), then became the foundation of nearly all software. Relational
databases themselves came from Codd's 1970 work on data independence, not from
banking. The accurate parallel: agent transactions also start with payments,
but the general primitive is atomically capturing everything that led to an
action, plus knowing which actions can be undone.

## Open questions

1. Does `undo()` need kernel support, or is the inverse granted with the action?
2. How are projections represented and granted (CEL, like refinements)?
3. What is the input class that reversibility evidence attaches to?
4. Where does escape tracking live: slots, records, or the Horizon?
5. Is idempotency the next law in the same mold, for retries after unknown
   outcomes?
6. What does provisional `remember()` look like, and who promotes it?
7. What is the description-length measure for `dream()` in this runtime
   (records, CEL nodes, procedure count)?
