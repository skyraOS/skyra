# The kernel protocol

**Date:** 2026-08-15
**Status: PROPOSAL.** The surface. This is where decision D1 lives.

> **Seven semantic distinctions. Five methods. Two protocols with public result types and no
> caller-facing method, because nothing ever calls them.**

---

## 1 · Five, not seven — decision D1

The set says three different things, in three documents that are all current:

```text
primitives.md          "They are kernel surface operations" · "the kernel owns the seven-method
                        protocol"
coherence/README.md    "The seven meanings are not seven peer methods."
think-interface        "Reach is not directly exposed to the possibility as a peer of Think. It is
                        the hidden realization phase of Decide."
```

A type surface cannot hold all three. The test that resolves it is **who calls it**, and the set
answers that consistently even where it disagrees about the word:

```text
                WHO CALLS IT              → METHOD?
─────────────────────────────────────────────────────────────────────────────
APPROACH        an outside party            YES. but not possibility-facing.
THINK           a possibility               YES.
DECIDE          a possibility               YES.
REMEMBER        a possibility               YES.
ORIENT          a possibility               YES.  primitives.md §7: "The caller invokes Orient
                                            on the kernel." Unambiguous.
─────────────────────────────────────────────────────────────────────────────
REACH           THE KERNEL, inside Decide   NO. "hidden realization phase."
RECOGNIZE       THE KERNEL, at the return   NO. "the kernel delivers it, calls the relevant
                boundary                    entity and engram methods, records the result."
```

So `Reach` and `Recognize` are **protocols, not operations**: they have public result types
(`Realization`, `Verdict`) that appear in the signatures of the methods that drive them, and no
entry point of their own. That preserves seven distinct meanings — each still asks a different
question, each still owns a failure class — while making the surface honest about the two nobody
can invoke.

```text
WHAT THIS COSTS   `Recognize` was the one method os-parallels.md §17 could not sort, and its note
                  is worth keeping: it is where LEARNING enters, which is the half the OS analogy
                  covers not at all. Demoting it to an internal protocol risks under-building the
                  most important thing in the system. Held against that: it is still a named
                  protocol with a typed result, and the entity/engram composition beneath it is
                  where the learning actually happens.
```

---

## 2 · The surface

```go
package kernel

// Kernel HOLDS EVERYTHING HAPPENING. The graph holds only what happened.
// It has no policy, reads no content, and performs no durable write except the one projection at
// outer Remember (07 §2).
type Kernel interface {
	// ─── the boundary ────────────────────────────────────────────────────────
	Approach(req Admission) (*Session, error)
	Unmount(ids.NamespaceID)

	// ─── the possibility-facing surface ──────────────────────────────────────
	// Every one takes the calling possibility's identity, which the kernel supplies to itself;
	// a possibility cannot name another.
	Think(p PossibilityID, q Query) ([]Candidate, error)
	Decide(p PossibilityID, d Decision) (Realization, error)
	Remember(p PossibilityID, o ids.ObservationID) (graph.Verdict, error)
	Orient(p PossibilityID) ([]Head, error)
}
```

```text
GONE FROM v5   Ask · Drive · Ingress · Accept · Reject · rootFor · Hold · Spiral · Introspect

               Ask         a composition written for a test rig.
               Drive       A SELECT OF CARDINALITY 1, AT DEPTH 0, WITH NO PARENT. it is Decide.
               Ingress     renamed Approach, and the rename carries the gate.
               Accept      Remember at the outer boundary. same operation, different surface.
               Reject      §6. it survives as an operation, not as a mode.
               rootFor     internal.
               Hold        an implementation leak — parking an unminted engram is a detail of a
                           RULE, and it was exported. Under 07 §2 nothing is minted before
                           projection, so there is nothing to park.
               Spiral      a global count with no node to start from. not introspection.
               Introspect  a COUNTER that counts the wrong thing. → §5.
```

```text
SOURCE   the-collapse.md §"The collapse", whole; terminology.md §Authority (Mount as "the ONLY
         exported kernel method" — superseded here by five, see Q5.1); primitives.md §§1–7.
```

> **QUESTION 5.1 — `terminology.md` says Mount is the only exported method.**
> That was written when the reflexive names were argued not to be a surface. `primitives.md`
> supersedes it and puts Think/Decide/Remember/Orient on the kernel explicitly. I have followed the
> newer document. Noted because the older one is still marked AUTHORITATIVE on vocabulary, and this
> is the one place I contradict it.

---

## 3 · `Approach` — the boundary, and the gate that does not exist

