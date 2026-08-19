# contracts/02 · The Child Contract

**Package:** `contract`

Two contracts per child position, both owned by the parent, because the parent is the behavior
whose future depends on the return.

---

## 1 · The shape

```go
type Child struct {
    Signature MethodSignature   // expectation + ordered param types. §3.
    Return    string            // THE CHILD TYPE. what this parent expects back.
}
```

**Two contracts, not three.**

```text
REVISED 08-15. engram-contracts-and-composition.md's "fundamental shape" lists THREE, adding an
expectation contract tested at composition time. That test was failure kind 1, and
../creator-and-compiler.md §4 removed it — Recognize runs on the RETURN PATH ONLY.

An ExpectationContract with no evaluation point is a type nothing calls. DELETED.
See ../coherence.md §2.1.
```

Nothing else is part of a child contract.

**There is no `Child` field on a `Child`.** The contract describes a *position*, not an edge to a
particular engram:

> *"A contract says what may occupy one position in the parent's continuation. It does not create a
> durable relationship to a particular child."*

A `Child ids.EngramID` field would be the rejected persistent type edge — it would go stale when
either side changed, and would read as evidence that composition succeeded. Which engram occupies
the position is decided per invocation, and recorded by Decide.

---

## 2 · Ownership is structural

The contract is reached only through the parent that owns it:

```go
func (e *Engram) Contract(position string) (*Child, bool)
```

There is no `func (e *Engram) ContractFor(caller ids.EngramID)` and no method on a child returning
what its callers must accept. **The direction of the accessor is the enforcement of P2.** A child
cannot declare what a parent must accept because there is no method through which it could say so.

```text
NO ENGRAM OWNS the contracts by which every possible caller in the system must use it.
The same engram may participate under different parent-owned contracts in different compositions.
```

That is also why age is positional (`../identity/04-position.md`): the same behavior composed by a
different parent under a different contract ages independently.

---

## 3 · The signature, and the contract it is made of

### MethodSignature — the signature IS the contract

> **The signature IS the contract.** Not a thing derived from it, not a thing checked against it.
> The parameter list, in order, with its types, is the whole of what may be bound at this position.

```go
type MethodSignature struct {
    Expectation Expression   // what behavior is wanted at this position
    Params      []string     // type names. ORDERED, and the order is SEMANTIC. see below.
}
```

### Why a slice and not a set — DECIDED 08-15

A set was considered and rejected. Under nominal typing the usual argument favours it: two params
sharing a contract are indistinguishable to the type system, so either they are interchangeable or
the contract is under-specified and should be two contracts. By that argument ordering is
redundant.

**It is not redundant, because a compiler may be an LLM, and an LLM weights earlier context more
heavily than later.** Parameter order is therefore a *salience declaration* by the parent — a
statement about what matters most at this position — not merely a binding mechanism.

```text
ORDER IS NOT INCIDENTAL. IT IS THE PARENT SAYING WHAT TO ATTEND TO FIRST.
```

Three consequences:

```text
1  A REORDER IS A REAL VERSION CHANGE.  ../../../reach/age.md increments on a signature change.
   Under a set that increment would be spurious — nothing observable changed. Under a slice it is
   correct: reordering changes what the compiler attends to, so it changes behavior.

2  Assemble MUST PRESERVE ORDER.  durable/03-crossing.md has the entity's own marshalling build
   what it reads. If Assemble reorders params for convenience, it silently destroys a parent's
   declaration. This is a constraint on the crossing, not a style note.

3  EVERY READER OF A SIGNATURE DEPENDS ON THE ORDER — for one of two reasons, never neither:

       THE AUTHORING MODEL   reads the frame. order is SALIENCE.
       THE COMPILER          compiles the body against it. order is BINDING.

   There is no third kind of reader that could ignore it. A python PROCESS is not one — it
   receives resolved arguments, not a signature, and by then binding has already happened. It
   never sees a MethodSignature at all.
```

