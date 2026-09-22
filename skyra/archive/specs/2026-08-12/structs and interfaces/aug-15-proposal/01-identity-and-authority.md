# Identity and authority

**Date:** 2026-08-15
**Status: PROPOSAL.**

> **`From` says where you are. `Horizon` says how far you may go. `Ceiling` says how far you may
> ever have been allowed to go. All three are supplied by the substrate, and the only operation on
> the second is narrowing.**

---

## 1 · Identifiers

```go
package ids

type (
	NamespaceID  string  // a principal's top engram. durable.
	EngramID     string  // durable behavior.
	ObservationID string // one witnessed result.
	EntityID     string  // WHO interprets. an identity and nothing else.
	HostID       string  // WHERE it lands. physical, including devices with no OS.
	RuntimeID    string  // one per device that can run one. AN AUTHORITY DOMAIN.
	SpanID       string  // one branch. the transaction key.
	InvocationID string  // one occurrence of one child position. the causal event.
	EventID      string  // one entry in the live causal record.
)
```

**Globally unique, never scoped.** `skyra-dna-v5/ids/ids.go` already argues this and the argument
survives the rewrite: `Provenance` resolves across trees, and the credit walk is bounded by cites
being empty at a crossing, never by ID scope. `os-parallels.md` §17 records that Linux takes the
opposite route — a pid namespace renumbers — and that the divergence is deliberate.

```text
NEW SINCE v5   RuntimeID     From needs a third coordinate and there is no type for it.
               InvocationID  observable-causality.md needs the invocation to be the identity of
                             the causal event, distinct from the engram and from the return.
               EventID       causality-and-recognition.md's live causal record. see 07 §1.
               NamespaceID   distinct from EngramID even though a namespace IS an engram, because
                             every authority signature takes one and none should accept the other.
```

```text
SOURCE   terminology.md §"The deployment layer"; ids.go's own comment; Where-outcomes 2 (device
         identity is derived, a hash over hardware fields read at boot — so RuntimeID has a
         derivation, not a minter).
```

> **QUESTION 1.1 — is `NamespaceID` a distinct type or an alias?**
> A namespace *is* an engram (`terminology.md`), so a distinct type is a lie the compiler tells for
> safety. The alternative is `EngramID` everywhere plus a `Ceiling` field that is only non-nil on
> namespaces, which is the "optional field means two types" smell. **Taken: distinct type**, with a
> total `func (n NamespaceID) Engram() EngramID`. Reversible either way; flagged because it is the
> kind of thing that is annoying to change late.

> **QUESTION 1.2 — does `RuntimeID` have a minter at all?**
> `Where-outcomes` decision 2 says identity is *derived* — re-hashed at boot, no state file, no
> orphaning. But the same decision admits a fallback to mint-and-persist for devices with no
> readable hardware fact. So the type has two provenances and no way to tell them apart, and
> "derived" is the property the rest of the design leans on. **Not answered here.** It belongs in
> the platform layer, and this proposal only needs the opaque string.

---

## 2 · `From` — the origin, filled by the runtime

```go
package identity

// From is grounding. NO CONSTRUCTOR AND NO SETTERS — the runtime fills it, an engram may reason
// from it, and nothing inside a possibility can declare or alter it.
type From struct {
	Namespace  ids.NamespaceID // whose memory and authority is acting
	Requesting ids.EngramID    // which DURABLE BEHAVIOR requested this. not the occurrence.
	Runtime    ids.RuntimeID   // where "here" is
}
```

Three coordinates and not four. `reach-boundaries-proposal.md` is explicit that the invocation
remains the identity of the event in the causal record and **does not become a fourth semantic
coordinate inside `From`** — which is why `InvocationID` exists separately in §1.

```text
WHY IT CANNOT BE INFERRED

    not from the prose of the expectation
    not from where the entity that emitted the preceding decision sits
    not from a resource named in an observation
    not from the machine an earlier invocation happened to land on

A nested reach into another runtime receives a NEW runtime-supplied From. Until then, `here` is
the current runtime's host and authority domain.
```

```text
SOURCE   reach-boundaries-proposal.md §"`FROM` supplies the origin"; think-interface-proposal.md
         §"`From` and `Horizon`"; distance.md §"NO ORIENT — distance is a RELATION and needs two
         operands."
```

---

## 3 · `Horizon` — one narrowing object, three alphabets

**This is decision D2, and it is the largest call in the proposal.**

`terminology.md` specifies `GRANT · PROVISION · SCOPE · CEILING` and marks the model as carrying
whole because the 08-12/13 work never addressed it. The 08-12/13 work then specified `Horizon`,
independently, with the same properties. `energy/README.md` proposes energy as a third quantity
provisioned by the same law and says so outright: *"it is the shape authority already has."*

```text
THREE DOCUMENTS, ONE OBJECT

terminology.md          "PROVISION — the act of handing down access from access you hold.
                         IT CAN ONLY NARROW."
reach-boundaries        "CHILD HORIZON ⊆ PARENT HORIZON"
energy/README.md        "A PARENT PROVISIONS a subset of its energy to a child position.
                         A CHILD MAY SPEND OR SUBDIVIDE IT. IT MAY NOT MINT."
```

Three vocabularies for one incapacity is exactly the defect `the-collapse.md` names about the
control plane — *exporting them again made one vocabulary look like two*. So:

