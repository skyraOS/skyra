# kernel/06 · Remember

**Package:** `kernel`

The upward causal form, and the only path to the durable graph.

---

## 1 · One operation at two boundaries

```go
func (k *Kernel) Remember(f Frame, o observation.Observation) error
```

`Remember` is the single operation previously split between handing a result upward and `Accept`.

```text
LOCAL REMEMBER     a parent binds an accepted child observation under a name in its continuation.
                   EPHEMERAL. free in the durable sense.

OUTWARD REMEMBER   an engram presents observations according to its own return contract.
                   still ephemeral.

OUTERMOST          the outer caller remembers a result. THIS closes live execution and makes the
REMEMBRANCE        recursively remembered causal cone durable.
```

The same operation repeats at every level. **One method, and which boundary it is at is a fact
about the frame, not a parameter.** A `Remember(o, durable bool)` would let a possibility ask for
durability, which is not its call to make — durability is a consequence of position in the chain.

```text
leaf observation
      ↓ remembered by its parent
parent observation
      ↓ remembered by its parent
outer observation
      ↓ remembered by the outermost caller
durable causal cone
```

> *"An observation is not durable merely because an inner parent remembered it into a live frame. It
> becomes durable only when an unbroken chain of remembrance reaches the outer boundary."*

---

## 2 · What Remember must preserve

Four things, from `primitives.md` §5, and each is a field somewhere in this surface:

```text
SEQUENCE           order within its possibility.              Provenance.Invocation + ordering
CAUSE              which invocation produced it.              Provenance.Invocation
EMITTED POSITION   its place in a concurrent group.           Provenance.Position
PROVENANCE         the typed distinction between data and     Provenance.Origin
                   delegated reach.
```

The fourth is the one that used to be `content.EngramRef`. See `../contracts/01-observation.md` §4 —
**this is the line in `primitives.md` that most directly authorizes moving reach into provenance**,
because it calls the distinction *provenance* rather than *type*.

`observable-causality.md` adds a longer list of what the live and durable records must preserve
together — expectation, entity, frame membership, selection cause, parentage, sequence, emitted
position, remembrance, revision, outer acceptance. All of it is reachable from the cone Remember
projects; none of it is a separate durable structure.

---

## 3 · Projection, not deletion

```text
EPHEMERAL CAUSAL RECORD
          ↓ outermost Remember projects one retained cone
DURABLE GRAPH
```

**Remember is selective projection, not deletion from an already-written graph.** Work outside the
cone evaporates when live execution closes; it was never written.

This is a genuine departure from the runtime, and the sharpest one in this directory after the
parser question. `../../03-graph.md` records the current behavior plainly:

> *"In the current runtime it is **not the only durable mutation**: `Ingress` and group completion
> call `Store`, closure calls `Mint` or `MintRoot`, effect handling calls `RecordDiscard`, and
> `Apply` stores the returned observation."*

```text
CURRENT     durable writes happen at every commit, from five call sites.
PROPOSED    one durable write, at the outer boundary, projecting a cone.
```

`../../../semantics/README.md` notes a critique that called `Remember` a change to the design and
corrects it: *"it is a change to the CODE, which still writes at every commit."* The decision is
08-11's, from `closure.md` — *closure is the only durable write*.

**This is also why `../../../overnightBuild/README.md` fenced the loop off**: *"DO NOT CHANGE when
durable writes happen. graph.Apply is still called from the commit path."* That work order was not
the place to make this change. This outline is the shape it would take when someone does.

### The backward walk

```text
chosen observation
        ↓ backward causal walk
translations · decisions · remembrances · expectations that contributed
        ↓ projection
durable graph records and materialized reached effects
```

**The kernel does not interpret which content was good.** Each parent chose what to remember, and
the kernel follows those structural causal edges. That is the whole of the policy, and it is why
this operation does not breach P16 — following edges is not reading content.

---

## 4 · Effects

```text
REMEMBER DECIDES which staged effects materialize: effects belonging to reached producers may
land; EVERY OTHER STAGED EFFECT MUST BE EXPLICITLY DISCARDED.
```

"Explicitly" is load-bearing and carries a type with it — `../../03-graph.md`'s `Discard`, which exists
because `kernel.md` demands a discard be a recorded **event** rather than an absence:

```go
type Discard struct {
    Span   ids.SpanID
    Engram ids.EngramID
    Detail string
}
```

Carried forward unchanged. An effect that simply fails to materialize is indistinguishable from one
that never existed, and that ambiguity is what the type refuses.

---

## 5 · Reject

```text
REJECT   the opposite of Remember. Remember selects a causal chain to survive; Reject establishes
         that it does not survive. Both exist; specifying either defines the complementary
         outcome of the other.
```

Not given a signature here. `primitives.md` lists it among words absorbed by the primitives and
defines it only by complement — which is enough to say it exists and not enough to type it. The
question it turns on is the same one `Approach`'s `Avoid` turns on: whether a refusal is durably
appended. Recorded with H8.

---

## 6 · The two laws

`observable-causality.md`, and the reason Remember cannot be the whole story:

```text
OCCURRENCE DOES NOT IMPLY REMEMBRANCE    live work may be ignored and evaporate.
REMEMBRANCE IMPLIES OCCURRENCE           every remembrance must cite a witnessed live event.

REMEMBERED ⊆ OCCURRED
```

P17. The second law is what prevents memory from becoming a self-authored story: a principal may
choose which actual result to remember, and **neither a principal nor an entity may use Remember to
manufacture an occurrence the substrate did not witness.**

It has a consequence the type surface should not paper over:

> *"Causality cannot be created by remembrance; energy spent creating possibilities cannot be
> unspent. Remembrance determines which causal work survives, and efficiency measures the cost of
> that selection."*

The causal proof may project away unremembered work. **Cost accounting may not pretend it never
occurred.** Those are two records with different subjects, and only the first is the graph. What
holds the second is unspecified — see `../../energy/` — and is not proposed here.