**The design already has one order-is-meaning case**, which this parallels exactly —
`../../../continuation-language/primitives.md` §3: *"For a concurrent invocation group, emitted
member order is meaningful and completion order is not."* Emitted order is preserved because it
means something. Parameter order is the second instance of the same rule.

**No parameter names.** A local name is how the *callee's* continuation refers to an observation,
and the continuation is interior — `../../../reach/age.md`: *"the continuation is the part no caller
can see."* `int f(int x)` and `int f(int y)` are the same signature, and so are these.

**`Expectation` is an expression, not a predicate.** This does not resurrect the
`ExpectationContract` deleted in §1 — that had an `Admits(Candidate) bool` method with no
evaluation point. This is `graph.Expectation` as it already exists:

```go
type Expression struct {
    Expression string     // NO LANGUAGE TAG. a tag would be a self-report.
}
```

It states what the position is *for*. Nothing checks it, so nothing needs to call it.

```text
BEARS ON H12   ../../06-entity.md and durable/01-engram.md §2 removed Expectation from the ENGRAM,
               on the P4 grounds that a child stating what it is cannot be typing. Putting it in
               the parent's SIGNATURE resolves that: the expectation of what is wanted at a
               position is the PARENT'S, and a child never describes itself.
               H12 asked whether the engram keeps an expectation. This says NO — the POSITION does.
```

`continuation-language.md` already says this in both directions:

```text
Destination getDestination( TripRequest request );
│           │               │           │
│           │               │           └─ interior. the callee's local name. NOT in the contract.
│           │               └───────────── Params[0]  — a type name, "TripRequest"
│           └───────────────────────────── Expectation — what this position is FOR
└───────────────────────────────────────── Child.Return — a type name, "Destination"

"METHOD SIGNATURE — supplied from the parent-owned input contract for this child position."
        — continuation-language.md, which still uses the older name for it
```

Two consequences follow from being explicit about it.

**1 · It has no runtime predicate.** The old shape here was `InputContract.Permits([]Observation) bool`. There is
no moment at which that could run: `../creator-and-compiler.md` §4 puts Recognize on the return path
only, so nothing checks a binding before the child runs.

It does not need to. Binding is safe **by construction**:

```text
continuation-language.md   every value traces to exactly one causal source —
                           A METHOD PARAMETER, or A CHILD ASSIGNMENT.

BOTH WERE ALREADY CHECKED WHEN THEY ARRIVED.
    a parameter was parsed on the way in. a child return was parsed by this parent's own contract.
    so a continuation can only pass observations that were type-checked on entry.
```

Which is the same guarantee `../../../continuation-language/conditions-and-loops.md` states for
reads: *"Field access cannot fail. The parser runs before any read, and contract satisfaction
guarantees those locations exist. No missing-field case, no null, no defensive read."*

```text
SIGNATURE   checked WHEN THE BODY COMPILES — arity and param contracts must line up.
            continuation-language.md acceptance test 3.
RETURN      checked AT RUNTIME, on the way up. the only role needing a runtime predicate.
```

**2 · A parameter and a return are the same kind of thing.** Both are type names — strings the
substrate holds and compares. `InputContract`, `ReturnContract`, and `ObservationContract` are all
gone; there is a `string` in two positions.

#### What it still does

Both jobs its source document gives it survive the loss of the predicate:

```text
COMPATIBILITY   the proposed child composition must be meaningful for these inputs.
                → now discharged AT COMPILE, by the signature lining up.
CAUSALITY       the substrate records the exact observations actually supplied.
                → unchanged, and never depended on a predicate. this is a RECORDING duty.
```

The second is the one to protect. `observable-causality.md` §5's *there is no uncited observational
influence* rests on the record naming exactly what was bound — which the signature plus the actual
invocation gives without anything being checked at runtime.

