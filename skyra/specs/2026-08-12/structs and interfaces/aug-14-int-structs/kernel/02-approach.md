# kernel/02 · Approach

**Package:** `kernel`

The outside-in admission boundary. The only way anything enters.

---

## 1 · Not on the possibility-facing interface

```go
// NOT a method on Kernel. Approach is encountered from OUTSIDE.
func (k *Kernel) Approach(a Arrival) (Admission, error)
```

`primitives.md` §1: Approach is *"always at a namespace, and always outside-in."* A possibility
already inside the system has nothing to call it about. Putting it on the `Kernel` interface that a
possibility holds would make the trust boundary reachable from inside it.

> **CONTRADICTED 08-15 — UNRATIFIED.** See [`../approach-and-root.md`](../approach-and-root.md) §9.
> **Approach and sudo are the same thing**: one gate with two uses, admitting an outside party and
> elevating an inside one. Elevation is requested from inside, so Approach must be callable from
> there, and the paragraph above is wrong as written.
>
> What survives: elevation opens a **new branch** with a raised initial Horizon rather than widening
> an existing one — so `CHILD HORIZON ⊆ PARENT HORIZON` is untouched, and the raise is still bounded
> by `grants ⊆ ceiling`. The unification is what keeps one credential endpoint and one trust
> boundary; splitting elevation into its own primitive would need a second holder of the
> `authenticate` endpoint.
>
> The cost is amending `primitives.md` §1's *"always outside-in"* to something like *always crosses
> a trust boundary* — which re-authentication does. That amendment is not made, and the same
> contradiction applies to [`01-protocol.md`](01-protocol.md) §1, which omits Approach from the
> possibility-facing interface for this reason.

---

## 2 · The shape

```go
type Arrival struct {
    Namespace ids.EngramID       // a namespace IS an engram. terminology.md.
    Party     Party              // WHO is reaching.        ← HOLE H2
    With      content.Content    // and with what.
}

type Admission struct {
    Namespace ids.EngramID
    From      From               // the initial From this admission establishes
    Horizon   Horizon            // the initial Horizon
}
```

Approach is **where bootstrap becomes expressible**, and that is why it produces `From` and
`Horizon` rather than receiving them:

> *"`Decide` cannot express the first entry: it is a parent choosing, and the first decision has no
> parent. Approach admits the namespace and establishes the initial `From` and Horizon."*

Every other call in this directory takes `From` and `Horizon` as runtime-supplied. This is the one
place they are created. Once inside, the maximally empty `Think` call asks what potential is
available from that grounded origin.

---

## 3 · The two outcomes

```text
APPROACH   admitted. a namespace opens the possibility, and the observation enters with
           Origin == OriginApproach, because no possibility produced it.

AVOID      refused.
```

`Avoid` is not an error return. It is the complementary outcome, the way `Reject` is the
complement of `Remember`. Whether a refusal is durably recorded, and by which append path, is
explicitly left open by `primitives.md` §1 and is not decided here.

```go
type Admission struct {
    // ...
    Avoided bool     // PLACEHOLDER. see below.
}
```

```text
FOR REVIEW    modelling avoidance as a bool on the success type is almost certainly wrong — it
              makes the refusal a variant of admission. The honest shapes are a sum type or a
              second return, and choosing between them depends on whether avoidance is durably
              appended, which is the open question. LEFT DELIBERATELY UGLY so it is not mistaken
              for settled.
```

---

## 4 · HOLE — `Party`, and the authenticate gap

`Party` above is a placeholder for the one hole both `../../../semantics/terminology.md` and
`2026-08-09/questions.md` §5 identify, from opposite directions:

```text
APPROACH        the primitive. should this be let in?
  authenticate  answers WHO — and this is the half §5 says is missing
  Mount         answers WHAT, already designed: grants ⊆ ceiling
```

> *"The ceiling bounds what an admitted session may do. Nothing bounds who may open one. That is
> `authenticate`, and it is specified here and built nowhere."*

`terminology.md` specified it on 08-11 — *reads the WORLD, WHO IS THIS* — and `primitives.md` does
not mention it. **The missing gate and the unplaced word are one hole with two descriptions.**

Nothing in this outline fills it. `Party` is a named placeholder so the shape of `Arrival` is not
silently written as though the question were settled. Recorded as **HOLE H2**.

---

## 5 · Approach is recursive

It applies at every boundary where something outside contacts something inside:

```text
a human       → approaches the runtime
the runtime   → approaches an entity it composes with
that entity   → is a whole system from inside, one party from outside
```

Which means `Approach` is not a boot-time-only method, and `Arrival` is not a login. The type must
work for a runtime approaching another runtime as well as for a person signing in — that is what
`terminology.md`'s *"nothing crosses a runtime boundary except through Reach"* pairs with on the
inbound side.

---

## 6 · Approach opens a namespace, not a branch

A distinction `terminology.md`'s addendum makes explicitly, and easy to get wrong in types:

```text
APPROACH     outside party → namespace     the admission boundary
BRANCH       one top-level possibility inside that admitted namespace   an execution position
```

They are not aliases. `Admission` therefore carries a namespace and no span — a span appears when
an ask inside that namespace opens a top-level possibility, which is a separate event.

```text
RETIRED, and the reason this matters: "ROOT" implied a special node and was not one. Every wrong
turn — root-as-engram, root-as-namespace, root-as-hinge — came from that implication.
```
