# Observation and contract

**Date:** 2026-08-15
**Status: PROPOSAL.**

> **An observation carries content. The substrate carries its provenance. The parent carries the
> contract that decides whether that known observation may be consumed at a position.
> Three separate facts, three separate owners, and the type must not let any of them merge.**

---

## 1 · `Content` — carried forward, one member removed

```go
package content

type Content map[string]Value

// Closed rather than any. Cites is DERIVED by scanning for Refs, and a scan over an open set
// cannot be exhaustive.
type Value interface{ isValue() }

type (
	Str  string
	Num  float64
	Bool bool
	List []Value
	Obj  map[string]Value
)

// Ref is an observation embedded in another's content. EVERY Ref IN A COMMIT IS A CITATION —
// there is no declared citation list to disagree with it.
type Ref ids.ObservationID

// WorldRef points at something in the world that is not an observation: a file, a rendered page,
// a spawned process. It is NOT an address — an address resolves and a WorldRef does not.
type WorldRef struct {
	Host   ids.HostID
	Handle string
}
```

```text
REMOVED   content.EngramRef      → §5. this is decision D3 and it is the load-bearing removal.
KEPT      Shape() · Fingerprint() · Cites()   the recognition primitives already exist and
                                              causality-and-recognition.md names them as
                                              "beginnings, not the whole mechanism."
```

```text
SOURCE   skyra-dna-v5/content/content.go, carried; causality-and-recognition.md §"Method two";
         conditions-and-loops.md on dotted paths (`plan.total` already means a location in content
         because collectShape builds them).
```

---

## 2 · `Provenance` — substrate-held, unforgeable

The single most important type in the proposal, because it is what replaces every declaration the
design has dissolved.

```go
package graph

// Provenance is what the SUBSTRATE knows about where an observation came from. It is not in
// Content and cannot be written by an entity. Zero value means INGRESS — no possibility produced
// it, so asking whether it was originated or composed is a category error.
type Provenance struct {
	Engram     ids.EngramID     // which durable behavior produced it
	Invocation ids.InvocationID // which OCCURRENCE. the causal event's identity.
	Entity     ids.EntityID     // who interpreted it
	Version    int              // which age of that position was acting. see 01 §4.
}

type Observation struct {
	ID   ids.ObservationID
	By   Provenance // ZERO AT INGRESS. the walk terminates here with no special case.
	Content content.Content
}

// Cites is derived, always. There is no field and no setter.
func (o *Observation) Cites() []ids.ObservationID
```

```text
WHY Provenance IS A STRUCT AND NOT FOUR FIELDS ON Observation

    it is the thing a contract matches against (§4), so it must be passable as one value;
    it is what the causal proof enumerates (observable-causality.md §"What the record proves");
    and keeping it separate makes "content cannot forge it" visually obvious at every use site.
```

```text
SOURCE   engram-contracts-and-composition.md §"Typing is known observation provenance";
         observable-causality.md §6 "Citation begins at invocation";
         primitives.md §1 (an admitted observation "enters with an empty producer").
```

> **QUESTION 2.1 — should `Version` be in `Provenance` at all?**
> It makes the record complete and it makes ablation reproducible, which `observable-causality.md`
> §2 requires. But `age.md` is emphatic that **age is a resolution coordinate, never a selection
> input**, and putting it on every observation puts it inside every frame. **Taken: included**, on
> the grounds that P9 is about what `Frame` exposes (04 §3) rather than what the record holds — but
> this is exactly how a measure leaks into selection, and it deserves a second look.

---

## 3 · What replaced the old `Expectation`

```go
// Expectation is the program. IMMUTABLE — authored once and re-run, and everything that
// accumulates accumulates AROUND it.
type Expectation struct {
	Expression string
}

// Draft is an expectation IN FLIGHT. It is not an Expectation, and that is the point: an
// improvised expression is a draft until closure, so immutability applies to COMMITTED
// expectations and not to expressions being refined.
type Draft struct {
	Expression string
	Revisions  int // how many times the parent restated it. the only signal that it was hard to say.
}
```