**It does not grant the child the parent's frame.** It identifies what may cross this boundary; the
compiled continuation selects a permitted subset per invocation. The distinction matters for
`../../../causality/observable-causality.md` §5 — *there is no uncited observational influence* — which
depends on frame membership being exact rather than permissive.

**The signature is not Horizon.** Both bound the child and they travel together, but:

```text
SIGNATURE   what observational MATERIAL may be supplied.
HORIZON     how far outward the child may SEE or ACT.
```

Collapsing them produces one of two known defects: authority becoming caller-authored semantic data,
or a second frame. See `../identity/03-horizon.md`.

### The contract is a type name — REVISED 08-15

An earlier pass in this session built `ObservationContract` as an interface with a predicate.
**Deleted.** A contract is a **string**, and the check is a comparison the kernel performs between
two strings it already holds:

```text
Contracts[position].Return    what the parent expects.   substrate-held.
child.Returns                 what the child produces.   substrate-held.
```

```text
THE KERNEL NEVER TOUCHES CONTENT TO TYPE-CHECK. it compares two declarations it holds.
```

**And this does not breach P4.** The label lives on the **engram** — durable, substrate-held — not
written into observation content by the child. A child cannot forge what it does not author, which
is the distinction P4 is actually about.

`Recognize` is a kernel method. See [`../kernel/07-recognize-orient.md`](../kernel/07-recognize-orient.md).

---

## 4 · Satisfaction is contextual

```text
Y ⊨ X

Y   the returned observation, with substrate-known provenance
X   the childType string this parent stores for this position
⊨   Y is admissible to the continuation governed by X
```

Both sides are substrate-held, which is what keeps the comparison honest:

```text
WHAT THE PARENT EXPECTS   Contracts[position].Return   on the parent engram.
WHAT THE CHILD PRODUCES   child.Returns                on the child engram.
NEITHER IS IN CONTENT     so neither is forgeable by whatever wrote the observation.
```

The same observation may satisfy a contract under one parent and fail under another. **A type name
therefore creates no global class hierarchy.** `continuation-language.md` states this as the reason
the familiar syntax does not imply familiar semantics: `Destination` names a contract meaningful in
the continuation that owns it, not a class.

This preserves both halves of typing:

```text
NOMINAL GROUNDING   the observation came from known behavior.
CONTEXTUAL USE      this parent accepts it in this position.
```

---

## 5 · What is deliberately absent

```text
NO ContractID          a contract is identified by its Position — (owning parent, position name).
                       a separate ID would be a second name for one thing.

NO CONTRACT IN THE     P11 and durable/02-graph.md. the graph records what composition HAPPENED.
DURABLE GRAPH          a stored contract edge would say only that two things MIGHT compose.

NO Error IN ANY         P5, P15.
SIGNATURE HERE

NO SELF-DESCRIPTION     no String(), no Describe(), no Explain(). the kernel holds contracts and
METHODS                 must not be able to read them into a log — that is exactly the breach
                        ../../../semantics/the-crossing.md identifies in shortExpr and oneLine.
```

The last one is a real constraint on the interface design and not a stylistic preference. Whatever
concrete form a contract takes, **the kernel must be able to hold it and invoke it without being
able to render it.**

---

## 6 · Open

Carried from `engram-contracts-and-composition.md`, unresolved and not resolved here:

```text
the vocabulary by which a parent contract identifies accepted observation provenance
whether contracts admit one producer, a bounded set, or a composable relation among producers
whether one child position may accept several return contracts
how contract revision behaves while a possibility remains live
```

```text
REMOVED FROM THIS LIST, 08-15:

"how structural content requirements combine with provenance requirements" — this was the tension
that decided whether the parser could run in the kernel at all. It is answered by
../primitives-as-engrams.md §5: NOMINAL checks are kernel-owned and read only Provenance;
STRUCTURAL checks are performed by an engram handed the primitive checkers it may use. The two
compose without the kernel reading content, so they no longer need combining rules.
```
