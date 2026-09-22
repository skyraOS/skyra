# kernel/04 · Decide

**Package:** `kernel`

The downward causal form. The choice **is** the binding.

---

## 1 · The signature

```go
func (k *Kernel) Decide(f Frame, c Candidate) (Decision, error)
```

One candidate. P8 — *Decide chooses one complete composition* — is enforced by arity, not by a
check. There is no `Decide([]Candidate)`.

## 2 · What a decision binds

`primitives.md` §3 lists six things, and the type must account for all of them:

```text
ENGRAM        the behavior being opened.
EXPECTATION   the selected candidate behavior, still provisional until retained.
ENTITY        who will interpret it.
INPUTS        observations passed into its frame.
CITATION      which visible need prompted the choice.
HORIZON       a subset of the parent's Horizon, provisioned downward.
```

Five of the six are already inside the `Candidate` or the `Frame`:

```text
ENGRAM · EXPECTATION · ENTITY   in the candidate observation.
INPUTS                          in the candidate, drawn from the frame.
CITATION                        the candidate observation ITSELF is the cite —
                                that is what §5 of 03-think.md establishes.
HORIZON                         narrowed from Frame.Horizon by the kernel.
```

**Which is why the signature has two parameters and not seven.** `engram-contracts-and-composition.md`
makes the argument for the one that would be most tempting to add back:

> *"The reach observation already identifies the entity. A second entity operand would create two
> places for one fact and permit contradiction:*
> `Decide(child, entity A, reachObservation naming entity B)`
> *Removing it makes that disagreement inexpressible."*

The same argument removes the position operand. The current position in the parent's compiled
continuation supplies the child contract. **Position is causal context, not another semantic operand
the entity must declare.**

---

## 3 · `Decision` has no exported fields

```go
type Decision struct {
    // unexported. constructed only by Decide, consumed only by Reach.
}
```

This is how P9 becomes structural. `Reach` may realize *exactly* what Decide selected and may not
substitute another entity, expectation, or input — and the way to make substitution inexpressible is
to leave nothing settable.

```text
NO CONSTRUCTOR      a Decision cannot be assembled by a caller.
NO SETTERS          it cannot be amended between Decide and Reach.
NO PUBLIC READ      what a possibility learns from a decision arrives as a returned observation,
                    not by inspecting the Decision.
```

A `Decision` that could be built by hand would let a possibility request a composition Think never
proposed, which reopens the unattributed-authoring hole from the other end.

---

## 4 · Decide is `SELECT` under a kernel-facing name

`grammar.md`, quoted in `primitives.md` §3:

> *"Select **is** invoke. There is no separate call form, which makes 'selection happens at the call
> site' literal rather than descriptive."*

```text
THE ENTITY EMITS   a select.
THE PRIMITIVE IT   Decide.
DRIVES
```

Whether two names for one act is worth keeping is open in the source and stays open here. In the
type surface it shows up as: `grammar.Select` and `grammar.Member` survive as the **wire form** an
entity emits, and `Decision` is what the kernel makes of one. They are not the same type and should
not be merged.

```go
// carried from ../../04-grammar.md, unchanged in shape
type Select struct{ Members []Member }

type Member struct {
    Engram       string
    Cite         ids.ObservationID
    Entity       ids.EntityID
    Observations []ids.ObservationID
    Host         ids.HostID           // optional; only where the host is the TARGET
}
```

```text
FOR REVIEW    Member.Entity and Member.Cite are the two-places-for-one-fact problem in wire form.
              Under the new model the cite is the candidate observation and the entity is inside
              it, so a Member that named both could contradict the candidate it came from. Either
              the wire form narrows to (candidate, inputs) or the kernel must reject disagreement
              explicitly. NOT DECIDED — recorded as H7.
```

---

## 5 · Concurrent groups

```text
EMITTED ORDER      MEANINGFUL. preserved by the kernel.
COMPLETION ORDER   NOT MEANINGFUL.
```

`primitives.md` §3 and `continuation-language.md` agree, and `observable-causality.md` lists
`EMITTED POSITION` among what the substrate must preserve — *"deterministic position inside a
concurrent group."*

So `Provenance` carries the position (see `../contracts/01-observation.md` §2) and the graph records
it. **Completion order may never rewrite declared order or observation dependencies.**

```text
STILL OPEN, from continuation-language.md: whether independent calls run concurrently must be
EXPLICIT in the eventual language semantics. "Source layout alone must not accidentally make
scheduler timing semantic."
```

---

## 6 · Decide has no failure of its own — REVISED 08-15

This section previously held failure kind 1, *incompatible potential*, checked at composition time.
[`../creator-and-compiler.md`](../creator-and-compiler.md) §4 removed it: **Recognize runs on the
return path only**, so nothing tests a candidate before Decide binds it.

```text
BEFORE                                    AFTER
1  INCOMPATIBLE POTENTIAL  the contract   1  UNREALIZABLE DECISION   REACH
2  UNREALIZABLE DECISION   Reach          2  FAILURE TO COMPILE      THE CHILD
3  UNSATISFIED RETURN      the parent     3  UNSATISFIED RETURN      THE PARENT
```

Decide either forms a composition from the candidate it was given, or the frame offered none. The
distinction this section used to draw — *the composition was never valid* versus *it could not be
made live* — collapses into the second.

The cost is accepted deliberately: an incompatible composition now costs a realization before
anyone notices. That matches the design's posture — *"the system records what happened when an
entity was tried."*

`ExpectationContract` was deleted with this check. See
[`../contracts/02-child-contract.md`](../contracts/02-child-contract.md) §1 and
[`../coherence.md`](../coherence.md) §2.1.
