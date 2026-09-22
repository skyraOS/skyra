# kernel/07 · Recognize and Orient

**Package:** `kernel`

Two kernel methods with opposite internal structure. One coordinates a composition it does not
perform; the other answers directly.

---

## 1 · The pair

```go
func (k *Kernel) Recognize(p Position, child ids.EngramID) bool
func (k *Kernel) Orient(f Frame) (Topology, error)
```

```text
REVISED 08-15 — RECOGNIZE IS A STRING COMPARISON, AND IT IS A KERNEL METHOD.

The child does a return statement. THE FIRST THING THE KERNEL DOES IS CALL RECOGNIZE.
It already holds both sides:

    parent.Contracts[position].Return    the childType expected.
    child.Returns                        the type produced.

It never touches content to do this. Both strings live on ENGRAMS — durable, substrate-held — so
neither is forgeable by whatever wrote the observation, which is what P4 is actually about.

This supersedes primitives-as-engrams.md §2's `recognize(primitives, observation)` injection model
AND §5's structural/nominal split. Neither is needed if the comparison never reads content.
See ../coherence.md.
```

```text
THINK       PAST             durable graph      what might address this need?
RECOGNIZE   PRESENT × PAST   live × durable     have I encountered this kind of thing?
ORIENT      PRESENT          live kernel        where am I, and what is happening?
```

Three epistemic relationships, and the internal work differs:

```text
THINK       the interface a possibility invokes.
RECOGNIZE   kernel protocol which CALLS entity and engram methods to compose a judgment.
ORIENT      kernel method reporting bounded live topology DIRECTLY.
```

The kernel owns both protocols. **Protocol ownership does not make the kernel a semantic
interpreter** — within Recognize, entities and engrams supply the interpretation. Orient may report
structure directly only because the kernel already owns live topology.

---

## 2 · Recognize is the parser boundary

Its smallest instance is the parent-owned return type boundary, which makes it the busiest path in
the system:

```text
kernel.Recognize(parent contract, returned observation)
        ↓
calls entity and engram methods
        ↓
their composition returns match or mismatch
```

The `bool` return is the same bit `ReturnContract.Parse` produces — see
`../contracts/03-parser.md`. **`Recognize` is how that parse is operationally performed when it
cannot be performed in the kernel.** Which is exactly the tension in that document's §3:

```text
IF PROVENANCE (or shape) SUFFICES     the kernel answers. no crossing. no entity consulted.
IF STRUCTURE MUST BE READ             Recognize crosses to entity and engram methods.
```

**So H1 is not only a question about the parser. It is a question about how often Recognize
crosses**, and therefore about whether solidification delivers its stated payoff. The two are one
decision.

### Richer recognition

Beyond the return boundary, Recognize may translate present evidence against durable experience:

```text
remembered observation
        ↓ kernel.Recognize
translator behavior
        ↓
canonical or class key
        ↓
structural lookup
        ↓
bounded expectation-observations describing applicable known behavior
```

The division is strict, and each part has a different owner:

```text
TRANSLATE   an engram interprets the observation and makes likeness explicit.
MATCH       entity and engram methods compose the comparison under Recognize.
PROPOSE     entity behavior may use Think to return associated expectation-observations.
```

Three strengths, which suggests the `bool` above is too narrow for the richer case:

```text
IDENTICAL   the same canonical material has been seen.
SIMILAR     the same explicit shape or partition has been seen.
FAMILIAR    translator behavior assigns a known class key.
```

```text
FOR REVIEW    the return-boundary instance needs exactly one bit; the richer instance needs a
              strength and possibly candidates. Whether these are one method with a wider return
              or two methods is unresolved. Making it one wide method risks the parser boundary
              paying for the richer case on every single return — which, per §2, is the cost the
              whole design is trying to avoid. RECORDED AS H9.
```

### What Recognize may never do

```text
RECOGNITION PROPOSES BEHAVIOR WHICH MIGHT APPLY NOW. IT DOES NOT CHOOSE OR RUN IT.
```

The caller still Decides, and Reach must still make the selected composition live. And the sharp
one:

> *"Recognition must never substitute an old observation for a fresh invocation merely because the
> inputs look familiar."*

That is a memoization prohibition, and it is worth stating as a type constraint: **nothing Recognize
returns may be an observation from a previous invocation presented as this one's result.** Anything
it proposes joins live causal state as an observation returned by Think — potential unless Decided,
ephemeral unless remembered to the outer boundary.

---

## 3 · Orient reports topology, never meaning

```go
type Topology struct {
    Heads []Head
}

type Head struct {
    Span   ids.SpanID
    Engram ids.EngramID     // AN ID. never the expectation.
    Entity ids.EntityID
    Host   ids.HostID
    Depth  int
    Since  time.Duration    // time in Step. this is what makes STALLED detectable.
}
```

Carried verbatim from `../../../semantics/terminology.md`, which specifies it and gives the reason for
every field. `Engram` being an ID and not the expectation is the whole discipline:

> *"The kernel hands back handles and the reader resolves them."*

The caller dereferences the engram against **their own graph**, which they may read because it is
their namespace. Same move as `Producer` being a field so `ObservationID` stays opaque. It is also
the fix for the two live `NO SEMANTICS` breaches — `shortExpr` and `oneLine` exist because the
kernel wanted to say something readable and had no other way.

### Head is derived, never counted

```text
head  ==  no live children  AND  no return under the current draft
```

Three node states, all inferable from structure, so **no state field is needed**:

```text
EXECUTING   inside Step, no live children.        A HEAD.
WAITING     children are live.                    not a head.
RETURNED    committed and still live, re-drivable. not a head until re-driven.
```

**A counter cannot express this**, and the current runtime's does not: `kernel.go` increments on
entering `open()` and decrements on leaving, so a depth-five chain with one leaf running reports
five — four of which are ancestors sitting in `runGroup`. That makes a stalled head
indistinguishable from a parent legitimately waiting, which is the eviction decision the number
exists for.

### The bound is structural

```text
ONE DESCENT, TWO FILTERS.  it is a TREE, not a stack — cardinality > 1 means several heads at
                           once, so the answer is a SET.

STOP DESCENDING       no live children.
"WHAT AM I DOING"     those leaves with Returned == nil.   THE HEADS.
"WHAT HAS THIS DONE"  those leaves with Returned != nil.   finished, still live, revisable.
```

Descending can only reach descendants, so `DOWN, NEVER UP` is a fact about which way the pointers
point rather than a rule anyone checks. Orient cannot reveal ancestors or siblings merely because
they share a kernel — which preserves the rule that **a possibility cannot observe its own reception
or optimize directly for being cited.**

```text
ORIENT MUST NOT EXPOSE   other expectations · observation content · reception history ·
                         raw kernel control.
```

---

## 4 · Neither opens a child

Stated once for both, because it is the constraint that keeps them out of the causal grammar:

> *"Recognition can produce potential through Think, and Orient can supply bounded structural
> observations to it; neither can open a child itself."*

`Decide` remains the only downward causal form. A `Topology` is material for Think; a recognition is
material for Think; neither is a `Decision` and neither can become one without passing back through
a candidate.

---

## 5 · Reconciliation with the runtime

```text
introspect        →  ORIENT. "reads the KERNEL, what is HAPPENING" carried verbatim across the
                     rename. ../../../semantics/terminology.md calls this pairing "the most durable
                     thing in this section."

kernel.Introspect →  survives as Orient with the Head shape above. the current implementation's
                     depth counter does not survive; see §3.

no counterpart    →  Recognize. it is new in 08-13 and has no 08-11 equivalent.
```
