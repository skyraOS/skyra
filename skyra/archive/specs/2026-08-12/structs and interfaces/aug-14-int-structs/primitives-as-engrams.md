# Primitives as engrams, and Recognize as compile

**Date:** 2026-08-15
**Status: CONVERSATION RECORD. UNRATIFIED.** A working session, written down before it is lost.
Nothing here is decided, and nothing here has been checked against the code beyond the citations
given. It is parked while H16 is worked.

> **Everything below came out of asking what `Shape()` is. It ends somewhere else: a type system
> whose primitives are engrams, whose recursion is floored by provenance, and whose checking never
> requires the kernel to read content.**

---

## 0 · What this replaces

If this is ratified, it **substantially resolves H1** and changes three documents:

```text
H1                       stops being "may the kernel read content." the kernel never reads it.
contracts/03-parser.md   §3's three options A/B/C are superseded by the injection model, §2.
kernel/07-recognize-…    Recognize gains a signature and a stated internal split.
open-questions.md        H1 downgrades; a new hole appears (§7).
```

---

## 1 · Kind and type collapse into one word

**Michael's call, and it stands.** Go does not distinguish primitive types from named types, and
neither should the continuation language. `str` `num` `bool` are types alongside `turn-tv-on`.
`continuation-language.md` already wants this — *"the syntax should feel familiar because novelty
belongs in the semantics, not punctuation."*

What was called kind-vs-type is really **nominal vs structural**, which is a distinction ordinary
languages do make and do not need two words for:

```text
STRUCTURAL   satisfied by SHAPE.    any content with these paths and kinds passes.
NOMINAL      satisfied by ORIGIN.   only observations the substrate saw come from this behavior.
```

The distinction cannot be dropped, only renamed. It is the delegation/injection boundary from
`../02-content.md` — *"same bytes, different provenance"* — and a purely structural `turn-tv-on`
would be forgeable by anything that composed the right fields.

```text
BUILT-IN      str · num · bool · list · obj     STRUCTURAL
USER-DEFINED  turn-tv-on · …                    NOMINAL
BOTH ARE CALLED "TYPE".
```

**A type is defined per parent and covers the entire return.** Self-contained:

```text
turn-tv-on {
    power:  bool                  ← structural, primitive
    source: some-other-engram     ← nominal, see §6
}
```

---

## 2 · The kernel passes the primitives in — this is the whole move

```text
recognize(nominal primitives, Obs observation)
```

The kernel does not inspect the observation. It **hands the engram trusted primitive checkers**, and
the engram performs the interpretation using them.

```text
KERNEL SUPPLIES   capability. the trusted primitive types.
ENGRAM PERFORMS   interpretation, using what it was given.
```

This is not a new mechanism. It is `../../semantics/terminology.md`'s SCOPE, one level in:

> *"SCOPE — what may be NAMED... this is the object-capability half: you can select what you can
> name, you can name what you were given."*

**You can check what you were given a checker for.**

So `the-crossing.md` is not bent, amended, or reinterpreted. The kernel never reads content because
the call direction is reversed — it is called *by* nobody and hands *out* capability instead of
reaching in.

---

## 3 · Primitives are leaf engrams, and the recursion is floored by provenance

A primitive type check is a **leaf engram**:

```text
DECIDE        none. a leaf has no children to decide about.
REACH         none. the kernel is the interpreting entity; nothing is contacted.
POSSIBILITY   the execution of the check.
REMEMBER      the return.
ENTITY        the runtime / kernel itself.
```

### The floor

The obvious objection is regress: if checking `num` is an engram invocation, its return needs a
check, which is an engram, forever. It terminates:

```text
type-check engram returns O
    ↓ the parent's contract for that position is NOMINAL
    ↓ "produced by the kernel's check behavior"
    ↓ the check is a FIELD COMPARISON on Provenance, not an invocation
    ↓ TERMINATES IN ONE STEP
```

> **The type check engram does not need a type check.**

Not an exemption — the nominal check meeting its most trusted possible producer. The floor is
derived from provenance, which already exists, rather than asserted.

```text
LOAD-BEARING CONDITION   the nominal check must NOT itself be an engram. if it is, the regress
                         restarts. this is why the split in §5 puts nominal on the kernel.
```

### Why this is cheaper than it first looks

A leaf invocation is not an orchestrator. No Decide, no Reach, no crossing, and — because the kernel
interprets it — **nothing ships to an entity's host**, so the frame-budget concern in
`contracts/04-continuation.md` §3 does not apply at all.

What remains is ephemeral bookkeeping and graph volume, and the second is handled by Remember:

```text
SUCCESSFUL CHECKS   cited by nothing. outside the retained cone. EVAPORATE at closure.
FAILED CHECKS       cited by the repair path. SURVIVE.
```

So the num-check engram does not accumulate a history of every number ever checked. It accumulates a
history of **exactly the positions that keep producing malformed output** — which `Partitions()`
then folds into a live map of where type failures concentrate. The objection inverts into a signal
the design had no other way to collect.

### What it buys

The type check becomes **visible in the causal cone**. A kernel-internal `kindOf` switch is
anonymous — nothing in the record proves a check ran, or which one. As an engram invocation it is
attributable, which is what `../../causality/observable-causality.md` asks the substrate to hold.

---

## 4 · Recognize is compile()

`../../continuation-language/surfaces-and-solidification.md` has this open by name:

> **WHO COMPILES** — *"`coherence/README.md` names a 'continuation compiler' as the owner of invalid
> source programs, but no primitive corresponds to it."*

Recognize is the missing correspondent. Compile-check and return-check are **one predicate over two
subjects**:

```text
RECOGNIZE(contract, CANDIDATE)     → may this occupy the position?    = COMPILE
RECOGNIZE(contract, OBSERVATION)   → does this satisfy the position?  = the return parse
```

P5 survives: those remain two distinct *failures* — incompatible potential vs unsatisfied return —
even as one *mechanism*. The invariant requires the failures be separable, not the machinery.

### Ownership, reconciled rather than contradicted

`primitives.md` calls Recognize kernel-owned; this session has the parent owning it. Both hold once
split, and the split is the document's own formula:

```text
THE KERNEL OWNS   that Recognize runs at every return boundary.   INVOCATION DISCIPLINE.
THE PARENT OWNS   what it checks.                                 THE CONTRACT.
```

P2 independently requires the second half, since Recognize executes the return contract and return
contracts are parent-owned. **The parent engram owns its child's structural types.**

---

## 5 · The two methods beneath Recognize

```text
Recognize                          an engram. composable, learnable, accumulates history.
├── STRUCTURAL   inspect the value's structure          FREE for engrams to negotiate
└── NOMINAL      inspect trusted provenance / identity   KERNEL-OWNED. not an engram. see §3.
```

The asymmetry is not stylistic. Nominal must stay kernel-owned and non-engram or the §3 floor
collapses. Structural can be negotiated because a wrong structural check is a contained failure,
while a forgeable nominal check breaks the injection boundary.

---

## 6 · Nested nominal types require `Ref`, and that has consequences

**Provenance lives on the Observation, not on the Values inside its Content.** A field is a `Value`;
values carry no producer. So a nominally-typed field has nothing to check against — unless it is a
`Ref`:

```text
{source: Ref("obs-7")}
       ↓ deref
   Observation obs-7 → Provenance.Engram == some-other-engram    ← the nominal check
```

Which forces a design consequence:

> **A nominally-typed field must be a `Ref`, not an inline value. Composite types are built from
> references to observations, not from nested objects.**

Arguably correct — a composite built from Refs preserves the causal structure of how its parts were
produced, where inline nesting would flatten it and lose who made what.

But note the asymmetry it creates with `Shape()`, which *refuses* to dereference — *"a nested `Ref`
is a shape element of kind `ref` and is not dereferenced."* So:

```text
STRUCTURAL CHECK   never traverses a Ref. sees `source:ref` and stops.
NOMINAL CHECK      MUST traverse it. one deref per nominally-typed field.
```

**Unresolved: the deref bound.** `Fingerprint` bounds at depth 4 and marks rather than degrading
silently. Nominal nesting needs the same treatment or it is unbounded and potentially cyclic.

---

## 7 · Leaf checking makes the stack safe — and the induction is already written down

Michael's claim: *if you check faithfully at the leaves, the entire stack becomes safe.* It holds,
and both premises are already in the set, in two documents that do not cite each other on this.

```text
surfaces-and-solidification.md   the language has NO ORIGINATION SYNTAX.
                                 "ORCHESTRATOR has a body. every value traces to a parameter or a
                                 child return. LEAF has NO body."

continuation-language.md         NOTHING MAY APPEAR FROM NOWHERE.
                                 "no unbound literals, no ambient graph reads, no hidden entity
                                 memory of this caller."
```

```text
BASE   all novel content enters at leaves. nowhere else can originate.
STEP   an orchestrator emits only what it received or what a child returned.
       it introduces no unchecked material.
∴      check the leaves faithfully → everything above is composed of checked material.
```

Confining origination to the crossing was done for a different stated reason — *"anything a leaf can
write an orchestrator can write"* — but this is the property it bought.