`closure.md` names this as the load-bearing gap — *"there is currently nowhere for that to live, so
refinement-in-place is not expressible at all."* Two types rather than a mutable one keeps
`interfaces.md`'s *immutable* literal.

```text
SOURCE   closure.md §"Expectations are drafts until closure"; type-surface-proposals.md A1.
```

---

## 4 · `Contract` — owned by the parent, contextual, not a class

```go
// A Contract is a CONSUMER-SIDE statement: "I know observations originating through this behavior,
// and I have a continuation for them here." It is not a universal classification of the producer.
//
// The same observation may satisfy a contract in one parent and not another. A contract name
// therefore does not create a global class hierarchy.
type Contract interface {
	// Admits is the whole of typing: Y ⊨ X. It receives the substrate's provenance and the
	// content, and it may read both — because the PARENT authored it, and a parent reading its
	// own child's return is not the kernel interpreting content.
	Admits(by Provenance, c content.Content) Verdict

	// Name is for diagnosis and for the compiled continuation's readability. It is NOT identity
	// and two contracts with one name are two contracts.
	Name() string
}

// ChildContract is everything a parent owns for one child position.
type ChildContract struct {
	Position Position

	Expectation Contract // WHAT behavior is being requested
	Input       Contract // WHICH observations may be given to it
	Return      Contract // WHICH observations this continuation can consume
	Entity      ids.EntityID // WHO, if the parent fixed it. zero = open.
	Horizon     Horizon      // HOW FAR. always narrower than the parent's.
}
```

```text
NO PERSISTENT TYPE EDGE.   engram-contracts-and-composition.md is explicit: a stored Parent──▶Child
                           type edge would say only that two things MIGHT compose, would go stale,
                           and is easy to mistake for evidence that they DID. The durable edge is
                           reserved for "this composition happened and survived Remember."
```

The four fields of `ChildContract` are exactly the four things `age.md` versions. That is not a
coincidence and it is worth stating as an invariant:

> **The contract IS the versioned surface. A change to any field is a new `Version`; the
> continuation is interior and does not increment.**

```text
SOURCE   engram-contracts-and-composition.md, whole; age.md §"The rule";
         parent-owns-the-return-boundary.md (DECIDED).
```

> **QUESTION 2.2 — how does a `Contract` identify accepted provenance?**
> The set leaves this open by name: *"whether contracts admit one producer, a bounded set, or a
> composable relation among producers."* `Admits` above ducks it by taking `Provenance` whole and
> letting the implementation decide. That is honest but it means the compiler can check nothing —
> `continuation-language.md` acceptance test 10 wants the dependency graph reconstructible from
> names and contracts, and an opaque predicate is not reconstructible. **Not answered.** → 09 F3.

> **QUESTION 2.3 — `parent-owns-the-return-boundary.md` says the comparison is "generic string
> matching."** That is far weaker than `Admits`. If the boundary really is `expected == returned`
> on a type name, then a type name is a *label*, and P3 forbids a child certifying its own type by
> writing one. The reconciliation must be that the name is assigned by the **substrate** from
> provenance, not carried in content — in which case string matching is fine and `Admits` is
> over-built. **Taken: `Admits`**, because it also has to express MISSING_FIELD and
> UNSUPPORTED_VERSION, which no string comparison can. Flagged because the DECIDED document says
> something narrower than what I have written.

---

## 5 · Reach does not travel in content — decision D3

**This removes `content.EngramRef` and it is the sharpest claim in the proposal.**

### The finding

`skyra-dna-v5/kernel/kernel.go` holds a contradiction inside one function:

```text
runGroup's COMMENT      "Reach narrows on the way down: a child gets its parent's reach, and a
                         select may drop from it but can never mint into it."

runGroup's CODE         for _, er := range ret.Content.Reach() { p.Delegated[er.Name] = child }
                         ← this MINTS a name into the PARENT'S frame, from a CHILD'S RETURN.
```

The comment states P2. The code violates it, upward, from the least trusted direction available.
The existing defenses — refuse an ancestor, ignore refs in arguments and at ingress — narrow the
hole without closing it, because whoever I invoked still chooses what I can subsequently touch.

### What the newer set already says instead