```go
type Admission struct {
	Namespace ids.NamespaceID
	Entity    ids.EntityID       // who is asking to be let in
	Requested identity.Horizon   // must satisfy ns.Ceiling.Admits
	Credential Credential        // ← THE HOLE. see below.
}

// Session is ALL LIVE WORK under one namespace. One per sign-in. Created by Approach, ejected by
// Unmount. It holds what admission provisions — ONCE, not per ask.
type Session struct {
	Namespace ids.NamespaceID
	Entity    ids.EntityID
	From      identity.From
	Horizon   identity.Horizon

	// Top-level possibilities, keyed by span. A BRANCH IS A POSITION, NOT A TYPE — a possibility
	// with no parent. Staged effects are partitioned by the SAME key, because accepting one ask
	// must neither land nor discard another's writes.
	branches map[ids.SpanID]*possibility
	staged   map[ids.SpanID][]*Effect
}

type Outcome uint8
const (
	Admitted Outcome = iota
	Avoided          // refused. whether this is durably appended is OPEN — primitives.md §1.
)
```

```text
THE RACK IS FLAT   admin is a ROW on it, not a level above it. Lineage (who provisioned my
                   namespace, DURABLE) and activity (who is signed in, EPHEMERAL) are different
                   questions, and if the rack mirrored the lineage, admin signing out would log
                   everyone out. What makes admin special is that BOOT wrote its ceiling.
                   → terminology.md §Authority.
```

### The gate

```text
APPROACH        the primitive. should this be let in at all?
  authenticate  answers WHO   ← SPECIFIED IN terminology.md SINCE 08-11, DROPPED BY primitives.md,
                                BUILT NOWHERE. questions.md §5 says nothing gates entry.
  Ceiling       answers WHAT  ← designed and typed above: Requested ⊆ Ceiling.
```

`Credential` above is a named hole, not a design. I have given it a type so that every call site
has to pass *something* and the absence is visible at the boundary rather than discovered later.

> **QUESTION 5.2 — this is the highest-privilege unguarded operation in the design, and now it also
> guards the budget.** `energy/README.md` states the compounding: *"A budget on an ungated
> admission point is a budget anyone can open."* So metering — the mechanism that would make
> eviction principled instead of a constant — cannot be trusted until `authenticate` exists.
> Two holes, one door. → 09 F13.

---

## 4 · `Think` · `Decide` · `Remember` · `Orient`

```go
// ─── THINK ───────────────────────────────────────────────────────────────────
// Creates potential. Performs no proposed work, contacts nothing, resolves no address.
type Query struct {
	Entities     []ids.EntityID     // EMPTY = open · ONE = fixed · SEVERAL = bounded
	Expectations []graph.Expectation
	Observations []ids.ObservationID
	// NO address operand. NO From. NO Horizon. The runtime supplies the last two.
}

// Candidate is an OBSERVATION carrying potential behavior. Not an executable handle: it holds no
// address, transport, or live relationship, and no claim that contact has been established.
type Candidate struct {
	Observation *graph.Observation // the candidate IS an observation. Decide binds THIS, not its text.
	Rank        int
}

// A no-candidate result is a CONSTRUCTED DIAGNOSIS, never an empty slice — absence cannot
// distinguish no match, no authority, and failure to think.
type Diagnosis uint8
const (
	NoKnownPotential Diagnosis = iota
	BeyondYourHorizon
)
```

```text
BOUNDED PLURALITY   ordinarily no more than five. ONE leaves no local selection; UNBOUNDED floods
                    the frame. Five is an operating bound, not a constant — the invariant is a
                    small ranked set, so it is not in the type.

THE REGRESS STOPS   calling Think does not require first thinking about which thinker to call.
AT THINK            Its internal behaviors are HIDDEN FROM CALLER SELECTION and VISIBLE IN THE
                    CAUSAL CONE — which is why Candidate carries an Observation with provenance
                    rather than a bare expectation.
```

```go
// ─── DECIDE ──────────────────────────────────────────────────────────────────
// The choice IS the binding. Reach happens inside it.
type Decision struct {
	Position  string                // WHERE in my continuation. causal context, not an operand.
	On        ids.ObservationID     // THE CANDIDATE OBSERVATION. preserved as a constituent.
	Inputs    []ids.ObservationID   // must all be InScope
	Cite      ids.ObservationID     // PER MEMBER, never per group.
	Horizon   identity.Horizon      // must Narrow from mine
}

// Realization — INVENTED. The three conclusions are a sum type in prose and nowhere in code.
type Realization struct {
	Kind    RealizationKind
	Opened  PossibilityID     // Kind == Live only
	Version identity.Version  // Kind == Live only. YOU HAVE TO REACH SOMETHING TO AGE IT.
	Why     *graph.Observation // bounded impossibility evidence. never a raw endpoint error.
}

type RealizationKind uint8
const (
	Live RealizationKind = iota
	BeyondCurrentReach // Horizon permitted it; no eligible relationship worked NOW.
	BeyondHorizon      // it exceeded the authority provisioned. more endpoints cannot repair it.
)
```