```go
// Horizon is THE OUTER BOUND ON ACTION. It is not a second frame and carries no application
// values: the frame holds knowledge and names, Horizon holds only how far outward this
// possibility may legitimately see and act.
//
// UNEXPORTED FIELDS, NO LITERAL CONSTRUCTION. The only way to obtain one is Narrow, from one you
// already hold, or Boot. That makes P2 an incapacity rather than a check.
type Horizon struct {
	grants    []Grant           // what may be DONE.   the world boundary.
	reach     []ids.EntityID    // who may be REACHED. the runtime boundary.
	energy    Budget            // what may be SPENT.  see 09 F4.
}

type Grant string // prefix form. fs:read:/p/** · spawn:<bin> · orient:self

// Narrow is the WHOLE of the object-capability guarantee, and it is one subset check.
// Used at every scale: boot → admin, Approach → session, Decide → child position.
func (h Horizon) Narrow(g []Grant, r []ids.EntityID, e Budget) (Horizon, error)

// Permits answers the world-boundary authorization. Arithmetic. No semantics.
func (h Horizon) Permits(g Grant) bool
func (h Horizon) CanReach(e ids.EntityID) bool
```

```text
THE DEFAULT IS NOTHING.   type-surface-proposals.md A4 argues this and the argument is complete:
                          inheritance-with-extra-syntax makes the act decorative; least privilege
                          by construction makes under-granting a loud failure the parent learns
                          from, and over-granting bounded by your own ceiling.
```

### `Ceiling` is a durable `Horizon`

```go
// On the namespace engram only. DURABLE, written by whoever provisioned it. Boot writes admin's;
// admin writes everyone else's.
type Ceiling Horizon

// Approach's one check. Arithmetic, no semantics, AND NO LIVE ADMIN REQUIRED — which is what keeps
// the rack flat. See terminology.md §Authority.
func (c Ceiling) Admits(h Horizon) bool
```

```text
SOURCE   terminology.md §Authority (CEILING · MOUNT · THE RACK, and the flat-rack argument);
         type-surface-proposals.md A4/A5; reach-boundaries-proposal.md §HORIZON;
         energy/README.md §Provisioning; os-parallels.md §17 "The bounding set arrives twice, at
         two scales" — CapBnd and grants⊆Ceiling are the same rule.
```

> **QUESTION 1.3 — is the unification right, or is it a collapse the set will regret?**
> The case against: `terminology.md` puts grants on the **branch** because *effects happen at
> leaves*, while Horizon narrows at **every Decide**. Those are different granularities, and
> fusing them means a parent can silently reduce a child's world authority as a side effect of an
> ordinary composition. That might be correct — it is least privilege — or it might make every
> intermediate position a place authority can be lost by omission, which is a new failure mode with
> no signal. **Taken: unified**, because two objects with one law will drift. **This is the fork I
> am least sure of.** → 09 F2.

> **QUESTION 1.4 — does `Budget` belong inside `Horizon` or beside it?**
> Grants and reach are *permissions* — checking one costs nothing and it does not deplete. Energy
> **depletes**, which makes it the one member with mutable state, inside a type whose whole claim
> is that it is immutable and only narrows. Spending is not narrowing. `energy/README.md` leaves
> "what a loop earns" and "whether energy is replenished at a boundary" open, and both questions
> are really this one. **Taken: inside, provisionally**, because the provisioning law is identical
> and separating them duplicates `Narrow`. → 09 F4.

---

## 4 · `Position` — INVENTED. the version coordinate

`age.md` defines the coordinate `(owning parent, child position, age)` and decides everything about
its behavior. Nothing types it.

```go
// Position identifies a child slot in one parent's continuation. It is the lineage identity, and
// the parent's map key already IS it — so this is a projection of existing structure, not a new
// durable object.
type Position struct {
	Parent ids.EngramID
	Name   string // the child position's name in the parent's continuation
}

// Version is what a Position resolves to. A POINTER, NOT A CONTAINER: the chain is the history.
type Version struct {
	Engram   ids.EngramID
	Age      int
	Previous *Version // nil at v0. THE CHAIN IS THE HISTORY.
}
```

**Four things increment age, and the continuation is not one of them.**

```text
signature      what observations this position receives
return type    what observations this position hands back
entity         who interprets it
horizon        how far outward it may see or act
               ─────────────────────────────────────────
               a body rewrite does NOT increment. the continuation is interior.
```

### The invariant that keeps `Previous` safe

```text
THE POINTER LIVES ON THE VERSION, NEVER IN CONTENT.
THE CREDIT WALK DESCENDS OBSERVATIONS ONLY.
```

`second-kind-of-credit.md` rejects *"a back pointer the walk can follow"* by name. This satisfies
that structurally rather than by policy: the walk enumerates `Cites` → `Provenance` → invocation,
and `Cites` is derived by scanning **content**, which cannot reach a non-content field. No citation
can name a predecessor, so no edit inherits standing.

```text
SOURCE   age.md, whole document. When-outcomes 2 and 3 supply AGE and CURRENT;
         second-kind-of-credit.md supplies the constraint the pointer must satisfy.
```

> **QUESTION 1.5 — what versions a namespace?**
> `age.md` flags this and it is unresolved: age is positional, a namespace sits in no parent's map,
> so the rule has nothing to attach to at the outer edge. And `When-outcomes` 4 makes GC
> eligibility "disconnected from all namespaces" — so if connection is to a *version* rather than
> to the namespace, every increment orphans everything attached to the previous one. **Not
> answered.** The type above simply cannot express a namespace version, which is at least honest.

> **QUESTION 1.6 — concurrent revision breaks total ordering.**
> Two live branches revise one position and both are remembered. `When-outcomes` 3 permits sibling
> versions, so `Previous` is a DAG, not a chain, and "the accepted version" is ambiguous.
> `*Version` as written cannot represent it. Options: `Previous []*Version` (honest, and makes
> every reader handle merge), or refuse concurrent revision at Remember (simple, and refuses
> something the outcomes permit). **Not taken.** → 09 F5.