`primitives.md` §3 lists among what `Decide` binds:

```text
HORIZON      a subset of the parent's Horizon, PROVISIONED DOWNWARD.
```

and §"Words absorbed" records `provision` as *absorbed by DECIDE, which binds "whatever reach and
grants are explicitly provisioned downward."* So under the 08-13 model **delegation is an act of
the parent at `Decide`, not a value in a child's return.** `EngramRef` is the mechanism of the
superseded model, still present because nothing removed it.

### What this dissolves

```text
THE E1/E2 STOP     overnightBuild/stops.md: "Two of the three adapters do not produce text, and
                   one of them cannot… the grammar has no surface form for an EngramRef." The
                   three exits it lists are all bad. THE STOP DISSOLVES: if no entity ever needs
                   to WRITE reach, `Raw` can be text and the kernel can parse one grammar.

THE INJECTION      content.go's "same bytes, different provenance, and only the type keeps them
DISTINCTION        apart" stops needing a type to keep them apart, because reach is no longer
                   bytes at all.

os-parallels §14   the symlink parallel is about a type that no longer exists. It flags this
                   itself, contingent on H13.
```

### What it costs, stated plainly

A child that *discovers* something reachable can no longer hand that capability up. It can only
return an **observation describing** it, and the parent can act on that description only if its own
Horizon already permits the reach — or by `Approach` at that boundary.

```text
BEFORE   the child hands up a capability.        capability flows UP. violates P2.
AFTER    the child hands up EVIDENCE.            capability flows DOWN only. P2 holds.
```

This is strictly correct under monotone narrowing and it removes a real capability from the system.
`primitives.md` §1 covers the case — *"Approach is recursive, for the same reason reach is"* — but
nothing types the recursive Approach, so the replacement path is specified and unbuilt.

> **QUESTION 2.4 — is losing upward delegation acceptable?**
> The scenario: a reached entity tells me about a better entity behind it. Under D3 I cannot use
> that unless I was already provisioned for it, which means a namespace must anticipate every
> entity its subtree might usefully reach. That is the ocap answer and it is also the annoying one.
> **Taken: accepted**, because the alternative demonstrably violates the narrowing law and the
> violation is currently *unnoticed in a comment that asserts the opposite*. But if upward
> delegation is required, D3 is wrong and `EngramRef` should come back as an explicit, recorded
> **Approach request** rather than a value the parent silently folds in. → 09 F6.

---

## 6 · `Verdict` — INVENTED. binary result, classified evidence

```go
// Verdict is what a parent's return contract produces. The BINARY controls execution; the
// classification is evidence for the awakened parent, and never the thing that causes failure.
type Verdict struct {
	OK    bool
	Class MismatchClass // meaningful only when OK is false
	Note  *Observation  // OPTIONAL explanatory observation. never required, never causal.
}

type MismatchClass uint8

const (
	Match MismatchClass = iota
	MissingField
	WrongProvenance
	UnsupportedVersion
	AuthorityMismatch
	MalformedObservation
)
```

```text
WHY NOT A THIRD OUTCOME   the parent's continuation either has a next state or it does not.
                          language-improvements.md is explicit that the improvement is NOT to
                          replace 1|0 with a third result — it is to preserve structured evidence
                          about why the parser reached the outcome it did.

WHY THE CLASS IS          a child cannot choose its own mismatch classification or make an
PARENT-OWNED              incompatible observation compatible by writing a label. The class is
                          derived by the parent's parser from the expected contract, the actual
                          observation, its provenance, and the authority it returned under.
```

```text
SOURCE   language-improvements.md, whole; primitives-and-continuation-summary.md §"What a type
         failure means"; P12.
```

> **QUESTION 2.5 — is `Note` a hole in P5?**
> An optional explanatory observation is exactly the shape an `Error` grows out of. The set is
> careful — *"the explanation reports on the mismatch, it does not create it"* — but a field that
> is present on every failure will be read as one, and the first `if v.Note != nil` in a repair
> path makes it load-bearing. **Taken: kept, as `*Observation` rather than a string**, so that it
> is an ordinary observation with provenance and no privileged status. Watch it.
