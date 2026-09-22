# kernel/01 · The Protocol

**Package:** `kernel`

The seven words are seven semantic boundaries. They are **not** seven peer methods, and the type
surface has to say which is which.

---

## 1 · What a possibility can call

```go
type Kernel interface {
    Think(Frame, ThinkArgs) ([]Candidate, error)
    Decide(Frame, Candidate) (Decision, error)
    Remember(Frame, observation.Observation) error
    Recognize(Position, ids.EngramID) bool          // a string compare. 07-…md.
    Orient(Frame) (Topology, error)
}
```

Five methods. Two of the seven are absent, and their absence is the design:

```text
APPROACH   NOT ON THIS INTERFACE. it is encountered from outside; a possibility already inside
           the system cannot call it. see 02-approach.md.

REACH      NOT ON THIS INTERFACE. it is the hidden realization phase of Decide. exposing it would
           make it a peer of Think, which primitives.md §4 refuses by name.
```

`primitives.md`'s categories, and where each lands:

```text
CAUSAL FORMS                        DECIDE · REMEMBER          both exposed
BOUNDARIES                          APPROACH · REACH           NEITHER exposed to a possibility
UNIVERSAL POSSIBILITY-FACING        THINK                      exposed
RECOGNITION PROTOCOL                RECOGNIZE                  exposed, kernel-owned
STRUCTURAL OBSERVATION              ORIENT                     exposed, kernel-owned
```

---

## 2 · Why Reach is not a method

`Decide` returns a `Decision`, and realization happens behind it:

```go
type Decision struct {
    // NO EXPORTED FIELDS. see 04-decide.md §3.
}
```

If `Reach` were callable, a possibility could reach something it had not decided, or decide
something and reach a substitute. P9 forbids both, and the way to make them inexpressible is to
leave `Reach` off the interface and give `Decision` no constructor.

The version coordinate a possibility learns from Reach — `../../../reach/age.md` — therefore arrives
on Decide's result, not from a separate call. See `05-reach.md` §4.

---

## 2a · The compiler registry — 08-15, UNRATIFIED

The kernel holds what compilers exist; it never holds what they mean.

`Engram.Compiler` is an **address** — where that compiler lives. Whether the kernel needs a registry
to resolve it, or the address suffices, is open; if a registry exists it must be
**constructor-keyed**, `map[…]func() Compiler`, for the reason `../../06-entity.md` already gives:

> *"that is what makes statelessness structural... if state hid inside an interpreter, two
> invocations that look identical would give different answers and **same question becomes a
> lie**."*

### The direction, corrected 08-15

An earlier pass put `Decide`, `Recognize`, and `Remember` on a `Compiler` interface. **Reverted.**

```text
THE KERNEL OWNS THE METHODS.   Decide · Remember · Recognize are kernel methods. all three.
THE COMPILER COMPILES.         it turns the causal syntax into calls on them —
                               the way any language compiles to syscalls.
```

A compiler holding the causal forms would be extending causal structure, which `primitives.md`
reserves: *"Decide and Remember... are the only two primitives that change the live causal
structure."*

```text
COMPILERS OWN   their syntax.
THE KERNEL OWNS scope · live topology · causal order · routing · closure · ephemeral→durable.
                NOT ONE OF THOSE NEEDS SYNTAX.
```

The kernel still never reads a body — it is called into by compiled code, and type-checks by
comparing two strings it holds.

### The registry is not Horizon

Two different facts, and both are needed:

```text
THE REGISTRY   what compilers EXIST on this runtime.   INSTALLED.
HORIZON        which of them THIS CHAIN may use.       PERMITTED. identity/03-horizon.md §1a.
```

A compiler can be present and unreachable. Registering one grants it to nobody — the same
relationship `terminology.md` draws between the rack and the provisioning lineage.

---

## 3 · The kernel owns the protocol; it does not own the semantics

The distinction runs through every method here and is the thing most easily lost in an
implementation:

```text
KERNEL OWNS      scope · live topology · routing · causal order · exact keys · closure ·
                 the transition from ephemeral to durable · the protocol itself

ENGRAMS AND      interpret needs · embody hidden thinking behaviors · rank potential ·
ENTITIES OWN     translate observations · compose within kernel protocols · recognize meaning ·
                 produce results

GRAPH            witnesses only what Remember retained
```

`Recognize` is the clearest case and the one most likely to be built wrong. It is a **kernel
method** whose body calls entity and engram methods:

```text
kernel.Recognize(parent contract, returned observation)
        ↓
calls entity and engram methods
        ↓
their composition returns match or mismatch
```

The kernel delivers the observation, invokes the composition, and records the result. It does not
invent the judgment. **Protocol ownership does not make the kernel a semantic interpreter.**

`Orient` is the opposite case and is allowed to answer directly, because the kernel already holds
live topology and is reporting structure rather than meaning.

---

## 4 · What is NOT on the kernel

Named here because each was on some earlier surface and its removal is a decision, not an omission.

```text
Root · Ask · Drive · Accept · Reject       the current runtime's exported kernel methods.
Ingress · Hold · Spiral · Introspect       ../../../overnightBuild/README.md pins them as untouchable
                                           for that work order. THIS outline is not that work
                                           order, and the mapping is in ../../../semantics/README.md:
                                           Accept → Remember, Ingress → Approach,
                                           Introspect → Orient, Drive → a select.

Mount / Unmount                            ../../../semantics/terminology.md calls Mount "the ONLY
                                           exported kernel method" and the authority model around
                                           it — GRANT · PROVISION · SCOPE · CEILING · THE RACK —
                                           "carries whole," untouched by the 08-12/13 work.
                                           IT IS NOT RECONCILED WITH THE SEVEN HERE. see H3.

Resolve · Match · Route · Forget           primitives.md "words absorbed by the primitives."
                                           mechanism inside Reach, inside recognition, inside
                                           Reach, and unnecessary respectively.
```

The `Mount` gap is not cosmetic. `terminology.md` is authoritative on vocabulary and says the
authority model is untouched by the seven; `primitives.md` puts admission under `Approach`. Both
are current. Reconciling them is **HOLE H3**.

---

## 5 · Error, and why there is so little of it

Every method above returns `error` except `Recognize`. That asymmetry is the point.

```text
error HERE MEANS         the protocol could not be performed — a malformed call, a Horizon
                         violation, a kernel-level refusal.

error DOES NOT MEAN      a child returned something the parent cannot use. that is Parse → 0,
                         it is not an error value, and it does not travel up.
```

P5 and P15 are about the second. The three failure kinds each keep their own channel:

```text
INVALID SOURCE PROGRAM      owner: the continuation compiler.   result: no decision.
UNREALIZABLE DECISION       owner: the Reach boundary.          result: a bounded conclusion.
UNSATISFIED CHILD RETURN    owner: the parent that defined it.  result: Parse → 0, parent wakes.
```

Collapsing these into one `error` is the specific defect the failure-ownership model exists to
prevent: transport failure, type failure, and source-language failure becoming one generic channel
that teaches nothing.
