# The entity crossing

**Date:** 2026-08-15
**Status: PROPOSAL.**

> **The kernel must HOLD the crossing. It still may not READ it.**
> Four methods, two opaque types, and one string that is now safe to be a string.

---

## 1 · `Interpreter`

```go
package entity

type Interpreter interface {
	// ASSEMBLE — engram state → what this entity reads. Bounded by the ENTITY: how much it can
	// take. THE KERNEL CALLS IT, so the kernel holds both the full form and the delta form, and
	// can re-assemble in full after a host dies.
	Assemble(d graph.Draft, f frame.Frame) (Rendering, error)

	// PROJECT — rendering → the host's medium. Bounded by the HOST: what form it can realize.
	// MUST FAIL LOUDLY. An unsupported modality must ERROR, never silently drop — otherwise the
	// invocation "succeeds", nobody perceived anything, and the record reads as success.
	Project(r Rendering, h ids.HostID) (Delivery, error)

	// STEP — still the only method that PRODUCES anything.
	Step(d Delivery) (Raw, error)

	// MONITOR — the delta form of assemble. RETURNS NOTHING, by constraint: a return value would
	// be a third emission channel outside the two forms.
	Monitor(o *graph.Observation)
}

// Opaque. The kernel holds them and cannot look inside. Each adapter defines its own unexported
// implementation. A `string` would let the kernel read it, and it would.
type Rendering interface {
	isRendering()
	// Comparing two opaque things needs a predicate from whoever owns them. The kernel supplies
	// BOTH SIDES and chooses WHEN — see §3.
	Equal(other Rendering) bool
}

type Delivery interface{ isDelivery() }

type Raw string // the entity's response, unparsed. THE KERNEL PARSES IT.
```

```text
SOURCE   the-crossing.md, which specifies all four and both opaque types;
         type-surface-proposals.md B2 for Raw; world.md's loud-projection requirement.
```

**Marshalling stays entity-side.** *Marshalling is what interpreting is* — a model's assemble builds
a prompt, python's builds stdin and argv, a human's builds a message. What moved is **who calls
it**, and therefore who holds the result.

---

## 2 · Why `Raw` can be a string now

`overnightBuild/stops.md` E1/E2 stopped on exactly this: `Raw string` + kernel-side parsing means
every emission must be expressible as text, and `content.EngramRef` is not — three resolvers commit
one, the grammar has no surface form for one, and inventing one puts reach in the data channel,
which inverts the delegation/injection invariant.

Under decision D3 (02 §5) **no entity ever writes reach.** Delegation is provisioned by the parent
at `Decide`, from Horizon, downward. So:

```text
THE THREE EXITS stops.md COULD SEE          WHAT D3 DOES
────────────────────────────────────────────────────────────────────────────────────
a  the grammar gains an EngramRef form      not needed. nothing emits one.
b  Raw stops being text                     not needed. Raw stays `string`.
c  code entities skip the parse path        REFUSED, correctly — two classes of entity
                                            would break the kernel's hold on what produced
                                            a decision. Now nothing requires it.
```

The stop is not worked around. **Its cause is removed**, and the removal was already implied by
`primitives.md` §3 folding provisioning into Decide.

---

## 3 · The monitor invariant, checked by the kernel

```text
assemble(frame) == assemble(frame₀) + monitor*
```

The kernel holds one `Rendering` per step. To compare it against anything it needs a second, and
`Monitor` returns nothing — so the second comes from **a second interpreter**.

```text
I₁   the live interpreter. has been fed Monitor(o) for every return.
     Assemble(d, f_now)  →  R_delta        the cheap path

I₂   a FRESH interpreter. has been fed nothing.
     Assemble(d, f_now)  →  R_full         the reference path

     R_delta.Equal(R_full)
```

Both sides are objects **the kernel constructed, from inputs the kernel chose, at a moment the
kernel picked.** That is what makes it not a self-report. The entity supplies only the equality
predicate, and it must, because only the entity knows what its own rendering means.

```text
THE REGISTRY ALREADY SUPPORTS THIS   it stores `func() Interpreter`, so a second costs nothing
                                     structural.

RESIDUAL, STATED HONESTLY            a broken adapter can return true from Equal. That is not
                                     fixed and is not what this fixes. What is fixed is that the
                                     entity no longer chooses WHAT is compared or WHEN — which is
                                     what v5's codex adapter does today, holding base, delta and
                                     Divergent itself.

WHEN TO RUN IT                       behind a flag, or in tests. A second full assemble per step
                                     is not free, and for a model adapter it is not free at all.
```

---