```text
NO ENTITY OPERAND.   the candidate observation already identifies it. A second entity field would
                     create two places for one fact and permit contradiction — Decide(child,
                     entity A, observation naming entity B) — and removing it makes the
                     disagreement inexpressible.

REACH MAY NOT        it resolves volatile location for the ONE composition Decide selected. Any
SUBSTITUTE           substitution creates a different possibility from the one decided.

THE VERSION COMES    not from Decide. A version returned by Decide would be a version for
BACK FROM REACH      something that may never exist.
```

```go
// ─── REMEMBER ────────────────────────────────────────────────────────────────
// Receives a return into the parent's live causal history. Runs Recognize as its first act.
// At the OUTER boundary it projects (07 §2) and closes.
Remember(p PossibilityID, o ids.ObservationID) (graph.Verdict, error)
```

```text
ONE OPERATION, EVERY LEVEL   a parent remembers a child's observation; its parent may remember the
                             composition; the outermost caller remembers a result. That outermost
                             remembrance closes the branch and makes the recursively remembered
                             cone durable.

VERDICT.OK == false          the parent WAKES at that position with a focused frame. That is not a
                             separate method — it is what the kernel does with the result.
```

```go
// ─── ORIENT ──────────────────────────────────────────────────────────────────
// Reports TOPOLOGY, never meaning. Bounded by the subtree, which is a fact about which way
// pointers point rather than a rule anyone checks.
type Head struct {
	Span   ids.SpanID
	Engram ids.EngramID  // AN ID. never the expectation.
	Entity ids.EntityID
	Host   ids.HostID
	Depth  int
	Since  time.Duration // time in Step. THIS IS WHAT MAKES STALLED DETECTABLE.
}
```

**The caller dereferences the engram against their own graph.** Same move as `Provenance` being a
field so `ObservationID` stays opaque: the kernel hands back handles and the reader resolves them.
That is also what retires `shortExpr` and `oneLine` — they existed because the kernel wanted to say
something readable and had no other channel.

---

## 5 · Heads are derived, and the counter was wrong

```go
// head == no live children AND no return under the current draft
func (k *kernel) heads(from ids.SpanID) []Head // ONE DESCENT, TWO FILTERS
```

```text
EXECUTING   inside Step, no live children.       A HEAD.        "what am I doing"
WAITING     children are live.                   not a head.
RETURNED    committed and still live.            not a head.    "what has this done"
```

v5's `heads map[SpanID]int` increments on entering `open()` and decrements on leaving, so a
depth-five chain with one leaf running reports **five** — four of which are ancestors sitting in
`runGroup`. That makes a *stalled head* indistinguishable from *a parent legitimately waiting*,
which is the eviction decision the number exists for.

```text
SOURCE   terminology.md §Head, which supplies the Head struct verbatim; type-surface-proposals.md
         A1 and D2.
```

---

## 6 · `Reject` — an operation, not a boolean

```go
// Not a mode on Remember and not an empty Remember. Remember selects a causal chain to survive;
// Reject establishes that it does not. Accomplishing either defines the complement.
Reject(p PossibilityID, why string) error
```

**Not on the `Kernel` interface above** — it is a `Session` operation at the outer boundary,
because only the outermost perspective can refuse a branch. Recorded here so the omission is
deliberate rather than an oversight.

```text
REMEMBER   the walk reaches the chain   → staged effects may LAND
REJECT     the chain does not survive   → staged effects are DISCARDED, and the discard is an
                                          EVENT. citation was designed with cheap failure and a
                                          lost write is not that.
```

> **QUESTION 5.3 — the rootless commit is still open, and it is now the last thing between this
> surface and one vocabulary.** If acceptance IS Remember, it has nowhere to land: the outermost
> parent has no engram, so no invocation record owns its verdict. Either the principal **gets an
> engram** and acceptance becomes literally an ordinary Remember, or acceptance stays a kernel-side
> event and the uniformity is approximate. `terminology.md` makes a namespace an engram, which
> points hard at the first — a principal commits INTO the namespace, and `closure.md` notes that is
> consistent and *written nowhere*. **Taken: not taken.** The signature above is the approximate
> version. → 09 F14.