### It terminates at both ends

```text
BOTTOM   leaf return, checked by its parent's contract.
TOP      the outermost return, checked by ACCEPTANCE, not by a contract.
```

`observable-causality.md` §4: *"The outermost caller is the final authority for its branch."* The top
of the stack has no parent to type-check it — the same structural fact as a namespace sitting in no
parent's map (H11). What closes it is a verdict, not a type.

### Leaf checks are not a different mechanism

A leaf's return is checked by its parent's contract like any child's. What is special is that it is
the only place the mechanism faces **novel** material; above it, checking re-verifies composition of
already-checked parts.

```text
CONSEQUENCE WORTH TAKING:  STRICT AT THE LEAVES, CHEAPER ABOVE.
```

### The limit, stated so it is not forgotten

Faithful leaf checking gives **type** safety, not correctness. `{"temp_celsius": 40000}` is a valid
`num` and nonsense. The design already assigns that elsewhere rather than asking the type system to
carry it:

```text
observable-causality.md §3   "Understanding is not a system property."
observable-causality.md §4   "Correctness is local to acceptance."
```

**The type system's job ends at type safety. Correctness is Remember's job.**

---

## 8 · Failure tiering — who got it wrong

A kind mismatch and a composition failure should not both cost an entity crossing. The axis is not
primitive-vs-nominal; it is **who erred**:

```text
THE CHILD PERFORMED WRONG   the composition was right; the producer malformed its output.
                            num where str was required.
                            → the decision STANDS. RE-DRIVE the child. nobody wakes.

THE PARENT CHOSE WRONG      no valid producer, provenance mismatch, unsound composition.
                            → WAKE the parent. it must decide again.

NO COMPOSITION EXISTS       repair over known behavior is exhausted.
                            → ESCALATE TO A CREATOR. authoring is the only move left.
```

This matters because of what leaves are: a leaf's terminal behavior is a model interpreting an
expectation, and **models malform output constantly**. The most frequent runtime type failure in the
system is a correct composition producing slightly-wrong-shaped content. Answering that with a full
wake, a focused frame, and a Think → Decide → Reach cycle spends an entity crossing to conclude
*"you said 12, I needed a string"* — a question with one answer that needed no judgment.

The mechanism half-exists. `../../continuation-language/conditions-and-loops.md`:

```text
LOOP    one position, re-occupied, gated by a condition over the child's return.
RETRY   one position, re-occupied, gated by the repair frame after a parser returns 0.
```

What this adds is a **trigger condition** for retry: a kind mismatch re-drives without waking,
because there is nothing at the parent for a model to decide.

```text
UNRESOLVED   the retry bound. N identical malformed returns must eventually escalate, or it loops.
             continuation-language.md lists "retry and whether it reuses or creates a child position
             occurrence" as open; this gives that question urgency it did not have.

UNRESOLVED   where the escalation line sits. a kind mismatch surviving N retries stops being
             "the child performed wrong" and becomes evidence the composition was wrong.
             the tiering is about the FIRST response, not a permanent classification.
```

---

## 9 · Open edges from this session

Recorded so they are not rediscovered:

```text
1  DOES NOMINAL PIN STRUCTURAL?    two parents may accept the same producer and require different
                                   structures. probably the feature — contextual satisfaction is
                                   already the design's position — but it means a type name has a
                                   global half and a per-parent half that can drift. RAISED, NOT
                                   ANSWERED.

2  THE NOMINAL DEREF BOUND         §6. depth limit, cycle handling, and whether it marks like
                                   Fingerprint does.

3  THE RETRY BOUND                 §8.

4  IS Recognize AN ENGRAM?         this session says yes, which makes it learnable and gives it
                                   history. primitives.md calls it a kernel protocol. §4 reconciles
                                   the OWNERSHIP but not the QUESTION of whether the thing itself
                                   is an engram.

5  EPHEMERAL COST ON THE HOT PATH  §3 argues graph volume is handled by evaporation. per-check
                                   allocation is not addressed and is the one cost that survives.
```

---

## 10 · What is NOT from this session

Marked because the rest of this directory holds itself to it:

```text
CITED FROM SOURCE   the two premises in §7 · the SCOPE quote in §2 · the WHO COMPILES hole in §4 ·
                    the Shape non-deref rule in §6 · the LOOP/RETRY framing in §8 ·
                    the correctness division in §7.

PROPOSED HERE       primitives as leaf engrams · the provenance floor · passing primitives into
                    recognize · Recognize as compile · the structural/nominal split beneath it ·
                    the failure tiering · strict-at-leaves.
```

Nothing in the second column has been checked against the code.