## 4 · `Registry`, and the live bug it currently holds

```go
type Registry struct {
	interpreters map[ids.EntityID]func() Interpreter
}

func (r *Registry) Open(id ids.EntityID) (Interpreter, error)
```

```text
THE BUG   kernel.go:173 passes k.Reg.Known() as a chain's reach — EVERY PRINCIPAL CAN NAME EVERY
          REGISTERED ENTITY. Reach must come from provisioning, never from the registry.

THE FIX   Horizon.CanReach is consulted before Open. The registry becomes what it should have
          been: HOW the kernel reaches an entity it was already authorized to name.
```

```text
SOURCE   type-surface-proposals.md A4, which identifies this as the live bug the permission model
         closes.
```

An entity is **an identity and nothing else** — every declaration dissolved, because a declaration
is made by whoever registered the entity, not by the entity. `type Entity struct{ ID EntityID }`
was proposed in `interfaces.md` and never implemented; `01-removals.md` confirms it should not be.
The ID is the whole type.

---

## 5 · `Events` — the sink, and the two lines it retires

```go
// Handed to the kernel. Never reaches into state. NIL-SAFE: tests pass nothing and nothing traces.
type Events interface {
	Opened(span ids.SpanID, e ids.EngramID, ent ids.EntityID, depth int)
	Crossed(span ids.SpanID, e ids.EngramID, ent ids.EntityID, r Rendering, raw Raw, err error)
	Emitted(span ids.SpanID, e ids.EngramID, em grammar.Emission)
	Closed(span ids.SpanID, e ids.EngramID, out ids.ObservationID)
	Refused(span ids.SpanID, e ids.EngramID, why error)
}
```

```text
span ON EVERY METHOD   the correlation key the current trace lacks. Concurrent possibilities on one
                       engram are indistinguishable without it — three independent handles with no
                       shared schema, and nothing joins them.

WHAT IT DELETES        kernel.go's indent · shortExpr · oneLine, and nine Logf calls.
                       shortExpr reads Expectation.Expression; oneLine reads Observation.Content.
                       THEY ARE THE ONLY TWO PLACES THE KERNEL READS CONTENT — stops.md E3
                       confirms it — and they exist only because the kernel wanted to say
                       something readable and had no other channel.
```

`E3` stopped because `Crossed` needs `Rendering` and `Raw`, which were E1's types. With §1 landed,
the signature is writable as specified.

```text
THE TRACE IS NOT PART OF THE DESIGN.   It is an operational facility, out of band, unavailable to
                                       anything inside the model. Named so nobody later mistakes
                                       it for structure and builds a dashboard out of engrams.
```

> **QUESTION 6.1 — `Emitted` takes a `grammar.Emission`, and I have not typed the grammar.**
> Under a compiled continuation there may be no emission at all on the happy path: the kernel
> drives `Steps` and parses returns itself. So the grammar is the **primitive-mode** vocabulary
> only, and `Emitted` fires in one of two modes. Either the sink takes both shapes, or the grammar
> survives as the universal wire form and the continuation compiles *to* it. **Not taken.** It is
> really F1 (what primitive mode is) wearing a trace-shaped hat. → 09 F1.

---

## 6 · The five entity criteria are a definition, not a type

Recorded because it is tempting to make them fields, and every one of them would be a declaration.

```text
1  UNIQUE                 globally.                          knowable at first contact
2  INDEPENDENT EXISTENCE  it can fail without taking you      REQUIRES A FAILURE, AND REQUIRES
                          with it.                            SURVIVING IT.
3  TAKES AN EXPRESSION    polymorphic, declares nothing.      knowable at first contact
4  RETURNS SOMETHING      evidence comes back.                knowable at first contact
5  THROUGH THE HISTORY    with its observer/parent.           accumulates
```

Criterion 2 cannot be established by asking or by one successful call, which is precisely why it is
not a field. It is observable — *two entities that always fail together share a failure domain* —
so co-location is **learned from correlated failure** rather than declared.

```text
SOURCE   Who-outcomes 2 and 4; Who-evidence E2 and E3.
```

> **QUESTION 6.2 — composition puts the middle entity in the trust path, and nothing types it.**
> `exec.CommandContext(ctx, "codex", …)` is one `EntityID` covering two entities — the CLI binary
> and the model behind it. If codex starts failing, nothing in the graph can say which regressed.
> `world.md`'s *start coarse and split on evidence* makes that the right starting state, and
> `Who-outcomes` 3 says what the split would be into. **No type here supports the split**, and
> adding one before the evidence exists would be declaring what should be learned. Recorded as an
> accepted cost, not a gap.
